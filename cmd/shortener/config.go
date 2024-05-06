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
	Addr             string
	ResultAddrPrefix string
	ResultPathPrefix string
}

func (c *Config) parseFlags() error {
	flag.StringVar(&c.Addr, "a", Addr, "Server address")
	flag.StringVar(&c.ResultAddrPrefix, "b", ResultAddrPrefix, "Prefix of the resulting address")
	flag.Parse()

	switch {
	case c.Addr != Addr: // ввели -a
		if c.ResultAddrPrefix != ResultAddrPrefix { // ввели -a и ввели -b
			u, err := url.ParseRequestURI(c.ResultAddrPrefix)
			if err != nil {
				return err
			}
			if u.Host != c.Addr || u.Path == "" {
				return ErrInvalidResultAddrPrefix
			}
			c.ResultPathPrefix = u.Path
		} else { // ввели -a и не ввели -b
			c.ResultAddrPrefix = c.Addr
			if !strings.HasPrefix(c.ResultAddrPrefix, "http://") || !strings.HasPrefix(c.ResultAddrPrefix, "https://") {
				c.ResultAddrPrefix = "http://" + c.ResultAddrPrefix
			}
			if !strings.HasSuffix(c.ResultAddrPrefix, "/") {
				c.ResultAddrPrefix += "/"
				c.ResultPathPrefix = "/"
			}
		}
	case c.ResultAddrPrefix != ResultAddrPrefix: // ввели -b
		u, err := url.ParseRequestURI(c.ResultAddrPrefix)
		if err != nil {
			return err
		}
		c.ResultPathPrefix = u.Path
		if c.Addr != Addr { // ввели -a и ввели -b (возможно ненужная часть кода, т.к. это проверяется выше)
			if u.Host != c.Addr || u.Path == "" {
				return ErrInvalidResultAddrPrefix
			}
		} else { // не ввели -a и ввели -b
			c.Addr = u.Host
		}
	default: // значения по-умолчанию - не ввели -a и не ввели -b
		u, err := url.ParseRequestURI(c.ResultAddrPrefix)
		if err != nil {
			return err
		}
		if u.Host != c.Addr || u.Path == "" {
			return ErrInvalidResultAddrPrefix
		}
		c.ResultPathPrefix = u.Path
	}

	return nil
}
