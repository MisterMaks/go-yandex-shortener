package repo

import (
	"database/sql"

	"github.com/MisterMaks/go-yandex-shortener/internal/user/usecase"
)

func NewUserRepo(db *sql.DB, filename string) (usecase.UserRepoInterface, error) {
	var userRepo usecase.UserRepoInterface
	var err error

	switch db {
	case nil:
		userRepo, err = NewUserRepoInmem(filename)
		if err != nil {
			return nil, err
		}
	default:
		userRepo, err = NewUserRepoPostgres(db)
		if err != nil {
			return nil, err
		}
	}

	return userRepo, nil
}
