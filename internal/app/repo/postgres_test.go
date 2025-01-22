package repo

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
	userRepoInternal "github.com/MisterMaks/go-yandex-shortener/internal/user/repo"
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
	return goose.RunContext(ctx, "down-to", db, "../../../migrations/", "0")
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
	if testing.Short() || DSN == "" {
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

func TestNewAppRepoPostgres(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	_, err := NewAppRepoPostgres(te.DB)
	assert.NoError(t, err, "Failed to run NewAppRepoPostgres()")
}

func TestAppRepoPostgres_GetOrCreateURL(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	r, err := NewAppRepoPostgres(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepoPostgres()")

	ur, err := userRepoInternal.NewUserRepoPostgres(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepoPostgres()")

	user, err := ur.CreateUser()
	require.NoError(t, err)

	testID := "1"
	testURLStr := "https://test.ru"
	testUserID := user.ID

	testURL := &app.URL{ID: testID, URL: testURLStr, UserID: testUserID, IsDeleted: false}

	actualURL, err := r.GetOrCreateURL(testID, testURLStr, testUserID)
	require.NoError(t, err)
	assert.Equal(t, testURL, actualURL)

	user2, err := ur.CreateUser()
	require.NoError(t, err)

	actualURL, err = r.GetOrCreateURL("2", testURLStr, user2.ID)
	require.NoError(t, err)
	assert.Equal(t, testURL, actualURL)

	_, err = r.GetOrCreateURL("1", "https://test2.ru", user2.ID)
	require.Error(t, err)
}

func TestAppRepoPostgres_GetURL(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	r, err := NewAppRepoPostgres(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepoPostgres()")

	ur, err := userRepoInternal.NewUserRepoPostgres(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepoPostgres()")

	user, err := ur.CreateUser()
	require.NoError(t, err)

	testID := "1"
	testURLStr := "https://test.ru"
	testUserID := user.ID

	testURL := &app.URL{ID: testID, URL: testURLStr, UserID: testUserID, IsDeleted: false}

	_, err = r.GetOrCreateURL(testID, testURLStr, testUserID)
	require.NoError(t, err)

	actualURL, err := r.GetURL(testID)
	require.NoError(t, err)
	assert.Equal(t, testURL, actualURL)

	_, err = r.GetURL("2")
	require.Error(t, err)
}

func TestAppRepoPostgres_CheckIDExistence(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	r, err := NewAppRepoPostgres(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepoPostgres()")

	ur, err := userRepoInternal.NewUserRepoPostgres(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepoPostgres()")

	user, err := ur.CreateUser()
	require.NoError(t, err)

	testID := "1"
	testURLStr := "https://test.ru"
	testUserID := user.ID

	_, err = r.GetOrCreateURL(testID, testURLStr, testUserID)
	require.NoError(t, err)

	ok, err := r.CheckIDExistence(testID)
	require.NoError(t, err)
	require.True(t, ok)

	ok, err = r.CheckIDExistence("2")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestAppRepoPostgres_Ping(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	r, err := NewAppRepoPostgres(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepoPostgres()")

	err = r.Ping()
	require.NoError(t, err)

	err = te.DB.Close()
	require.NoError(t, err)

	err = r.Ping()
	require.Error(t, err)
}

func TestAppRepoPostgres_GetOrCreateURLs(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	r, err := NewAppRepoPostgres(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepoPostgres()")

	ur, err := userRepoInternal.NewUserRepoPostgres(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepoPostgres()")

	user, err := ur.CreateUser()
	require.NoError(t, err)

	user2, err := ur.CreateUser()
	require.NoError(t, err)

	user3, err := ur.CreateUser()
	require.NoError(t, err)

	testURLs := []*app.URL{
		{ID: "1", URL: "https://test.ru", UserID: user.ID, IsDeleted: false},
		{ID: "2", URL: "https://test2.ru", UserID: user.ID, IsDeleted: false},
		{ID: "3", URL: "https://test3.ru", UserID: user2.ID, IsDeleted: false},
	}

	actualURLs, err := r.GetOrCreateURLs(testURLs)
	require.NoError(t, err)
	assert.Equal(t, testURLs, actualURLs)

	testURLs2 := []*app.URL{
		{ID: "4", URL: "https://test.ru", UserID: user3.ID, IsDeleted: false},
		{ID: "5", URL: "https://test2.ru", UserID: user3.ID, IsDeleted: false},
		{ID: "6", URL: "https://test3.ru", UserID: user3.ID, IsDeleted: false},
	}

	actualURLs, err = r.GetOrCreateURLs(testURLs2)
	require.NoError(t, err)
	assert.Equal(t, testURLs, actualURLs)
}

func TestAppRepoPostgres_GetUserURLs(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	r, err := NewAppRepoPostgres(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepoPostgres()")

	ur, err := userRepoInternal.NewUserRepoPostgres(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepoPostgres()")

	user, err := ur.CreateUser()
	require.NoError(t, err)

	user2, err := ur.CreateUser()
	require.NoError(t, err)

	testURLs := []*app.URL{
		{ID: "1", URL: "https://test.ru", UserID: user.ID, IsDeleted: false},
		{ID: "2", URL: "https://test2.ru", UserID: user.ID, IsDeleted: false},
		{ID: "3", URL: "https://test3.ru", UserID: user2.ID, IsDeleted: false},
	}

	actualURLs, err := r.GetOrCreateURLs(testURLs)
	require.NoError(t, err)
	assert.Equal(t, testURLs, actualURLs)

	userURLs, err := r.GetUserURLs(user.ID)
	require.NoError(t, err)
	assert.Equal(t, testURLs[:2], userURLs)

	user2URLs, err := r.GetUserURLs(user2.ID)
	require.NoError(t, err)
	assert.Equal(t, testURLs[2:], user2URLs)
}

func TestAppRepoPostgres_DeleteUserURLs(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	r, err := NewAppRepoPostgres(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepoPostgres()")

	ur, err := userRepoInternal.NewUserRepoPostgres(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepoPostgres()")

	user, err := ur.CreateUser()
	require.NoError(t, err)

	testURLs := []*app.URL{
		{ID: "1", URL: "https://test.ru", UserID: user.ID, IsDeleted: false},
		{ID: "2", URL: "https://test2.ru", UserID: user.ID, IsDeleted: false},
	}

	actualURLs, err := r.GetOrCreateURLs(testURLs)
	require.NoError(t, err)
	assert.Equal(t, testURLs, actualURLs)

	err = r.DeleteUserURLs(testURLs[:1])
	require.NoError(t, err)

	u, err := r.GetURL(testURLs[0].ID)
	require.NoError(t, err)
	assert.True(t, u.IsDeleted)

	u, err = r.GetURL(testURLs[1].ID)
	require.NoError(t, err)
	assert.False(t, u.IsDeleted)
}

func TestAppRepoPostgres_Close(t *testing.T) {
	te := newTestEnvironment(DSN, t)
	defer te.clean()

	r, err := NewAppRepoPostgres(te.DB)
	require.NoError(t, err, "Failed to run NewAppRepoPostgres()")

	err = r.Close()
	assert.NoError(t, err)
}
