package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	coreerrors "server/internal/core/errors"
)

func (s *Service) AuthenticateSession(ctx context.Context, token string) (uint64, error) {
	tokenHash := sha256.Sum256([]byte(token))

	session, err := s.repo.GetSessionByTokenHash(ctx, tokenHash[:])
	if err != nil {
		if errors.Is(err, coreerrors.ErrNotFound) {
			return 0, coreerrors.ErrUnauthorized
		}

		return 0, fmt.Errorf("get session: %w", err)
	}

	if err := s.repo.UpdateSession(ctx, session.ID, time.Now().UTC().Add(sessionTTL)); err != nil {
		return 0, fmt.Errorf("prolong session: %w", err)
	}

	return session.UserID, nil
}
