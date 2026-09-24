package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SendPasswordResetCodeDTO struct {
	PasswordResetID uuid.UUID
	Email           string
	Code            string `json:"-"`
	ExpiresAt       time.Time
}

type PasswordResetSender interface {
	SendPasswordResetCode(
		ctx context.Context,
		input SendPasswordResetCodeDTO,
	) error
}
