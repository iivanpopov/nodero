package v1

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/google/uuid"
)

type OAuthRedirectOutput struct {
	Location  string      `header:"Location"`
	SetCookie http.Cookie `header:"Set-Cookie"`
}

func (h *Handler) HandleGoogle(ctx context.Context, input *struct{}) (*OAuthRedirectOutput, error) {
	state := uuid.NewString()

	url := h.authSvc.GetGoogleOAuthURL(state)

	return &OAuthRedirectOutput{
		Location: url,
		SetCookie: http.Cookie{
			Name:     "oauth_state",
			Value:    state,
			Path:     "/",
			MaxAge:   int((10 * time.Minute).Seconds()),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
	}, nil
}

type GoogleCallbackInput struct {
	SessionToken string `cookie:"session_token" required:"false"`
	Code         string `query:"code" required:"true"`
	State        string `query:"state" required:"true"`
	StateCookie  string `cookie:"oauth_state"`
}

func (h *Handler) HandleGoogleCallback(ctx context.Context, input *GoogleCallbackInput) (*OAuthRedirectOutput, error) {
	if input.StateCookie == "" || input.State != input.StateCookie {
		return nil, huma.Error400BadRequest("invalid state param")
	}

	res, err := h.authSvc.AuthWithGoogle(ctx, input.Code)
	if err != nil {
		return nil, huma.Error401Unauthorized("error authorization via Google", err)
	}

	return &OAuthRedirectOutput{
		Location: res.RedirectURL,
		SetCookie: http.Cookie{
			Name:     "session_token",
			Value:    res.SessionToken,
			Path:     "/",
			MaxAge:   int((7 * 24 * time.Hour).Seconds()),
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		},
	}, nil
}
