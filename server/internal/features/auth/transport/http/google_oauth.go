package http

import "context"

type OAuthRedirectOutput struct {
	Location string `header:"Location"`
}

func (h *Handler) HandleGoogle(_ context.Context, _ *struct{}) (*OAuthRedirectOutput, error) {
	return &OAuthRedirectOutput{
		Location: h.svc.GetGoogleOAuthURL(),
	}, nil
}
