package storage

import (
	"bytes"
	"container/list"
	"encoding/hex"
	"github.com/mpetavy/tresor/service/index"
	"io"
	"io/ioutil"
	"reflect"
	"runtime"
	"strings"

	"github.com/mpetavy/tresor/models"
	"github.com/mpetavy/tresor/service/database"

	"github.com/mpetavy/tresor/service/cluster"

	"path/filepath"
	"sync"

	"os"

	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/cache"
	"github.com/mpetavy/tresor/hash"
)

const (
	TYPE_FS   = "fs"
	FS_VOLUME = "FS_VOLUME_"
)

type FsUID struct {
	Path string
}

func NewFsUID(path string) *FsUID {
	uid, _ := ParseFsUID(path)

	return uid
}

func ParseFsUID(path string) (*FsUID, error) {
	path = strings.Replace(path, "\\", "/", -1)
	path = strings.Replace(path, "@", string(filepath.Separator), -1)

	return &FsUID{path}, nil
}

func (uid *FsUID) String() string {
	return uid.Path
}

type Fs struct {
	volumes map[string]*FsVolume
	mu      *sync.Mutex
}

type FsVolume struct {
	Name string
	Path string
}

func NewFsVolume(name string, path string) (*FsVolume, error) {
	volume := FsVolume{Name: name, Path: common.CleanPath(path)}

	if !common.FileExists(path) {
		return nil, &common.ErrFileNotFound{path}
	}

	return &volume, nil
}

func NewFs() (*Fs, error) {
	fs := &Fs{volumes: make(map[string]*FsVolume), mu: new(sync.Mutex)}

	return fs, nil
}

func (fs *Fs) Init(cfg *Cfg) error {
	for i := 0; i < len(cfg.Volumes); i++ {
		path := common.CleanPath(cfg.Volumes[i].Path)

		if !common.FileExists(path) {
			return &ErrVolumePathNotFound{volume: cfg.Volumes[i].Name, path: path}
		}

		vol, err := NewFsVolume(cfg.Volumes[i].Name, path)
		if common.Error(err) {
			return err
		}

		fs.AddVolume(vol)
	}

	return nil
}

func (fs *Fs) Start() error {
	return nil
}

func (fs *Fs) Stop() error {
	return nil
}

func (fs *Fs) Volume(n string) *FsVolume {
	return fs.volumes[n]
}

func (fs *Fs) Volumes() []string {
	l := make([]string, 0)

	for _, v := range fs.volumes {
		l = append(l, v.Name)
	}

	return l
}

func (fs *Fs) AddVolume(v *FsVolume) {
	fs.volumes[v.Name] = v
}

func (fs *Fs) RemoveVolume(v *FsVolume) {
	delete(fs.volumes, v.Name)
}

func (fs *Fs) CurrentVersion(uid *FsUID) (int, error) {
	return 1, nil
}

func createFsPath(rootDir string, uid *FsUID) (string, error) {
	var path string

	if rootDir != "" {
		path = strings.Join([]string{rootDir, uid.Path}, string(filepath.Separator))
	} else {
		path = uid.Path
	}
	return common.CleanPath(path), nil
}

func (fs *Fs) find(uid *FsUID, options *Options) (*FsVolume, string, error) {
	var listVolumes list.List

	volumeName, ok := cache.Get(FS_VOLUME, uid.Path)

	if ok {
		_, valid := fs.volumes[volumeName.(string)]

		if valid {
			listVolumes.PushFront(volumeName)
		} else {
			ok = false
		}
	}

	for volume := range fs.volumes {
		if !ok || volumeName != volume {
			listVolumes.PushBack(volume)
		}
	}

	for i := 0; i < listVolumes.Len(); i++ {
		vn := getFromList(listVolumes, i).(string)
		volume := fs.volumes[vn]

		path, err := createFsPath(volume.Path, uid)
		if common.Error(err) {
			return nil, "", err
		}

		if common.FileExists(path) {
			cache.Put(FS_VOLUME, uid.Path, volume.Name)

			return volume, path, nil
		}
	}

	return nil, "", &ErrObjectNotFound{"??", uid.String()}
}

func (fs *Fs) Store(suid string, source io.Reader, options *Options) (string, *[]byte, error) {
	uid, err := ParseFsUID(suid)
	if common.Error(err) {
		return "", nil, err
	}

	cluster.Lock(cluster.ByStorageUid(uid.Path))
	defer cluster.Unlock(cluster.ByStorageUid(uid.Path))

	var volume *FsVolume

	if options != nil && len(options.VolumeName) > 0 {
		var ok bool

		volume, ok = fs.volumes[options.VolumeName]
		if !ok {
			return "", nil, &ErrInvalidVolumeName{options.VolumeName}
		}
	} else {
		if len(fs.volumes) > 0 {
			volume = fs.volumes[reflect.ValueOf(fs.Volumes).MapKeys()[0].String()]
		} else {
			return "", nil, &ErrNoVolumesDefined{}
		}
	}

	uid.Path = suid

	cluster.Lock(cluster.ByStorageVolume(volume.Name))
	defer cluster.Unlock(cluster.ByStorageVolume(volume.Name))

	path, err := createFsPath(volume.Path, uid)
	if common.Error(err) {
		return "", nil, err
	}

	if common.FileExists(path) {
		fs, err := common.FileSize(path)
		if common.Error(err) {
			return "", nil, err
		}

		if fs > 0 {
			return "", nil, &ErrObjectAlreadyExists{volume.Name, uid.String()}
		}
	}

	err = os.MkdirAll(filepath.Dir(path), common.DefaultDirMode)
	if common.Error(err) {
		return "", nil, err
	}

	dest, err := os.Create(path)
	if common.Error(err) {
		return "", nil, err
	}
	defer func() {
		common.Error(dest.Close())
	}()

	h, err := hash.New(hash.MD5)
	if common.Error(err) {
		return "", nil, err
	}

	_, err = io.Copy(io.MultiWriter(dest, h), source)
	if common.Error(err) {
		return "", nil, err
	}

	digest := h.Sum(nil)

	cache.Put(FS_VOLUME, uid.Path, volume.Name)

	return uid.String(), &digest, nil
}

