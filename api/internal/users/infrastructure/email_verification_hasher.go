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
	emailVerificationCodeContext  = "foundry-stack:email-verification:code"
	emailVerificationTokenContext = "foundry-stack:email-verification:token"
)

type HMACEmailVerificationHasher struct {
	codeSecret  []byte
	tokenSecret []byte
}

var _ domain.EmailVerificationHasher = (*HMACEmailVerificationHasher)(nil)

func NewHMACEmailVerificationHasher(
	codeSecret []byte,
	tokenSecret []byte,
) (*HMACEmailVerificationHasher, error) {
	if bytes.Equal(codeSecret, tokenSecret) {
		return nil, errors.New(
			"email verification code and token secrets must be distinct",
		)
	}

	return &HMACEmailVerificationHasher{
		codeSecret:  append([]byte(nil), codeSecret...),
		tokenSecret: append([]byte(nil), tokenSecret...),
	}, nil
}

func (h *HMACEmailVerificationHasher) HashCode(
	verificationID uuid.UUID,
	email string,
	code string,
) []byte {
	return emailVerificationHash(
		h.codeSecret,
		emailVerificationCodeContext,
		verificationID.String(),
		email,
		code,
	)
}

func (h *HMACEmailVerificationHasher) HashToken(
	email string,
	token string,
) []byte {
	return emailVerificationHash(
		h.tokenSecret,
		emailVerificationTokenContext,
		email,
		token,
	)
}

func (h *HMACEmailVerificationHasher) MatchesCode(
	expectedHash []byte,
	verificationID uuid.UUID,
	email string,
	code string,
) bool {
	actualHash := h.HashCode(
		verificationID,
		email,
		code,
	)

	return hmac.Equal(expectedHash, actualHash)
}

func (h *HMACEmailVerificationHasher) MatchesToken(
	expectedHash []byte,
	email string,
	token string,
) bool {
	actualHash := h.HashToken(email, token)

	return hmac.Equal(expectedHash, actualHash)
}

func emailVerificationHash(
	secret []byte,
	values ...string,
) []byte {
	mac := hmac.New(sha256.New, secret)

	for _, value := range values {
		writeHashValue(mac, value)
	}

	return mac.Sum(nil)
}

func writeHashValue(destination hash.Hash, value string) {
	_, _ = destination.Write([]byte(value))
	_, _ = destination.Write([]byte{0})
}
