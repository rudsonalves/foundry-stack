package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ReplaceEmailVerificationChallengeDTO struct {
	VerificationID uuid.UUID
	Email          string
	CodeHash       []byte `json:"-"`
	CodeExpiresAt  time.Time
	CreatedAt      time.Time
}

type ConfirmEmailVerificationChallengeDTO struct {
	VerificationID             uuid.UUID
	VerificationTokenHash      []byte `json:"-"`
	VerificationTokenExpiresAt time.Time
	ConfirmedAt                time.Time
	MaxAttempts                int
}

type RecordEmailVerificationFailureDTO struct {
	VerificationID uuid.UUID
	FailedAt       time.Time
	MaxAttempts    int
}

type EmailVerificationRepository interface {
	Replace(
		ctx context.Context,
		input ReplaceEmailVerificationChallengeDTO,
	) (EmailVerificationChallenge, error)

	FindByID(
		ctx context.Context,
		verificationID uuid.UUID,
	) (EmailVerificationChallenge, error)

	RecordFailedAttempt(
		ctx context.Context,
		input RecordEmailVerificationFailureDTO,
	) (EmailVerificationChallenge, error)

	Confirm(
		ctx context.Context,
		input ConfirmEmailVerificationChallengeDTO,
	) (EmailVerificationChallenge, error)

	DeleteStale(
		ctx context.Context,
		before time.Time,
	) (int64, error)

	Invalidate(
		ctx context.Context,
		verificationID uuid.UUID,
		invalidatedAt time.Time,
	) error
}
