package middleware

import (
	"context"
	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
)

// InterceptorLogger adapts slog logger to interceptor logger.
// https://github.com/grpc-ecosystem/go-grpc-middleware/blob/4679fb12b6915f8f7a682a525073fe3810d5c64e/interceptors/logging/examples/slog/example_test.go#L15C1-L21C2
func InterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}
