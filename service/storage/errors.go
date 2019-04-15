package storage

import (
	"fmt"

	"github.com/mpetavy/common"
)

type ErrInvalidUID struct {
	uid string
}

func (e *ErrInvalidUID) Error() string {
	return fmt.Sprintf("invalid uid: %s", e.uid)
}

type ErrNoVolumesDefined struct {
}

func (e *ErrNoVolumesDefined) Error() string {
	return fmt.Sprintf("no volumes defined")
}

type ErrVolumePathNotFound struct {
	volume string
	path   string
}

func (e *ErrVolumePathNotFound) Error() string {
	return fmt.Sprintf("ShaVolume %s path not found: %s", e.volume, e.path)
}

type ErrInvalidVolumeName struct {
	volume string
}

func (e *ErrInvalidVolumeName) Error() string {
	return fmt.Sprintf("ShaVolume name not allowed: %s", e.volume)
}

type ErrObjectAlreadyExists struct {
	volume string
	uid    string
}

func (e *ErrObjectAlreadyExists) Error() string {
	uid := common.Eval(e.uid != "", func() interface{} { return e.uid }, "??")
	return fmt.Sprintf("Object already exists: ShaVolume %s, Value %v", e.volume, uid)
}

type ErrObjectNotFound struct {
	volume string
	uid    string
}

func (e *ErrObjectNotFound) Error() string {
	uid := common.Eval(e.uid != "", func() interface{} { return e.uid }, "??")
	return fmt.Sprintf("Object not found: ShaVolume %s, Value %v", e.volume, uid)
}
