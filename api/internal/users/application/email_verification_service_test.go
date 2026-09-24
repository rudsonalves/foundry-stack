package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

func TestEmailVerificationServiceRequestRejectsRegisteredEmail(t *testing.T) {
	users := &requestUserRepositoryStub{}
	codes := &requestCodeGeneratorStub{code: "123456"}
	service := newRequestEmailVerificationService(users, codes, nil, nil, nil)

	_, err := service.Request(context.Background(), RequestEmailVerificationInput{
		Email:    "registered@example.com",
		ClientIP: "203.0.113.1",
	})

	var appErr *sharederrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Request() error type = %T, want *AppError", err)
	}
	if appErr.Code != sharederrors.ErrCodeEmailAlreadyRegistered {
		t.Fatalf("Request() code = %q", appErr.Code)
	}
	if codes.calls != 0 {
		t.Fatal("code generator called for registered email")
	}
}

func TestEmailVerificationServiceRequestMapsRateLimit(t *testing.T) {
	users := &requestUserRepositoryStub{findByEmailErr: domain.ErrUserNotFound}
	codes := &requestCodeGeneratorStub{code: "123456"}
	limiter := &requestRateLimiterStub{reserveErr: domain.ErrEmailVerificationRateLimit}
	service := newRequestEmailVerificationService(users, codes, nil, nil, limiter)

	_, err := service.Request(context.Background(), RequestEmailVerificationInput{
		Email:    "user@example.com",
		ClientIP: "203.0.113.1",
	})

	var appErr *sharederrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("Request() error type = %T, want *AppError", err)
	}
	if appErr.Code != sharederrors.ErrCodeRateLimitExceeded {
		t.Fatalf("Request() code = %q", appErr.Code)
	}
	if codes.calls != 0 {
		t.Fatal("code generator called after rate limit rejection")
	}
}

func TestEmailVerificationServiceRequestAllowsRetryAfterDeliveryFailure(t *testing.T) {
	users := &requestUserRepositoryStub{findByEmailErr: domain.ErrUserNotFound}
	codes := &requestCodeGeneratorStub{code: "000042"}
	verifications := &requestVerificationRepositoryStub{}
	sender := &requestSenderStub{
		failuresRemaining: 1,
		err:               errors.New("SMTP unavailable"),
	}
	limiter := &requestRateLimiterStub{}
	service := newRequestEmailVerificationService(
		users,
		codes,
		verifications,
		sender,
		limiter,
	)

	input := RequestEmailVerificationInput{
		Email:    "  USER@EXAMPLE.COM ",
		ClientIP: "203.0.113.1",
	}

	_, err := service.Request(context.Background(), input)
	var appErr *sharederrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("first Request() error type = %T, want *AppError", err)
	}
	if appErr.Code != sharederrors.ErrCodeEmailDeliveryUnavailable {
		t.Fatalf("first Request() code = %q", appErr.Code)
	}
	if len(verifications.invalidated) != 1 {
		t.Fatalf("Invalidate() calls = %d, want 1", len(verifications.invalidated))
	}
	if len(limiter.released) != 1 {
		t.Fatalf("ReleaseEmail() calls = %d, want 1", len(limiter.released))
	}

	result, err := service.Request(context.Background(), input)
	if err != nil {
		t.Fatalf("second Request() error = %v", err)
	}
	if result.VerificationID == uuid.Nil {
		t.Fatal("second Request() returned nil verification ID")
	}
	if sender.calls != 2 {
		t.Fatalf("SendVerificationCode() calls = %d, want 2", sender.calls)
	}
	if verifications.lastReplace.Email != "user@example.com" {
		t.Fatalf("Replace() email = %q", verifications.lastReplace.Email)
	}
	if sender.lastInput.Code != "000042" {
		t.Fatalf("sent code = %q", sender.lastInput.Code)
	}
}

func newRequestEmailVerificationService(
	users domain.UserRepository,
	codes domain.EmailVerificationCodeGenerator,
	verifications domain.EmailVerificationRepository,
	sender domain.EmailVerificationSender,
	limiter domain.EmailVerificationRateLimiter,
) *EmailVerificationService {
	if verifications == nil {
		verifications = &requestVerificationRepositoryStub{}
	}
	if sender == nil {
		sender = &requestSenderStub{}
	}
	if limiter == nil {
		limiter = &requestRateLimiterStub{}
	}

	return NewEmailVerificationService(
		users,
		verifications,
		codes,
		nil,
		requestHasherStub{},
		sender,
		requestClockStub{now: time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)},
		EmailVerificationServiceConfig{
			CodeTTL:            15 * time.Minute,
			TokenTTL:           24 * time.Minute,
			ResendCooldown:     time.Minute,
			MaxConfirmAttempts: 5,
		},
		limiter,
	)
}

