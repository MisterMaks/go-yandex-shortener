package usecase

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/MisterMaks/go-yandex-shortener/internal/logger"
	"github.com/MisterMaks/go-yandex-shortener/internal/user"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
)

// UserIDKeyType is type for UserIDKey constant.
type UserIDKeyType string

// Constants for usecase.
const (
	UserIDKey      UserIDKeyType = "user_id"
	AccessTokenKey string        = "accessToken"
)

// Claims is jwt.RegisteredClaims with UserID field.
type Claims struct {
	jwt.RegisteredClaims
	UserID uint
}

// UserRepoInterface contains the necessary functions for storage.
type UserRepoInterface interface {
	CreateUser() (*user.User, error)
	Close() error
}

// UserUsecase business logic struct.
type UserUsecase struct {
	UserRepo UserRepoInterface

	SecretKey string
	TokenExp  time.Duration
}

// NewUserUsecase creates *UserUsecase.
func NewUserUsecase(userRepo UserRepoInterface, sk string, te time.Duration) (*UserUsecase, error) {
	return &UserUsecase{
		UserRepo: userRepo,

		SecretKey: sk,
		TokenExp:  te,
	}, nil
}

// buildJWTString creates token and return it in string format.
func (uu *UserUsecase) buildJWTString(userID uint) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(uu.TokenExp)),
		},
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(uu.SecretKey))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func (uu *UserUsecase) getUserID(tokenString string) (uint, error) {
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
	_, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(uu.SecretKey), nil
	})
	if err != nil {
		return 0, err
	}

	// возвращаем ID пользователя в читаемом виде
	return claims.UserID, nil
}

// CreateUser create user.
func (uu *UserUsecase) CreateUser() (*user.User, error) {
	return uu.UserRepo.CreateUser()
}

// AuthenticateOrRegister auths or registers user using JWT token in Cookie.
func (uu *UserUsecase) AuthenticateOrRegister(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxLogger := logger.GetContextLogger(r.Context())

		cookie, err := r.Cookie(AccessTokenKey)

		var u *user.User
		var accessToken string

		if err != nil {
			u, err = uu.CreateUser()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			accessToken, err = uu.buildJWTString(u.ID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{Name: AccessTokenKey, Value: accessToken, Path: "/"})

			ctx := context.WithValue(r.Context(), UserIDKey, u.ID)

			ctxLogger = ctxLogger.With(zap.Uint(string(UserIDKey), u.ID))
			ctx = context.WithValue(ctx, logger.LoggerKey, ctxLogger)

			h.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		value := cookie.Value
		userID, err := uu.getUserID(value)
		if err != nil {
			u, err = uu.CreateUser()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			accessToken, err := uu.buildJWTString(u.ID)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			http.SetCookie(w, &http.Cookie{Name: AccessTokenKey, Value: accessToken, Path: "/"})

			ctx := context.WithValue(r.Context(), UserIDKey, u.ID)

			ctxLogger = ctxLogger.With(zap.Uint(string(UserIDKey), u.ID))
			ctx = context.WithValue(ctx, logger.LoggerKey, ctxLogger)

			h.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		ctxLogger = ctxLogger.With(zap.Uint(string(UserIDKey), userID))
		ctx = context.WithValue(ctx, logger.LoggerKey, ctxLogger)

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Authenticate auths user using JWT token in Cookie.
func (uu *UserUsecase) Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("accessToken")
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		value := cookie.Value
		userID, err := uu.getUserID(value)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		ctxLogger := logger.GetContextLogger(r.Context())
		ctxLogger = ctxLogger.With(zap.Uint(string(UserIDKey), userID))
		ctx = context.WithValue(ctx, logger.LoggerKey, ctxLogger)

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetContextUserID gets user ID from context.
func GetContextUserID(ctx context.Context) (uint, error) {
	if ctx == nil {
		return 0, fmt.Errorf("no context")
	}
	userID, ok := ctx.Value(UserIDKey).(uint)
	if !ok {
		return 0, fmt.Errorf("no %v", UserIDKey)
	}
	return userID, nil
}
