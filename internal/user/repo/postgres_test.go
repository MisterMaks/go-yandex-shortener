package repo

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var DSN = os.Getenv("TEST_DATABASE_URI")

func upMigrations(dsn string) error {
	db, err := goose.OpenDBWithDriver("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx := context.Background()
	return goose.RunContext(ctx, "up", db, "../../../migrations/")
}

func downMigrations(dsn string) error {
	db, err := goose.OpenDBWithDriver("postgres", dsn)
	if err != nil {
		return err
	}
	defer db.Close()
	ctx := context.Background()
	return goose.RunContext(ctx, "down", db, "../../../migrations/")
}

func connectPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func newTestEnvironment(dsn string, t *testing.T) *TestEnvironment {
	if testing.Short() {
		t.Skip()
	}

	err := upMigrations(dsn)
	require.NoError(t, err, "Failed to apply migrations")

	db, err := connectPostgres(DSN)
	require.NoError(t, err, "Failed to connect to Postgres")

	return &TestEnvironment{
		DSN: dsn,
		T:   t,
		DB:  db,
	}
}

type TestEnvironment struct {
	DSN string
	T   *testing.T
	DB  *sql.DB
}

func (te *TestEnvironment) clean() {
	err := te.DB.Close()
	assert.NoError(te.T, err, "Failed to close DB")
	err = downMigrations(te.DSN)
	require.NoError(te.T, err, "Failed to roll back migrations")
}

func TestNewUserRepoPostgres(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	_, err := NewUserRepoPostgres(te.DB)
	assert.NoError(t, err, "Failed to run NewUserRepoPostgres()")
}

func TestAppRepo_CreateUser(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	userRepo, err := NewUserRepoPostgres(te.DB)
	require.NoError(t, err, "Failed to run NewUserRepoPostgres()")

	u, err := userRepo.CreateUser()
	require.NoError(t, err)
	assert.NotNil(t, u)
}
