package application

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

func TestEmailVerificationServiceConfirmSuccess(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	verificationID := uuid.New()
	code := "000042"
	hasher := requestHasherStub{}
	repository := newConfirmationRepositoryStub(
		domain.EmailVerificationChallenge{
			VerificationID: verificationID,
			Email:          "user@example.com",
			CodeHash: hasher.HashCode(
				verificationID,
				"user@example.com",
				code,
			),
			CodeExpiresAt: now.Add(15 * time.Minute),
		},
	)
	tokens := &confirmationTokenGeneratorStub{
		tokens: []string{"opaque-token"},
	}
	service := newConfirmationService(
		repository,
		tokens,
		now,
	)

	result, err := service.Confirm(
		context.Background(),
		ConfirmEmailVerificationInput{
			VerificationID: verificationID,
			Code:           code,
		},
	)
	if err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}

	if result.EmailVerificationToken != "opaque-token" {
		t.Errorf(
			"token = %q",
			result.EmailVerificationToken,
		)
	}

	if result.ExpiresAt != now.Add(24*time.Hour) {
		t.Errorf("ExpiresAt = %v", result.ExpiresAt)
	}

	inputs := repository.confirmationInputs()
	if len(inputs) != 1 {
		t.Fatalf("Confirm repository calls = %d", len(inputs))
	}

	input := inputs[0]
	if input.VerificationID != verificationID {
		t.Errorf("repository verification ID = %s", input.VerificationID)
	}
	if input.MaxAttempts != 5 {
		t.Errorf("repository MaxAttempts = %d", input.MaxAttempts)
	}

	expectedHash := hasher.HashToken(
		"user@example.com",
		"opaque-token",
	)
	if string(input.VerificationTokenHash) != string(expectedHash) {
		t.Fatal("repository received unexpected token hash")
	}
}

func TestEmailVerificationServiceConfirmRotatesToken(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	verificationID := uuid.New()
	code := "123456"
	hasher := requestHasherStub{}
	repository := newConfirmationRepositoryStub(
		domain.EmailVerificationChallenge{
			VerificationID: verificationID,
			Email:          "user@example.com",
			CodeHash: hasher.HashCode(
				verificationID,
				"user@example.com",
				code,
			),
			CodeExpiresAt: now.Add(15 * time.Minute),
		},
	)
	tokens := &confirmationTokenGeneratorStub{
		tokens: []string{"first-token", "second-token"},
	}
	service := newConfirmationService(repository, tokens, now)

	first, err := service.Confirm(
		context.Background(),
		ConfirmEmailVerificationInput{
			VerificationID: verificationID,
			Code:           code,
		},
	)
	if err != nil {
		t.Fatalf("first Confirm() error = %v", err)
	}

	second, err := service.Confirm(
		context.Background(),
		ConfirmEmailVerificationInput{
			VerificationID: verificationID,
			Code:           code,
		},
	)
	if err != nil {
		t.Fatalf("second Confirm() error = %v", err)
	}

	if first.EmailVerificationToken == second.EmailVerificationToken {
		t.Fatal("repeated confirmation returned the same token")
	}

	inputs := repository.confirmationInputs()
	if len(inputs) != 2 {
		t.Fatalf("repository Confirm() calls = %d", len(inputs))
	}

	if string(inputs[0].VerificationTokenHash) ==
		string(inputs[1].VerificationTokenHash) {
		t.Fatal("token hash was not rotated")
	}
}

func TestEmailVerificationServiceConfirmRejectsExpiredChallenge(
	t *testing.T,
) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	repository := newConfirmationRepositoryStub(
		domain.EmailVerificationChallenge{
			VerificationID: uuid.New(),
			Email:          "user@example.com",
			CodeHash:       []byte("hash"),
			CodeExpiresAt:  now,
		},
	)
	tokens := &confirmationTokenGeneratorStub{
		tokens: []string{"must-not-be-generated"},
	}
	service := newConfirmationService(repository, tokens, now)

	_, err := service.Confirm(
		context.Background(),
		ConfirmEmailVerificationInput{
			VerificationID: repository.challenge.VerificationID,
			Code:           "123456",
		},
	)

	assertInvalidEmailVerification(t, err)

	if tokens.calls() != 0 {
		t.Fatal("token generated for expired challenge")
	}
	if len(repository.confirmationInputs()) != 0 {
		t.Fatal("repository Confirm called for expired challenge")
	}
}

