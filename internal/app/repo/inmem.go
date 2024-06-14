package repo

import (
	"errors"
	"sync"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
)

var ErrURLNotFound = errors.New("url not found")

type AppRepoInmem struct {
	urls     []*app.URL
	mu       sync.RWMutex
	producer *producer
}

func NewAppRepoInmem(filename string) (*AppRepoInmem, error) {
	if filename == "" {
		return &AppRepoInmem{
			urls:     []*app.URL{},
			mu:       sync.RWMutex{},
			producer: nil,
		}, nil
	}

	consumer, err := newConsumer(filename)
	if err != nil {
		return nil, err
	}
	defer consumer.close()
	urls, err := consumer.readURLs()
	if err != nil {
		return nil, err
	}
	producer, err := newProducer(filename)
	if err != nil {
		return nil, err
	}
	return &AppRepoInmem{
		urls:     urls,
		mu:       sync.RWMutex{},
		producer: producer,
	}, nil
}

func (ari *AppRepoInmem) GetOrCreateURL(id, rawURL string) (*app.URL, error) {
	ari.mu.Lock()
	defer ari.mu.Unlock()
	for _, url := range ari.urls {
		if rawURL == url.URL {
			return url, nil
		}
	}
	url := &app.URL{ID: id, URL: rawURL}
	ari.urls = append(ari.urls, url)

	if ari.producer != nil {
		ari.producer.writeURL(url)
	}

	return url, nil
}

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

func (ari *AppRepoInmem) Close() error {
	return ari.producer.close()
}
