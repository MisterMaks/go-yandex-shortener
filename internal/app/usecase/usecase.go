package usecase

import (
	"errors"
	"math/rand"

	app "github.com/MisterMaks/go-yandex-shortener/internal/app"
)

const (
	Symbols      string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	CountSymbols int    = len(Symbols)
)

var (
	ErrZeroLengthID            = errors.New("length ID == 0")
	ErrZeroMaxLengthID         = errors.New("max length ID == 0")
	ErrMaxLengthIDLessLengthID = errors.New("max length ID is less length ID")
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

func generateShortURL(addr, id string) string {
	return "http://" + addr + "/" + id
}

type AppRepoInterface interface {
	GetOrCreateURL(id, rawURL string) (*app.URL, error)
	GetURL(id string) (*app.URL, error)
	CheckIDExistence(id string) (bool, error)
}

type AppUsecase struct {
	AppRepo AppRepoInterface

	CountRegenerationsForLengthID uint
	LengthID                      uint
	MaxLengthID                   uint
}

func NewAppUsecase(appRepo AppRepoInterface, countRegenerationsForLengthID, lengthID, maxLengthID uint) (*AppUsecase, error) {
	if lengthID == 0 {
		return nil, ErrZeroLengthID
	}
	if maxLengthID == 0 {
		return nil, ErrZeroMaxLengthID
	}
	if maxLengthID < lengthID {
		return nil, ErrMaxLengthIDLessLengthID
	}
	return &AppUsecase{
		AppRepo:                       appRepo,
		CountRegenerationsForLengthID: countRegenerationsForLengthID,
		LengthID:                      lengthID,
		MaxLengthID:                   maxLengthID,
	}, nil
}

func (au *AppUsecase) GetOrCreateURL(rawURL string) (*app.URL, error) {
	if au.LengthID > au.MaxLengthID {
		return nil, ErrMaxLengthIDLessLengthID
	}

	var err error
	var checked bool
	var id string
	for i := 0; i < int(au.CountRegenerationsForLengthID); i++ {
		id, err = generateID(au.LengthID)
		if err != nil {
			return nil, err
		}
		checked, err = au.AppRepo.CheckIDExistence(id)
		if err != nil || checked {
			continue
		}
		break
	}

	if err != nil || checked {
		au.LengthID++
		au.GetOrCreateURL(rawURL)
	}

	url, err := au.AppRepo.GetOrCreateURL(id, rawURL)
	return url, err
}

func (au *AppUsecase) GetURL(id string) (*app.URL, error) {
	return au.AppRepo.GetURL(id)
}

func (au *AppUsecase) GenerateShortURL(addr, id string) string {
	return generateShortURL(addr, id)
}
