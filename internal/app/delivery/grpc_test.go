package delivery

import (
	"context"
	"fmt"
	"testing"

	pb "github.com/MisterMaks/go-yandex-shortener/api/proto"
	"github.com/MisterMaks/go-yandex-shortener/internal/app"
	"github.com/MisterMaks/go-yandex-shortener/internal/app/delivery/mocks"
	"github.com/MisterMaks/go-yandex-shortener/internal/user/usecase"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestNewAppGRPCHandler(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppUsecaseInterface(ctrl)

	appGRPCHandler := NewAppGRPCHandler(m)
	assert.Equal(t, &AppGRPCHandler{
		UnimplementedAppServer: pb.UnimplementedAppServer{},
		AppUsecase:             m,
	}, appGRPCHandler)
}

func TestAppGRPCHandler_GetOrCreateURL(t *testing.T) {
	url := &app.URL{
		ID:        TestID,
		URL:       TestValidURL,
		UserID:    TestUserID,
		IsDeleted: false,
	}

	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// создаём объект-заглушку
	m := mocks.NewMockAppUsecaseInterface(ctrl)

	m.EXPECT().GetOrCreateURL(TestValidURL, TestUserID).Return(url, false, nil).AnyTimes()
	m.EXPECT().GetOrCreateURL(gomock.Any(), gomock.Any()).Return(nil, false, ErrTestInvalidURL).AnyTimes()

	m.EXPECT().GenerateShortURL(gomock.Any()).DoAndReturn(
		func(id string) string {
			return id
		},
	)

	type request struct {
		ctx context.Context
		in  *pb.GetOrCreateURLRequest
	}

	type want struct {
		out *pb.GetOrCreateURLResponse
		err error
	}

	tests := []struct {
		name    string
		request request
		want    want
	}{
		{
			name: "ok",
			request: request{
				ctx: context.WithValue(context.Background(), usecase.UserIDKey, TestUserID),
				in:  &pb.GetOrCreateURLRequest{Url: TestValidURL},
			},
			want: want{
				out: &pb.GetOrCreateURLResponse{ShortUrl: TestID},
				err: nil,
			},
		},
		{
			name: "invalid URL",
			request: request{
				ctx: context.WithValue(context.Background(), usecase.UserIDKey, TestUserID),
				in:  &pb.GetOrCreateURLRequest{Url: TestInvalidURL},
			},
			want: want{
				out: nil,
				err: status.Error(codes.Unknown, "Bad request"),
			},
		},
		{
			name: "user unauthorized",
			request: request{
				ctx: context.Background(),
				in:  &pb.GetOrCreateURLRequest{Url: TestValidURL},
			},
			want: want{
				out: nil,
				err: status.Error(codes.Unauthenticated, "User unauthorized"),
			},
		},
	}

	appGRPCHandler := NewAppGRPCHandler(m)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out, err := appGRPCHandler.GetOrCreateURL(tt.request.ctx, tt.request.in)

			assert.ErrorIs(t, err, tt.want.err)
			assert.Equal(t, tt.want.out, out)
		})
	}
}

func TestAppGRPCHandler_Ping(t *testing.T) {
	// создаём контроллер
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	type request struct {
		ctx context.Context
		in  *pb.PingRequest
	}

	type want struct {
		out *pb.PingResponse
		err error
	}

	tests := []struct {
		name             string
		request          request
		usecasePingError error
		want             want
	}{
		{
			name: "ok",
			request: request{
				ctx: context.Background(),
				in:  &pb.PingRequest{},
			},
			usecasePingError: nil,
			want: want{
				out: &pb.PingResponse{},
				err: nil,
			},
		},
		{
			name: "internal server error",
			request: request{
				ctx: context.Background(),
				in:  &pb.PingRequest{},
			},
			usecasePingError: fmt.Errorf("internal server error"),
			want: want{
				out: nil,
				err: status.Error(codes.Internal, "Internal error"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// создаём объект-заглушку
			m := mocks.NewMockAppUsecaseInterface(ctrl)
			m.EXPECT().Ping().Return(tt.usecasePingError)

			appGRPCHandler := NewAppGRPCHandler(m)

			out, err := appGRPCHandler.Ping(tt.request.ctx, tt.request.in)

			assert.ErrorIs(t, err, tt.want.err)
			assert.Equal(t, tt.want.out, out)
		})
	}
}
