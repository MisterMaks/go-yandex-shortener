package delivery

import (
	"context"

	pb "github.com/MisterMaks/go-yandex-shortener/api/proto"
	"github.com/MisterMaks/go-yandex-shortener/internal/app"
	"github.com/MisterMaks/go-yandex-shortener/internal/logger"
	"github.com/MisterMaks/go-yandex-shortener/internal/user/usecase"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AppGRPCHandler struct {
	pb.UnimplementedAppServer

	AppUsecase AppUsecaseInterface
}

// NewAppGRPCHandler creates *AppGRPCHandler
func NewAppGRPCHandler(appUsecase AppUsecaseInterface) *AppGRPCHandler {
	return &AppGRPCHandler{AppUsecase: appUsecase}
}

// GetOrCreateURL Get (if URL existed) or create URL.
func (agh *AppGRPCHandler) GetOrCreateURL(ctx context.Context, in *pb.GetOrCreateURLRequest) (*pb.GetOrCreateURLResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Creating or getting URL")

	userID, err := usecase.GetContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn("No user ID",
			zap.Any(RequestKey, in),
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, "User unauthorized")
	}

	url, _, err := agh.AppUsecase.GetOrCreateURL(in.Url, userID)
	if err != nil {
		handlerLogger.Warn("Bad request",
			zap.Any(RequestKey, in),
			zap.Error(err),
		)
		return nil, status.Error(codes.Unknown, "Bad request")
	}

	shortURL := agh.AppUsecase.GenerateShortURL(url.ID)

	handlerLogger.Info("Short URL created",
		zap.String(URLIDKey, url.ID),
		zap.String(URLKey, url.URL),
		zap.String(ShortURLKey, shortURL),
	)

	return &pb.GetOrCreateURLResponse{ShortUrl: shortURL}, nil
}

// rpc RedirectToURL(RedirectToURLRequest) returns (RedirectToURLResponse);

// Ping Ping database.
func (agh *AppGRPCHandler) Ping(ctx context.Context, _ *pb.PingRequest) (*pb.PingResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Ping DB")

	err := agh.AppUsecase.Ping()
	if err != nil {
		handlerLogger.Error("Failed to ping DB",
			zap.Error(err),
		)
		return nil, status.Error(codes.Internal, "Internal error")
	}

	return nil, nil
}

// GetOrCreateURLs Get (if URLs existed) or create URLs.
func (agh *AppGRPCHandler) GetOrCreateURLs(ctx context.Context, in *pb.GetOrCreateURLsRequest) (*pb.GetOrCreateURLsResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Creating or getting URLs batch")

	userID, err := usecase.GetContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn("No user ID",
			zap.Any(RequestKey, in),
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, "User unauthorized")
	}

	countURLs := len(in.Urls)

	requestBatchURLs := make([]app.RequestBatchURL, countURLs)
	for i, url := range in.Urls {
		requestBatchURLs[i] = app.RequestBatchURL{
			CorrelationID: url.CorrelationId,
			OriginalURL:   url.OriginalUrl,
		}
	}

	responseBatchURLs, err := agh.AppUsecase.GetOrCreateURLs(requestBatchURLs, userID)
	if err != nil {
		handlerLogger.Warn("Bad request",
			zap.Any(URLsKey, requestBatchURLs),
			zap.Error(err),
		)
		return nil, status.Error(codes.Unknown, "Bad request")
	}

	shortURLs := make([]*pb.GetOrCreateURLsResponse_URL, countURLs)

	for i, url := range responseBatchURLs {
		shortURLs[i] = &pb.GetOrCreateURLsResponse_URL{
			CorrelationId: url.CorrelationID,
			ShortUrl:      url.ShortURL,
		}
	}

	return &pb.GetOrCreateURLsResponse{ShortUrls: shortURLs}, nil
}

// GetUserURLs Get user URLs.
func (agh *AppGRPCHandler) GetUserURLs(ctx context.Context, in *pb.GetUserURLsRequest) (*pb.GetUserURLsResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Getting user URLs")

	userID, err := usecase.GetContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn("No user ID",
			zap.Any(RequestKey, in),
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, "User unauthorized")
	}

	userURLs, err := agh.AppUsecase.GetUserURLs(userID)
	if err != nil {
		handlerLogger.Warn("Bad request", zap.Error(err))
		return nil, status.Error(codes.Unknown, "Bad request")
	}

	response := &pb.GetUserURLsResponse{Urls: make([]*pb.GetUserURLsResponse_URL, len(userURLs))}
	for i, url := range userURLs {
		response.Urls[i] = &pb.GetUserURLsResponse_URL{
			OriginalUrl: url.OriginalURL,
			ShortUrl:    url.ShortURL,
		}
	}

	return response, nil
}

// DeleteUserURLs Delete user URLs.
func (agh *AppGRPCHandler) DeleteUserURLs(ctx context.Context, in *pb.DeleteUserURLsRequest) (*pb.DeleteUserURLsResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Deleting user URLs")

	userID, err := usecase.GetContextUserID(ctx)
	if err != nil {
		handlerLogger.Warn("No user ID",
			zap.Any(RequestKey, in),
			zap.Error(err),
		)
		return nil, status.Error(codes.Unauthenticated, "User unauthorized")
	}

	agh.AppUsecase.SendDeleteUserURLsInChan(userID, in.UrlIds)

	return nil, nil
}

// GetInternalStats Get internal stats.
func (agh *AppGRPCHandler) GetInternalStats(ctx context.Context, _ *pb.GetInternalStatsRequest) (*pb.GetInternalStatsResponse, error) {
	handlerLogger := logger.GetContextLogger(ctx)

	handlerLogger.Info("Getting internal stats")

	internalStats, err := agh.AppUsecase.GetInternalStats()
	if err != nil {
		handlerLogger.Warn("Bad request", zap.Error(err))
		return nil, status.Error(codes.Unknown, "Bad request")
	}

	return &pb.GetInternalStatsResponse{
		Urls:  int64(internalStats.URLs),
		Users: int64(internalStats.Users),
	}, nil
}
