package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SendEmailVerificationCodeDTO struct {
	VerificationID uuid.UUID
	Email          string
	Code           string `json:"-"`
	ExpiresAt      time.Time
}

type EmailVerificationSender interface {
	SendVerificationCode(
		ctx context.Context,
		input SendEmailVerificationCodeDTO,
	) error
}
