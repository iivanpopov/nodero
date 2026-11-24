package service

import (
	"context"
	"time"

	"server/internal/core/domain"
)

const sessionTTL = 30 * 24 * time.Hour

type TxManager interface {
	WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type Repository interface {
	GetOAuthAccount(ctx context.Context, provider domain.OAuthProvider, providerUserID string) (*domain.OAuthAccount, error)
	CreateOAuthAccount(ctx context.Context, userID uint64, provider domain.OAuthProvider, providerUserID string) error
	CreateSession(ctx context.Context, userID uint64, tokenHash []byte, expiresAt time.Time) error
	GetSessionByTokenHash(ctx context.Context, tokenHash []byte) (*domain.Session, error)
	UpdateSession(ctx context.Context, sessionID uint64, expiresAt time.Time) error
}

type UserService interface {
	GetByID(ctx context.Context, userID uint64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, email string) (uint64, error)
}

type OAuthProvider interface {
	AuthCodeURL() string
	FetchUserInfo(ctx context.Context, code string) (*domain.OAuthUserInfo, error)
}

type Service struct {
	txManager TxManager
	repo      Repository
	users     UserService
	google    OAuthProvider
}

func NewService(
	txManager TxManager,
	repo Repository,
	users UserService,
	google OAuthProvider,
) *Service {
	return &Service{
		txManager: txManager,
		repo:      repo,
		users:     users,
		google:    google,
	}
}