func TestEmailVerificationServiceConfirmRecordsIncorrectCode(
	t *testing.T,
) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	verificationID := uuid.New()
	hasher := requestHasherStub{}
	repository := newConfirmationRepositoryStub(
		domain.EmailVerificationChallenge{
			VerificationID: verificationID,
			Email:          "user@example.com",
			CodeHash: hasher.HashCode(
				verificationID,
				"user@example.com",
				"123456",
			),
			CodeExpiresAt: now.Add(15 * time.Minute),
		},
	)
	service := newConfirmationService(
		repository,
		&confirmationTokenGeneratorStub{},
		now,
	)

	_, err := service.Confirm(
		context.Background(),
		ConfirmEmailVerificationInput{
			VerificationID: verificationID,
			Code:           "654321",
		},
	)

	assertInvalidEmailVerification(t, err)

	failures := repository.recordedFailures()
	if len(failures) != 1 {
		t.Fatalf("RecordFailedAttempt() calls = %d", len(failures))
	}
	if failures[0].MaxAttempts != 5 {
		t.Errorf("MaxAttempts = %d", failures[0].MaxAttempts)
	}
	if failures[0].FailedAt != now {
		t.Errorf("FailedAt = %v", failures[0].FailedAt)
	}
}

func TestEmailVerificationServiceConfirmHidesUnavailableChallenge(
	t *testing.T,
) {
	repository := newConfirmationRepositoryStub(
		domain.EmailVerificationChallenge{},
	)
	repository.findErr = domain.ErrEmailVerificationNotFound

	service := newConfirmationService(
		repository,
		&confirmationTokenGeneratorStub{},
		time.Now().UTC(),
	)

	_, err := service.Confirm(
		context.Background(),
		ConfirmEmailVerificationInput{
			VerificationID: uuid.New(),
			Code:           "123456",
		},
	)

	assertInvalidEmailVerification(t, err)
}

func TestEmailVerificationServiceConfirmPreservesUnexpectedError(
	t *testing.T,
) {
	repositoryError := errors.New("database unavailable")
	repository := newConfirmationRepositoryStub(
		domain.EmailVerificationChallenge{},
	)
	repository.findErr = repositoryError

	service := newConfirmationService(
		repository,
		&confirmationTokenGeneratorStub{},
		time.Now().UTC(),
	)

	_, err := service.Confirm(
		context.Background(),
		ConfirmEmailVerificationInput{
			VerificationID: uuid.New(),
			Code:           "123456",
		},
	)

	if !errors.Is(err, repositoryError) {
		t.Fatalf("Confirm() error = %v", err)
	}
}

func TestEmailVerificationServiceConfirmConcurrentRotation(
	t *testing.T,
) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	verificationID := uuid.New()
	code := "123456"
	hasher := requestHasherStub{}
	repository := newConfirmationRepositoryStub(
		domain.EmailVerificationChallenge{
			VerificationID: verificationID,
			Email:          "user@example.com",
			CodeHash: hasher.HashCode(
				verificationID,
				"user@example.com",
				code,
			),
			CodeExpiresAt: now.Add(15 * time.Minute),
		},
	)

	const confirmations = 20
	generatedTokens := make([]string, confirmations)
	for index := range generatedTokens {
		generatedTokens[index] = fmt.Sprintf(
			"token-%02d",
			index,
		)
	}

	tokens := &confirmationTokenGeneratorStub{
		tokens: generatedTokens,
	}
	service := newConfirmationService(repository, tokens, now)

	results := make(chan ConfirmEmailVerificationResult, confirmations)
	errs := make(chan error, confirmations)
	var waitGroup sync.WaitGroup

	for range confirmations {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			result, err := service.Confirm(
				context.Background(),
				ConfirmEmailVerificationInput{
					VerificationID: verificationID,
					Code:           code,
				},
			)
			results <- result
			errs <- err
		}()
	}

	waitGroup.Wait()
	close(results)
	close(errs)

	for err := range errs {
		if err != nil {
			t.Errorf("concurrent Confirm() error = %v", err)
		}
	}

	returnedTokens := make(map[string]struct{}, confirmations)
	for result := range results {
		returnedTokens[result.EmailVerificationToken] = struct{}{}
	}

	if len(returnedTokens) != confirmations {
		t.Errorf(
			"distinct returned tokens = %d, want %d",
			len(returnedTokens),
			confirmations,
		)
	}

	if len(repository.confirmationInputs()) != confirmations {
		t.Errorf(
			"repository Confirm() calls = %d, want %d",
			len(repository.confirmationInputs()),
			confirmations,
		)
	}
}

