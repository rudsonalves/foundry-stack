package domain

import (
	"time"

	"github.com/google/uuid"
)

type PasswordResetChallengeStatus string

const (
	PasswordResetChallengePending     PasswordResetChallengeStatus = "pending"
	PasswordResetChallengeConfirmed   PasswordResetChallengeStatus = "confirmed"
	PasswordResetChallengeConsumed    PasswordResetChallengeStatus = "consumed"
	PasswordResetChallengeInvalidated PasswordResetChallengeStatus = "invalidated"
)

type PasswordResetChallenge struct {
	PasswordResetID             uuid.UUID
	UserID                      uuid.UUID
	CodeHash                    []byte `json:"-"`
	PasswordResetTokenHash      []byte `json:"-"`
	FailedAttempts              int
	CreatedAt                   time.Time
	UpdatedAt                   time.Time
	CodeExpiresAt               time.Time
	PasswordResetTokenExpiresAt *time.Time
	ConfirmedAt                 *time.Time
	ConsumedAt                  *time.Time
	InvalidatedAt               *time.Time
}

func (c PasswordResetChallenge) Status() PasswordResetChallengeStatus {
	switch {
	case c.InvalidatedAt != nil:
		return PasswordResetChallengeInvalidated
	case c.ConsumedAt != nil:
		return PasswordResetChallengeConsumed
	case c.ConfirmedAt != nil:
		return PasswordResetChallengeConfirmed
	default:
		return PasswordResetChallengePending
	}
}

func (c PasswordResetChallenge) IsCodeExpired(now time.Time) bool {
	return !now.UTC().Before(c.CodeExpiresAt.UTC())
}

func (c PasswordResetChallenge) IsPasswordResetTokenExpired(
	now time.Time,
) bool {
	if c.PasswordResetTokenExpiresAt == nil {
		return false
	}

	return !now.UTC().Before(c.PasswordResetTokenExpiresAt.UTC())
}
