package storage

import (
	"bytes"
	"container/list"
	"encoding/hex"
	"github.com/mpetavy/tresor/service/index"
	"io"
	"io/ioutil"
	"reflect"
	"strings"

	"github.com/mpetavy/tresor/models"
	"github.com/mpetavy/tresor/service/database"

	"github.com/mpetavy/common/generics"
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
	name    string
	volumes map[string]*FsVolume
	mu      *sync.Mutex
}

type FsVolume struct {
	Name string
	Path string
}

func NewFsVolume(name string, path string) (*FsVolume, error) {
	if name == UNZIP {
		return nil, &ErrInvalidVolumeName{name}
	}

	volume := FsVolume{Name: name, Path: common.CleanPath(path)}

	b, err := common.FileExists(path)
	if err != nil {
		return nil, err
	}

	if !b {
		return nil, &common.ErrFileNotFound{path}
	}

	return &volume, nil
}

func NewFs() (*Fs, error) {
	fs := &Fs{volumes: make(map[string]*FsVolume), mu: new(sync.Mutex)}

	path, err := common.CreateTempDir()
	if err != nil {
		return nil, err
	}

	fs.AddVolume(&FsVolume{Name: UNZIP, Path: common.CleanPath(path)})

	return fs, nil
}

func (fs *Fs) Init(cfg *common.Jason) error {
	name, err := cfg.String("name")
	if err != nil {
		return err
	}
	fs.name = name

	for i := 0; i < cfg.ArrayCount("volumes"); i++ {
		v, err := cfg.Array("volumes", i)
		if err != nil {
			return err
		}

		volumeName, err := v.String("name")
		if err != nil {
			return err
		}
		path, err := v.String("path")
		if err != nil {
			return err
		}

		path = common.CleanPath(path)

		b, err := common.FileExists(path)
		if err != nil {
			return err
		}

		if !b {
			return &ErrVolumePathNotFound{volumeName, path}
		}

		vol, err := NewFsVolume(volumeName, path)
		if err != nil {
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
		if volume != UNZIP && (!ok || volumeName != volume) {
			listVolumes.PushBack(volume)
		}
	}
	listVolumes.PushBack(UNZIP)

	for i := 0; i < listVolumes.Len(); i++ {
		vn := generics.GetFromList(listVolumes, i).(string)
		volume := fs.volumes[vn]

		path, err := createFsPath(volume.Path, uid)
		if err != nil {
			return nil, "", err
		}

		b, err := common.FileExists(path)
		if err != nil {
			return nil, "", err
		}

		if b {
			cache.Put(FS_VOLUME, uid.Path, volume.Name)

			return volume, path, nil
		}
	}

	return nil, "", &ErrObjectNotFound{"??", uid.String()}
}

func (fs *Fs) Store(suid string, source io.Reader, options *Options) (string, *[]byte, error) {
	uid, err := ParseFsUID(suid)
	if err != nil {
		return "", nil, err
	}

	cluster.Lock(cluster.STORAGE_UID(fs.name, uid.Path))
	defer cluster.Unlock(cluster.STORAGE_UID(fs.name, uid.Path))

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

	cluster.Lock(cluster.STORAGE_VOLUME(fs.name, volume.Name))
	defer cluster.Unlock(cluster.STORAGE_VOLUME(fs.name, volume.Name))

	path, err := createFsPath(volume.Path, uid)
	if err != nil {
		return "", nil, err
	}

	b, err := common.FileExists(path)
	if err != nil {
		return "", nil, err
	}

	if b {
		fs, err := common.FileSize(path)
		if err != nil {
			return "", nil, err
		}

		if fs > 0 {
			return "", nil, &ErrObjectAlreadyExists{volume.Name, uid.String()}
		}
	}

	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return "", nil, err
	}

	dest, err := os.Create(path)
	if err != nil {
		return "", nil, err
	}
	defer dest.Close()

	h, err := hash.New(hash.MD5)
	if err != nil {
		return "", nil, err
	}

	_, err = io.Copy(io.MultiWriter(dest, h), source)
	if err != nil {
		return "", nil, err
	}

	digest := h.Sum(nil)

	cache.Put(FS_VOLUME, uid.Path, volume.Name)

	return uid.String(), &digest, nil
}

