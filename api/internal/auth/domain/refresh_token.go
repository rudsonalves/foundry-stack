package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	UserID    uuid.UUID
	ClientID  string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

type CreateRefreshTokenDTO struct {
	TokenHash []byte
	UserID    uuid.UUID
	ClientID  string
	ExpiresAt time.Time
}

type RefreshTokenRepository interface {
	Create(
		ctx context.Context,
		input CreateRefreshTokenDTO,
	) error

	FindActive(
		ctx context.Context,
		tokenHash []byte,
		now time.Time,
	) (RefreshToken, error)

	Revoke(
		ctx context.Context,
		tokenHash []byte,
		revokedAt time.Time,
	) error
}
