package usecase

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/MisterMaks/go-yandex-shortener/internal/user"
	"github.com/MisterMaks/go-yandex-shortener/internal/user/usecase/mocks"
	"github.com/go-chi/chi/v5"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestNewUserUsecase(t *testing.T) {
	u, err := NewUserUsecase(nil, "secretkey", time.Second, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, &UserUsecase{
		UserRepo:  nil,
		SecretKey: "secretkey",
		TokenExp:  time.Second,
		grpcMethodsForAuthenticateUnaryInterceptor:           map[string]struct{}{},
		grpcMethodsForAuthenticateOrRegisterUnaryInterceptor: map[string]struct{}{},
	}, u)
}

func TestBuildJWTStringAndGetUserID(t *testing.T) {
	testUserID := uint(1)

	u, err := NewUserUsecase(nil, "secretkey", time.Second, nil, nil)
	require.NoError(t, err)

	jwtString, err := u.buildJWTString(testUserID)
	require.NoError(t, err)

	actualUserID, err := u.getUserID(jwtString)
	require.NoError(t, err)
	assert.Equal(t, testUserID, actualUserID)
}

func TestUserUsecase_CreateUser(t *testing.T) {
	testUser := &user.User{ID: 1}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockUserRepoInterface(ctrl)
	m.EXPECT().CreateUser().Return(testUser, nil)

	u, err := NewUserUsecase(m, "secretkey", time.Second, nil, nil)
	require.NoError(t, err)

	actualUser, err := u.CreateUser()
	require.NoError(t, err)
	assert.Equal(t, testUser, actualUser)
}

func TestGetContextUserID(t *testing.T) {
	var ctx context.Context

	ctx = nil

	userID, err := GetContextUserID(ctx)
	assert.Error(t, err)
	assert.Equal(t, uint(0), userID)

	ctx = context.Background()

	userID, err = GetContextUserID(ctx)
	assert.Error(t, err)
	assert.Equal(t, uint(0), userID)

	ctx = context.WithValue(context.Background(), UserIDKey, uint(1))

	userID, err = GetContextUserID(ctx)
	require.NoError(t, err)
	assert.Equal(t, uint(1), userID)
}

func TestUserUsecase_AuthenticateOrRegister(t *testing.T) {
	existingUserID := uint(1)
	newUserID := uint(2)

	newUser := &user.User{ID: newUserID}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockUserRepoInterface(ctrl)
	m.EXPECT().CreateUser().Return(newUser, nil).AnyTimes()

	u, err := NewUserUsecase(m, "secretkey", time.Second, nil, nil)
	require.NoError(t, err)

	jwt, err := u.buildJWTString(existingUserID)
	require.NoError(t, err)

	tests := []struct {
		name           string
		cookie         *http.Cookie
		expectedUserID uint
	}{
		{
			name:           "new user",
			cookie:         nil,
			expectedUserID: newUserID,
		},
		{
			name: "existed user",
			cookie: &http.Cookie{
				Name:     AccessTokenKey,
				Value:    jwt,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
			},
			expectedUserID: existingUserID,
		},
		{
			name: "bad cookie",
			cookie: &http.Cookie{
				Name:     AccessTokenKey,
				Value:    "tratata",
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
			},
			expectedUserID: newUserID,
		},
	}

	var expectedUserID uint

	handler := func(w http.ResponseWriter, r *http.Request) {
		actualUserID, ok := r.Context().Value(UserIDKey).(uint)
		require.True(t, ok)
		assert.Equal(t, expectedUserID, actualUserID)
	}

	r := chi.NewRouter()
	r.Use(u.AuthenticateOrRegister)
	r.Get(`/`, handler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedUserID = tt.expectedUserID

			request, err := http.NewRequest(http.MethodGet, ts.URL+"/", nil)
			require.NoError(t, err)

			if tt.cookie != nil {
				request.AddCookie(tt.cookie)
			}

			response, err := ts.Client().Do(request)
			require.NoError(t, err)
			err = response.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, http.StatusOK, response.StatusCode)
		})
	}
}

