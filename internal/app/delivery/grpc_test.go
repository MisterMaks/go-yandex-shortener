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

func TestAppGRPCHandler_GetOrCreateURLs(t *testing.T) {
	type request struct {
		ctx context.Context
		in  *pb.GetOrCreateURLsRequest
	}

	type want struct {
		out *pb.GetOrCreateURLsResponse
		err error
	}

	type usecaseReturn struct {
		responseBatchURLs []app.ResponseBatchURL
		err               error
	}

	tests := []struct {
		name          string
		request       request
		usecaseReturn usecaseReturn
		want          want
	}{
		{
			name: "ok",
			request: request{
				ctx: context.WithValue(context.Background(), usecase.UserIDKey, TestUserID),
				in: &pb.GetOrCreateURLsRequest{
					Urls: []*pb.GetOrCreateURLsRequest_URL{
						{
							CorrelationId: TestID,
							OriginalUrl:   TestValidURL,
						},
					},
				},
			},
			usecaseReturn: usecaseReturn{
				responseBatchURLs: []app.ResponseBatchURL{{CorrelationID: TestID, ShortURL: TestID}},
				err:               nil,
			},
			want: want{
				out: &pb.GetOrCreateURLsResponse{
					ShortUrls: []*pb.GetOrCreateURLsResponse_URL{
						{
							CorrelationId: TestID,
							ShortUrl:      TestID,
						},
					},
				},
				err: nil,
			},
		},
		{
			name: "user unauthorized",
			request: request{
				ctx: context.Background(),
				in: &pb.GetOrCreateURLsRequest{
					Urls: []*pb.GetOrCreateURLsRequest_URL{
						{
							CorrelationId: TestID,
							OriginalUrl:   TestValidURL,
						},
					},
				},
			},
			usecaseReturn: usecaseReturn{
				responseBatchURLs: []app.ResponseBatchURL{{CorrelationID: TestID, ShortURL: TestID}},
				err:               nil,
			},
			want: want{
				out: nil,
				err: status.Error(codes.Unauthenticated, "User unauthorized"),
			},
		},
		{
			name: "internal server error",
			request: request{
				ctx: context.WithValue(context.Background(), usecase.UserIDKey, TestUserID),
				in: &pb.GetOrCreateURLsRequest{
					Urls: []*pb.GetOrCreateURLsRequest_URL{
						{
							CorrelationId: TestID,
							OriginalUrl:   TestValidURL,
						},
					},
				},
			},
			usecaseReturn: usecaseReturn{
				responseBatchURLs: nil,
				err:               fmt.Errorf("internal server error"),
			},
			want: want{
				out: nil,
				err: status.Error(codes.Unknown, "Bad request"),
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks.NewMockAppUsecaseInterface(ctrl)
			m.EXPECT().GetOrCreateURLs(gomock.Any(), gomock.Any()).Return(
				tt.usecaseReturn.responseBatchURLs,
				tt.usecaseReturn.err,
			).AnyTimes()

			appGRPCHandler := NewAppGRPCHandler(m)

			out, err := appGRPCHandler.GetOrCreateURLs(tt.request.ctx, tt.request.in)

			assert.ErrorIs(t, err, tt.want.err)
			assert.Equal(t, tt.want.out, out)
		})
	}
}

func TestAppGRPCHandler_GetUserURLs(t *testing.T) {
	type request struct {
		ctx context.Context
		in  *pb.GetUserURLsRequest
	}

	type want struct {
		out *pb.GetUserURLsResponse
		err error
	}

	type usecaseReturn struct {
		responseUserURLs []app.ResponseUserURL
		err              error
	}

	tests := []struct {
		name          string
		request       request
		usecaseReturn usecaseReturn
		want          want
	}{
		{
			name: "ok",
			request: request{
				ctx: context.WithValue(context.Background(), usecase.UserIDKey, TestUserID),
				in:  &pb.GetUserURLsRequest{},
			},
			usecaseReturn: usecaseReturn{
				responseUserURLs: []app.ResponseUserURL{
					{
						ShortURL:    TestID,
						OriginalURL: TestValidURL,
					},
				},
				err: nil,
			},
			want: want{
				out: &pb.GetUserURLsResponse{
					Urls: []*pb.GetUserURLsResponse_URL{{ShortUrl: TestID, OriginalUrl: TestValidURL}},
				},
				err: nil,
			},
		},
		{
			name: "user unauthorized",
			request: request{
				ctx: context.Background(),
				in:  &pb.GetUserURLsRequest{},
			},
			usecaseReturn: usecaseReturn{
				responseUserURLs: []app.ResponseUserURL{
					{
						ShortURL:    TestID,
						OriginalURL: TestValidURL,
					},
				},
				err: nil,
			},
			want: want{
				out: nil,
				err: status.Error(codes.Unauthenticated, "User unauthorized"),
			},
		},
		{
			name: "internal server error",
			request: request{
				ctx: context.WithValue(context.Background(), usecase.UserIDKey, TestUserID),
				in:  &pb.GetUserURLsRequest{},
			},
			usecaseReturn: usecaseReturn{
				responseUserURLs: nil,
				err:              fmt.Errorf("internal server error"),
			},
			want: want{
				out: nil,
				err: status.Error(codes.Unknown, "Bad request"),
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks.NewMockAppUsecaseInterface(ctrl)
			m.EXPECT().GetUserURLs(gomock.Any()).Return(
				tt.usecaseReturn.responseUserURLs,
				tt.usecaseReturn.err,
			).AnyTimes()

			appGRPCHandler := NewAppGRPCHandler(m)

			out, err := appGRPCHandler.GetUserURLs(tt.request.ctx, tt.request.in)

			assert.ErrorIs(t, err, tt.want.err)
			assert.Equal(t, tt.want.out, out)
		})
	}
}

