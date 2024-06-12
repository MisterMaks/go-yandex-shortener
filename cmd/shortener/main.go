package main

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"

	appDeliveryInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/delivery"
	appRepoInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/repo"
	appUsecaseInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/usecase"
	"github.com/MisterMaks/go-yandex-shortener/internal/gzip"
	"github.com/MisterMaks/go-yandex-shortener/internal/logger"
)

const (
	Addr                          string = "localhost:8080"
	ResultAddrPrefix              string = "http://localhost:8080/"
	FileStoragePath               string = "/tmp/short-url-db.json"
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
	Ping(w http.ResponseWriter, r *http.Request)
}

func shortenerRouter(appHandler AppHandlerInterface, redirectPathPrefix string) chi.Router {
	r := chi.NewRouter()
	r.Use(logger.RequestLogger)
	redirectPathPrefix = strings.TrimPrefix(redirectPathPrefix, "/")
	r.Get(`/`+redirectPathPrefix+`{id}`, appHandler.RedirectToURL)
	r.Get(`/ping`, appHandler.Ping)
	r.Route(`/`, func(r chi.Router) {
		r.Use(gzip.GzipMiddleware)
		r.Post(`/`, appHandler.GetOrCreateURL)
		r.Post(`/api/shorten`, appHandler.APIGetOrCreateURL)
	})
	return r
}

func connectPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return db, nil
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

	db, err := connectPostgres(config.DatabaseDSN)
	if err != nil {
		logger.Log.Fatal("Failed to connect to Postgres",
			zap.Error(err),
		)
	}

	appRepo, err := appRepoInternal.NewAppRepoInmem(config.FileStoragePath)
	if err != nil {
		logger.Log.Fatal("Failed to create appRepo",
			zap.Error(err),
		)
	}
	defer appRepo.Close()

	appRepoPostgres, err := appRepoInternal.NewAppRepoPostgres(db)
	if err != nil {
		logger.Log.Fatal("Failed to create appRepoPostgres",
			zap.Error(err),
		)
	}

	appUsecase, err := appUsecaseInternal.NewAppUsecase(
		appRepo,
		config.BaseURL,
		CountRegenerationsForLengthID,
		LengthID,
		MaxLengthID,
		appRepoPostgres,
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
