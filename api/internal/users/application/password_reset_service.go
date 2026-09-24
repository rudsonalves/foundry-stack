package application

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

type PasswordResetServiceConfig struct {
	CodeTTL            time.Duration
	TokenTTL           time.Duration
	ResendCooldown     time.Duration
	MaxConfirmAttempts int
}

type PasswordResetService struct {
	users       domain.UserRepository
	resets      domain.PasswordResetRepository
	codes       domain.PasswordResetCodeGenerator
	tokens      domain.PasswordResetTokenGenerator
	hasher      domain.PasswordResetHasher
	sender      domain.PasswordResetSender
	clock       domain.Clock
	config      PasswordResetServiceConfig
	rateLimiter domain.PasswordResetRateLimiter
	logger      *log.Logger
	completer   domain.PasswordResetCompleter
	passwords   domain.PasswordHasher
}

type RequestPasswordResetInput struct{ Email, ClientIP string }
type RequestPasswordResetResult struct {
	PasswordResetID   uuid.UUID
	CodeExpiresAt     time.Time
	ResendAvailableAt time.Time
}
type ConfirmPasswordResetInput struct {
	PasswordResetID uuid.UUID
	Code            string
}
type ConfirmPasswordResetResult struct {
	PasswordResetToken string
	ExpiresAt          time.Time
}

type PasswordResetCompletionDependencies struct {
	Completer domain.PasswordResetCompleter
	Passwords domain.PasswordHasher
}

type CompletePasswordResetInput struct {
	PasswordResetID    uuid.UUID
	PasswordResetToken string
	NewPassword        string
}

func NewPasswordResetService(users domain.UserRepository, resets domain.PasswordResetRepository, codes domain.PasswordResetCodeGenerator, tokens domain.PasswordResetTokenGenerator, hasher domain.PasswordResetHasher, sender domain.PasswordResetSender, clock domain.Clock, config PasswordResetServiceConfig, rateLimiter domain.PasswordResetRateLimiter, logger *log.Logger, completion ...PasswordResetCompletionDependencies) *PasswordResetService {
	service := &PasswordResetService{users: users, resets: resets, codes: codes, tokens: tokens, hasher: hasher, sender: sender, clock: clock, config: config, rateLimiter: rateLimiter, logger: logger}
	if len(completion) > 0 {
		service.completer = completion[0].Completer
		service.passwords = completion[0].Passwords
	}
	return service
}

func (s *PasswordResetService) Request(ctx context.Context, input RequestPasswordResetInput) (RequestPasswordResetResult, error) {
	email := normalizeEmail(input.Email)
	violations := validateEmail(email)
	if len(email) > 254 {
		violations = append(violations, sharederrors.FieldViolation{Field: "email", Message: "email must contain at most 254 characters"})
	}
	if len(violations) > 0 {
		return RequestPasswordResetResult{}, sharederrors.NewValidation("invalid input", violations)
	}
	now := s.clock.Now().UTC()
	reservationID, err := s.rateLimiter.Reserve(ctx, domain.ReservePasswordResetRequestDTO{Email: email, ClientIP: input.ClientIP, CreatedAt: now})
	if errors.Is(err, domain.ErrPasswordResetCooldown) || errors.Is(err, domain.ErrPasswordResetRateLimit) {
		return RequestPasswordResetResult{}, &sharederrors.AppError{Code: sharederrors.ErrCodeRateLimitExceeded, Message: "too many password reset requests", Err: err}
	}
	if err != nil {
		return RequestPasswordResetResult{}, err
	}

	resetID := uuid.New()
	result := RequestPasswordResetResult{PasswordResetID: resetID, CodeExpiresAt: now.Add(s.config.CodeTTL), ResendAvailableAt: now.Add(s.config.ResendCooldown)}
	user, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, domain.ErrUserNotFound) {
		code, generationErr := s.codes.GenerateCode()
		if generationErr != nil {
			return RequestPasswordResetResult{}, s.release(ctx, reservationID, generationErr)
		}
		_ = s.hasher.HashCode(resetID, uuid.Nil, code)
		return result, nil
	}
	if err != nil {
		return RequestPasswordResetResult{}, s.release(ctx, reservationID, err)
	}
	code, err := s.codes.GenerateCode()
	if err != nil {
		return RequestPasswordResetResult{}, s.release(ctx, reservationID, err)
	}
	challenge, err := s.resets.Replace(ctx, domain.ReplacePasswordResetChallengeDTO{PasswordResetID: resetID, UserID: user.ID, CodeHash: s.hasher.HashCode(resetID, user.ID, code), CodeExpiresAt: result.CodeExpiresAt, CreatedAt: now})
	if err != nil {
		return RequestPasswordResetResult{}, s.release(ctx, reservationID, err)
	}
	if err := s.sender.SendPasswordResetCode(ctx, domain.SendPasswordResetCodeDTO{PasswordResetID: resetID, Email: email, Code: code, ExpiresAt: result.CodeExpiresAt}); err != nil {
		cleanup := context.WithoutCancel(ctx)
		invalidateErr := s.resets.Invalidate(cleanup, domain.InvalidatePasswordResetChallengeDTO{PasswordResetID: challenge.PasswordResetID, InvalidatedAt: s.clock.Now().UTC()})
		if s.logger != nil {
			s.logger.Printf("password reset delivery failed: password_reset_id=%s challenge_invalidated=%t", resetID, invalidateErr == nil)
		}
		return result, nil
	}
	return result, nil
}

