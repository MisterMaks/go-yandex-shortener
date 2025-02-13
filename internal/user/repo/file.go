package repo

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/MisterMaks/go-yandex-shortener/internal/user"
)

type producer struct {
	file *os.File
	// добавляем Writer в Producer
	writer *bufio.Writer
}

func newProducer(filename string) (*producer, error) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &producer{
		file: file,
		// создаём новый Writer
		writer: bufio.NewWriter(file),
	}, nil
}

func (p *producer) close() error {
	// закрываем файл
	return p.file.Close()
}

func (p *producer) writeUser(u *user.User) error {
	data, err := json.Marshal(&u)
	if err != nil {
		return err
	}

	// записываем событие в буфер
	if _, err := p.writer.Write(data); err != nil {
		return err
	}

	// добавляем перенос строки
	if err := p.writer.WriteByte('\n'); err != nil {
		return err
	}

	// записываем буфер в файл
	return p.writer.Flush()
}

type consumer struct {
	file *os.File
	// заменяем Reader на Scanner
	scanner *bufio.Scanner
}

func newConsumer(filename string) (*consumer, error) {
	file, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &consumer{
		file: file,
		// создаём новый scanner
		scanner: bufio.NewScanner(file),
	}, nil
}

func (c *consumer) close() error {
	return c.file.Close()
}

func (c *consumer) readUser() (*user.User, error) {
	// одиночное сканирование до следующей строки
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}
	// читаем данные из scanner
	data := c.scanner.Bytes()

	u := user.User{}
	err := json.Unmarshal(data, &u)
	if err != nil {
		return nil, err
	}

	return &u, nil
}

func (c *consumer) readUsers() ([]*user.User, error) {
	users := []*user.User{}
	for {
		u, err := c.readUser()
		if err != nil {
			return nil, err
		}
		if u == nil {
			return users, nil
		}
		users = append(users, u)
	}
}
