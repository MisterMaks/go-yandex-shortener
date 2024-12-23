package repo

import (
	"database/sql"
	"fmt"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
)

// AppRepoPostgres application data storage in PostgreSQL.
type AppRepoPostgres struct {
	db *sql.DB
}

// NewAppRepoPostgres creates *AppRepoPostgres.
func NewAppRepoPostgres(db *sql.DB) (*AppRepoPostgres, error) {
	return &AppRepoPostgres{db: db}, nil
}

// GetOrCreateURL insert new URL in DB or get existed URL.
func (arp *AppRepoPostgres) GetOrCreateURL(id, rawURL string, userID uint) (*app.URL, error) {
	query := `INSERT INTO url (url, url_id, user_id) 
VALUES ($1, $2, $3) 
ON CONFLICT (url) DO UPDATE SET url = EXCLUDED.url, user_id = COALESCE(url.user_id, EXCLUDED.user_id) 
RETURNING url_id, user_id;`
	err := arp.db.QueryRow(query, rawURL, id, userID).Scan(&id, &userID)
	if err != nil {
		return nil, err
	}
	url := &app.URL{ID: id, URL: rawURL, UserID: userID}
	return url, nil
}

// GetURL get URL from DB.
func (arp *AppRepoPostgres) GetURL(id string) (*app.URL, error) {
	query := `SELECT url, url_id, user_id, is_deleted FROM url WHERE url_id = $1;`
	url := &app.URL{}
	err := arp.db.QueryRow(query, id).Scan(&url.URL, &url.ID, &url.UserID, &url.IsDeleted)
	if err != nil {
		return nil, err
	}
	return url, nil
}

// CheckIDExistence check ID existence in DB.
func (arp *AppRepoPostgres) CheckIDExistence(id string) (bool, error) {
	query := `SELECT true FROM url WHERE url_id = $1;`
	var exists bool
	err := arp.db.QueryRow(query, id).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// Ping ping DB.
func (arp *AppRepoPostgres) Ping() error {
	return arp.db.Ping()
}

// GetOrCreateURLs insert batch URLs or get existed URLs from DB.
func (arp *AppRepoPostgres) GetOrCreateURLs(urls []*app.URL) ([]*app.URL, error) {
	query := `INSERT INTO url (url, url_id, user_id) VALUES `
	args := make([]interface{}, 0, len(urls)*3)
	lenURLs := len(urls)
	for i, url := range urls {
		query += fmt.Sprintf("($%d, $%d, $%d)", i*3+1, i*3+2, i*3+3)
		args = append(args, url.URL, url.ID, url.UserID)
		if i < lenURLs-1 {
			query += ", "
		}
	}
	query += ` ON CONFLICT (url) 
DO UPDATE SET url = EXCLUDED.url, user_id = COALESCE(url.user_id, EXCLUDED.user_id) 
RETURNING url, url_id, user_id;`

	rows, err := arp.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls = nil
	for rows.Next() {
		var (
			id     string
			url    string
			userID uint
		)
		err = rows.Scan(&url, &id, &userID)
		if err != nil {
			return nil, err
		}
		urls = append(urls, &app.URL{ID: id, URL: url, UserID: userID})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return urls, err
}

// GetUserURLs get user URLs from DB.
func (arp *AppRepoPostgres) GetUserURLs(userID uint) ([]*app.URL, error) {
	query := `SELECT url, url_id, user_id, is_deleted FROM url WHERE user_id = $1;`

	rows, err := arp.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	urls := []*app.URL{}
	for rows.Next() {
		var (
			id        string
			url       string
			isDeleted bool
		)
		err = rows.Scan(&url, &id, &userID, &isDeleted)
		if err != nil {
			return nil, err
		}
		urls = append(urls, &app.URL{ID: id, URL: url, UserID: userID, IsDeleted: isDeleted})
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return urls, err
}

// DeleteUserURLs delete user URLs from DB.
func (arp *AppRepoPostgres) DeleteUserURLs(urls []*app.URL) error {
	query := `UPDATE url SET is_deleted = true WHERE `
	args := make([]interface{}, 0, len(urls)*2)
	lenURLs := len(urls)
	for i, url := range urls {
		query += fmt.Sprintf("(url_id = $%d AND user_id = $%d)", i*2+1, i*2+2)
		args = append(args, url.ID, url.UserID)
		if i < lenURLs-1 {
			query += " OR "
		}
	}
	query += ";"

	_, err := arp.db.Exec(query, args...)

	return err
}