func (s *PasswordResetService) release(ctx context.Context, id uuid.UUID, cause error) error {
	if err := s.rateLimiter.Release(context.WithoutCancel(ctx), id); err != nil {
		return errors.Join(cause, err)
	}
	return cause
}

func (s *PasswordResetService) Complete(ctx context.Context, input CompletePasswordResetInput) error {
	violations := make([]sharederrors.FieldViolation, 0, 3)
	if input.PasswordResetID == uuid.Nil {
		violations = append(violations, sharederrors.FieldViolation{Field: "password_reset_id", Message: "password reset ID is required"})
	}
	if input.PasswordResetToken == "" {
		violations = append(violations, sharederrors.FieldViolation{Field: "password_reset_token", Message: "password reset token is required"})
	}
	violations = append(violations, validatePassword(input.NewPassword)...)
	if len(violations) > 0 {
		return sharederrors.NewValidation("invalid input", violations)
	}
	if s.completer == nil || s.passwords == nil {
		return errors.New("password reset completion is not configured")
	}

	challenge, err := s.resets.FindByID(ctx, input.PasswordResetID)
	if errors.Is(err, domain.ErrPasswordResetNotFound) || errors.Is(err, domain.ErrInvalidPasswordReset) {
		return invalidPasswordResetError(err)
	}
	if err != nil {
		return err
	}
	now := s.clock.Now().UTC()
	if challenge.Status() != domain.PasswordResetChallengeConfirmed || challenge.IsPasswordResetTokenExpired(now) {
		return invalidPasswordResetError(domain.ErrInvalidPasswordReset)
	}

	passwordHash, err := s.passwords.Hash(input.NewPassword)
	if err != nil {
		return err
	}
	err = s.completer.Complete(ctx, domain.CompletePasswordResetDTO{
		PasswordResetID:        input.PasswordResetID,
		PasswordResetTokenHash: s.hasher.HashToken(challenge.PasswordResetID, challenge.UserID, input.PasswordResetToken),
		PasswordHash:           passwordHash,
		CompletedAt:            now,
	})
	if errors.Is(err, domain.ErrInvalidPasswordReset) || errors.Is(err, domain.ErrPasswordResetNotFound) {
		return invalidPasswordResetError(err)
	}
	return err
}

func (s *PasswordResetService) Confirm(ctx context.Context, input ConfirmPasswordResetInput) (ConfirmPasswordResetResult, error) {
	var violations []sharederrors.FieldViolation
	if input.PasswordResetID == uuid.Nil {
		violations = append(violations, sharederrors.FieldViolation{Field: "password_reset_id", Message: "password reset ID is required"})
	}
	if !isSixDigitCode(input.Code) {
		violations = append(violations, sharederrors.FieldViolation{Field: "code", Message: "code must contain exactly 6 digits"})
	}
	if len(violations) > 0 {
		return ConfirmPasswordResetResult{}, sharederrors.NewValidation("invalid input", violations)
	}
	challenge, err := s.resets.FindByID(ctx, input.PasswordResetID)
	if errors.Is(err, domain.ErrPasswordResetNotFound) || errors.Is(err, domain.ErrInvalidPasswordReset) {
		return ConfirmPasswordResetResult{}, invalidPasswordResetError(err)
	}
	if err != nil {
		return ConfirmPasswordResetResult{}, err
	}
	now := s.clock.Now().UTC()
	if challenge.Status() == domain.PasswordResetChallengeConsumed || challenge.Status() == domain.PasswordResetChallengeInvalidated || challenge.IsCodeExpired(now) || challenge.FailedAttempts >= s.config.MaxConfirmAttempts {
		return ConfirmPasswordResetResult{}, invalidPasswordResetError(domain.ErrInvalidPasswordReset)
	}
	if !s.hasher.MatchesCode(challenge.CodeHash, challenge.PasswordResetID, challenge.UserID, input.Code) {
		_, err := s.resets.RecordFailedAttempt(ctx, domain.RecordPasswordResetFailureDTO{PasswordResetID: challenge.PasswordResetID, FailedAt: now, MaxAttempts: s.config.MaxConfirmAttempts})
		if err != nil && !errors.Is(err, domain.ErrInvalidPasswordReset) && !errors.Is(err, domain.ErrPasswordResetNotFound) {
			return ConfirmPasswordResetResult{}, err
		}
		return ConfirmPasswordResetResult{}, invalidPasswordResetError(domain.ErrInvalidPasswordReset)
	}
	token, err := s.tokens.GenerateToken()
	if err != nil {
		return ConfirmPasswordResetResult{}, err
	}
	expiresAt := now.Add(s.config.TokenTTL)
	_, err = s.resets.Confirm(ctx, domain.ConfirmPasswordResetChallengeDTO{PasswordResetID: challenge.PasswordResetID, PasswordResetTokenHash: s.hasher.HashToken(challenge.PasswordResetID, challenge.UserID, token), PasswordResetTokenExpiresAt: expiresAt, ConfirmedAt: now, MaxAttempts: s.config.MaxConfirmAttempts})
	if errors.Is(err, domain.ErrInvalidPasswordReset) || errors.Is(err, domain.ErrPasswordResetNotFound) {
		return ConfirmPasswordResetResult{}, invalidPasswordResetError(err)
	}
	if err != nil {
		return ConfirmPasswordResetResult{}, err
	}
	return ConfirmPasswordResetResult{PasswordResetToken: token, ExpiresAt: expiresAt}, nil
}
