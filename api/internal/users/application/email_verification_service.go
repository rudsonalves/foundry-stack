package application

import (
	"context"
	"errors"
	"net/mail"
	"strings"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

type EmailVerificationServiceConfig struct {
	CodeTTL            time.Duration
	TokenTTL           time.Duration
	ResendCooldown     time.Duration
	MaxConfirmAttempts int
}

type EmailVerificationService struct {
	users         domain.UserRepository
	verifications domain.EmailVerificationRepository
	codes         domain.EmailVerificationCodeGenerator
	tokens        domain.EmailVerificationTokenGenerator
	hasher        domain.EmailVerificationHasher
	sender        domain.EmailVerificationSender
	clock         domain.Clock
	config        EmailVerificationServiceConfig
	rateLimiter   domain.EmailVerificationRateLimiter
}

type ConfirmEmailVerificationInput struct {
	VerificationID uuid.UUID
	Code           string
}

type ConfirmEmailVerificationResult struct {
	EmailVerificationToken string
	ExpiresAt              time.Time
}

func NewEmailVerificationService(
	users domain.UserRepository,
	verifications domain.EmailVerificationRepository,
	codes domain.EmailVerificationCodeGenerator,
	tokens domain.EmailVerificationTokenGenerator,
	hasher domain.EmailVerificationHasher,
	sender domain.EmailVerificationSender,
	clock domain.Clock,
	config EmailVerificationServiceConfig,
	rateLimiter domain.EmailVerificationRateLimiter,
) *EmailVerificationService {
	return &EmailVerificationService{
		users:         users,
		verifications: verifications,
		codes:         codes,
		tokens:        tokens,
		hasher:        hasher,
		sender:        sender,
		clock:         clock,
		config:        config,
		rateLimiter:   rateLimiter,
	}
}

type RequestEmailVerificationInput struct {
	Email    string
	ClientIP string
}

type RequestEmailVerificationResult struct {
	VerificationID    uuid.UUID
	CodeExpiresAt     time.Time
	ResendAvailableAt time.Time
}

func (s *EmailVerificationService) Request(
	ctx context.Context,
	input RequestEmailVerificationInput,
) (RequestEmailVerificationResult, error) {
	email := normalizeEmail(input.Email)

	violations := validateEmail(email)
	if len(violations) > 0 {
		return RequestEmailVerificationResult{},
			sharederrors.NewValidation(
				"invalid input",
				violations,
			)
	}

	_, err := s.users.FindByEmail(ctx, email)

	switch {
	case err == nil:
		return RequestEmailVerificationResult{},
			&sharederrors.AppError{
				Code:    sharederrors.ErrCodeEmailAlreadyRegistered,
				Message: "email already registered",
				Err:     domain.ErrEmailAlreadyExists,
			}

	case errors.Is(err, domain.ErrUserNotFound):
		// The email is available.

	default:
		return RequestEmailVerificationResult{}, err
	}

	now := s.clock.Now().UTC()
	verificationID := uuid.New()

	reservationID, err := s.rateLimiter.Reserve(
		ctx,
		domain.ReserveEmailVerificationRequestDTO{
			Email:     email,
			ClientIP:  input.ClientIP,
			CreatedAt: now,
		},
	)
	if errors.Is(err, domain.ErrEmailVerificationCooldown) ||
		errors.Is(err, domain.ErrEmailVerificationRateLimit) {
		return RequestEmailVerificationResult{},
			&sharederrors.AppError{
				Code:    sharederrors.ErrCodeRateLimitExceeded,
				Message: "too many email verification requests",
				Err:     err,
			}
	}
	if err != nil {
		return RequestEmailVerificationResult{}, err
	}

	code, err := s.codes.GenerateCode()
	if err != nil {
		cleanupCtx := context.WithoutCancel(ctx)

		if releaseErr := s.rateLimiter.ReleaseEmail(
			cleanupCtx,
			reservationID,
		); releaseErr != nil {
			return RequestEmailVerificationResult{},
				errors.Join(err, releaseErr)
		}

		return RequestEmailVerificationResult{}, err
	}

	challenge, err := s.verifications.Replace(
		ctx,
		domain.ReplaceEmailVerificationChallengeDTO{
			VerificationID: verificationID,
			Email:          email,
			CodeHash: s.hasher.HashCode(
				verificationID,
				email,
				code,
			),
			CodeExpiresAt: now.Add(s.config.CodeTTL),
			CreatedAt:     now,
		},
	)
	if err != nil {
		cleanupCtx := context.WithoutCancel(ctx)

		if releaseErr := s.rateLimiter.ReleaseEmail(
			cleanupCtx,
			reservationID,
		); releaseErr != nil {
			return RequestEmailVerificationResult{},
				errors.Join(err, releaseErr)
		}

		return RequestEmailVerificationResult{}, err
	}

	if err := s.sender.SendVerificationCode(
		ctx,
		domain.SendEmailVerificationCodeDTO{
			VerificationID: challenge.VerificationID,
			Email:          challenge.Email,
			Code:           code,
			ExpiresAt:      challenge.CodeExpiresAt,
		},
	); err != nil {
		cleanupCtx := context.WithoutCancel(ctx)
		causes := []error{err}

		if invalidationErr := s.verifications.Invalidate(
			cleanupCtx,
			challenge.VerificationID,
			s.clock.Now().UTC(),
		); invalidationErr != nil {
			causes = append(causes, invalidationErr)
		}

		if releaseErr := s.rateLimiter.ReleaseEmail(
			cleanupCtx,
			reservationID,
		); releaseErr != nil {
			causes = append(causes, releaseErr)
		}

		return RequestEmailVerificationResult{},
			&sharederrors.AppError{
				Code:    sharederrors.ErrCodeEmailDeliveryUnavailable,
				Message: "email delivery temporarily unavailable",
				Err:     errors.Join(causes...),
			}
	}

	return RequestEmailVerificationResult{
		VerificationID:    challenge.VerificationID,
		CodeExpiresAt:     challenge.CodeExpiresAt,
		ResendAvailableAt: now.Add(s.config.ResendCooldown),
	}, nil
}

func (s *EmailVerificationService) Confirm(
	ctx context.Context,
	input ConfirmEmailVerificationInput,
) (ConfirmEmailVerificationResult, error) {
	violations := validateEmailVerificationConfirmation(input)
	if len(violations) > 0 {
		return ConfirmEmailVerificationResult{},
			sharederrors.NewValidation(
				"invalid input",
				violations,
			)
	}

	challenge, err := s.verifications.FindByID(
		ctx,
		input.VerificationID,
	)
	if errors.Is(err, domain.ErrEmailVerificationNotFound) ||
		errors.Is(err, domain.ErrInvalidEmailVerification) {
		return ConfirmEmailVerificationResult{},
			invalidEmailVerificationError(err)
	}
	if err != nil {
		return ConfirmEmailVerificationResult{}, err
	}

	now := s.clock.Now().UTC()

	if challenge.Status() == domain.EmailVerificationChallengeConsumed ||
		challenge.Status() == domain.EmailVerificationChallengeInvalidated ||
		challenge.IsCodeExpired(now) ||
		challenge.FailedAttempts >= s.config.MaxConfirmAttempts {
		return ConfirmEmailVerificationResult{},
			invalidEmailVerificationError(
				domain.ErrInvalidEmailVerification,
			)
	}

	if !s.hasher.MatchesCode(
		challenge.CodeHash,
		challenge.VerificationID,
		challenge.Email,
		input.Code,
	) {
		_, err := s.verifications.RecordFailedAttempt(
			ctx,
			domain.RecordEmailVerificationFailureDTO{
				VerificationID: challenge.VerificationID,
				FailedAt:       s.clock.Now().UTC(),
				MaxAttempts:    s.config.MaxConfirmAttempts,
			},
		)
		if errors.Is(err, domain.ErrInvalidEmailVerification) ||
			errors.Is(err, domain.ErrEmailVerificationNotFound) {
			return ConfirmEmailVerificationResult{},
				invalidEmailVerificationError(err)
		}
		if err != nil {
			return ConfirmEmailVerificationResult{}, err
		}

		return ConfirmEmailVerificationResult{},
			invalidEmailVerificationError(
				domain.ErrInvalidEmailVerification,
			)
	}

	token, err := s.tokens.GenerateToken()
	if err != nil {
		return ConfirmEmailVerificationResult{}, err
	}

	expiresAt := now.Add(s.config.TokenTTL)

	_, err = s.verifications.Confirm(
		ctx,
		domain.ConfirmEmailVerificationChallengeDTO{
			VerificationID: input.VerificationID,
			VerificationTokenHash: s.hasher.HashToken(
				challenge.Email,
				token,
			),
			VerificationTokenExpiresAt: expiresAt,
			ConfirmedAt:                now,
			MaxAttempts:                s.config.MaxConfirmAttempts,
		},
	)
	if errors.Is(err, domain.ErrInvalidEmailVerification) ||
		errors.Is(err, domain.ErrEmailVerificationNotFound) {
		return ConfirmEmailVerificationResult{},
			invalidEmailVerificationError(err)
	}
	if err != nil {
		return ConfirmEmailVerificationResult{}, err
	}

	return ConfirmEmailVerificationResult{
		EmailVerificationToken: token,
		ExpiresAt:              expiresAt,
	}, nil
}

// normalizeEmail converts the given email to lowercase and trims any
// surrounding whitespace.
func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// validateEmail checks if the given email is valid and returns a list
// of field violations if any.
func validateEmail(
	email string,
) []sharederrors.FieldViolation {
	if email == "" {
		return []sharederrors.FieldViolation{
			{
				Field:   "email",
				Message: "email is required",
			},
		}
	}

	address, err := mail.ParseAddress(email)
	if err != nil || address.Address != email {
		return []sharederrors.FieldViolation{
			{
				Field:   "email",
				Message: "invalid email format",
			},
		}
	}

	return nil
}

// validateEmailVerificationConfirmation checks the validity of the email
// verification confirmation input and returns a list of field violations
// if any.
func validateEmailVerificationConfirmation(
	input ConfirmEmailVerificationInput,
) []sharederrors.FieldViolation {
	violations := make(
		[]sharederrors.FieldViolation,
		0,
		2,
	)

	if input.VerificationID == uuid.Nil {
		violations = append(
			violations,
			sharederrors.FieldViolation{
				Field:   "verification_id",
				Message: "verification ID is required",
			},
		)
	}

	if !isSixDigitCode(input.Code) {
		violations = append(
			violations,
			sharederrors.FieldViolation{
				Field:   "code",
				Message: "code must contain exactly 6 digits",
			},
		)
	}

	return violations
}

// isSixDigitCode checks if the given code consists of exactly six digits.
func isSixDigitCode(code string) bool {
	if len(code) != 6 {
		return false
	}

	for _, character := range code {
		if character < '0' || character > '9' {
			return false
		}
	}

	return true
}

// invalidEmailVerificationError creates a new AppError representing an
// invalid email verification error.
// It takes the underlying cause as an argument and returns a pointer to
// the AppError.
func invalidEmailVerificationError(
	cause error,
) *sharederrors.AppError {
	return &sharederrors.AppError{
		Code:    sharederrors.ErrCodeInvalidEmailVerification,
		Message: "invalid email verification",
		Err:     cause,
	}
}
