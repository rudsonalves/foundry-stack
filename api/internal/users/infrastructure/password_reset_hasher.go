package infrastructure

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"hash"

	"github.com/google/uuid"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

const (
	passwordResetCodeContext  = "foundry-stack:password-reset:code"
	passwordResetTokenContext = "foundry-stack:password-reset:token"
)

type HMACPasswordResetHasher struct {
	codeSecret  []byte
	tokenSecret []byte
}

var _ domain.PasswordResetHasher = (*HMACPasswordResetHasher)(nil)

func NewHMACPasswordResetHasher(
	codeSecret []byte,
	tokenSecret []byte,
) (*HMACPasswordResetHasher, error) {
	if bytes.Equal(codeSecret, tokenSecret) {
		return nil, errors.New(
			"password reset code and token secrets must be distinct",
		)
	}

	return &HMACPasswordResetHasher{
		codeSecret:  append([]byte(nil), codeSecret...),
		tokenSecret: append([]byte(nil), tokenSecret...),
	}, nil
}

func (h *HMACPasswordResetHasher) HashCode(
	passwordResetID uuid.UUID,
	userID uuid.UUID,
	code string,
) []byte {
	return passwordResetHash(
		h.codeSecret,
		passwordResetCodeContext,
		passwordResetID.String(),
		userID.String(),
		code,
	)
}

func (h *HMACPasswordResetHasher) HashToken(
	passwordResetID uuid.UUID,
	userID uuid.UUID,
	token string,
) []byte {
	return passwordResetHash(
		h.tokenSecret,
		passwordResetTokenContext,
		passwordResetID.String(),
		userID.String(),
		token,
	)
}

func (h *HMACPasswordResetHasher) MatchesCode(
	expectedHash []byte,
	passwordResetID uuid.UUID,
	userID uuid.UUID,
	code string,
) bool {
	actualHash := h.HashCode(passwordResetID, userID, code)

	return hmac.Equal(expectedHash, actualHash)
}

func (h *HMACPasswordResetHasher) MatchesToken(
	expectedHash []byte,
	passwordResetID uuid.UUID,
	userID uuid.UUID,
	token string,
) bool {
	actualHash := h.HashToken(passwordResetID, userID, token)

	return hmac.Equal(expectedHash, actualHash)
}

func passwordResetHash(secret []byte, values ...string) []byte {
	mac := hmac.New(sha256.New, secret)

	for _, value := range values {
		writePasswordResetHashValue(mac, value)
	}

	return mac.Sum(nil)
}

func writePasswordResetHashValue(destination hash.Hash, value string) {
	_, _ = destination.Write([]byte(value))
	_, _ = destination.Write([]byte{0})
}
