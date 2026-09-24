package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ReserveEmailVerificationRequestDTO struct {
	Email     string
	ClientIP  string
	CreatedAt time.Time
}

type EmailVerificationRateLimiter interface {
	Reserve(
		ctx context.Context,
		input ReserveEmailVerificationRequestDTO,
	) (uuid.UUID, error)

	ReleaseEmail(
		ctx context.Context,
		reservationID uuid.UUID,
	) error
}
