package service

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"server/internal/core/domain"
	coreerrors "server/internal/core/errors"

	"github.com/google/uuid"
)

func (s *Service) AuthWithGoogle(ctx context.Context, code string) (*domain.AuthSession, error) {
	info, err := s.google.FetchUserInfo(ctx, code)
	if err != nil {
		return nil, err
	}

	if err := info.Validate(); err != nil {
		return nil, err
	}

	token := uuid.NewString()
	tokenHash := sha256.Sum256([]byte(token))
	expiresAt := time.Now().UTC().Add(sessionTTL)

	err = s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		var userID uint64

		oauthAcc, err := s.repo.GetOAuthAccount(txCtx, domain.OAuthProviderGoogle, info.Sub)
		if err == nil {
			userID = oauthAcc.UserID
		} else if errors.Is(err, coreerrors.ErrNotFound) {
			user, err := s.users.GetByEmail(txCtx, info.Email)
			if errors.Is(err, coreerrors.ErrNotFound) {
				id, err := s.users.Create(txCtx, info.Email)
				if err != nil {
					return fmt.Errorf("create user: %w", err)
				}

				userID = id
			} else if err != nil {
				return fmt.Errorf("get user: %w", err)
			} else {
				userID = user.ID
			}

			if err := s.repo.CreateOAuthAccount(txCtx, userID, domain.OAuthProviderGoogle, info.Sub); err != nil {
				return fmt.Errorf("create oauth account: %w", err)
			}
		} else {
			return fmt.Errorf("get oauth account: %w", err)
		}

		if err := s.repo.CreateSession(txCtx, userID, tokenHash[:], expiresAt); err != nil {
			return fmt.Errorf("create session: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &domain.AuthSession{
		Token:     token,
		ExpiresAt: expiresAt,
	}, nil
}
