package auth

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"server/internal/db"
	"time"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

var (
	ErrInvalidOAuthCode = errors.New("invalid oauth code")
)

type Service struct {
	db        *db.Queries
	googleCfg *oauth2.Config
}

func NewService(
	db *db.Queries,
	googleCfg *oauth2.Config,
) *Service {
	return &Service{
		db:        db,
		googleCfg: googleCfg,
	}
}

type GoogleUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Session struct {
	Token     string
	ExpiresAt time.Time
}

func (s *Service) GetGoogleOAuthURL(state string) string {
	return s.googleCfg.AuthCodeURL(state)
}

func (s *Service) AuthWithGoogle(ctx context.Context, code string) (*Session, error) {
	oauthToken, err := s.googleCfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidOAuthCode, err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://openidconnect.googleapis.com/v1/userinfo", nil)
	if err != nil {
		return nil, fmt.Errorf("build google userinfo request: %w", err)
	}

	resp, err := s.googleCfg.Client(ctx, oauthToken).Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch google userinfo: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch google userinfo: unexpected status %d", resp.StatusCode)
	}

	var info GoogleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode google userinfo: %w", err)
	}
	if info.Email == "" {
		return nil, fmt.Errorf("google userinfo missing email")
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
	expiresAt := time.Now().UTC().Add(sessionTTL)

	err = s.db.CreateSession(ctx, db.CreateSessionParams{
		UserID:    userID,
		TokenHash: tokenHash[:],
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &Session{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}
