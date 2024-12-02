package repo

import (
	"database/sql"

	"github.com/MisterMaks/go-yandex-shortener/internal/user"
)

// UserRepoPostgres user data storage in PostgreSQL.
type UserRepoPostgres struct {
	db *sql.DB
}

// NewUserRepoPostgres creates *UserRepoPostgres.
func NewUserRepoPostgres(db *sql.DB) (*UserRepoPostgres, error) {
	return &UserRepoPostgres{db: db}, nil
}

// CreateUser create user in DB.
func (urp *UserRepoPostgres) CreateUser() (*user.User, error) {
	query := `INSERT INTO "user" DEFAULT VALUES RETURNING id;`
	var id uint
	err := urp.db.QueryRow(query).Scan(&id)
	if err != nil {
		return nil, err
	}
	u := &user.User{ID: id}
	return u, nil
}
