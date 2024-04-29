package repo

import (
	"fmt"
	"sync"

	app "github.com/MisterMaks/go-yandex-shortener/internal/app"
)

type AppRepo struct {
	URLs []*app.URL
	mu   *sync.RWMutex
}

func NewAppRepo() *AppRepo {
	return &AppRepo{
		URLs: []*app.URL{},
		mu:   &sync.RWMutex{},
	}
}

func (ar *AppRepo) Create(id, rawURL string) (*app.URL, error) {
	ar.mu.RLock()
	for _, url := range ar.URLs {
		if rawURL == url.URL {
			ar.mu.RUnlock()
			return url, nil
		}
	}
	ar.mu.RUnlock()
	ar.mu.Lock()
	defer ar.mu.Unlock()
	url := app.NewURL(id, rawURL)
	ar.URLs = append(ar.URLs, url)
	return url, nil
}

func (ar *AppRepo) Get(id string) (*app.URL, error) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	for _, url := range ar.URLs {
		if id == url.ID {
			return url, nil
		}
	}
	return nil, fmt.Errorf("url not found")
}

func (ar *AppRepo) IsExistID(id string) (bool, error) {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	for _, url := range ar.URLs {
		if id == url.ID {
			return true, nil
		}
	}
	return false, nil
}
