package repo

import (
	"sync"

	"github.com/MisterMaks/go-yandex-shortener/internal/user"
)

// UserRepoInmem in-memory data storage for users.
type UserRepoInmem struct {
	users    []*user.User
	mu       sync.RWMutex
	producer *producer
	maxID    uint
}

// NewUserRepoInmem creates *NewUserRepoInmem and loads saved data from file.
func NewUserRepoInmem(filename string) (*UserRepoInmem, error) {
	if filename == "" {
		return &UserRepoInmem{
			users:    []*user.User{},
			mu:       sync.RWMutex{},
			producer: nil,
			maxID:    0,
		}, nil
	}

	consumer, err := newConsumer(filename)
	if err != nil {
		return nil, err
	}
	if err = consumer.close(); err != nil {
		return nil, err
	}
	users, err := consumer.readUsers()
	if err != nil {
		return nil, err
	}
	producer, err := newProducer(filename)
	if err != nil {
		return nil, err
	}

	var maxID uint
	for _, u := range users {
		if u.ID > maxID {
			maxID = u.ID
		}
	}

	return &UserRepoInmem{
		users:    users,
		mu:       sync.RWMutex{},
		producer: producer,
		maxID:    maxID,
	}, nil
}

// Close finishes working with the file.
func (uri *UserRepoInmem) Close() error {
	if uri.producer != nil {
		return uri.producer.close()
	}

	return nil
}

// CreateUser creates new user.
func (uri *UserRepoInmem) CreateUser() (*user.User, error) {
	uri.mu.Lock()
	defer uri.mu.Unlock()

	uri.maxID++
	u := &user.User{ID: uri.maxID}
	uri.users = append(uri.users, u)

	if uri.producer != nil {
		err := uri.producer.writeUser(u)
		if err != nil {
			return nil, err
		}
	}

	return u, nil
}
