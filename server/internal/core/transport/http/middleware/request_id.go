package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-ID"

type requestIDContextKey struct{}

func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDContextKey{}).(string)
	return id
}

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get(RequestIDHeader)
			if reqID == "" {
				reqID = uuid.NewString()
			}

			w.Header().Set(RequestIDHeader, reqID)

			ctx := context.WithValue(r.Context(), requestIDContextKey{}, reqID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
