package repo

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUserRepo(t *testing.T) {
	r, err := NewUserRepo(nil, "")
	assert.NoError(t, err)
	assert.NotNil(t, r)

	_, ok := r.(*UserRepoInmem)
	assert.True(t, ok)

	db := &sql.DB{}
	r, err = NewUserRepo(db, "")
	assert.NoError(t, err)
	assert.NotNil(t, r)

	_, ok = r.(*UserRepoPostgres)
	assert.True(t, ok)
}
