package infrastructure

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

const emailVerificationCodeLimit = 1_000_000

type SecureEmailVerificationCodeGenerator struct{}

var _ domain.EmailVerificationCodeGenerator = (*SecureEmailVerificationCodeGenerator)(nil)

func NewSecureEmailVerificationCodeGenerator() *SecureEmailVerificationCodeGenerator {
	return &SecureEmailVerificationCodeGenerator{}
}

func (g *SecureEmailVerificationCodeGenerator) GenerateCode() (string, error) {
	value, err := rand.Int(
		rand.Reader,
		big.NewInt(emailVerificationCodeLimit),
	)
	if err != nil {
		return "", fmt.Errorf("generate email verification code: %w", err)
	}

	return formatEmailVerificationCode(value.Int64()), nil
}

func formatEmailVerificationCode(value int64) string {
	return fmt.Sprintf("%06d", value)
}
