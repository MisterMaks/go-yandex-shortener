package main

import (
	"context"
	"database/sql"
	"github.com/pressly/goose/v3"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	appDeliveryInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/delivery"
	appRepoInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/repo"
	appUsecaseInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/usecase"
	"github.com/MisterMaks/go-yandex-shortener/internal/gzip"
	"github.com/MisterMaks/go-yandex-shortener/internal/logger"
	userRepoInternal "github.com/MisterMaks/go-yandex-shortener/internal/user/repo"
	userUsecaseInternal "github.com/MisterMaks/go-yandex-shortener/internal/user/usecase"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
)

const (
	Addr                          string = "localhost:8080"
	ResultAddrPrefix              string = "http://localhost:8080/"
	URLsFileStoragePath           string = "/tmp/short-url-db.json"
	UsersFileStoragePath          string = "/tmp/user-db.json"
	CountRegenerationsForLengthID uint   = 5
	LengthID                      uint   = 5
	MaxLengthID                   uint   = 20
	LogLevel                      string = "INFO"
	SecretKey                     string = "supersecretkey"
	TokenExp                             = time.Hour * 3

	ConfigKey string = "config"
	AddrKey   string = "addr"
)

func migrate(dsn string) error {
	db, err := goose.OpenDBWithDriver("postgres", dsn)
	if err != nil {
		return err
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Log.Fatal("Failed to close DB",
				zap.Error(err),
			)
		}
	}()
	ctx := context.Background()
	return goose.RunContext(ctx, "up", db, "./migrations/")
}

type AppHandlerInterface interface {
	GetOrCreateURL(w http.ResponseWriter, r *http.Request)
	APIGetOrCreateURL(w http.ResponseWriter, r *http.Request)
	RedirectToURL(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
	APIGetOrCreateURLs(w http.ResponseWriter, r *http.Request)
	APIGetUserURLs(w http.ResponseWriter, r *http.Request)
}

type Middlewares struct {
	RequestLogger          func(http.Handler) http.Handler
	GzipMiddleware         func(http.Handler) http.Handler
	Authenticate           func(http.Handler) http.Handler
	AuthenticateOrRegister func(http.Handler) http.Handler
}

func shortenerRouter(
	appHandler AppHandlerInterface,
	redirectPathPrefix string,
	middlewares *Middlewares,
) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares.RequestLogger)
	r.With(middlewares.Authenticate, middlewares.GzipMiddleware).Get(`/api/user/urls`, appHandler.APIGetUserURLs)
	r.Use(middlewares.AuthenticateOrRegister)
	redirectPathPrefix = strings.TrimPrefix(redirectPathPrefix, "/")
	r.Get(`/`+redirectPathPrefix+`{id}`, appHandler.RedirectToURL)
	r.Get(`/ping`, appHandler.Ping)
	r.Route(`/`, func(r chi.Router) {
		r.Use(middlewares.GzipMiddleware)
		r.Post(`/`, appHandler.GetOrCreateURL)
		r.Post(`/api/shorten`, appHandler.APIGetOrCreateURL)
		r.Post(`/api/shorten/batch`, appHandler.APIGetOrCreateURLs)
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
		logger.Log.Error("Failed to ping DB Postgres",
			zap.Error(err),
		)
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
	defer db.Close()

	var appRepo appUsecaseInternal.AppRepoInterface
	var userRepo userUsecaseInternal.UserRepoInterface
	switch config.DatabaseDSN {
	case "":
		appRepoInmem, err := appRepoInternal.NewAppRepoInmem(config.FileStoragePath)
		if err != nil {
			logger.Log.Fatal("Failed to create appRepoInmem",
				zap.Error(err),
			)
		}
		defer appRepoInmem.Close()
		appRepo = appRepoInmem

		userRepoInmem, err := userRepoInternal.NewUserRepoInmem(UsersFileStoragePath)
		if err != nil {
			logger.Log.Fatal("Failed to create userRepoInmem",
				zap.Error(err),
			)
		}
		defer userRepoInmem.Close()
		userRepo = userRepoInmem
	default:
		logger.Log.Info("Applying migrations")
		err = migrate(config.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("Failed to apply migrations",
				zap.Error(err),
			)
		}

		appRepoPostgres, err := appRepoInternal.NewAppRepoPostgres(db)
		if err != nil {
			logger.Log.Fatal("Failed to create appRepoPostgres",
				zap.Error(err),
			)
		}
		appRepo = appRepoPostgres

		userRepoPostgres, err := userRepoInternal.NewUserRepoPostgres(db)
		if err != nil {
			logger.Log.Fatal("Failed to create userRepoPostgres",
				zap.Error(err),
			)
		}
		userRepo = userRepoPostgres
	}

	appUsecase, err := appUsecaseInternal.NewAppUsecase(
		appRepo,
		config.BaseURL,
		CountRegenerationsForLengthID,
		LengthID,
		MaxLengthID,
		db,
	)
	if err != nil {
		logger.Log.Fatal("Failed to create appUsecase",
			zap.Error(err),
		)
	}

	userUsecase, err := userUsecaseInternal.NewUserUsecase(
		userRepo,
		SecretKey,
		TokenExp,
	)
	if err != nil {
		logger.Log.Fatal("Failed to create userUsecase",
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

	middlewares := &Middlewares{
		RequestLogger:          logger.RequestLogger,
		GzipMiddleware:         gzip.GzipMiddleware,
		AuthenticateOrRegister: userUsecase.AuthenticateOrRegister,
		Authenticate:           userUsecase.Authenticate,
	}

	r := shortenerRouter(appHandler, redirectPathPrefix, middlewares)

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
