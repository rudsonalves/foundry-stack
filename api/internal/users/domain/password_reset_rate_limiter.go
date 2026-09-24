package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ReservePasswordResetRequestDTO struct {
	Email     string
	ClientIP  string
	CreatedAt time.Time
}

type PasswordResetRateLimiter interface {
	Reserve(
		ctx context.Context,
		input ReservePasswordResetRequestDTO,
	) (uuid.UUID, error)

	Release(
		ctx context.Context,
		reservationID uuid.UUID,
	) error
}
