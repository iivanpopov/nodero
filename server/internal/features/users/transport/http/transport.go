package http

import (
	"context"

	"server/internal/core/domain"

	"github.com/danielgtaylor/huma/v2"
)

type UserService interface {
	GetByID(ctx context.Context, userID uint64) (*domain.User, error)
}

type Handler struct {
	svc UserService
}

func NewHandler(svc UserService) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) Register(_ huma.API) {
}
