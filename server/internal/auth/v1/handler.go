package v1

import (
	"net/http"
	"server/internal/auth"

	"github.com/danielgtaylor/huma/v2"
)

type Handler struct {
	authSvc *auth.Service
}

func NewHandler(authSvc *auth.Service) *Handler {
	return &Handler{
		authSvc: authSvc,
	}
}

func (h *Handler) Register(api huma.API) {
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
		DefaultStatus: http.StatusTemporaryRedirect,
	}, h.HandleGoogleCallback)
}
