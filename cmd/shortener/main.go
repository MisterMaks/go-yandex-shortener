package main

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/MisterMaks/go-yandex-shortener/api"
	appDeliveryInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/delivery"
	appRepoInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/repo"
	appUsecaseInternal "github.com/MisterMaks/go-yandex-shortener/internal/app/usecase"
	"github.com/MisterMaks/go-yandex-shortener/internal/certcreator"
	"github.com/MisterMaks/go-yandex-shortener/internal/gzip"
	"github.com/MisterMaks/go-yandex-shortener/internal/logger"
	userRepoInternal "github.com/MisterMaks/go-yandex-shortener/internal/user/repo"
	userUsecaseInternal "github.com/MisterMaks/go-yandex-shortener/internal/user/usecase"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	httpSwagger "github.com/swaggo/http-swagger"
	"go.uber.org/zap"
)

// App constants.
const (
	Addr                          string = "localhost:8080"
	ResultAddrPrefix              string = "localhost:8080"
	URLsFileStoragePath           string = "/tmp/short-url-db.json"
	DeletedURLsFileStoragePath    string = "/tmp/deleted-url-db.json"
	UsersFileStoragePath          string = "/tmp/user-db.json"
	CountRegenerationsForLengthID uint   = 5
	LengthID                      uint   = 5
	MaxLengthID                   uint   = 20
	LogLevel                      string = "INFO"
	SecretKey                     string = "supersecretkey"
	TokenExp                             = time.Hour * 3
	DeleteURLsWaitingTime                = 5 * time.Second
	DeleteURLsChanSize            uint   = 1024

	ConfigKey string = "config"
	AddrKey   string = "addr"
)

var buildVersion string
var buildDate string
var buildCommit string

func printBuildInfo() {
	if buildVersion == "" {
		fmt.Println("Build version: N/A")
	} else {
		fmt.Println("Build version:", buildVersion)
	}

	if buildDate == "" {
		fmt.Println("Build date: N/A")
	} else {
		fmt.Println("Build date:", buildDate)
	}

	if buildCommit == "" {
		fmt.Println("Build commit: N/A")
	} else {
		fmt.Println("Build commit:", buildCommit)
	}
}

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

// AppHandlerInterface contains the necessary functions for the handlers of app.
type AppHandlerInterface interface {
	GetOrCreateURL(w http.ResponseWriter, r *http.Request)
	APIGetOrCreateURL(w http.ResponseWriter, r *http.Request)
	RedirectToURL(w http.ResponseWriter, r *http.Request)
	Ping(w http.ResponseWriter, r *http.Request)
	APIGetOrCreateURLs(w http.ResponseWriter, r *http.Request)
	APIGetUserURLs(w http.ResponseWriter, r *http.Request)
	APIDeleteUserURLs(w http.ResponseWriter, r *http.Request)
}

// Middlewares used middlewares.
type Middlewares struct {
	RequestLogger          func(http.Handler) http.Handler
	GzipMiddleware         func(http.Handler) http.Handler
	Authenticate           func(http.Handler) http.Handler
	AuthenticateOrRegister func(http.Handler) http.Handler
}

