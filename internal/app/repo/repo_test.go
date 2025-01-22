package repo

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAppRepo(t *testing.T) {
	r, err := NewAppRepo(nil, "", "")
	assert.NoError(t, err)
	assert.NotNil(t, r)

	_, ok := r.(*AppRepoInmem)
	assert.True(t, ok)

	db := &sql.DB{}
	r, err = NewAppRepo(db, "", "")
	assert.NoError(t, err)
	assert.NotNil(t, r)

	_, ok = r.(*AppRepoPostgres)
	assert.True(t, ok)
}
