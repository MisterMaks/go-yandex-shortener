package main

import (
	"errors"
	"flag"
	"net/url"
	"strings"
)

var (
	ErrInvalidResultAddrPrefix = errors.New("invalid prefix of the resulting address")
)

type Config struct {
	// Адрес запуска HTTP-сервера. Пример: localhost:8080
	Addr string
	// Базовый адрес результирующего сокращённого URL.
	// Требования:
	//     - Должен быть указан протокол: http/http
	//     - Адрес должен быть равен адресу в поле Addr
	//     - Путь URL Path должен быть (по-умолчанию /)
	// Пример: http:localhost:8080/blablabla
	ResultAddrPrefix string
	// Базовый путь URL Path (проставляется автоматически из ResultAddrPrefix)
	// Требования:
	//     - Должен совпадать с путем URL Path в ResultAddrPrefix
	//     - Начинается с /
	// Пример: /blablabla
	ResultPathPrefix string
}

func (c *Config) parseFlags() error {
	flag.StringVar(&c.Addr, "a", "", "Server address")
	flag.StringVar(&c.ResultAddrPrefix, "b", "", "Prefix of the resulting address")
	flag.Parse()

	switch {
	case c.Addr != "": // ввели -a
		if c.ResultAddrPrefix != "" { // ввели -a и ввели -b
			u, err := url.ParseRequestURI(c.ResultAddrPrefix)
			if err != nil {
				return err
			}
			if u.Host != c.Addr {
				return ErrInvalidResultAddrPrefix
			}
			if u.Path != "" {
				c.ResultPathPrefix = u.Path
			}
		} else { // ввели -a и не ввели -b
			c.ResultAddrPrefix = c.Addr
			if !strings.HasPrefix(c.ResultAddrPrefix, "http://") || !strings.HasPrefix(c.ResultAddrPrefix, "https://") {
				c.ResultAddrPrefix = "http://" + c.ResultAddrPrefix
			}
			if !strings.HasSuffix(c.ResultAddrPrefix, "/") {
				c.ResultAddrPrefix += "/"
			}
		}
	case c.ResultAddrPrefix != "": // ввели -b
		u, err := url.ParseRequestURI(c.ResultAddrPrefix)
		if err != nil {
			return err
		}
		if u.Path != "" {
			c.ResultPathPrefix = u.Path
		}
		if c.Addr != "" { // ввели -a и ввели -b (возможно ненужная часть кода, т.к. это проверяется выше)
			if u.Host != c.Addr {
				return ErrInvalidResultAddrPrefix
			}
		} else { // не ввели -a и ввели -b
			c.Addr = u.Host
		}
	default: // значения по-умолчанию - не ввели -a и не ввели -b
		c.Addr = Addr
		c.ResultAddrPrefix = ResultAddrPrefix
		u, err := url.ParseRequestURI(c.ResultAddrPrefix)
		if err != nil {
			return err
		}
		if u.Host != c.Addr {
			return ErrInvalidResultAddrPrefix
		}
		if u.Path != "" {
			c.ResultPathPrefix = u.Path
		}
	}

	return nil
}
