package domain

import (
	"fmt"
	"time"

	coreerrors "server/internal/core/errors"
)

type OAuthProvider string

const (
	OAuthProviderGoogle OAuthProvider = "google"
	OAuthProviderGitHub OAuthProvider = "github"
)

type OAuthAccount struct {
	ID             uint64
	UserID         uint64
	Provider       OAuthProvider
	ProviderUserID string
	CreatedAt      time.Time
}

type OAuthUserInfo struct {
	Sub           string
	Email         string
	EmailVerified bool
	Name          string
}

func (info OAuthUserInfo) Validate() error {
	if info.Sub == "" {
		return fmt.Errorf("%w: oauth userinfo missing sub", coreerrors.ErrInvalidArgument)
	}

	if info.Email == "" {
		return fmt.Errorf("%w: oauth userinfo missing email", coreerrors.ErrInvalidArgument)
	}

	if !info.EmailVerified {
		return fmt.Errorf("%w: oauth userinfo email not verified", coreerrors.ErrInvalidArgument)
	}

	return nil
}
