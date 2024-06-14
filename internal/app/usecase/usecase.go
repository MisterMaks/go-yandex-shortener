package usecase

import (
	"errors"
	"math/rand"
	"net/url"
	"regexp"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
	appRepo "github.com/MisterMaks/go-yandex-shortener/internal/app/repo"
)

const (
	Symbols      string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	CountSymbols        = len(Symbols)
)

var (
	ErrZeroLengthID            = errors.New("length ID == 0")
	ErrZeroMaxLengthID         = errors.New("max length ID == 0")
	ErrMaxLengthIDLessLengthID = errors.New("max length ID is less length ID")
	ErrInvalidBaseURL          = errors.New("invalid Base URL")
)

func generateID(length uint) (string, error) {
	if length == 0 {
		return "", ErrZeroLengthID
	}
	b := make([]byte, length)
	for i := range b {
		b[i] = Symbols[rand.Intn(CountSymbols)]
	}
	return string(b), nil
}

func parseURL(rawURL string) (string, error) {
	matched, err := regexp.MatchString("^https?://", rawURL)
	if err != nil {
		return "", err
	}
	parsedRequestURI := rawURL
	if !matched {
		parsedRequestURI = "http://" + rawURL
	}
	_, err = url.ParseRequestURI(parsedRequestURI)
	if err != nil {
		return "", err
	}
	return parsedRequestURI, nil
}

type AppRepoInterface interface {
	GetOrCreateURL(id, rawURL string) (*app.URL, error)
	GetURL(id string) (*app.URL, error)
	CheckIDExistence(id string) (bool, error)
}

type AppUsecase struct {
	AppRepo AppRepoInterface

	BaseURL                       string
	CountRegenerationsForLengthID uint
	LengthID                      uint
	MaxLengthID                   uint

	AppRepoPostgres *appRepo.AppRepoPostgres
}

func NewAppUsecase(appRepo AppRepoInterface, baseURL string, countRegenerationsForLengthID, lengthID, maxLengthID uint, appRepoPostgres *appRepo.AppRepoPostgres) (*AppUsecase, error) {
	if lengthID == 0 {
		return nil, ErrZeroLengthID
	}
	if maxLengthID == 0 {
		return nil, ErrZeroMaxLengthID
	}
	if maxLengthID < lengthID {
		return nil, ErrMaxLengthIDLessLengthID
	}
	u, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, err
	}
	if u.Path == "" {
		return nil, ErrInvalidBaseURL
	}
	return &AppUsecase{
		AppRepo:                       appRepo,
		BaseURL:                       baseURL,
		CountRegenerationsForLengthID: countRegenerationsForLengthID,
		LengthID:                      lengthID,
		MaxLengthID:                   maxLengthID,
		AppRepoPostgres:               appRepoPostgres,
	}, nil
}

func (au *AppUsecase) GetOrCreateURL(rawURL string) (*app.URL, error) {
	if au.LengthID > au.MaxLengthID {
		return nil, ErrMaxLengthIDLessLengthID
	}

	_, err := parseURL(rawURL)
	if err != nil {
		return nil, err
	}

	var checked bool
	var id string
	for i := 0; i < int(au.CountRegenerationsForLengthID); i++ {
		id, err = generateID(au.LengthID)
		if err != nil {
			return nil, err
		}
		checked, err = au.AppRepo.CheckIDExistence(id)
		if err != nil {
			return nil, err
		}
		if checked {
			continue
		}
		break
	}

	if checked {
		au.LengthID++
		return au.GetOrCreateURL(rawURL)
	}

	appURL, err := au.AppRepo.GetOrCreateURL(id, rawURL)
	return appURL, err
}

func (au *AppUsecase) GetURL(id string) (*app.URL, error) {
	return au.AppRepo.GetURL(id)
}

func (au *AppUsecase) GenerateShortURL(id string) string {
	return au.BaseURL + id
}

func (au *AppUsecase) Ping() error {
	return au.AppRepoPostgres.Ping()
}
