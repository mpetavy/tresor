package hash

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
)

const (
	MD5    = "md5"
	SHA1   = "sha1"
	SHA256 = "sha256"
	SHA512 = "sha512"
)

type ErrUnknownHash struct {
	Algorithm string
}

func (e *ErrUnknownHash) Error() string {
	return fmt.Sprintf("Unknown algorithmn algorithm: %s", e.Algorithm)
}

func New(alg string) (hash.Hash, error) {
	switch alg {
	case MD5:
		return md5.New(), nil
	case SHA1:
		return sha1.New(), nil
	case SHA256:
		return sha256.New(), nil
	case SHA512:
		return sha512.New(), nil
	default:
		return nil, &ErrUnknownHash{alg}
	}
}
