package repo

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
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

func (p *producer) writeURL(url *app.URL) error {
	data, err := json.Marshal(&url)
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

func (c *consumer) readURL() (*app.URL, error) {
	// одиночное сканирование до следующей строки
	if !c.scanner.Scan() {
		return nil, c.scanner.Err()
	}
	// читаем данные из scanner
	data := c.scanner.Bytes()

	url := app.URL{}
	err := json.Unmarshal(data, &url)
	if err != nil {
		return nil, err
	}

	return &url, nil
}

func (c *consumer) readURLs() ([]*app.URL, error) {
	urls := []*app.URL{}
	for {
		url, err := c.readURL()
		if err != nil {
			return nil, err
		}
		if url == nil {
			return urls, nil
		}
		urls = append(urls, url)
	}
}
