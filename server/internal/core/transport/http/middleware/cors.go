package middleware

import (
	"net/http"
	"strings"
)

func CORS(allowedOrigins []string) Middleware {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	allowAll := false

	for _, origin := range allowedOrigins {
		normalized := strings.TrimRight(strings.TrimSpace(origin), "/")
		if normalized == "" {
			continue
		}
		if normalized == "*" {
			allowAll = true
			continue
		}
		allowed[normalized] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := strings.TrimRight(r.Header.Get("Origin"), "/")
			if origin != "" {
				w.Header().Add("Vary", "Origin")

				if _, ok := allowed[origin]; ok {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					setCORSHeaders(w)
				} else if allowAll {
					w.Header().Set("Access-Control-Allow-Origin", "*")
					setCORSHeaders(w)
				}
			}

			if r.Method == http.MethodOptions {
				w.Header().Add("Vary", "Access-Control-Request-Method")
				w.Header().Add("Vary", "Access-Control-Request-Headers")
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Request-ID")
	w.Header().Set("Access-Control-Expose-Headers", "X-Request-ID")
}
