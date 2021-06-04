package storage

import (
	"container/list"
	"encoding/hex"
	"io"
	"reflect"
	"strings"

	"github.com/mpetavy/tresor/models"
	"github.com/mpetavy/tresor/service/database"
	"github.com/mpetavy/tresor/service/index"

	"github.com/mpetavy/tresor/service/cluster"

	"fmt"

	"path/filepath"
	"sync"

	"math"
	"os"
	"strconv"

	"github.com/mpetavy/common"
	"github.com/mpetavy/tresor/cache"
	"github.com/mpetavy/tresor/hash"
)

const (
	TYPE_SHA   = "sha"
	SHA_VOLUME = "SHA_VOLUME_"
)

type ShaUID struct {
	Id      int
	Version int
	Object  string
}

func (uid *ShaUID) withoutObject() *ShaUID {
	return &ShaUID{Id: uid.Id, Version: uid.Version}
}

func NewShaUID(id int, version int, object string) *ShaUID {
	return &ShaUID{Id: id, Version: version, Object: object}
}

func ParseShaUID(s string) (*ShaUID, error) {
	var parseErr = &ErrInvalidUID{s}
	var err error

	if s == "" {
		return nil, parseErr
	}

	uid := &ShaUID{}

	elements := strings.Split(s, "|")

	if strings.Contains(elements[0], ".") {
		elements := strings.Split(elements[0], ".")

		uid.Id, err = strconv.Atoi(elements[0])
		if common.Error(err) {
			return nil, parseErr
		}

		if len(elements) > 1 {
			uid.Version, err = strconv.Atoi(elements[1])
			if common.Error(err) {
				return nil, parseErr
			}
		}
	}

	if len(elements) > 1 {
		uid.Object = elements[1]
	}

	return uid, nil
}

func (uid *ShaUID) String() string {
	sb := strings.Builder{}

	sb.WriteString(strconv.Itoa(uid.Id))
	sb.WriteString(".")
	sb.WriteString(strconv.Itoa(uid.Version))
	if uid.Object != "" {
		sb.WriteString("|")
		sb.WriteString(uid.Object)
	}

	return sb.String()
}

type Sha struct {
	volumes map[string]*ShaVolume
	uid     int
	mu      *sync.Mutex
}

type ShaVolume struct {
	Name string
	Path string
	Flat bool
	Zip  bool
}

func NewShaVolume(name string, path string, flat bool, zip bool) (*ShaVolume, error) {
	if name == UNZIP {
		return nil, &ErrInvalidVolumeName{name}
	}

	volume := ShaVolume{Name: name, Path: common.CleanPath(path), Flat: flat, Zip: zip}

	if !common.FileExists(path) {
		return nil, &common.ErrFileNotFound{path}
	}

	return &volume, nil
}

func NewSha() (*Sha, error) {
	sha := &Sha{volumes: make(map[string]*ShaVolume), uid: 0, mu: new(sync.Mutex)}

	path, err := common.CreateTempDir()
	if common.Error(err) {
		return nil, err
	}

	sha.AddVolume(&ShaVolume{Name: UNZIP, Path: common.CleanPath(path), Flat: false, Zip: false})

	return sha, nil
}

func (sha *Sha) Init(cfg *Cfg) error {
	for i := 0; i < len(cfg.Volumes); i++ {
		path := common.CleanPath(cfg.Volumes[i].Path)

		if !common.FileExists(path) {
			return &ErrVolumePathNotFound{Volume: cfg.Volumes[i].Name, Path: path}
		}

		vol, err := NewShaVolume(cfg.Volumes[i].Name, path, cfg.Volumes[i].Flat, cfg.Volumes[i].Zip)
		if common.Error(err) {
			return err
		}

		sha.AddVolume(vol)
	}

	return nil
}

func (sha *Sha) Start() error {
	return nil
}

func (sha *Sha) Stop() error {
	return nil
}

func (sha *Sha) Volume(n string) *ShaVolume {
	return sha.volumes[n]
}

func (sha *Sha) Volumes() []string {
	l := make([]string, 0)

	for _, v := range sha.volumes {
		l = append(l, v.Name)
	}

	return l
}

func (sha *Sha) AddVolume(v *ShaVolume) {
	sha.volumes[v.Name] = v
}

func (sha *Sha) RemoveVolume(v *ShaVolume) {
	delete(sha.volumes, v.Name)
}

func (sha *Sha) nextUID() int {
	cluster.Lock(cluster.ByStorage())
	defer cluster.Unlock(cluster.ByStorage())

	sha.uid++

	return sha.uid
}

func (sha *Sha) CurrentVersion(uid *ShaUID) (int, error) {
	return sha.currentVersion(uid, "")
}

