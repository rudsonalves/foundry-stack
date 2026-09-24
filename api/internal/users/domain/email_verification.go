package domain

import (
	"time"

	"github.com/google/uuid"
)

type EmailVerificationChallengeStatus string

const (
	EmailVerificationChallengePending     EmailVerificationChallengeStatus = "pending"
	EmailVerificationChallengeConfirmed   EmailVerificationChallengeStatus = "confirmed"
	EmailVerificationChallengeConsumed    EmailVerificationChallengeStatus = "consumed"
	EmailVerificationChallengeInvalidated EmailVerificationChallengeStatus = "invalidated"
)

type EmailVerificationChallenge struct {
	VerificationID             uuid.UUID
	Email                      string
	CodeHash                   []byte `json:"-"`
	VerificationTokenHash      []byte `json:"-"`
	FailedAttempts             int
	CreatedAt                  time.Time
	UpdatedAt                  time.Time
	CodeExpiresAt              time.Time
	VerificationTokenExpiresAt *time.Time
	ConfirmedAt                *time.Time
	ConsumedAt                 *time.Time
	InvalidatedAt              *time.Time
}

func (c EmailVerificationChallenge) Status() EmailVerificationChallengeStatus {
	switch {
	case c.InvalidatedAt != nil:
		return EmailVerificationChallengeInvalidated
	case c.ConsumedAt != nil:
		return EmailVerificationChallengeConsumed
	case c.ConfirmedAt != nil:
		return EmailVerificationChallengeConfirmed
	default:
		return EmailVerificationChallengePending
	}
}

func (c EmailVerificationChallenge) IsCodeExpired(now time.Time) bool {
	return !now.UTC().Before(c.CodeExpiresAt.UTC())
}

func (c EmailVerificationChallenge) IsVerificationTokenExpired(
	now time.Time,
) bool {
	if c.VerificationTokenExpiresAt == nil {
		return false
	}

	return !now.UTC().Before(c.VerificationTokenExpiresAt.UTC())
}