func shortenerRouter(
	appHandler AppHandlerInterface,
	baseURL *url.URL,
	middlewares *Middlewares,
) (chi.Router, error) {
	if baseURL == nil {
		return nil, fmt.Errorf("baseURL == nil")
	}
	if middlewares == nil {
		return nil, fmt.Errorf("middlewares == nil")
	}

	r := chi.NewRouter()
	r.Use(middlewares.RequestLogger)

	api.SwaggerInfo.Host = baseURL.Host
	api.SwaggerInfo.Schemes = []string{"http", "https"}
	r.Get("/swagger/*", httpSwagger.Handler())

	redirectPathPrefix := strings.TrimPrefix(baseURL.Path, "/")
	r.Get(`/`+redirectPathPrefix+`{id}`, appHandler.RedirectToURL)
	r.Get(`/ping`, appHandler.Ping)
	r.Route(`/`, func(r chi.Router) {
		r.Use(middlewares.GzipMiddleware, middlewares.AuthenticateOrRegister)
		r.Post(`/`, appHandler.GetOrCreateURL)
		r.Post(`/api/shorten`, appHandler.APIGetOrCreateURL)
		r.Post(`/api/shorten/batch`, appHandler.APIGetOrCreateURLs)
	})
	r.Route(`/api/user/urls`, func(r chi.Router) {
		r.Use(middlewares.Authenticate)
		r.Get(`/`, appHandler.APIGetUserURLs)
		r.Delete(`/`, appHandler.APIDeleteUserURLs)
	})

	return r, nil
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

func runServer(server *http.Server, enableHTTPS bool) {
	var err error

	if enableHTTPS {
		var certPEMBytes, privateKeyPEMBytes []byte

		certPEMBytes, privateKeyPEMBytes, err = certcreator.Create()
		if err != nil {
			logger.Log.Fatal("Failed to create certificate",
				zap.Error(err),
			)
		}

		var cert tls.Certificate
		cert, err = tls.X509KeyPair(certPEMBytes, privateKeyPEMBytes)
		if err != nil {
			logger.Log.Fatal("Failed to parse a public/private key pair from a pair of PEM encoded data",
				zap.Error(err),
			)
		}

		tlsConfig := &tls.Config{
			Certificates: []tls.Certificate{cert},
			MinVersion:   tls.VersionTLS12,
		}

		server.TLSConfig = tlsConfig

		err = server.ListenAndServeTLS("", "")
	} else {
		err = server.ListenAndServe()
	}

	if err != nil && err != http.ErrServerClosed {
		logger.Log.Fatal("Failed to start server",
			zap.Error(err),
		)
	}
}

func main() {
	printBuildInfo()

	config, err := NewConfig()
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to create config. Error:", err)
	}

	err = logger.Initialize(config.LogLevel)
	if err != nil {
		log.Fatalln("CRITICAL\tFailed to init logger. Error:", err)
	}

	logger.Log.Debug("Config data",
		zap.Any(ConfigKey, config),
	)

	var db *sql.DB

	if config.DatabaseDSN != "" {
		logger.Log.Info("Applying migrations")
		err = migrate(config.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("Failed to apply migrations",
				zap.Error(err),
			)
		}

		db, err = connectPostgres(config.DatabaseDSN)
		if err != nil {
			logger.Log.Fatal("Failed to connect to Postgres",
				zap.Error(err),
			)
		}
		defer func() {
			err = db.Close()
			if err != nil {
				logger.Log.Fatal("Failed to close Postgres",
					zap.Error(err),
				)
			}
		}()
	}

	appRepo, err := appRepoInternal.NewAppRepo(
		db,
		config.FileStoragePath,
		DeletedURLsFileStoragePath,
	)
	if err != nil {
		logger.Log.Fatal("Failed to create appRepo",
			zap.Error(err),
		)
	}
	defer func() {
		err = appRepo.Close()
		if err != nil {
			logger.Log.Fatal("Failed to close appRepo",
				zap.Error(err),
			)
		}
	}()

	userRepo, err := userRepoInternal.NewUserRepo(db, UsersFileStoragePath)
	if err != nil {
		logger.Log.Fatal("Failed to create userRepo",
			zap.Error(err),
		)
	}
	defer func() {
		err = userRepo.Close()
		if err != nil {
			logger.Log.Fatal("Failed to close userRepo",
				zap.Error(err),
			)
		}
	}()

	appUsecase, err := appUsecaseInternal.NewAppUsecase(
		appRepo,
		config.BaseURL,
		CountRegenerationsForLengthID,
		LengthID,
		MaxLengthID,
		db,
		DeleteURLsChanSize,
		DeleteURLsWaitingTime,
	)
	if err != nil {
		logger.Log.Fatal("Failed to create appUsecase",
			zap.Error(err),
		)
	}

	defer func() {
		err = appUsecase.Close()
		if err != nil {
			logger.Log.Fatal("Failed to close userRepoInmem",
				zap.Error(err),
			)
		}
	}()

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

	middlewares := &Middlewares{
		RequestLogger:          logger.RequestLogger,
		GzipMiddleware:         gzip.GzipMiddleware,
		AuthenticateOrRegister: userUsecase.AuthenticateOrRegister,
		Authenticate:           userUsecase.Authenticate,
	}

	r, err := shortenerRouter(appHandler, u, middlewares)
	if err != nil {
		logger.Log.Fatal("Failed to create router",
			zap.Error(err),
		)
	}

	logger.Log.Info("Server running",
		zap.String(AddrKey, config.ServerAddress),
	)

	server := &http.Server{
		Addr:    config.ServerAddress,
		Handler: r,
	}

	go runServer(server, config.EnableHTTPS)

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	exitSyg := <-exitChan
	logger.Log.Info("terminating: via signal", zap.Any("signal", exitSyg))
	err = server.Shutdown(context.Background())
	if err != nil {
		logger.Log.Fatal("Failed to HTTP server shutdown", zap.Error(err))
	}
}
