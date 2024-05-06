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
	// Базовый адрес результирующего сокращённого URL.
	// Требования:
	//     - Должен быть указан протокол: http/http
	//     - Адрес должен быть равен адресу в поле ServerAddress
	//     - Путь URL Path должен быть (по-умолчанию /)
	// Пример: http:localhost:8080/blablabla
	BaseURL string `env:"BASE_URL"`
}

func (c *Config) parseFlags() error {
	flag.StringVar(&c.ServerAddress, "a", "", "Server address")
	flag.StringVar(&c.BaseURL, "b", "", "Base URL")
	flag.Parse()

	err := env.Parse(c)
	if err != nil {
		return err
	}

	// Если не ввели -a и -b, то значения по-умолчанию
	if c.ServerAddress == "" && c.BaseURL == "" {
		c.ServerAddress = Addr
		c.BaseURL = ResultAddrPrefix
	}

	switch {
	case c.ServerAddress != "" && c.BaseURL != "": // ввели -a и -b ИЛИ значения по-умолчанию
		u, err := url.ParseRequestURI(c.BaseURL)
		if err != nil {
			return err
		}
		if u.Host != c.ServerAddress {
			return ErrInvalidBaseURL
		}
	case c.ServerAddress != "": // ввели только -a
		c.BaseURL = c.ServerAddress
	case c.BaseURL != "": // ввели только -b
		u, err := url.ParseRequestURI(c.BaseURL)
		if err != nil {
			return err
		}
		c.ServerAddress = u.Host
	}

	if !strings.HasPrefix(c.BaseURL, "http://") && !strings.HasPrefix(c.BaseURL, "https://") {
		c.BaseURL = "http://" + c.BaseURL
	}
	if !strings.HasSuffix(c.BaseURL, "/") {
		c.BaseURL += "/"
	}

	return nil
}
