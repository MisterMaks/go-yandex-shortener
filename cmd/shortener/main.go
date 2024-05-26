package main

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	appDeliveryInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/delivery"
	appRepoInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/repo"
	appUsecaseInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/usecase"
	"github.com/MisterMaks/go-yandex-shortener/internal/logger"
)

const (
	Addr                          string = "localhost:8080"
	ResultAddrPrefix              string = "http://localhost:8080/"
	CountRegenerationsForLengthID uint   = 5
	LengthID                      uint   = 5
	MaxLengthID                   uint   = 20
	LogLevel                      string = "INFO"

	ConfigKey string = "config"
	AddrKey   string = "addr"
)

type AppHandlerInterface interface {
	GetOrCreateURL(w http.ResponseWriter, r *http.Request)
	APIGetOrCreateURL(w http.ResponseWriter, r *http.Request)
	RedirectToURL(w http.ResponseWriter, r *http.Request)
}

func shortenerRouter(appHandler AppHandlerInterface, redirectPathPrefix string) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	redirectPathPrefix = strings.TrimPrefix(redirectPathPrefix, "/")
	r.Route(`/`, func(r chi.Router) {
		r.Post(`/`, appHandler.GetOrCreateURL)
		r.Post(`/api/shorten`, appHandler.APIGetOrCreateURL)
		r.Get(`/`+redirectPathPrefix+`{id}`, appHandler.RedirectToURL)
	})
	return r
}

func main() {
	config := &Config{}
	err := config.parseFlags()
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to parse flags. Error:", err)
	}

	err = logger.Initialize(config.LogLevel)
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to init logger. Error:", err)
	}

	logger.Log.Info("Config data",
		zap.Any(ConfigKey, config),
	)

	appRepo := appRepoInternal.NewAppRepoInmem()
	appUsecase, err := appUsecaseInternal.NewAppUsecase(
		appRepo,
		config.BaseURL,
		CountRegenerationsForLengthID,
		LengthID,
		MaxLengthID,
	)
	if err != nil {
		logger.Log.Fatal("Failed to create appUsecase",
			zap.Error(err),
		)
	}

	appHandler := appDeliveryInternal.NewAppHandler(appUsecase)

	u, err := url.ParseRequestURI(config.BaseURL)
	if err != nil {
		logger.Log.Fatal("Failed to parse config result addr prefix",
			zap.Error(err),
		)
	}
	redirectPathPrefix := u.Path

	r := shortenerRouter(appHandler, redirectPathPrefix)

	logger.Log.Info("Server running",
		zap.String(AddrKey, config.ServerAddress),
	)
	err = http.ListenAndServe(config.ServerAddress, r)
	if err != nil {
		logger.Log.Fatal("Failed to start server",
			zap.Error(err),
		)
	}
}
