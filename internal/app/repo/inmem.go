package repo

import (
	"errors"
	"sync"

	app "github.com/MisterMaks/go-yandex-shortener/internal/app"
)

var ErrURLNotFound = errors.New("url not found")

type AppRepoInmem struct {
	urls []*app.URL
	mu   *sync.RWMutex
}

func NewAppRepoInmem() *AppRepoInmem {
	return &AppRepoInmem{
		urls: []*app.URL{},
		mu:   &sync.RWMutex{},
	}
}

func (ari *AppRepoInmem) GetOrCreateURL(id, rawURL string) (*app.URL, error) {
	if ari.mu == nil {
		ari.mu = &sync.RWMutex{}
	}

	ari.mu.RLock()
	for _, url := range ari.urls {
		if rawURL == url.URL {
			ari.mu.RUnlock()
			return url, nil
		}
	}
	ari.mu.RUnlock()
	ari.mu.Lock()
	defer ari.mu.Unlock()
	url, err := app.NewURL(id, rawURL)
	if err != nil {
		return nil, err
	}
	ari.urls = append(ari.urls, url)
	return url, nil
}

func (ari *AppRepoInmem) GetURL(id string) (*app.URL, error) {
	if ari.mu == nil {
		ari.mu = &sync.RWMutex{}
	}

	ari.mu.RLock()
	defer ari.mu.RUnlock()
	for _, url := range ari.urls {
		if id == url.ID {
			return url, nil
		}
	}
	return nil, ErrURLNotFound
}

func (ari *AppRepoInmem) CheckIDExistence(id string) (bool, error) {
	if ari.mu == nil {
		ari.mu = &sync.RWMutex{}
	}

	ari.mu.RLock()
	defer ari.mu.RUnlock()
	for _, url := range ari.urls {
		if id == url.ID {
			return true, nil
		}
	}
	return false, nil
}
