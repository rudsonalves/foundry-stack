package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CreateUserDTO struct {
	Name                  string
	Email                 string
	PasswordHash          string `json:"-"`
	VerificationTokenHash []byte `json:"-"`
	Now                   time.Time
}

type UserRepository interface {
	Create(
		ctx context.Context,
		input CreateUserDTO,
	) (User, error)

	FindByID(
		ctx context.Context,
		id uuid.UUID,
	) (User, error)

	FindByEmail(
		ctx context.Context,
		email string,
	) (User, error)
}
