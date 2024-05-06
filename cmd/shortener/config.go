package main

import (
	"errors"
	"flag"
	"net/url"
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
	u, err := url.ParseRequestURI(c.ResultAddrPrefix)
	if err != nil {
		return err
	}
	if u.Host != c.Addr || u.Path == "" {
		return ErrInvalidResultAddrPrefix
	}
	c.ResultPathPrefix = u.Path
	return nil
}
