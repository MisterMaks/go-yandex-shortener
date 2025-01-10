package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/MisterMaks/go-yandex-shortener/internal/user"
	"github.com/MisterMaks/go-yandex-shortener/internal/user/usecase/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserUsecase(t *testing.T) {
	u, err := NewUserUsecase(nil, "secretkey", time.Second, nil, nil)
	require.NoError(t, err)
	assert.Equal(t, &UserUsecase{
		UserRepo:  nil,
		SecretKey: "secretkey",
		TokenExp:  time.Second,
		GRPCMethodsForAuthenticateUnaryInterceptor:           map[string]struct{}{},
		GRPCMethodsForAuthenticateOrRegisterUnaryInterceptor: map[string]struct{}{},
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
