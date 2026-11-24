package mysql

import (
	"server/internal/core/domain"
	"server/internal/features/auth/repository/mysql/sqlc"
)

func toDomainOAuthAccount(acc sqlc.OauthAccount) *domain.OAuthAccount {
	return &domain.OAuthAccount{
		ID:             acc.ID,
		UserID:         acc.UserID,
		Provider:       domain.OAuthProvider(acc.Provider),
		ProviderUserID: acc.ProviderUserID,
		CreatedAt:      acc.CreatedAt,
	}
}

func toDomainSession(s sqlc.GetSessionByTokenHashRow) *domain.Session {
	return &domain.Session{
		ID:         s.ID,
		UserID:     s.UserID,
		ExpiresAt:  s.ExpiresAt,
		LastUsedAt: s.LastUsedAt,
		CreatedAt:  s.CreatedAt,
	}
}
