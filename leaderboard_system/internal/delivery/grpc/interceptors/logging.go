package interceptors

import (
	"context"
	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
)

// slogLogger adapts slog.Logger to the go-grpc-middleware logging.Logger interface.
type slogLogger struct {
	logger *slog.Logger
}

func (l *slogLogger) Log(ctx context.Context, level logging.Level, msg string, fields ...any) {
	slogLevel := slog.LevelInfo
	switch level {
	case logging.LevelDebug:
		slogLevel = slog.LevelDebug
	case logging.LevelInfo:
		slogLevel = slog.LevelInfo
	case logging.LevelWarn:
		slogLevel = slog.LevelWarn
	case logging.LevelError:
		slogLevel = slog.LevelError
	}
	l.logger.Log(ctx, slogLevel, msg, fields...)
}

// LoggingUnaryInterceptor creates a gRPC unary interceptor for structured request logging
// using go-grpc-middleware/v2.
func LoggingUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return logging.UnaryServerInterceptor(&slogLogger{logger: logger})
}
