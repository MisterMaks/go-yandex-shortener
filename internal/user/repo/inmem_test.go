package repo

import (
	"os"
	"sync"
	"testing"

	"github.com/MisterMaks/go-yandex-shortener/internal/user"
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

	r, err := NewUserRepoInmem(tmpFile.Name())
	assert.NoError(t, err)
	assert.NotNil(t, r)

	r, err = NewUserRepoInmem("")
	assert.NoError(t, err)
	assert.Equal(t, &UserRepoInmem{
		users:    []*user.User{},
		mu:       sync.RWMutex{},
		producer: nil,
		maxID:    0,
	}, r)
}

func TestUserRepoInmem_Close(t *testing.T) {
	tmpFile, err := os.CreateTemp("", TestFilenamePattern)
	require.NoError(t, err)
	defer func() {
		err = os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	r, err := NewUserRepoInmem(tmpFile.Name())
	require.NoError(t, err)
	assert.NotNil(t, r)

	err = r.Close()
	assert.NoError(t, err)
}

func TestUserRepoInmem_CreateUser(t *testing.T) {
	tmpFile, err := os.CreateTemp("", TestFilenamePattern)
	require.NoError(t, err)
	defer func() {
		err = os.Remove(tmpFile.Name())
		require.NoError(t, err)
	}()

	r, err := NewUserRepoInmem(tmpFile.Name())
	require.NoError(t, err)
	assert.NotNil(t, r)

	u, err := r.CreateUser()
	require.NoError(t, err)
	assert.NotNil(t, u)
}