func newConfirmationService(
	repository domain.EmailVerificationRepository,
	tokens domain.EmailVerificationTokenGenerator,
	now time.Time,
) *EmailVerificationService {
	return NewEmailVerificationService(
		nil,
		repository,
		nil,
		tokens,
		requestHasherStub{},
		nil,
		requestClockStub{now: now},
		EmailVerificationServiceConfig{
			CodeTTL:            15 * time.Minute,
			TokenTTL:           24 * time.Hour,
			ResendCooldown:     time.Minute,
			MaxConfirmAttempts: 5,
		},
		nil,
	)
}

func assertInvalidEmailVerification(
	t *testing.T,
	err error,
) {
	t.Helper()

	var appErr *sharederrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error type = %T, want *AppError", err)
	}

	if appErr.Code != sharederrors.ErrCodeInvalidEmailVerification {
		t.Fatalf("error code = %q", appErr.Code)
	}

	if appErr.Message != "invalid email verification" {
		t.Fatalf("error message = %q", appErr.Message)
	}
}

type confirmationRepositoryStub struct {
	mu            sync.Mutex
	challenge     domain.EmailVerificationChallenge
	findErr       error
	confirmInputs []domain.ConfirmEmailVerificationChallengeDTO
	failureInputs []domain.RecordEmailVerificationFailureDTO
}

func newConfirmationRepositoryStub(
	challenge domain.EmailVerificationChallenge,
) *confirmationRepositoryStub {
	return &confirmationRepositoryStub{
		challenge: challenge,
	}
}

func (s *confirmationRepositoryStub) Replace(
	context.Context,
	domain.ReplaceEmailVerificationChallengeDTO,
) (domain.EmailVerificationChallenge, error) {
	return domain.EmailVerificationChallenge{},
		errors.New("not implemented")
}

func (s *confirmationRepositoryStub) FindByID(
	context.Context,
	uuid.UUID,
) (domain.EmailVerificationChallenge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.challenge, s.findErr
}

func (s *confirmationRepositoryStub) RecordFailedAttempt(
	_ context.Context,
	input domain.RecordEmailVerificationFailureDTO,
) (domain.EmailVerificationChallenge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.failureInputs = append(s.failureInputs, input)
	s.challenge.FailedAttempts++

	if s.challenge.FailedAttempts >= input.MaxAttempts {
		invalidatedAt := input.FailedAt
		s.challenge.InvalidatedAt = &invalidatedAt
	}

	return s.challenge, nil
}

func (s *confirmationRepositoryStub) Confirm(
	_ context.Context,
	input domain.ConfirmEmailVerificationChallengeDTO,
) (domain.EmailVerificationChallenge, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.confirmInputs = append(s.confirmInputs, input)
	s.challenge.VerificationTokenHash =
		append([]byte(nil), input.VerificationTokenHash...)
	s.challenge.VerificationTokenExpiresAt =
		&input.VerificationTokenExpiresAt

	if s.challenge.ConfirmedAt == nil {
		confirmedAt := input.ConfirmedAt
		s.challenge.ConfirmedAt = &confirmedAt
	}

	return s.challenge, nil
}

func (s *confirmationRepositoryStub) DeleteStale(
	context.Context,
	time.Time,
) (int64, error) {
	return 0, errors.New("not implemented")
}

func (s *confirmationRepositoryStub) Invalidate(
	context.Context,
	uuid.UUID,
	time.Time,
) error {
	return errors.New("not implemented")
}

func (s *confirmationRepositoryStub) confirmationInputs() []domain.ConfirmEmailVerificationChallengeDTO {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(
		[]domain.ConfirmEmailVerificationChallengeDTO,
		len(s.confirmInputs),
	)
	copy(result, s.confirmInputs)

	return result
}

func (s *confirmationRepositoryStub) recordedFailures() []domain.RecordEmailVerificationFailureDTO {
	s.mu.Lock()
	defer s.mu.Unlock()

	result := make(
		[]domain.RecordEmailVerificationFailureDTO,
		len(s.failureInputs),
	)
	copy(result, s.failureInputs)

	return result
}

type confirmationTokenGeneratorStub struct {
	mu     sync.Mutex
	tokens []string
	index  int
}

func (s *confirmationTokenGeneratorStub) GenerateToken() (
	string,
	error,
) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.index >= len(s.tokens) {
		return "", errors.New("no token configured")
	}

	token := s.tokens[s.index]
	s.index++

	return token, nil
}

func (s *confirmationTokenGeneratorStub) calls() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.index
}
