package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"server/internal/db"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type Service struct {
	db          *db.Queries
	googleCfg   *oauth2.Config
	githubCfg   *oauth2.Config
	frontendURL string
}

func NewService(
	db *db.Queries,
	googleCfg *oauth2.Config,
	githubCfg *oauth2.Config,
	frontendURL string,
) *Service {
	return &Service{
		db,
		googleCfg,
		githubCfg,
		frontendURL,
	}
}

type GoogleUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type OAuthResult struct {
	RedirectURL  string
	SessionToken string
}

func (s *Service) GetGoogleOAuthURL(state string) string {
	return s.googleCfg.AuthCodeURL(state)
}

func (s *Service) AuthWithGoogle(ctx context.Context, code string) (*OAuthResult, error) {
	oauthToken, err := s.googleCfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("google token exchange %w", err)
	}

	client := s.googleCfg.Client(ctx, oauthToken)
	resp, err := client.Get("https://openidconnect.googleapis.com/v1/userinfo")
	if err != nil {
		return nil, fmt.Errorf("fetch google userinfo: %w", err)
	}
	defer resp.Body.Close()

	var info GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode google userinfo: %w", err)
	}

	const sessionTTL = 30 * 24 * time.Hour

	var userID uint64

	user, err := s.db.GetUserByEmail(ctx, info.Email)
	if errors.Is(err, sql.ErrNoRows) {
		id, err := s.db.CreateUser(ctx, info.Email)
		if err != nil {
			return nil, fmt.Errorf("create user: %w", err)
		}

		userID = uint64(id)
	} else if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	} else {
		userID = user.ID
	}

	token := uuid.NewString()
	tokenHash := sha256.Sum256([]byte(token))

	err = s.db.CreateSession(ctx, db.CreateSessionParams{
		UserID:    userID,
		TokenHash: tokenHash[:],
		ExpiresAt: time.Now().UTC().Add(sessionTTL),
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &OAuthResult{
		RedirectURL:  s.frontendURL,
		SessionToken: token,
	}, nil
}
