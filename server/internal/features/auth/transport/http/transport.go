package http

import (
	"context"
	"net/http"

	"server/internal/core/domain"
	"server/internal/core/transport/http/middleware"

	"github.com/danielgtaylor/huma/v2"
)

type AuthService interface {
	GetGoogleOAuthURL() string
	AuthWithGoogle(ctx context.Context, code string) (*domain.AuthSession, error)
	GetCurrentUser(ctx context.Context, userID uint64) (*domain.User, error)
}

type Handler struct {
	svc AuthService
}

func NewHandler(svc AuthService) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) Register(api huma.API, authMiddleware *middleware.Auth) {
	huma.Register(api, huma.Operation{
		Method:        http.MethodGet,
		Tags:          []string{"Auth"},
		Path:          "/auth/me",
		DefaultStatus: http.StatusOK,
		Security: []map[string][]string{
			{middleware.BearerAuthScheme: {}},
		},
		Middlewares: huma.Middlewares{authMiddleware.Handle},
	}, h.HandleMe)

	huma.Register(api, huma.Operation{
		Method:        http.MethodGet,
		Tags:          []string{"Auth"},
		Path:          "/auth/google",
		DefaultStatus: http.StatusTemporaryRedirect,
	}, h.HandleGoogle)

	huma.Register(api, huma.Operation{
		Method:        http.MethodGet,
		Tags:          []string{"Auth"},
		Path:          "/auth/google/callback",
		DefaultStatus: http.StatusOK,
	}, h.HandleGoogleCallback)
}
