package middleware

import (
	"net/http"
	"runtime/debug"

	"server/internal/core/logger"
)

func Recover() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					logger.FromContext(r.Context()).ErrorContext(
						r.Context(),
						"panic recovered",
						"panic", rec,
						"stack", string(debug.Stack()),
					)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