type requestUserRepositoryStub struct {
	findByEmailErr error
}

func (s *requestUserRepositoryStub) Create(context.Context, domain.CreateUserDTO) (domain.User, error) {
	return domain.User{}, errors.New("not implemented")
}

func (s *requestUserRepositoryStub) FindByID(context.Context, uuid.UUID) (domain.User, error) {
	return domain.User{}, errors.New("not implemented")
}

func (s *requestUserRepositoryStub) FindByEmail(context.Context, string) (domain.User, error) {
	return domain.User{}, s.findByEmailErr
}

type requestVerificationRepositoryStub struct {
	lastReplace domain.ReplaceEmailVerificationChallengeDTO
	invalidated []uuid.UUID
}

func (s *requestVerificationRepositoryStub) Replace(
	_ context.Context,
	input domain.ReplaceEmailVerificationChallengeDTO,
) (domain.EmailVerificationChallenge, error) {
	s.lastReplace = input
	return domain.EmailVerificationChallenge{
		VerificationID: input.VerificationID,
		Email:          input.Email,
		CodeHash:       input.CodeHash,
		CreatedAt:      input.CreatedAt,
		UpdatedAt:      input.CreatedAt,
		CodeExpiresAt:  input.CodeExpiresAt,
	}, nil
}

func (s *requestVerificationRepositoryStub) FindByID(
	context.Context,
	uuid.UUID,
) (domain.EmailVerificationChallenge, error) {
	return domain.EmailVerificationChallenge{}, errors.New("not implemented")
}

func (s *requestVerificationRepositoryStub) RecordFailedAttempt(
	context.Context,
	domain.RecordEmailVerificationFailureDTO,
) (domain.EmailVerificationChallenge, error) {
	return domain.EmailVerificationChallenge{}, errors.New("not implemented")
}

func (s *requestVerificationRepositoryStub) Confirm(
	context.Context,
	domain.ConfirmEmailVerificationChallengeDTO,
) (domain.EmailVerificationChallenge, error) {
	return domain.EmailVerificationChallenge{}, errors.New("not implemented")
}

func (s *requestVerificationRepositoryStub) DeleteStale(
	context.Context,
	time.Time,
) (int64, error) {
	return 0, errors.New("not implemented")
}

func (s *requestVerificationRepositoryStub) Invalidate(
	_ context.Context,
	verificationID uuid.UUID,
	_ time.Time,
) error {
	s.invalidated = append(s.invalidated, verificationID)
	return nil
}

type requestCodeGeneratorStub struct {
	code  string
	calls int
}

func (s *requestCodeGeneratorStub) GenerateCode() (string, error) {
	s.calls++
	return s.code, nil
}

type requestHasherStub struct{}

func (requestHasherStub) HashCode(
	verificationID uuid.UUID,
	email string,
	code string,
) []byte {
	return []byte(
		verificationID.String() + ":" + email + ":" + code,
	)
}

func (h requestHasherStub) MatchesCode(
	hash []byte,
	verificationID uuid.UUID,
	email string,
	code string,
) bool {
	return string(hash) == string(
		h.HashCode(verificationID, email, code),
	)
}

func (requestHasherStub) HashToken(email string, token string) []byte {
	return []byte(email + ":" + token)
}

func (h requestHasherStub) MatchesToken(
	hash []byte,
	email string,
	token string,
) bool {
	return string(hash) == string(h.HashToken(email, token))
}

type requestSenderStub struct {
	failuresRemaining int
	err               error
	calls             int
	lastInput         domain.SendEmailVerificationCodeDTO
}

func (s *requestSenderStub) SendVerificationCode(
	_ context.Context,
	input domain.SendEmailVerificationCodeDTO,
) error {
	s.calls++
	s.lastInput = input
	if s.failuresRemaining > 0 {
		s.failuresRemaining--
		return s.err
	}
	return nil
}

type requestClockStub struct {
	now time.Time
}

func (s requestClockStub) Now() time.Time {
	return s.now
}

type requestRateLimiterStub struct {
	reserveErr error
	released   []uuid.UUID
}

func (s *requestRateLimiterStub) Reserve(
	context.Context,
	domain.ReserveEmailVerificationRequestDTO,
) (uuid.UUID, error) {
	if s.reserveErr != nil {
		return uuid.Nil, s.reserveErr
	}
	return uuid.New(), nil
}

func (s *requestRateLimiterStub) ReleaseEmail(
	_ context.Context,
	reservationID uuid.UUID,
) error {
	s.released = append(s.released, reservationID)
	return nil
}
