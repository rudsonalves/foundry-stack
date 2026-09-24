package application

import (
	"context"
	"errors"
	"strings"

	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

type Service struct {
	users                   domain.UserRepository
	passwords               domain.PasswordHasher
	emailVerificationHasher domain.EmailVerificationHasher
	clock                   domain.Clock
}

func NewService(
	users domain.UserRepository,
	passwords domain.PasswordHasher,
	emailVerificationHasher domain.EmailVerificationHasher,
	clock domain.Clock,
) *Service {
	return &Service{
		users:                   users,
		passwords:               passwords,
		emailVerificationHasher: emailVerificationHasher,
		clock:                   clock,
	}
}

var ErrInvalidCredentials = errors.New(
	"invalid credentials",
)

func (s *Service) Authenticate(
	ctx context.Context,
	email string,
	password string,
) (domain.User, error) {
	email = normalizeEmail(email)

	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrUserNotFound) {
		return domain.User{}, ErrInvalidCredentials
	}
	if err != nil {
		return domain.User{}, err
	}

	err = s.passwords.Compare(user.PasswordHash, password)
	if err != nil {
		return domain.User{}, ErrInvalidCredentials
	}

	return user, nil
}

type RegisterInput struct {
	Name                   string
	Email                  string
	Password               string
	EmailVerificationToken string
}

func (s *Service) Register(
	ctx context.Context,
	input RegisterInput,
) (domain.User, error) {
	name := strings.TrimSpace(input.Name)
	email := normalizeEmail(input.Email)

	violations := validateRegistration(
		name,
		email,
		input.Password,
	)
	if len(violations) > 0 {
		return domain.User{}, sharederrors.NewValidation(
			"invalid input",
			violations,
		)
	}
	if strings.TrimSpace(input.EmailVerificationToken) == "" {
		return domain.User{}, invalidEmailVerificationError(
			domain.ErrInvalidEmailVerification,
		)
	}

	verificationTokenHash := s.emailVerificationHasher.HashToken(
		email,
		input.EmailVerificationToken,
	)

	passwordHash, err := s.passwords.Hash(input.Password)
	if err != nil {
		return domain.User{}, err
	}

	user, err := s.users.Create(ctx, domain.CreateUserDTO{
		Name:                  name,
		Email:                 email,
		PasswordHash:          passwordHash,
		VerificationTokenHash: verificationTokenHash,
		Now:                   s.clock.Now().UTC(),
	})
	if errors.Is(err, domain.ErrEmailAlreadyExists) {
		return domain.User{}, &sharederrors.AppError{
			Code:    sharederrors.ErrCodeEmailAlreadyRegistered,
			Message: "email already registered",
			Err:     err,
		}
	}
	if errors.Is(err, domain.ErrInvalidEmailVerification) {
		return domain.User{}, invalidEmailVerificationError(err)
	}
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func validateRegistration(
	name string,
	email string,
	password string,
) []sharederrors.FieldViolation {
	violations := []sharederrors.FieldViolation{}

	if name == "" {
		violations = append(
			violations,
			sharederrors.FieldViolation{
				Field:   "name",
				Message: "name is required",
			},
		)
	}

	violations = append(
		violations,
		validateEmail(email)...,
	)

	violations = append(violations, validatePassword(password)...)

	return violations
}

func validatePassword(password string) []sharederrors.FieldViolation {
	violations := make([]sharederrors.FieldViolation, 0, 1)
	if len(password) < 8 {
		violations = append(
			violations,
			sharederrors.FieldViolation{
				Field:   "password",
				Message: "password must be at least 8 characters",
			},
		)
	} else if len(password) > 72 {
		violations = append(
			violations,
			sharederrors.FieldViolation{
				Field:   "password",
				Message: "password must be at most 72 characters",
			},
		)
	}

	return violations
}