func TestAppGRPCHandler_DeleteUserURLs(t *testing.T) {
	type request struct {
		ctx context.Context
		in  *pb.DeleteUserURLsRequest
	}

	type want struct {
		out *pb.DeleteUserURLsResponse
		err error
	}

	tests := []struct {
		name             string
		request          request
		usecaseCallTimes int
		want             want
	}{
		{
			name: "ok",
			request: request{
				ctx: context.WithValue(context.Background(), usecase.UserIDKey, TestUserID),
				in:  &pb.DeleteUserURLsRequest{UrlIds: []string{TestID}},
			},
			usecaseCallTimes: 1,
			want: want{
				out: &pb.DeleteUserURLsResponse{},
				err: nil,
			},
		},
		{
			name: "user unauthorized",
			request: request{
				ctx: context.Background(),
				in:  &pb.DeleteUserURLsRequest{UrlIds: []string{TestID}},
			},
			usecaseCallTimes: 0,
			want: want{
				out: nil,
				err: status.Error(codes.Unauthenticated, "User unauthorized"),
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks.NewMockAppUsecaseInterface(ctrl)
			m.EXPECT().SendDeleteUserURLsInChan(gomock.Any(), gomock.Any()).Return().Times(tt.usecaseCallTimes)

			appGRPCHandler := NewAppGRPCHandler(m)

			out, err := appGRPCHandler.DeleteUserURLs(tt.request.ctx, tt.request.in)

			assert.ErrorIs(t, err, tt.want.err)
			assert.Equal(t, tt.want.out, out)
		})
	}
}

func TestAppGRPCHandler_GetInternalStats(t *testing.T) {
	type request struct {
		ctx context.Context
		in  *pb.GetInternalStatsRequest
	}

	type want struct {
		out *pb.GetInternalStatsResponse
		err error
	}

	type usecaseReturn struct {
		internalStats app.InternalStats
		err           error
	}

	tests := []struct {
		name          string
		request       request
		usecaseReturn usecaseReturn
		want          want
	}{
		{
			name: "ok",
			request: request{
				ctx: context.Background(),
				in:  &pb.GetInternalStatsRequest{},
			},
			usecaseReturn: usecaseReturn{
				internalStats: app.InternalStats{URLs: 1, Users: 1},
				err:           nil,
			},
			want: want{
				out: &pb.GetInternalStatsResponse{
					Urls:  1,
					Users: 1,
				},
				err: nil,
			},
		},
		{
			name: "internal server error",
			request: request{
				ctx: context.Background(),
				in:  &pb.GetInternalStatsRequest{},
			},
			usecaseReturn: usecaseReturn{
				internalStats: app.InternalStats{},
				err:           fmt.Errorf("internal server error"),
			},
			want: want{
				out: nil,
				err: status.Error(codes.Unknown, "Bad request"),
			},
		},
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := mocks.NewMockAppUsecaseInterface(ctrl)
			m.EXPECT().GetInternalStats().Return(tt.usecaseReturn.internalStats, tt.usecaseReturn.err).AnyTimes()

			appGRPCHandler := NewAppGRPCHandler(m)

			out, err := appGRPCHandler.GetInternalStats(tt.request.ctx, tt.request.in)

			assert.ErrorIs(t, err, tt.want.err)
			assert.Equal(t, tt.want.out, out)
		})
	}
}
