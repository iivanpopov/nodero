package google

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"server/internal/core/domain"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const userInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"

type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

type Provider struct {
	oauthCfg *oauth2.Config
}

type userInfoResponse struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
}

func NewProvider(cfg Config) *Provider {
	return &Provider{
		oauthCfg: &oauth2.Config{
			Scopes:       []string{"openid", "email", "profile"},
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Endpoint:     google.Endpoint,
		},
	}
}

func (p *Provider) AuthCodeURL() string {
	return p.oauthCfg.AuthCodeURL("")
}

func (p *Provider) FetchUserInfo(ctx context.Context, code string) (*domain.OAuthUserInfo, error) {
	oauthToken, err := p.oauthCfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google token exchange: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build google userinfo request: %w", err)
	}

	resp, err := p.oauthCfg.Client(ctx, oauthToken).Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch google userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch google userinfo: unexpected status %d", resp.StatusCode)
	}

	var info userInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode google userinfo: %w", err)
	}

	return &domain.OAuthUserInfo{
		Sub:           info.Sub,
		Email:         info.Email,
		EmailVerified: info.EmailVerified,
		Name:          info.Name,
	}, nil
}
