package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	coreerrors "server/internal/core/errors"
	"server/internal/core/logger"

	"github.com/danielgtaylor/huma/v2"
)

const BearerAuthScheme = "bearerAuth"

type userIDContextKey struct{}

func UserIDFromContext(ctx context.Context) (uint64, bool) {
	id, ok := ctx.Value(userIDContextKey{}).(uint64)
	return id, ok
}

type SessionAuthenticator interface {
	AuthenticateSession(ctx context.Context, token string) (uint64, error)
}

type Auth struct {
	api  huma.API
	auth SessionAuthenticator
}

func NewAuth(api huma.API, auth SessionAuthenticator) *Auth {
	return &Auth{
		api:  api,
		auth: auth,
	}
}

func (m *Auth) Handle(ctx huma.Context, next func(huma.Context)) {
	token, ok := extractBearerToken(ctx.Header("Authorization"))
	if !ok {
		_ = huma.WriteErr(m.api, ctx, http.StatusUnauthorized, "missing or invalid authorization header")
		return
	}

	userID, err := m.auth.AuthenticateSession(ctx.Context(), token)
	if err != nil {
		if errors.Is(err, coreerrors.ErrUnauthorized) {
			_ = huma.WriteErr(m.api, ctx, http.StatusUnauthorized, "invalid or expired session")
			return
		}

		logger.FromContext(ctx.Context()).ErrorContext(ctx.Context(), "failed to authenticate session", "err", err)
		_ = huma.WriteErr(m.api, ctx, http.StatusInternalServerError, "unexpected error occurred")
		return
	}

	ctx = huma.WithValue(ctx, userIDContextKey{}, userID)
	next(ctx)
}

func extractBearerToken(header string) (string, bool) {
	scheme, token, ok := strings.Cut(strings.TrimSpace(header), " ")
	if !ok || !strings.EqualFold(scheme, "Bearer") {
		return "", false
	}

	token = strings.TrimSpace(token)
	if token == "" || strings.ContainsAny(token, " \t") {
		return "", false
	}

	return token, true
}
