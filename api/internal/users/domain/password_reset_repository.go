package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ReplacePasswordResetChallengeDTO struct {
	PasswordResetID uuid.UUID
	UserID          uuid.UUID
	CodeHash        []byte `json:"-"`
	CodeExpiresAt   time.Time
	CreatedAt       time.Time
}

type ConfirmPasswordResetChallengeDTO struct {
	PasswordResetID             uuid.UUID
	PasswordResetTokenHash      []byte `json:"-"`
	PasswordResetTokenExpiresAt time.Time
	ConfirmedAt                 time.Time
	MaxAttempts                 int
}

type RecordPasswordResetFailureDTO struct {
	PasswordResetID uuid.UUID
	FailedAt        time.Time
	MaxAttempts     int
}

type CompletePasswordResetDTO struct {
	PasswordResetID        uuid.UUID
	PasswordResetTokenHash []byte `json:"-"`
	PasswordHash           string `json:"-"`
	CompletedAt            time.Time
}

type InvalidatePasswordResetChallengeDTO struct {
	PasswordResetID uuid.UUID
	InvalidatedAt   time.Time
}

type PasswordResetRepository interface {
	Replace(
		ctx context.Context,
		input ReplacePasswordResetChallengeDTO,
	) (PasswordResetChallenge, error)

	FindByID(
		ctx context.Context,
		passwordResetID uuid.UUID,
	) (PasswordResetChallenge, error)

	RecordFailedAttempt(
		ctx context.Context,
		input RecordPasswordResetFailureDTO,
	) (PasswordResetChallenge, error)

	Confirm(
		ctx context.Context,
		input ConfirmPasswordResetChallengeDTO,
	) (PasswordResetChallenge, error)

	Invalidate(
		ctx context.Context,
		input InvalidatePasswordResetChallengeDTO,
	) error

	DeleteStale(
		ctx context.Context,
		before time.Time,
	) (int64, error)
}

type PasswordResetCompleter interface {
	Complete(
		ctx context.Context,
		input CompletePasswordResetDTO,
	) error
}
