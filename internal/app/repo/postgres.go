package repo

import (
	"database/sql"
)

type AppRepoPostgres struct {
	db *sql.DB
}

func NewAppRepoPostgres(db *sql.DB) (*AppRepoPostgres, error) {
	return &AppRepoPostgres{db: db}, nil
}

// TODO
// func (arp *AppRepoPostgres) GetOrCreateURL(id, rawURL string) (*app.URL, error) {
// 	return nil, nil
// }

// TODO
// func (arp *AppRepoPostgres) GetURL(id string) (*app.URL, error) {
// 	return nil, nil
// }

// TODO
// func (arp *AppRepoPostgres) CheckIDExistence(id string) (bool, error) {
// 	return false, nil
// }

func (arp *AppRepoPostgres) Ping() error {
	return arp.db.Ping()
}
