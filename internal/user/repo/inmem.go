package repo

import (
	"github.com/MisterMaks/go-yandex-shortener/internal/user"
	"sync"
)

type UserRepoInmem struct {
	users    []*user.User
	mu       sync.RWMutex
	producer *producer
	maxID    uint
}

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
	defer consumer.close()
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

func (uri *UserRepoInmem) Close() error {
	return uri.producer.close()
}

func (uri *UserRepoInmem) CreateUser() (*user.User, error) {
	uri.mu.Lock()
	defer uri.mu.Unlock()

	uri.maxID++
	u := &user.User{ID: uri.maxID}
	uri.users = append(uri.users, u)

	if uri.producer != nil {
		uri.producer.writeUser(u)
	}

	return u, nil
}
