package middleware

import (
	"log/slog"
	"net/http"

	"server/internal/core/logger"
)

func Logger(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqLog := log
			if reqID := RequestIDFromContext(r.Context()); reqID != "" {
				reqLog = reqLog.With("request_id", reqID)
			}

			ctx := logger.ToContext(r.Context(), reqLog)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
