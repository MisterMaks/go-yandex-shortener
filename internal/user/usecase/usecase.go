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
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
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
	GetCountUsers() (int, error) // get count users
	Close() error
}

// UserUsecase business logic struct.
type UserUsecase struct {
	UserRepo UserRepoInterface

	SecretKey string
	TokenExp  time.Duration

	grpcMethodsForAuthenticateOrRegisterUnaryInterceptor map[string]struct{}
	grpcMethodsForAuthenticateUnaryInterceptor           map[string]struct{}
}

// NewUserUsecase creates *UserUsecase.
func NewUserUsecase(
	userRepo UserRepoInterface,
	sk string,
	te time.Duration,
	grpcMethodsForAuthenticateOrRegisterUnaryInterceptorSl []string,
	grpcMethodsForAuthenticateUnaryInterceptorSl []string,
) (*UserUsecase, error) {
	grpcMethodsForAuthenticateOrRegisterUnaryInterceptor := map[string]struct{}{}
	grpcMethodsForAuthenticateUnaryInterceptor := map[string]struct{}{}

	for _, grpcMethod := range grpcMethodsForAuthenticateOrRegisterUnaryInterceptorSl {
		grpcMethodsForAuthenticateOrRegisterUnaryInterceptor[grpcMethod] = struct{}{}
	}

	for _, grpcMethod := range grpcMethodsForAuthenticateUnaryInterceptorSl {
		grpcMethodsForAuthenticateUnaryInterceptor[grpcMethod] = struct{}{}
	}

	return &UserUsecase{
		UserRepo: userRepo,

		SecretKey: sk,
		TokenExp:  te,

		grpcMethodsForAuthenticateOrRegisterUnaryInterceptor: grpcMethodsForAuthenticateOrRegisterUnaryInterceptor,
		grpcMethodsForAuthenticateUnaryInterceptor:           grpcMethodsForAuthenticateUnaryInterceptor,
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

		ctx := context.WithValue(r.Context(), UserIDKey, userID)

		ctxLogger = ctxLogger.With(zap.Uint(string(UserIDKey), userID))
		ctx = context.WithValue(ctx, logger.LoggerKey, ctxLogger)

		h.ServeHTTP(w, r.WithContext(ctx))
	})
}

// Authenticate auths user using JWT token in Cookie.
func (uu *UserUsecase) Authenticate(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(AccessTokenKey)
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

// AuthenticateOrRegisterUnaryInterceptor is unary interceptor for auths or registers user.
func (uu *UserUsecase) AuthenticateOrRegisterUnaryInterceptor(ctx context.Context, req any, si *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if _, ok := uu.grpcMethodsForAuthenticateOrRegisterUnaryInterceptor[si.FullMethod]; !ok {
		return handler(ctx, req)
	}

	ctxLogger := logger.GetContextLogger(ctx)

	var token string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("accessToken")
		if len(values) > 0 {
			token = values[0]
		}
	}

	var u *user.User
	var accessToken string
	var err error
	var userID uint

	if len(token) != 0 {
		userID, err = uu.getUserID(token)
	}

	if len(token) == 0 || err != nil {
		u, err = uu.CreateUser()
		if err != nil {
			return nil, status.Error(codes.Unknown, "Bad request")
		}

		accessToken, err = uu.buildJWTString(u.ID)
		if err != nil {
			return nil, status.Error(codes.Internal, "Internal error")
		}

		header := metadata.Pairs(AccessTokenKey, accessToken)
		err = grpc.SetHeader(ctx, header)
		if err != nil {
			ctxLogger.Error("Internal error", zap.Error(err))
			return nil, status.Error(codes.Internal, "Internal error")
		}

		ctx = context.WithValue(ctx, UserIDKey, u.ID)

		ctxLogger = ctxLogger.With(zap.Uint(string(UserIDKey), u.ID))
		ctx = context.WithValue(ctx, logger.LoggerKey, ctxLogger)

		return handler(ctx, req)
	}

	ctx = context.WithValue(ctx, UserIDKey, userID)

	ctxLogger = ctxLogger.With(zap.Uint(string(UserIDKey), userID))
	ctx = context.WithValue(ctx, logger.LoggerKey, ctxLogger)

	return handler(ctx, req)
}

// AuthenticateUnaryInterceptor is unary interceptor for auths user.
func (uu *UserUsecase) AuthenticateUnaryInterceptor(ctx context.Context, req any, si *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	if _, ok := uu.grpcMethodsForAuthenticateUnaryInterceptor[si.FullMethod]; !ok {
		return handler(ctx, req)
	}

	var token string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		values := md.Get("accessToken")
		if len(values) > 0 {
			token = values[0]
		}
	}

	if len(token) == 0 {
		return nil, status.Error(codes.Unauthenticated, "missing accessToken")
	}

	userID, err := uu.getUserID(token)

	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid accessToken")
	}

	ctx = context.WithValue(ctx, UserIDKey, userID)

	ctxLogger := logger.GetContextLogger(ctx)
	ctxLogger = ctxLogger.With(zap.Uint(string(UserIDKey), userID))
	ctx = context.WithValue(ctx, logger.LoggerKey, ctxLogger)

	return handler(ctx, req)
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

// GetCountUsers get count users.
func (uu *UserUsecase) GetCountUsers() (int, error) {
	return uu.UserRepo.GetCountUsers()
}