func TestUserUsecase_Authenticate(t *testing.T) {
	existingUserID := uint(1)

	u, err := NewUserUsecase(nil, "secretkey", time.Second, nil, nil)
	require.NoError(t, err)

	jwt, err := u.buildJWTString(existingUserID)
	require.NoError(t, err)

	handler := func(w http.ResponseWriter, r *http.Request) {
		actualUserID, ok := r.Context().Value(UserIDKey).(uint)
		require.True(t, ok)
		assert.Equal(t, existingUserID, actualUserID)
	}

	r := chi.NewRouter()
	r.Use(u.Authenticate)
	r.Get(`/`, handler)

	ts := httptest.NewServer(r)
	defer ts.Close()

	tests := []struct {
		name       string
		cookie     *http.Cookie
		statusCode int
	}{
		{
			name: "user authenticated",
			cookie: &http.Cookie{
				Name:     AccessTokenKey,
				Value:    jwt,
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
			},
			statusCode: http.StatusOK,
		},
		{
			name: "invalid jwt",
			cookie: &http.Cookie{
				Name:     AccessTokenKey,
				Value:    "tratata",
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
			},
			statusCode: http.StatusUnauthorized,
		},
		{
			name: "invalid cookie",
			cookie: &http.Cookie{
				HttpOnly: true,
				Secure:   true,
			},
			statusCode: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(http.MethodGet, ts.URL+"/", nil)
			require.NoError(t, err)

			request.AddCookie(tt.cookie)

			response, err := ts.Client().Do(request)
			require.NoError(t, err)
			err = response.Body.Close()
			require.NoError(t, err)
			assert.Equal(t, tt.statusCode, response.StatusCode)
		})
	}
}

func TestUserUsecase_AuthenticateOrRegisterUnaryInterceptor(t *testing.T) {
	existingUserID := uint(1)
	newUserID := uint(2)

	newUser := &user.User{ID: newUserID}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockUserRepoInterface(ctrl)
	m.EXPECT().CreateUser().Return(newUser, nil).AnyTimes()

	u, err := NewUserUsecase(m, "secretkey", time.Second, []string{"ok"}, nil)
	require.NoError(t, err)

	jwt, err := u.buildJWTString(existingUserID)
	require.NoError(t, err)

	mockSTS := mocks.NewMockServerTransportStream(ctrl)
	mockSTS.EXPECT().SetHeader(gomock.Any()).AnyTimes()

	handler := func(ctx context.Context, _ any) (any, error) {
		handlerActualUserID, handlerErr := GetContextUserID(ctx)
		return handlerActualUserID, handlerErr
	}

	type want struct {
		userID uint
		err    error
	}

	tests := []struct {
		name       string
		ctx        context.Context
		methodName string
		want       want
	}{
		{
			name: "ok",
			ctx: metadata.NewIncomingContext(
				grpc.NewContextWithServerTransportStream(context.Background(), mockSTS),
				metadata.MD{AccessTokenKey: []string{jwt}},
			),
			methodName: "ok",
			want: want{
				userID: existingUserID,
				err:    nil,
			},
		},
		{
			name:       "ignoring method",
			ctx:        grpc.NewContextWithServerTransportStream(context.Background(), mockSTS),
			methodName: "ignoring method",
			want: want{
				userID: 0,
				err:    fmt.Errorf("no user_id"),
			},
		},
		{
			name:       "new user",
			ctx:        grpc.NewContextWithServerTransportStream(context.Background(), mockSTS),
			methodName: "ok",
			want: want{
				userID: newUserID,
				err:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualUserID, actualErr := u.AuthenticateOrRegisterUnaryInterceptor(tt.ctx, nil, &grpc.UnaryServerInfo{FullMethod: tt.methodName}, handler)

			if tt.want.err != nil {
				assert.Error(t, actualErr)
			} else {
				assert.NoError(t, actualErr)
			}

			assert.Equal(t, tt.want.userID, actualUserID)
		})
	}
}
