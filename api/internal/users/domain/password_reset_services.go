package domain

import "github.com/google/uuid"

type PasswordResetCodeGenerator interface {
	GenerateCode() (string, error)
}

type PasswordResetTokenGenerator interface {
	GenerateToken() (string, error)
}

type PasswordResetHasher interface {
	HashCode(
		passwordResetID uuid.UUID,
		userID uuid.UUID,
		code string,
	) []byte

	HashToken(
		passwordResetID uuid.UUID,
		userID uuid.UUID,
		token string,
	) []byte

	MatchesCode(
		hash []byte,
		passwordResetID uuid.UUID,
		userID uuid.UUID,
		code string,
	) bool

	MatchesToken(
		hash []byte,
		passwordResetID uuid.UUID,
		userID uuid.UUID,
		token string,
	) bool
}
