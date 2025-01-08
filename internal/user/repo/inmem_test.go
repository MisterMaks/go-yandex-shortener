package repo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	TestFilenamePattern string = "internal_user_repo_inmem_test_*.json"
)

func TestNewUserRepoInmem(t *testing.T) {
	tmpFile, err := os.CreateTemp("", TestFilenamePattern)
	require.NoError(t, err)
	defer func() {
		err = os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	appRepoInMem, err := NewUserRepoInmem(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, appRepoInMem)
}
