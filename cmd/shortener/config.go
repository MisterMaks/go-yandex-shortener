package main

import (
	"flag"
	"net/url"
	"os"
	"strings"

	"github.com/caarlos0/env/v11"
)

// Config config data for app.
type Config struct {
	// Адрес запуска HTTP-сервера. Пример: localhost:8080
	ServerAddress string `env:"SERVER_ADDRESS"` // address to start the server
	// Базовый адрес результирующего сокращённого URL
	// Требования:
	//     - Должен быть указан протокол (по умолчанию автоматически добавится http://): http/https
	//     - Путь URL Path должен быть (по-умолчанию автоматически добавится /)
	// Пример: http://localhost:8080/blablabla
	BaseURL         string `env:"BASE_URL"` // short URLs will be returned with this host
	LogLevel        string `env:"LOG_LEVEL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func (c *Config) parseFlags() error {
	flag.StringVar(&c.ServerAddress, "a", "", "Server address")
	flag.StringVar(&c.BaseURL, "b", "", "Base URL")
	flag.StringVar(&c.LogLevel, "l", "", "Log level")
	flag.StringVar(&c.FileStoragePath, "f", "", "File storage path")
	flag.StringVar(&c.DatabaseDSN, "d", "", "Database DSN")
	flag.Parse()

	foundFlagFileStoragePath := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "f" {
			foundFlagFileStoragePath = true
		}
	})

	_, foundEnvFileStoragePath := os.LookupEnv("FILE_STORAGE_PATH")

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
	if !foundFlagFileStoragePath && !foundEnvFileStoragePath {
		c.FileStoragePath = URLsFileStoragePath
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
