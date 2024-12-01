package shortener

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/pressly/goose/v3"
	httpSwagger "github.com/swaggo/http-swagger"

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
	DeleteURLsWaitingTime                = 5 * time.Second
	DeleteURLsChanSize            uint   = 1024

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
	APIDeleteUserURLs(w http.ResponseWriter, r *http.Request)
}

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
) chi.Router {
	r := chi.NewRouter()
	r.Use(middlewares.RequestLogger)

	SwaggerInfo.Host = baseURL.Host
	SwaggerInfo.Schemes = []string{"http", "https"}
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
		DeleteURLsChanSize,
		DeleteURLsWaitingTime,
	)
	if err != nil {
		logger.Log.Fatal("Failed to create appUsecase",
			zap.Error(err),
		)
	}
	defer appUsecase.Close()

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

	r := shortenerRouter(appHandler, u, middlewares)

	logger.Log.Info("Server running",
		zap.String(AddrKey, config.ServerAddress),
	)
	go func() {
		err = http.ListenAndServe(config.ServerAddress, r)
		if err != nil {
			logger.Log.Fatal("Failed to start server",
				zap.Error(err),
			)
		}
	}()

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)

	for exitSyg := range exitChan {
		logger.Log.Info("terminating: via signal", zap.Any("signal", exitSyg))
		break
	}
}
