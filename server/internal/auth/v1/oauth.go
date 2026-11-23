package v1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"server/internal/auth"
	"time"

	"github.com/google/uuid"
)

type OAuthRedirectOutput struct {
	Location  string        `header:"Location"`
	SetCookie []http.Cookie `header:"Set-Cookie"`
}

type OAuthCallbackInput struct {
	SessionToken string `cookie:"session_token" required:"false"`
	Code         string `query:"code"`
	State        string `query:"state"`
	OAuthError   string `query:"error"`
	StateCookie  string `cookie:"oauth_state"`
}

func clearOAuthStateCookie() http.Cookie {
	return http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
}

func (h *Handler) oauthErrorRedirect(errCode string) *OAuthRedirectOutput {
	return &OAuthRedirectOutput{
		Location:  "/?error=" + url.QueryEscape(errCode),
		SetCookie: []http.Cookie{clearOAuthStateCookie()},
	}
}

func (h *Handler) HandleGoogle(ctx context.Context, _ *struct{}) (*OAuthRedirectOutput, error) {
	state := uuid.NewString()

	return &OAuthRedirectOutput{
		Location: h.authSvc.GetGoogleOAuthURL(state),
		SetCookie: []http.Cookie{
			{
				Name:     "oauth_state",
				Value:    state,
				Path:     "/",
				MaxAge:   int((10 * time.Minute).Seconds()),
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			},
		},
	}, nil
}

func (h *Handler) HandleGoogleCallback(ctx context.Context, input *OAuthCallbackInput) (*OAuthRedirectOutput, error) {
	if input.OAuthError != "" {
		slog.WarnContext(ctx, "google oauth callback returned error", "oauth_error", input.OAuthError)
		return h.oauthErrorRedirect("oauth_failed"), nil
	}

	if input.StateCookie == "" || input.State == "" || input.State != input.StateCookie {
		slog.WarnContext(ctx, "google oauth state mismatch")
		return h.oauthErrorRedirect("oauth_failed"), nil
	}

	if input.Code == "" {
		slog.WarnContext(ctx, "google oauth missing code")
		return h.oauthErrorRedirect("oauth_failed"), nil
	}

	session, err := h.authSvc.AuthWithGoogle(ctx, input.Code)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidOAuthCode) {
			slog.WarnContext(ctx, "invalid google oauth code", "err", err)
		} else {
			slog.ErrorContext(ctx, "google auth failed", "err", err)
		}
		return h.oauthErrorRedirect("oauth_failed"), nil
	}

	return &OAuthRedirectOutput{
		Location: "/",
		SetCookie: []http.Cookie{
			clearOAuthStateCookie(),
			{
				Name:     "session_token",
				Value:    session.Token,
				Path:     "/",
				Expires:  session.ExpiresAt,
				MaxAge:   int(time.Until(session.ExpiresAt).Seconds()),
				HttpOnly: true,
				SameSite: http.SameSiteLaxMode,
			},
		},
	}, nil
}
