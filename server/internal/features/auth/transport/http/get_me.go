package http

import (
	"context"
	"errors"
	"time"

	coreerrors "server/internal/core/errors"
	"server/internal/core/logger"
	"server/internal/core/transport/http/middleware"

	"github.com/danielgtaylor/huma/v2"
)

type MeBody struct {
	ID        uint64    `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MeOutput struct {
	Body MeBody
}

func (h *Handler) HandleMe(ctx context.Context, _ *struct{}) (*MeOutput, error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, huma.Error401Unauthorized("unauthorized")
	}

	user, err := h.svc.GetCurrentUser(ctx, userID)
	if err != nil {
		if errors.Is(err, coreerrors.ErrNotFound) {
			return nil, huma.Error401Unauthorized("invalid or expired session")
		}

		logger.FromContext(ctx).ErrorContext(ctx, "failed to get current user", "user_id", userID, "err", err)
		return nil, huma.Error500InternalServerError("unexpected error occurred")
	}

	return &MeOutput{
		Body: MeBody{
			ID:        user.ID,
			Email:     user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		},
	}, nil
}
