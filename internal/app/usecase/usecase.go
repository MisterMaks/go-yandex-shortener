package usecase

import (
	"fmt"
	"math/rand"
	"net/url"
	"regexp"
	"strings"

	app "github.com/MisterMaks/go-yandex-shortener/internal/app"
)

const (
	Symbols      string = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	CountSymbols int    = len(Symbols)
)

func isURL(rawURL string) (bool, error) {
	matched, err := regexp.MatchString(`^.+\..+$`, rawURL) // Проверка наличия домена первого уровня (.com, .ru, .org, ...)
	if err != nil {
		return false, err
	}
	if !matched {
		return false, nil
	}
	_, err = url.Parse(rawURL)
	return err == nil, nil
}

func generateID(length uint) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = Symbols[rand.Intn(CountSymbols)]
	}
	return string(b)
}

type AppRepoInterface interface {
	Create(id, rawURL string) (*app.URL, error)
	Get(id string) (*app.URL, error)
	IsExistID(id string) (bool, error)
}

type AppUsecase struct {
	AppRepo AppRepoInterface

	CountRegenerationsForLengthID uint
	LengthID                      uint
	MaxLengthID                   uint
}

func NewAppUsecase(appRepo AppRepoInterface, countRegenerationsForLengthID, lengthID, maxLengthID uint) (*AppUsecase, error) {
	if lengthID == 0 {
		return nil, fmt.Errorf("lengthID == 0")
	}
	if maxLengthID == 0 {
		return nil, fmt.Errorf("maxLengthID == 0")
	}
	return &AppUsecase{
		AppRepo:                       appRepo,
		CountRegenerationsForLengthID: countRegenerationsForLengthID,
		LengthID:                      lengthID,
		MaxLengthID:                   maxLengthID,
	}, nil
}

func (au *AppUsecase) Create(s string) (*app.URL, error) {
	if au.LengthID > au.MaxLengthID {
		return nil, fmt.Errorf("length ID > max length ID")
	}

	checked, err := isURL(s)
	if err != nil {
		return nil, err
	}
	if !checked {
		return nil, fmt.Errorf("not URL")
	}

	id := ""
	for i := 0; i < int(au.CountRegenerationsForLengthID); i++ {
		id = generateID(au.LengthID)
		checked, err = au.AppRepo.IsExistID(id)
		if err != nil || checked {
			continue
		}
		break
	}

	if err != nil || checked {
		au.LengthID++
		au.Create(s)
	}

	s = strings.TrimPrefix(s, "http://")
	s = strings.TrimPrefix(s, "https://")
	s = "http://" + s

	url, err := au.AppRepo.Create(id, s)
	return url, err
}

func (au *AppUsecase) Get(id string) (*app.URL, error) {
	return au.AppRepo.Get(id)
}

func (au *AppUsecase) GenerateShortURL(addr, id string) string {
	return "http://" + addr + "/" + id
}
