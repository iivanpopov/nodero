package http

import (
	"context"
	"errors"
	"time"

	coreerrors "server/internal/core/errors"
	"server/internal/core/logger"

	"github.com/danielgtaylor/huma/v2"
)

type OAuthCallbackInput struct {
	Code string `query:"code"`
}

type AuthSessionBody struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

type AuthSessionOutput struct {
	Body AuthSessionBody
}

func (h *Handler) HandleGoogleCallback(ctx context.Context, input *OAuthCallbackInput) (*AuthSessionOutput, error) {
	log := logger.FromContext(ctx)

	if input.Code == "" {
		log.WarnContext(ctx, "google oauth missing code")
		return nil, huma.Error400BadRequest("missing authorization code")
	}

	session, err := h.svc.AuthWithGoogle(ctx, input.Code)
	if err != nil {
		if errors.Is(err, coreerrors.ErrInvalidArgument) {
			log.WarnContext(ctx, "google auth invalid user info", "err", err)
			return nil, huma.Error400BadRequest("invalid oauth user info")
		}

		log.ErrorContext(ctx, "google auth failed", "err", err)
		return nil, huma.Error500InternalServerError("unexpected error occurred")
	}

	return &AuthSessionOutput{
		Body: AuthSessionBody{
			Token:     session.Token,
			ExpiresAt: session.ExpiresAt,
		},
	}, nil
}