func (sha *Sha) currentVersion(_uid *ShaUID, path string) (int, error) {
	uid := _uid

	if path == "" {
		uid.Object = ""
		uid.Version = 0

		var err error

		_, path, err = sha.find(uid, nil)
		if common.Error(err) {
			return -1, err
		}
	}

	files, err := os.ReadDir(path)
	if common.Error(err) {
		return -1, err
	}

	if len(files) == 0 {
		return 1, nil
	}

	currentVersion := 0

	for _, file := range files {
		if file.IsDir() {
			v, err := strconv.Atoi(file.Name())
			if common.Error(err) {
				return -1, err
			}

			if v > currentVersion {
				currentVersion = v
			}
		}
	}

	return currentVersion + 1, nil
}

func createShaPath(rootDir string, uid *ShaUID, flat bool, zip bool) (string, error) {
	var path string

	if rootDir != "" {
		path = rootDir + string(filepath.Separator)
	}

	object := ""
	if uid.Object != "" {
		object = fmt.Sprintf("%s%s", string(filepath.Separator), uid.Object)
	}

	if flat {
		path += fmt.Sprintf("%d", uid.Id)
		path += string(filepath.Separator)

		if zip {
			if uid.Version > 1 {
				path += fmt.Sprintf("%012d.%012d.zip", uid.Id, uid.Version-1)
			} else {
				path += fmt.Sprintf("%012d.zip", uid.Id)
			}
		} else {
			if uid.Version > 1 {
				path += fmt.Sprintf("%012d", uid.Version-1)
			}
		}

		if uid.Object != "" {
			if zip {
				path += ":"
			}

			path += object
		}
	} else {
		for i := 3; i >= 0; i-- {
			if i < 3 {
				path += string(filepath.Separator)
			}

			t := math.Pow(float64(1000), float64(i))
			v := (uid.Id / int(t)) * int(t)

			path += fmt.Sprintf("%012d", v)
		}

		if uid.Version > 1 {
			if zip {
				path += "."
			} else {
				path += string(filepath.Separator)
			}
			path += fmt.Sprintf("%012d", uid.Version-1)
		}

		if zip {
			path += ".zip"
		}

		if object != "" {
			path += object
		}
	}

	return common.CleanPath(path), nil
}

func (sha *Sha) find(uid *ShaUID, options *Options) (*ShaVolume, string, error) {
	var listVolumes list.List

	volumeName, ok := cache.Get(SHA_VOLUME, strconv.Itoa(uid.Id))

	if ok {
		_, valid := sha.volumes[volumeName.(string)]

		if valid {
			listVolumes.PushFront(volumeName)
		} else {
			ok = false
		}
	}

	for key := range sha.volumes {
		if key != UNZIP && (!ok || volumeName != key) {
			listVolumes.PushBack(key)
		}
	}
	listVolumes.PushBack(UNZIP)

	for i := 0; i < listVolumes.Len(); i++ {
		vn := getFromList(listVolumes, i).(string)
		volume := sha.volumes[vn]

		path, err := createShaPath(volume.Path, uid, volume.Flat, false)
		if common.Error(err) {
			return nil, "", err
		}

		b := common.FileExists(path)

		if !b && volume.Zip {
			path, err = createShaPath(volume.Path, uid.withoutObject(), volume.Flat, true)
			if common.Error(err) {
				return nil, "", err
			}

			b = common.FileExists(path)

			if b {
				volume = sha.volumes[UNZIP]

				uidDir, err := createShaPath(volume.Path, uid.withoutObject(), volume.Flat, volume.Zip)
				if common.Error(err) {
					return nil, "", err
				}

				err = os.MkdirAll(uidDir, common.DefaultDirMode)
				if common.Error(err) {
					return nil, "", err
				}

				err = common.Unzip(uidDir, path)
				if common.Error(err) {
					return nil, "", err
				}

				path, err = createShaPath(volume.Path, uid, volume.Flat, volume.Zip)
				if common.Error(err) {
					return nil, "", err
				}

				b = common.FileExists(path)
			}
		}

		if b {
			cache.Put(SHA_VOLUME, strconv.Itoa(uid.Id), volume.Name)

			return volume, path, nil
		}
	}

	return nil, "", &ErrObjectNotFound{"??", uid.String()}
}

