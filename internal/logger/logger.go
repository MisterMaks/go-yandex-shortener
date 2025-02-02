package logger

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

// Log is global logger.
var Log *zap.Logger = zap.NewNop()

// LoggerKeyType is type for LoggerKey constant.
type LoggerKeyType string

// Constants for logget.
const (
	MethodKey            string        = "method"
	URIKey               string        = "uri"
	RequestIDKey         string        = "request_id"
	ExecutionDurationKey string        = "execution_duration"
	StatusCodeKey        string        = "status_code"
	StatusKey            string        = "status"
	ResponseBodySizeBKey string        = "response_body_size_B"
	LoggerKey            LoggerKeyType = "logger_key"
)

// Initialize init logger.
func Initialize(level string) error {
	// преобразуем текстовый уровень логирования в zap.AtomicLevel
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return err
	}
	// создаём новую конфигурацию логера
	cfg := zap.NewProductionConfig()
	// устанавливаем уровень
	cfg.Level = lvl
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)
	// создаём логер на основе конфигурации
	zl, err := cfg.Build()
	if err != nil {
		return err
	}
	// устанавливаем синглтон
	Log = zl
	return nil
}

func generateRequestID() string {
	id := uuid.New()
	return id.String()
}

// LoggerResponseWriter is logger for responses.
type LoggerResponseWriter struct {
	http.ResponseWriter
	bodySize   int
	statusCode int
}

// WriteHeader write status code in header.
func (lrw *LoggerResponseWriter) WriteHeader(statusCode int) {
	lrw.ResponseWriter.WriteHeader(statusCode)
	lrw.statusCode = statusCode
}

// WriteHeader write data in reponse and get body size.
func (lrw *LoggerResponseWriter) Write(bytes []byte) (int, error) {
	bodySize, err := lrw.ResponseWriter.Write(bytes)
	lrw.bodySize = bodySize
	return bodySize, err
}

// RequestLogger is logger middleware for app.
func RequestLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := generateRequestID()
		lrw := &LoggerResponseWriter{ResponseWriter: w}
		Log.Info("got incoming HTTP request",
			zap.String(MethodKey, r.Method),
			zap.Any(URIKey, r.RequestURI),
			zap.String(RequestIDKey, requestID),
		)
		ctxLogger := Log.With(
			zap.String(RequestIDKey, requestID),
		)
		ctx := context.WithValue(r.Context(), LoggerKey, ctxLogger)
		now := time.Now()
		h.ServeHTTP(lrw, r.WithContext(ctx))
		Log.Info("processed incoming HTTP request",
			zap.Int(StatusCodeKey, lrw.statusCode),
			zap.Int(ResponseBodySizeBKey, lrw.bodySize),
			zap.Duration(ExecutionDurationKey, time.Since(now)),
			zap.String(RequestIDKey, requestID),
		)
	})
}

// RequestLoggerUnaryInterceptor is logger unary interceptor for app.
func RequestLoggerUnaryInterceptor(ctx context.Context, req any, si *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	requestID := generateRequestID()
	Log.Info("got incoming GRPC request",
		zap.String(MethodKey, "POST"),
		zap.Any(URIKey, si.FullMethod),
		zap.String(RequestIDKey, requestID),
	)
	ctxLogger := Log.With(
		zap.String(RequestIDKey, requestID),
	)
	ctx = context.WithValue(ctx, LoggerKey, ctxLogger)
	now := time.Now()
	m, err := handler(ctx, req)

	responseSize := 0
	if m != nil {
		responseSize = proto.Size(m.(proto.Message))
	}

	Log.Info("processed incoming GRPC request",
		zap.Any(StatusCodeKey, uint32(status.Code(err))),
		zap.Any(StatusKey, status.Code(err)),
		zap.Int(ResponseBodySizeBKey, responseSize),
		zap.Duration(ExecutionDurationKey, time.Since(now)),
		zap.String(RequestIDKey, requestID),
	)
	return m, err
}

// GetContextLogger gets logger from context.
func GetContextLogger(ctx context.Context) *zap.Logger {
	if ctx == nil {
		return Log
	}
	logger, ok := ctx.Value(LoggerKey).(*zap.Logger)
	if !ok || logger == nil {
		return Log
	}
	return logger
}
