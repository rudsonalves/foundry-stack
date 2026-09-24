package infrastructure

import (
	"context"
	"log"
	"time"

	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

type LoggingEmailVerificationSender struct {
	next   domain.EmailVerificationSender
	logger *log.Logger
}

var _ domain.EmailVerificationSender = (*LoggingEmailVerificationSender)(nil)

func NewLoggingEmailVerificationSender(
	next domain.EmailVerificationSender,
	logger *log.Logger,
) *LoggingEmailVerificationSender {
	return &LoggingEmailVerificationSender{
		next:   next,
		logger: logger,
	}
}

func (s *LoggingEmailVerificationSender) SendVerificationCode(
	ctx context.Context,
	input domain.SendEmailVerificationCodeDTO,
) error {
	s.logger.Printf(
		"email verification code generated: verification_id=%s code=%s expires_at=%s",
		input.VerificationID,
		input.Code,
		input.ExpiresAt.UTC().Format(time.RFC3339),
	)

	return s.next.SendVerificationCode(ctx, input)
}
