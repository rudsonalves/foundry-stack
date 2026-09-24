package domain

import (
	"time"

	"github.com/google/uuid"
)

type Clock interface {
	Now() time.Time
}

type EmailVerificationCodeGenerator interface {
	GenerateCode() (string, error)
}

type EmailVerificationTokenGenerator interface {
	GenerateToken() (string, error)
}

type EmailVerificationHasher interface {
	HashCode(
		verificationID uuid.UUID,
		email string,
		code string,
	) []byte

	HashToken(email string, token string) []byte

	MatchesCode(
		hash []byte,
		verificationID uuid.UUID,
		email string,
		code string,
	) bool

	MatchesToken(hash []byte, email string, token string) bool
}
