package infrastructure

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

const passwordResetCodeLimit = 1_000_000

type SecurePasswordResetCodeGenerator struct{}

var _ domain.PasswordResetCodeGenerator = (*SecurePasswordResetCodeGenerator)(nil)

func NewSecurePasswordResetCodeGenerator() *SecurePasswordResetCodeGenerator {
	return &SecurePasswordResetCodeGenerator{}
}

func (g *SecurePasswordResetCodeGenerator) GenerateCode() (string, error) {
	value, err := rand.Int(
		rand.Reader,
		big.NewInt(passwordResetCodeLimit),
	)
	if err != nil {
		return "", fmt.Errorf("generate password reset code: %w", err)
	}

	return formatPasswordResetCode(value.Int64()), nil
}

func formatPasswordResetCode(value int64) string {
	return fmt.Sprintf("%06d", value)
}
