package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func LoggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		reqID := r.Header.Get("X-Request-ID")
		if reqID == "" {
			reqID = uuid.NewString()
		}

		w.Header().Set("X-Request-ID", reqID)

		ctx := context.WithValue(r.Context(), "request_id", reqID)

		next.ServeHTTP(w, r.WithContext(ctx))

		slog.Info("http request", "request id", reqID, "method", r.Method, "path", r.URL.Path, "duration", time.Since(start))
	})
}
