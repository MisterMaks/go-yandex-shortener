package repo

import (
	"errors"
	"sync"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
)

// ErrURLNotFound is error for not found URL.
var ErrURLNotFound = errors.New("url not found")

// DefaultCountURLs constant for in-mem repo.
const DefaultCountURLs = 256

// AppRepoInmem in-memory application data storage.
type AppRepoInmem struct {
	urls              []*app.URL
	mu                sync.RWMutex
	producer          *producer
	deleteURLProducer *producer
}

// NewAppRepoInmem creates *AppRepoInmem and loads saved data from file.
func NewAppRepoInmem(filename string, deletedURLsFilename string) (*AppRepoInmem, error) {
	if filename == "" {
		return &AppRepoInmem{
			urls:     make([]*app.URL, 0, DefaultCountURLs),
			mu:       sync.RWMutex{},
			producer: nil,
		}, nil
	}

	c, err := newConsumer(filename)
	if err != nil {
		return nil, err
	}
	urls, err := c.readURLs()
	if err != nil {
		return nil, err
	}
	if err = c.close(); err != nil {
		return nil, err
	}

	c, err = newConsumer(deletedURLsFilename)
	if err != nil {
		return nil, err
	}
	deletedURLs, err := c.readURLs()
	if err != nil {
		return nil, err
	}
	if err = c.close(); err != nil {
		return nil, err
	}

	for _, deletedURL := range deletedURLs {
		for _, url := range urls {
			if url.ID == deletedURL.URL {
				url.IsDeleted = true
				continue
			}
		}
	}

	p, err := newProducer(filename)
	if err != nil {
		return nil, err
	}

	deleteURLProducer, err := newProducer(deletedURLsFilename)
	if err != nil {
		return nil, err
	}

	return &AppRepoInmem{
		urls:              urls,
		mu:                sync.RWMutex{},
		producer:          p,
		deleteURLProducer: deleteURLProducer,
	}, nil
}

// GetOrCreateURL get saved URL or creates new URL and save it in file.
func (ari *AppRepoInmem) GetOrCreateURL(id, rawURL string, userID uint) (*app.URL, error) {
	ari.mu.Lock()
	defer ari.mu.Unlock()
	for _, url := range ari.urls {
		if rawURL == url.URL {
			return url, nil
		}
	}
	url := &app.URL{ID: id, URL: rawURL, UserID: userID}
	ari.urls = append(ari.urls, url)

	if ari.producer != nil {
		if err := ari.producer.writeURL(url); err != nil {
			return nil, err
		}
	}

	return url, nil
}

// GetURL get URL with ID.
func (ari *AppRepoInmem) GetURL(id string) (*app.URL, error) {
	ari.mu.RLock()
	defer ari.mu.RUnlock()
	for _, url := range ari.urls {
		if id == url.ID {
			return url, nil
		}
	}
	return nil, ErrURLNotFound
}

// CheckIDExistence check URL ID existence.
func (ari *AppRepoInmem) CheckIDExistence(id string) (bool, error) {
	ari.mu.RLock()
	defer ari.mu.RUnlock()
	for _, url := range ari.urls {
		if id == url.ID {
			return true, nil
		}
	}
	return false, nil
}

// Close finishes working with the file.
func (ari *AppRepoInmem) Close() error {
	if ari.producer != nil {
		return ari.producer.close()
	}

	if ari.deleteURLProducer != nil {
		return ari.deleteURLProducer.close()
	}

	return nil
}

// GetOrCreateURLs gets created URLs and saves new URLs and returns them.
func (ari *AppRepoInmem) GetOrCreateURLs(urls []*app.URL) ([]*app.URL, error) {
	ari.mu.Lock()
	defer ari.mu.Unlock()

	for _, url := range urls {
		for _, ariURL := range ari.urls {
			if url.URL == ariURL.URL {
				url.ID = ariURL.ID
				url.UserID = ariURL.UserID
				continue
			}
		}

		url = &app.URL{ID: url.ID, URL: url.URL, UserID: url.UserID}
		ari.urls = append(ari.urls, url)

		if ari.producer != nil {
			if err := ari.producer.writeURL(url); err != nil {
				return nil, err
			}
		}
	}

	return urls, nil
}

// GetUserURLs gets user URLs.
func (ari *AppRepoInmem) GetUserURLs(userID uint) ([]*app.URL, error) {
	ari.mu.RLock()
	defer ari.mu.RUnlock()

	userURLs := []*app.URL{}
	for _, url := range ari.urls {
		if url.UserID == userID {
			userURLs = append(userURLs, url)
		}
	}

	return userURLs, nil
}

// DeleteUserURLs delete user URLs.
func (ari *AppRepoInmem) DeleteUserURLs(urls []*app.URL) error {
	ari.mu.Lock()
	defer ari.mu.Unlock()

	for _, url := range urls {
		for _, ariURL := range ari.urls {
			if url.ID == ariURL.ID && url.UserID == ariURL.UserID {
				ariURL.IsDeleted = true

				if ari.deleteURLProducer != nil {
					if err := ari.deleteURLProducer.writeURL(url); err != nil {
						return err
					}
				}

				continue
			}
		}
	}

	return nil
}
