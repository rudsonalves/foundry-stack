package infrastructure

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

const emailVerificationTokenSize = 32

type SecureEmailVerificationTokenGenerator struct{}

var _ domain.EmailVerificationTokenGenerator = (*SecureEmailVerificationTokenGenerator)(nil)

func NewSecureEmailVerificationTokenGenerator() *SecureEmailVerificationTokenGenerator {
	return &SecureEmailVerificationTokenGenerator{}
}

func (g *SecureEmailVerificationTokenGenerator) GenerateToken() (string, error) {
	rawToken := make([]byte, emailVerificationTokenSize)

	if _, err := rand.Read(rawToken); err != nil {
		return "", fmt.Errorf("generate email verification token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(rawToken), nil
}
