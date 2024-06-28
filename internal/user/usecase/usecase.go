package usecase

import (
	"context"
	"github.com/MisterMaks/go-yandex-shortener/internal/logger"
	"github.com/MisterMaks/go-yandex-shortener/internal/user"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type UserIDKeyType string

const UserIDKey UserIDKeyType = "user_id"

type Claims struct {
	jwt.RegisteredClaims
	UserID uint
}

type UserRepoInterface interface {
	CreateUser() (*user.User, error)
}

type UserUsecase struct {
	UserRepo UserRepoInterface

	SecretKey string
	TokenExp  time.Duration
}

func NewUserUsecase(userRepo UserRepoInterface, sk string, te time.Duration) (*UserUsecase, error) {
	return &UserUsecase{
		UserRepo: userRepo,

		SecretKey: sk,
		TokenExp:  te,
	}, nil
}

// BuildJWTString создаёт токен и возвращает его в виде строки.
func (uu *UserUsecase) BuildJWTString(userID uint) (string, error) {
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

func (uu *UserUsecase) CreateUser() (*user.User, error) {
	return uu.UserRepo.CreateUser()
}

func (uu *UserUsecase) AuthentificateOrRegister(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("accessToken")
		if err != nil {
			u, err := uu.CreateUser()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, u.ID)
			h.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		value := cookie.Value
		userID, err := uu.getUserID(value)
		if err != nil {
			u, err := uu.CreateUser()
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, u.ID)
			h.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, userID)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (uu *UserUsecase) Authentificate(h http.Handler) http.Handler {
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
		ctxLogger = ctxLogger.With(
			zap.Uint(string(UserIDKey), userID),
		)
		ctx = context.WithValue(ctx, logger.LoggerKey, ctxLogger)

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
