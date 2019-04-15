package storage

import (
	"bytes"
	"strconv"
	"testing"

	"fmt"
	"os"
	"path/filepath"
	"sync"

	"encoding/hex"
	"flag"
	"io/ioutil"

	"github.com/mpetavy/common"
	"github.com/stretchr/testify/assert"
)

var count = flag.Int("count", 100, "amount of documents to test with")

func TestMain(m *testing.M) {
	defer common.Cleanup()
	flag.Parse()
	common.Exit(m.Run())
}

func TestCreatePath(t *testing.T) {
	tests := []struct {
		Value    *ShaUID
		Flat     bool
		Zip      bool
		Expected string
	}{
		{NewShaUID(1, 1, ""), false, false, common.CleanPath("000000000000/000000000000/000000000000/000000000001")},
		{NewShaUID(1001, 1, ""), false, false, common.CleanPath("000000000000/000000000000/000000001000/000000001001")},
		{NewShaUID(1000001, 1, ""), false, false, common.CleanPath("000000000000/000001000000/000001000000/000001000001")},
		{NewShaUID(1000000001, 1, ""), false, false, common.CleanPath("001000000000/001000000000/001000000000/001000000001")},
		{NewShaUID(1234567890, 1, ""), false, false, common.CleanPath("001000000000/001234000000/001234567000/001234567890")},
		{NewShaUID(1, 2, ""), false, false, common.CleanPath("000000000000/000000000000/000000000000/000000000001/000000000001")},
		{NewShaUID(44, 1, ""), false, false, common.CleanPath("000000000000/000000000000/000000000000/000000000044")},
		{NewShaUID(44, 2, ""), false, false, common.CleanPath("000000000000/000000000000/000000000000/000000000044/000000000001")},
		{NewShaUID(45, 1, ""), false, true, common.CleanPath("000000000000/000000000000/000000000000/000000000045.zip")},
		{NewShaUID(45, 2, ""), false, true, common.CleanPath("000000000000/000000000000/000000000000/000000000045.000000000001.zip")},
		{NewShaUID(44, 1, ""), true, false, common.CleanPath("44")},
		{NewShaUID(44, 2, ""), true, false, common.CleanPath("44/000000000001")},
		{NewShaUID(45, 1, ""), true, true, common.CleanPath("45/000000000045.zip")},
		{NewShaUID(45, 2, ""), true, true, common.CleanPath("45/000000000045.000000000001.zip")},
	}

	for _, test := range tests {
		p, err := createShaPath("", test.Value, test.Flat, test.Zip)
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, test.Expected, p)
	}
}

func TestBasicArchive(t *testing.T) {
	path, err := common.CreateTempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	fs, err := NewSha()
	if err != nil {
		t.Fatal(err)
	}

	v, err := NewShaVolume("test", path, false, false)
	if err != nil {
		t.Fatal(err)
	}

	fs.AddVolume(v)

	for id := 1; id <= 10; id++ {
		uid := NewShaUID(0, 0, "")
		for version := 1; version <= 10; version++ {
			uid.Version = 0
			for page := 1; page <= 10; page++ {
				s := fmt.Sprintf("Doc #%d.%d Page %d", id, version, page)

				uid.Object = PAGE + "." + strconv.Itoa(page)

				suid, _, err := fs.Store(uid.String(), bytes.NewReader([]byte(s)), &Options{VolumeName: "test"})
				if err != nil {
					t.Fatal(err)
				}

				uid, err = ParseShaUID(suid)
				if err != nil {
					t.Fatal(err)
				}

				assert.Equal(t, uid.Id, id, "Correct incremented ID")
			}

			v, err := fs.CurrentVersion(uid)
			if err != nil {
				t.Fatal(err)
			}

			if v != version {
				fs.CurrentVersion(uid)
			}

			assert.Equal(t, v, version, "Correct version")
		}
	}
}

func TestBasicIO(t *testing.T) {
	path, err := common.CreateTempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(path)

	fs, err := NewSha()
	if err != nil {
		t.Fatal(err)
	}

	v, err := NewShaVolume("test", path, false, false)
	if err != nil {
		t.Fatal(err)
	}

	fs.AddVolume(v)

	s := "Hello world!"

	suid, hs, err := fs.Store(NewShaUID(0, 0, PAGE+"."+strconv.Itoa(0)).String(), bytes.NewReader([]byte(s)), &Options{VolumeName: "test"})
	if err != nil {
		t.Fatal(err)
	}

	uid, err := ParseShaUID(suid)
	if err != nil {
		t.Fatal(err)
	}

	var w bytes.Buffer

	hl, _, err := fs.Load(suid, &w, nil)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, s, w.String(), "Content compare")
	assert.Equal(t, hs, hl, "Hash compare")

	err = fs.Delete(suid, nil)
	if err != nil {
		t.Fatal(err)
	}

	path, err = createShaPath(v.Path, uid, false, false)
	if err != nil {
		t.Fatal(err)
	}

	b, err := common.FileExists(path)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, false, b, "ShaUID with object shall not exist")

	path, err = createShaPath(v.Path, uid.withoutObject(), false, false)
	if err != nil {
		t.Fatal(err)
	}

	b, err = common.FileExists(path)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, b, "ShaUID without object shall exist")

	err = fs.Delete(uid.withoutObject().String(), nil)
	if err != nil {
		t.Fatal(err)
	}

	for path != v.Path {
		b, err = common.FileExists(path)
		if err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, false, b, "ShaUID shall not exist")

		path = filepath.Dir(path)
	}
}

