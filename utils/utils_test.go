package utils

import (
	"github.com/mpetavy/common"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCreateHierarchicalPath(t *testing.T) {
	p, err := CreateHierarchicalPath(false, 1)
	if common.Error(err) {
		t.Error(err)
	}
	require.Equal(t, common.CleanPath("000000000000/000000000000/000000000000/000000000001"), p)

	p, err = CreateHierarchicalPath(false, 1001)
	if common.Error(err) {
		t.Error(err)
	}
	require.Equal(t, common.CleanPath("000000000000/000000000000/000000001000/000000001001"), p)

	p, err = CreateHierarchicalPath(false, 1000001)
	if common.Error(err) {
		t.Error(err)
	}
	require.Equal(t, common.CleanPath("000000000000/000001000000/000001000000/000001000001"), p)

	p, err = CreateHierarchicalPath(false, 1000000001)
	if common.Error(err) {
		t.Error(err)
	}
	require.Equal(t, common.CleanPath("001000000000/001000000000/001000000000/001000000001"), p)

	p, err = CreateHierarchicalPath(false, 1234567890)
	if common.Error(err) {
		t.Error(err)
	}
	require.Equal(t, common.CleanPath("001000000000/001234000000/001234567000/001234567890"), p)
}
