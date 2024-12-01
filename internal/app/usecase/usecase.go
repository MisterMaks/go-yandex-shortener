package usecase

import (
	"database/sql"
	"errors"
	"math/rand"
	"net/url"
	"regexp"
	"time"

	"go.uber.org/zap"

	"github.com/MisterMaks/go-yandex-shortener/internal/app"
	loggerInternal "github.com/MisterMaks/go-yandex-shortener/internal/logger"
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

// AppRepoInterface contains the necessary functions for storage.
type AppRepoInterface interface {
	GetOrCreateURL(id, rawURL string, userID uint) (*app.URL, error) // get created or create short URL for request URL
	GetURL(id string) (*app.URL, error)                              // get original URL for short URL
	CheckIDExistence(id string) (bool, error)                        // check URL ID existence
	GetOrCreateURLs(urls []*app.URL) ([]*app.URL, error)             // get created or create URLs
	GetUserURLs(userID uint) ([]*app.URL, error)                     // get user URLs
	DeleteUserURLs(urls []*app.URL) error                            // delete urls
}

// AppUsecase business logic struct.
type AppUsecase struct {
	AppRepo AppRepoInterface // storage

	BaseURL                       string // base URL
	CountRegenerationsForLengthID uint   // count regenerations for length ID
	LengthID                      uint   // length ID
	MaxLengthID                   uint   // max length ID

	db *sql.DB

	deleteURLsChan   chan *app.URL
	deleteURLsTicker *time.Ticker

	doneCh chan struct{}
}

// NewAppUsecase creates *AppUsecase.
func NewAppUsecase(
	appRepo AppRepoInterface,
	baseURL string,
	countRegenerationsForLengthID, lengthID, maxLengthID uint,
	db *sql.DB,
	deleteURLsChanSize uint,
	deleteURLsWaitingTime time.Duration,
) (*AppUsecase, error) {
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

	doneCh := make(chan struct{})

	appUsecase := &AppUsecase{
		AppRepo:                       appRepo,
		BaseURL:                       baseURL,
		CountRegenerationsForLengthID: countRegenerationsForLengthID,
		LengthID:                      lengthID,
		MaxLengthID:                   maxLengthID,
		db:                            db,
		deleteURLsChan:                make(chan *app.URL, deleteURLsChanSize),
		deleteURLsTicker:              time.NewTicker(deleteURLsWaitingTime),

		doneCh: doneCh,
	}

	go appUsecase.deleteUserURLs()

	return appUsecase, nil
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

// GetOrCreateURL get created or create short URL for request URL.
// Func generate unique short URL for rawURL, save and return it or return short URL (if rawURL existed).
// Func return URL struct, true if rawURL is new or false if rawURL exists and error.
func (au *AppUsecase) GetOrCreateURL(rawURL string, userID uint) (*app.URL, bool, error) {
	_, err := parseURL(rawURL)
	if err != nil {
		return nil, false, err
	}
	id, err := au.generateID()
	if err != nil {
		return nil, false, err
	}
	appURL, err := au.AppRepo.GetOrCreateURL(id, rawURL, userID)
	if err != nil {
		return nil, false, err
	}
	return appURL, appURL.ID != id, err
}

// GetURL get original URL for short URL.
func (au *AppUsecase) GetURL(id string) (*app.URL, error) {
	return au.AppRepo.GetURL(id)
}

// GenerateShortURL generate short URL.
// Func concatenate BaseURL and URL ID.
func (au *AppUsecase) GenerateShortURL(id string) string {
	return au.BaseURL + id
}

// Ping ping database.
func (au *AppUsecase) Ping() error {
	return au.db.Ping()
}

// GetOrCreateURLs get created or create short URLs for request batch URLs.
// Func generate unique short URL for every OriginalURL (or get existed short URL for OriginalURL) in requestBatchURLs,
// save new URLs in repo and return []app.ResponseBatchURL.
func (au *AppUsecase) GetOrCreateURLs(requestBatchURLs []app.RequestBatchURL, userID uint) ([]app.ResponseBatchURL, error) {
	urls := []*app.URL{}
	for _, rbu := range requestBatchURLs {
		id, err := au.generateID()
		if err != nil {
			return nil, err
		}
		urls = append(urls, &app.URL{ID: id, URL: rbu.OriginalURL, UserID: userID})
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

// GetUserURLs get short and original URLs for user.
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

// SendDeleteUserURLsInChan send urls in delete chan.
func (au *AppUsecase) SendDeleteUserURLsInChan(userID uint, urlIDs []string) {
	go func() {
		for _, urlID := range urlIDs {
			select {
			case au.deleteURLsChan <- &app.URL{ID: urlID, UserID: userID}:
			case <-au.doneCh:
				return
			}
		}
	}()
}

func (au *AppUsecase) deleteUserURLs() {
	logger := loggerInternal.Log

	urls := make([]*app.URL, 0, 2*len(au.deleteURLsChan))

	for {
		select {
		case appURL := <-au.deleteURLsChan:
			urls = append(urls, appURL)
		case <-au.deleteURLsTicker.C:
			if len(urls) == 0 {
				continue
			}
			logger.Debug("Deleting user URLs",
				zap.Any("urls", urls),
			)
			err := au.AppRepo.DeleteUserURLs(urls)
			if err != nil {
				logger.Error("Failed to delete user URLs",
					zap.Error(err),
				)
				continue
			}
			urls = urls[:0]
		}
	}
}

// Close closing channels and stop executing requests/tasks.
func (au *AppUsecase) Close() error {
	close(au.deleteURLsChan)
	close(au.doneCh)
	return nil
}
