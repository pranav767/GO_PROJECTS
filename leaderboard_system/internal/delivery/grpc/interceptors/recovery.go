package interceptors

import (
	"context"
	"log/slog"
	"runtime/debug"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryUnaryInterceptor creates a gRPC unary interceptor that recovers from panics
// in downstream handlers, logs the stack trace, and returns codes.Internal.
func RecoveryUnaryInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return recovery.UnaryServerInterceptor(
		recovery.WithRecoveryHandlerContext(
			func(ctx context.Context, p any) error {
				logger.Error("panic recovered in gRPC handler",
					slog.Any("panic", p),
					slog.String("stack", string(debug.Stack())),
				)
				return status.Error(codes.Internal, "internal server error")
			},
		),
	)
}
