package repo

import (
	"database/sql"
	"github.com/MisterMaks/go-yandex-shortener/internal/user"
)

type UserRepoPostgres struct {
	db *sql.DB
}

func NewUserRepoPostgres(db *sql.DB) (*UserRepoPostgres, error) {
	return &UserRepoPostgres{db: db}, nil
}

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
