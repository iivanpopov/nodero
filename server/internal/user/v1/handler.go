package v1

import (
	"server/internal/user"

	"github.com/danielgtaylor/huma/v2"
)

type Handler struct {
	svc *user.Service
}

func NewHandler(svc *user.Service) *Handler {
	return &Handler{svc}
}

func (h *Handler) Register(api huma.API) {

}