func (sha *Sha) Store(suid string, source io.Reader, options *Options) (string, *[]byte, error) {
	uid, err := ParseShaUID(suid)
	if common.Error(err) {
		return "", nil, err
	}

	if uid.Id != 0 {
		cluster.Lock(cluster.ByStorageUid(strconv.Itoa(uid.Id)))
		defer cluster.Unlock(cluster.ByStorageUid(strconv.Itoa(uid.Id)))
	}

	var volume *ShaVolume

	if uid.Id != 0 {
		var err error
		var path string

		volume, path, err = sha.find(uid.withoutObject(), options)
		if common.Error(err) {
			return "", nil, err
		}

		if uid.Version == 0 {
			v, err := sha.currentVersion(nil, path)
			if common.Error(err) {
				return "", nil, err
			}

			uid.Version = v + 1
		}
	} else {
		if options != nil && len(options.VolumeName) > 0 {
			var ok bool

			volume, ok = sha.volumes[options.VolumeName]
			if !ok {
				return "", nil, &ErrInvalidVolumeName{options.VolumeName}
			}
		} else {
			if len(sha.volumes) > 0 {
				volume = sha.volumes[reflect.ValueOf(sha.Volumes).MapKeys()[0].String()]
			} else {
				return "", nil, &ErrNoVolumesDefined{}
			}
		}

		uid.Id = sha.nextUID()
		uid.Version = 1
	}

	cluster.Lock(cluster.ByStorageVolume(volume.Name))
	defer cluster.Unlock(cluster.ByStorageVolume(volume.Name))

	path, err := createShaPath(volume.Path, uid, volume.Flat, false)
	if common.Error(err) {
		return "", nil, err
	}

	if common.FileExists(path) {
		sha, err := common.FileSize(path)
		if common.Error(err) {
			return "", nil, err
		}

		if sha > 0 {
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
		common.Ignore(dest.Close())
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

	cache.Put(SHA_VOLUME, strconv.Itoa(uid.Id), volume.Name)

	return uid.String(), &digest, nil
}

func (sha *Sha) Load(suid string, dest io.Writer, options *Options) (string, *[]byte, int64, error) {
	uid, err := ParseShaUID(suid)
	if common.Error(err) {
		return "", nil, -1, err
	}

	cluster.Lock(cluster.ByStorageUid(strconv.Itoa(uid.Id)))
	defer cluster.Unlock(cluster.ByStorageUid(strconv.Itoa(uid.Id)))

	_, path, err := sha.find(uid, options)
	if common.Error(err) {
		return "", nil, -1, err
	}

	source, err := os.Open(path)
	if common.Error(err) {
		return "", nil, -1, err
	}
	defer func() {
		common.Ignore(source.Close())
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

func (sha *Sha) Delete(suid string, options *Options) error {
	uid, err := ParseShaUID(suid)
	if common.Error(err) {
		return err
	}

	cluster.Lock(cluster.ByStorageUid(strconv.Itoa(uid.Id)))
	defer cluster.Unlock(cluster.ByStorageUid(strconv.Itoa(uid.Id)))

	var volume *ShaVolume
	var path string

	volume, path, err = sha.find(uid, options)
	if common.Error(err) {
		return err
	}

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

			files, err := os.ReadDir(path)
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

	cache.Remove(SHA_VOLUME, strconv.Itoa(uid.Id))

	return nil
}

func (sha *Sha) rebuildBucket(wg *sync.WaitGroup, uid *ShaUID, version int) {
	defer wg.Done()

	bucket := models.NewBucket()
	bucket.Uid = uid.String()

	wgIndex := sync.WaitGroup{}
	mapIndex := make(map[int]index.IndexResult)

	page := 1

	for ; ; page++ {
		uid.Object = PAGE + "." + strconv.Itoa(page)

		path, h, n, err := sha.Load(uid.String(), io.Discard, nil)
		if common.Error(err) {
			if page == 1 {
				common.Error(err)
			}
			break
		}

		bucket.FileNames = append(bucket.FileNames, uid.Object)
		bucket.FileHashes = append(bucket.FileHashes, hex.EncodeToString(*h))

		wgIndex.Add(1)
		go func(page int, path string) {
			ir := index.IndexResult{}

			err = index.Exec(func(index index.Handle) error {
				ir.MimeType, ir.Mapping, ir.Thumbnail, ir.Fulltext, ir.Orientation, err = index.Index(path, nil)

				return err
			})
			if common.Error(err) {
				common.Error(err)
				return
			}

			mapIndex[page] = ir

			wgIndex.Done()
		}(page, path)

		bucket.FileSizes = append(bucket.FileSizes, n)

		common.Debug("%s: %s", (*uid).String(), hex.EncodeToString(*h))
	}

	wgIndex.Wait()

	for i := 1; i < page; i++ {
		ir := mapIndex[i]
		bucket.FileTypes = append(bucket.FileTypes, ir.MimeType)
	}

	err := database.Exec(func(db database.Handle) error {
		return db.SaveBucket(&bucket, nil)
	})
	if common.Error(err) {
		common.Error(err)
		return
	}
}

func (sha *Sha) Rebuild() (int, error) {
	cluster.Lock(cluster.ByStorage())
	defer cluster.Unlock(cluster.ByStorage())

	c := 0
	wg := sync.WaitGroup{}

loop_id:
	for id := 1; true; id++ {
	loop_version:
		for version := 1; ; version++ {
			uid := NewShaUID(id, version, "")

			_, _, err := sha.find(uid, nil)
			if _, ok := err.(*ErrObjectNotFound); ok {
				if version == 1 {
					break loop_id
				} else {
					break loop_version
				}
			}

			c++

			wg.Add(1)
			go sha.rebuildBucket(&wg, uid, version)
		}
	}

	wg.Wait()

	return c, nil
}
