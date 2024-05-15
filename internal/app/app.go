package app

import (
	"errors"
	"net/url"
	"regexp"
)

var (
	ErrInvalidURL = errors.New("invalid url")
)

type URL struct {
	ID  string
	URL string
}

func NewURL(id, rawURL string) (*URL, error) {
	matched, err := regexp.MatchString("^https?://", rawURL)
	if err != nil {
		return nil, err
	}
	parsedRequestURI := rawURL
	if !matched {
		parsedRequestURI = "http://" + rawURL
	}
	_, err = url.ParseRequestURI(parsedRequestURI)
	if err != nil {
		return nil, err
	}
	return &URL{ID: id, URL: rawURL}, nil
}