func TestFilestorage(t *testing.T) {
	tempDir, err := common.CreateTempDir()
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	t.Logf("test with %d documents", *count)

	t.Logf("create volumes")

	fs, err := NewSha()
	if err != nil {
		t.Fatal(err)
	}

	var volNames [5]string

	for i := 0; i < len(volNames); i++ {
		volNames[i] = fmt.Sprintf("volume #%d", i)
		path := filepath.Join(tempDir, volNames[i])

		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			t.Fatal(err)
		}

		v, err := NewShaVolume(volNames[i], path, false, false)
		if err != nil {
			t.Fatal(err)
		}

		fs.AddVolume(v)
	}

	t.Logf("create documents")

	wg := sync.WaitGroup{}

	type maker struct {
		uid  *ShaUID
		hash [10]*[]byte
	}

	uidsChan := make(chan maker, *count)

	for i := 0; i < *count; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			var m maker

			for page := 0; page < len(m.hash); page++ {
				var err error

				s := fmt.Sprintf("%s;%d", m.uid, page)

				if m.uid == nil {
					m.uid = NewShaUID(0, 0, PAGE+"."+strconv.Itoa(page))
				} else {
					m.uid.Object = PAGE + "." + strconv.Itoa(page)
				}

				suid, hash, err := fs.Store(m.uid.String(), bytes.NewReader([]byte(s)), &Options{volNames[common.Rnd(len(volNames))]})
				if err != nil {
					t.Fatal(err)
				}

				uid, err := ParseShaUID(suid)
				if err != nil {
					t.Fatal(err)
				}

				m.uid = uid
				m.hash[page] = hash
			}

			uidsChan <- m
		}()
	}

	wg.Wait()

	close(uidsChan)

	t.Logf("search documents")

	for m := range uidsChan {
		wg.Add(1)
		go func(m maker) {
			defer wg.Done()

			deletedPage := common.Rnd(len(m.hash)) + 1

			m.uid.Object = PAGE + "." + strconv.Itoa(deletedPage)

			err := fs.Delete(m.uid.String(), nil)
			if err != nil {
				t.Fatal(err)
			}

			for page := 0; page < len(m.hash); page++ {
				if err != nil {
					t.Fatal(err)
				}

				m.uid.Object = PAGE + "." + strconv.Itoa(page)

				hash, _, err := fs.Load(m.uid.String(), ioutil.Discard, nil)

				if page == deletedPage {
					hash, _, err = fs.Load(m.uid.String(), ioutil.Discard, nil)
					_, ok := err.(*ErrObjectNotFound)
					assert.True(t, ok, "load on deleted page gave no error")
				} else {
					if err != nil {
						t.Fatal(err)
					}

					assert.Equal(t, *m.hash[page], *hash, "hash values are different")
				}
			}

			err = fs.Delete(m.uid.withoutObject().String(), nil)
			if err != nil {
				t.Fatal(err)
			}

			_, _, err = fs.find(m.uid.withoutObject(), nil)
			_, ok := err.(*ErrObjectNotFound)

			if !ok {
				t.Fatal(err)
			}

			assert.True(t, ok, "load on deleted page gave no error")
		}(m)
	}

	wg.Wait()
}

func TestSample(t *testing.T) {
	fs, err := NewSha()
	if err != nil {
		t.Fatal(err)
	}

	v, err := NewShaVolume("sample", common.CleanPath("~/archive/sample"), true, false)
	if err != nil {
		t.Fatal(err)
	}

	fs.AddVolume(v)
loop_id:
	for id := 1; id < 46; id++ {
	loop_version:
		for version := 0; ; version++ {
			uid := NewShaUID(id, version, "")

			_, _, err := fs.find(uid, nil)
			if _, ok := err.(*ErrObjectNotFound); ok {
				break loop_version
			}
		loop_page:
			for page := 1; ; page++ {
				uid.Object = PAGE + "." + strconv.Itoa(page)

				hash, _, err := fs.Load(uid.String(), ioutil.Discard, nil)
				if err != nil {
					if page == 1 {
						break loop_id
					} else {
						break loop_page
					}
				}

				t.Logf("%s: %s", (*uid).String(), hex.EncodeToString(*hash))
			}
		}
	}
}
