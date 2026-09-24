package infrastructure

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

const passwordResetTokenSize = 32

type SecurePasswordResetTokenGenerator struct{}

var _ domain.PasswordResetTokenGenerator = (*SecurePasswordResetTokenGenerator)(nil)

func NewSecurePasswordResetTokenGenerator() *SecurePasswordResetTokenGenerator {
	return &SecurePasswordResetTokenGenerator{}
}

func (g *SecurePasswordResetTokenGenerator) GenerateToken() (string, error) {
	rawToken := make([]byte, passwordResetTokenSize)

	if _, err := rand.Read(rawToken); err != nil {
		return "", fmt.Errorf("generate password reset token: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(rawToken), nil
}
