package storage

import (
	"fmt"

	"github.com/mpetavy/common"
)

type ErrInvalidUID struct {
	Uid string
}

func (e *ErrInvalidUID) Error() string {
	return fmt.Sprintf("invalid uid: %s", e.Uid)
}

type ErrNoVolumesDefined struct {
}

func (e *ErrNoVolumesDefined) Error() string {
	return fmt.Sprintf("no volumes defined")
}

type ErrVolumePathNotFound struct {
	Volume string
	Path   string
}

func (e *ErrVolumePathNotFound) Error() string {
	return fmt.Sprintf("ShaVolume %s path not found: %s", e.Volume, e.Path)
}

type ErrInvalidVolumeName struct {
	Volume string
}

func (e *ErrInvalidVolumeName) Error() string {
	return fmt.Sprintf("ShaVolume name not allowed: %s", e.Volume)
}

type ErrObjectAlreadyExists struct {
	Volume string
	Uid    string
}

func (e *ErrObjectAlreadyExists) Error() string {
	uid := common.Eval(e.Uid != "", e.Uid, "??")
	return fmt.Sprintf("Object already exists: ShaVolume %s, Value %v", e.Volume, uid)
}

type ErrObjectNotFound struct {
	Volume string
	Uid    string
}

func (e *ErrObjectNotFound) Error() string {
	uid := common.Eval(e.Uid != "", e.Uid, "??")
	return fmt.Sprintf("Object not found: ShaVolume %s, Value %v", e.Volume, uid)
}