func (fs *Fs) Load(suid string, dest io.Writer, options *Options) (string, *[]byte, int64, error) {
	uid, err := ParseFsUID(suid)
	if err != nil {
		return "", nil, -1, err
	}

	cluster.Lock(cluster.STORAGE_UID(fs.name, uid.Path))
	defer cluster.Unlock(cluster.STORAGE_UID(fs.name, uid.Path))

	_, path, err := fs.find(uid, options)
	if err != nil {
		return "", nil, -1, err
	}

	source, err := os.Open(path)
	if err != nil {
		return "", nil, -1, err
	}
	defer source.Close()

	h, err := hash.New(hash.MD5)
	if err != nil {
		return "", nil, -1, err
	}

	n, err := io.Copy(io.MultiWriter(dest, h), source)
	if err != nil {
		return "", nil, -1, err
	}

	digest := h.Sum(nil)

	return path, &digest, n, nil
}

func (fs *Fs) Delete(suid string, options *Options) error {
	uid, err := ParseFsUID(suid)
	if err != nil {
		return err
	}

	cluster.Lock(cluster.STORAGE_UID(fs.name, uid.Path))
	defer cluster.Unlock(cluster.STORAGE_UID(fs.name, uid.Path))

	var volume *FsVolume
	var path string

	volume, path, err = fs.find(uid, options)
	if err != nil {
		return err
	}

	b,err := common.IsFile(path) 
	if err != nil {
		return err
	}

	if b {
		err := os.Remove(path)
		if err != nil {
			return err
		}
	} else {
		cluster.Lock(cluster.STORAGE_VOLUME(fs.name, volume.Name))
		defer cluster.Unlock(cluster.STORAGE_VOLUME(fs.name, volume.Name))

		err := os.RemoveAll(path)
		if err != nil {
			return err
		}

		for {
			path = filepath.Dir(path)

			if path == volume.Path {
				break
			}

			files, err := ioutil.ReadDir(path)
			if err != nil {
				break
			}

			if len(files) == 0 {
				err := os.RemoveAll(path)
				if err != nil {
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

func (fs *Fs) rebuildBucket(wg *sync.WaitGroup, uid *FsUID) {
	defer wg.Done()

	bucket := models.NewBucket()
	bucket.Uid = uid.String()

	buffer := bytes.Buffer{}

	path, h, n, err := fs.Load(uid.String(), &buffer, nil)
	if err != nil {
		common.Error(err)
		return
	}

	bucket.FileName = append(bucket.FileName, uid.String())
	bucket.FileHash = append(bucket.FileHash, hex.EncodeToString(*h))

	var mimeType string
	var mapping *index.Mapping
	var thumbnail *[]byte
	var fulltext string
	var orientation common.Orientation

	err = index.Exec("index", func(index *index.Index) error {
		mimeType, mapping, thumbnail, fulltext, orientation, err = (*index).Index(path, nil)

		return err
	})
	if err != nil {
		common.Error(err)
		return
	}

	if len(mimeType) > 0 {
		bucket.FileType = append(bucket.FileType, mimeType)
	}

	for k, v := range *mapping {
		bucket.Prop[k] = v
	}

	bucket.FileLen = append(bucket.FileLen, n)

	common.Debug("%s: %s", (*uid).String(), hex.EncodeToString(*h))

	err = database.Exec("db", func(db *database.Database) error {
		return (*db).SaveBucket(&bucket, nil)
	})
	if err != nil {
		common.Error(err)
		return
	}
}

func (fs *Fs) Rebuild() (int, error) {
	cluster.Lock(cluster.STORAGE(fs.name))
	defer cluster.Unlock(cluster.STORAGE(fs.name))

	err := database.Exec("db", func(db *database.Database) error {
		return (*db).SwitchIndices([]interface{}{models.NewBucket()}, false)
	})
	if err != nil {
		return -1, err
	}
	defer common.Error(database.Exec("db", func(db *database.Database) error {
		return (*db).SwitchIndices([]interface{}{models.NewBucket()}, true)
	}))

	c := 0
	wg := sync.WaitGroup{}

	for _, volume := range fs.volumes {
		if volume.Name != UNZIP {
			err := filepath.Walk(volume.Path, func(path string, info os.FileInfo, err error) error {
				if !info.IsDir() {
					c++
					path = path[len(volume.Path)+1:]

					wg.Add(1)
					go fs.rebuildBucket(&wg, NewFsUID(path))
				}

				return nil
			})

			if err != nil {
				return -1, err
			}
		}
	}

	wg.Wait()

	return c, nil
}
