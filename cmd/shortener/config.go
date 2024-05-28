package main

import (
	"errors"
	"flag"
	"net/url"
	"strings"

	"github.com/caarlos0/env/v11"
)

var (
	ErrInvalidBaseURL = errors.New("invalid Base URL")
)

type Config struct {
	// Адрес запуска HTTP-сервера. Пример: localhost:8080
	ServerAddress string `env:"SERVER_ADDRESS"`
	// Базовый адрес результирующего сокращённого URL
	// Требования:
	//     - Должен быть указан протокол (по умолчанию автоматически добавится http://): http/https
	//     - Путь URL Path должен быть (по-умолчанию автоматически добавится /)
	// Пример: http://localhost:8080/blablabla
	BaseURL         string `env:"BASE_URL"`
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func (c *Config) parseFlags() error {
	flag.StringVar(&c.ServerAddress, "a", "", "Server address")
	flag.StringVar(&c.BaseURL, "b", "", "Base URL")
	flag.StringVar(&c.LogLevel, "l", "", "Log level")
	flag.StringVar(&c.FileStoragePath, "f", "", "File storage path")
	flag.Parse()

	err := env.Parse(c)
	if err != nil {
		return err
	}

	// Если не ввели -a, -b, -l, -f то значения по-умолчанию
	if c.ServerAddress == "" {
		c.ServerAddress = Addr
	}
	if c.BaseURL == "" {
		c.BaseURL = ResultAddrPrefix
	}
	if c.LogLevel == "" {
		c.LogLevel = LogLevel
	}
	if c.FileStoragePath == "" {
		c.FileStoragePath = FileStoragePath
	}

	_, err = url.ParseRequestURI(c.BaseURL)
	if err != nil {
		return err
	}

	if !strings.HasPrefix(c.BaseURL, "http://") && !strings.HasPrefix(c.BaseURL, "https://") {
		c.BaseURL = "http://" + c.BaseURL
	}
	if !strings.HasSuffix(c.BaseURL, "/") {
		c.BaseURL += "/"
	}

	return nil
}
