package utils

import (
	"github.com/mpetavy/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateHierarchicalPath(t *testing.T) {
	p, err := CreateHierarchicalPath(false, 1)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, common.CleanPath("000000000000/000000000000/000000000000/000000000001"), p)

	p, err = CreateHierarchicalPath(false, 1001)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, common.CleanPath("000000000000/000000000000/000000001000/000000001001"), p)

	p, err = CreateHierarchicalPath(false, 1000001)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, common.CleanPath("000000000000/000001000000/000001000000/000001000001"), p)

	p, err = CreateHierarchicalPath(false, 1000000001)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, common.CleanPath("001000000000/001000000000/001000000000/001000000001"), p)

	p, err = CreateHierarchicalPath(false, 1234567890)
	if err != nil {
		t.Error(err)
	}
	assert.Equal(t, common.CleanPath("001000000000/001234000000/001234567000/001234567890"), p)
}