func (fs *Fs) Load(suid string, dest io.Writer, options *Options) (string, *[]byte, int64, error) {
	uid, err := ParseFsUID(suid)
	if common.Error(err) {
		return "", nil, -1, err
	}

	cluster.Lock(cluster.ByStorageUid(uid.Path))
	defer cluster.Unlock(cluster.ByStorageUid(uid.Path))

	_, path, err := fs.find(uid, options)
	if common.Error(err) {
		return "", nil, -1, err
	}

	source, err := os.Open(path)
	if common.Error(err) {
		return "", nil, -1, err
	}
	defer func() {
		common.Error(source.Close())
	}()

	h, err := hash.New(hash.MD5)
	if common.Error(err) {
		return "", nil, -1, err
	}

	n, err := io.Copy(io.MultiWriter(dest, h), source)
	if common.Error(err) {
		return "", nil, -1, err
	}

	digest := h.Sum(nil)

	return path, &digest, n, nil
}

func (fs *Fs) Delete(suid string, options *Options) error {
	uid, err := ParseFsUID(suid)
	if common.Error(err) {
		return err
	}

	var volume *FsVolume
	var path string

	volume, path, err = fs.find(uid, options)
	if common.Error(err) {
		return err
	}

	cluster.Lock(cluster.ByStorageUid(uid.Path))
	defer cluster.Unlock(cluster.ByStorageUid(uid.Path))

	if common.IsFile(path) {
		err := os.Remove(path)
		if common.Error(err) {
			return err
		}
	} else {
		cluster.Lock(cluster.ByStorageVolume(volume.Name))
		defer cluster.Unlock(cluster.ByStorageVolume(volume.Name))

		err := os.RemoveAll(path)
		if common.Error(err) {
			return err
		}

		for {
			path = filepath.Dir(path)

			if path == volume.Path {
				break
			}

			files, err := ioutil.ReadDir(path)
			if common.Error(err) {
				break
			}

			if len(files) == 0 {
				err := os.RemoveAll(path)
				if common.Error(err) {
					break
				}
			} else {
				break
			}
		}
	}

	cache.Remove(FS_VOLUME, uid.Path)

	return nil
}

func (fs *Fs) rebuildBucket(uid *FsUID) error {
	bucket := models.NewBucket()
	bucket.Uid = uid.String()

	buffer := bytes.Buffer{}

	path, h, n, err := fs.Load(uid.String(), &buffer, nil)
	if common.Error(err) {
		return err
	}

	bucket.FileNames = append(bucket.FileNames, uid.String())
	bucket.FileHashes = append(bucket.FileHashes, hex.EncodeToString(*h))

	var mimeType string
	var mapping index.Mapping

	err = index.Exec(func(index index.Handle) error {
		mimeType, mapping, _, _, _, err = index.Index(path, nil)

		return err
	})
	if common.Error(err) {
		return err
	}

	if len(mimeType) > 0 {
		bucket.FileTypes = append(bucket.FileTypes, mimeType)
	}

	for k, v := range mapping {
		bucket.Props[k] = v
	}

	bucket.FileSizes = append(bucket.FileSizes, n)

	common.Debug("%s: %s", (*uid).String(), hex.EncodeToString(*h))

	err = database.Exec(func(db database.Handle) error {
		return db.SaveBucket(&bucket, nil)
	})
	if common.Error(err) {
		return err
	}

	return nil
}

func (fs *Fs) Rebuild() (int, error) {
	cluster.Lock(cluster.ByStorage())
	defer cluster.Unlock(cluster.ByStorage())

	err := database.Exec(func(db database.Handle) error {
		return db.EnableIndices([]interface{}{models.NewBucket()}, false)
	})
	if common.Error(err) {
		return -1, err
	}
	defer common.Error(database.Exec(func(db database.Handle) error {
		return db.EnableIndices([]interface{}{models.NewBucket()}, true)
	}))

	c := 0
	workerChannel := make(chan struct{}, runtime.NumCPU()*5)

	for _, volume := range fs.volumes {
		err := filepath.Walk(volume.Path, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				c++
				path = path[len(volume.Path)+1:]

				workerChannel <- struct{}{}
				go func() {
					defer func() {
						<-workerChannel
					}()

					common.Error(fs.rebuildBucket(NewFsUID(path)))
				}()
			}

			return nil
		})

		if common.Error(err) {
			return -1, err
		}
	}

	return c, nil
}
