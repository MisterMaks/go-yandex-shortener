package repo

import (
	"fmt"
	"sync"

	app "github.com/MisterMaks/go-yandex-shortener/internal/app"
)

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

func (ari *AppRepoInmem) Create(id, rawURL string) (*app.URL, error) {
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
	url := app.NewURL(id, rawURL)
	ari.urls = append(ari.urls, url)
	return url, nil
}

func (ari *AppRepoInmem) Get(id string) (*app.URL, error) {
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
	return nil, fmt.Errorf("url not found")
}

func (ari *AppRepoInmem) IsExistID(id string) (bool, error) {
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
