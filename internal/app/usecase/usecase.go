package usecase

import (
	"database/sql"
	"errors"
	"math/rand"
	"net/url"
	"regexp"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
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
	GetOrCreateURLs(urls []*app.URL) ([]*app.URL, error)
	GetUserURLs(userID uint) ([]*app.URL, error)
}

type AppUsecase struct {
	AppRepo AppRepoInterface

	BaseURL                       string
	CountRegenerationsForLengthID uint
	LengthID                      uint
	MaxLengthID                   uint

	db *sql.DB
}

func NewAppUsecase(appRepo AppRepoInterface, baseURL string, countRegenerationsForLengthID, lengthID, maxLengthID uint, db *sql.DB) (*AppUsecase, error) {
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
		db:                            db,
	}, nil
}

func (au *AppUsecase) generateID() (string, error) {
	if au.LengthID > au.MaxLengthID {
		return "", ErrMaxLengthIDLessLengthID
	}

	var err error
	var checked bool
	var id string
	for i := 0; i < int(au.CountRegenerationsForLengthID); i++ {
		id, err = generateID(au.LengthID)
		if err != nil {
			return "", err
		}
		checked, err = au.AppRepo.CheckIDExistence(id)
		if err != nil {
			return "", err
		}
		if checked {
			continue
		}
		break
	}

	if checked {
		au.LengthID++
		return au.generateID()
	}

	return id, nil
}

func (au *AppUsecase) GetOrCreateURL(rawURL string) (*app.URL, bool, error) {
	_, err := parseURL(rawURL)
	if err != nil {
		return nil, false, err
	}
	id, err := au.generateID()
	if err != nil {
		return nil, false, err
	}
	appURL, err := au.AppRepo.GetOrCreateURL(id, rawURL)
	return appURL, appURL.ID != id, err
}

func (au *AppUsecase) GetURL(id string) (*app.URL, error) {
	return au.AppRepo.GetURL(id)
}

func (au *AppUsecase) GenerateShortURL(id string) string {
	return au.BaseURL + id
}

func (au *AppUsecase) Ping() error {
	return au.db.Ping()
}

func (au *AppUsecase) GetOrCreateURLs(requestBatchURLs []app.RequestBatchURL) ([]app.ResponseBatchURL, error) {
	urls := []*app.URL{}
	for _, rbu := range requestBatchURLs {
		id, err := au.generateID()
		if err != nil {
			return nil, err
		}
		urls = append(urls, &app.URL{ID: id, URL: rbu.OriginalURL})
	}

	urls, err := au.AppRepo.GetOrCreateURLs(urls)
	if err != nil {
		return nil, err
	}

	responseBatchURLs := []app.ResponseBatchURL{}
	for _, appURL := range urls {
		for _, rbu := range requestBatchURLs {
			if appURL.URL == rbu.OriginalURL {
				responseBatchURLs = append(responseBatchURLs, app.ResponseBatchURL{
					CorrelationID: rbu.CorrelationID,
					ShortURL:      au.GenerateShortURL(appURL.ID),
				})
			}
		}
	}

	return responseBatchURLs, nil
}

func (au *AppUsecase) GetUserURLs(userID uint) ([]app.ResponseUserURL, error) {
	urls, err := au.AppRepo.GetUserURLs(userID)
	if err != nil {
		return nil, err
	}

	responseUserURLs := []app.ResponseUserURL{}
	for _, appURL := range urls {
		responseUserURLs = append(responseUserURLs, app.ResponseUserURL{
			ShortURL:    au.GenerateShortURL(appURL.ID),
			OriginalURL: appURL.URL,
		})
	}

	return responseUserURLs, nil
}
