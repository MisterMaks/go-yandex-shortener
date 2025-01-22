package repo

import (
	"database/sql"

	"github.com/MisterMaks/go-yandex-shortener/internal/app/usecase"
)

func NewAppRepo(
	db *sql.DB,
	filename string,
	deletedURLsFilename string,
) (usecase.AppRepoInterface, error) {
	var appRepo usecase.AppRepoInterface
	var err error

	switch db {
	case nil:
		appRepo, err = NewAppRepoInmem(filename, deletedURLsFilename)
		if err != nil {
			return nil, err
		}
	default:
		appRepo, err = NewAppRepoPostgres(db)
		if err != nil {
			return nil, err
		}
	}

	return appRepo, nil
}
