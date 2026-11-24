package middleware

import (
	"net/http"
	"slices"
)

type Middleware func(http.Handler) http.Handler

func Chain(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, mw := range slices.Backward(middlewares) {
		handler = mw(handler)
	}

	return handler
}
