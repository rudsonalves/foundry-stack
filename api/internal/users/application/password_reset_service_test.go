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

type resetRepoStub struct {
	challenge   domain.PasswordResetChallenge
	findErr     error
	replaced    bool
	failed      bool
	confirmed   domain.ConfirmPasswordResetChallengeDTO
	invalidated bool
	completed   domain.CompletePasswordResetDTO
	completeErr error
}

func (r *resetRepoStub) Replace(_ context.Context, in domain.ReplacePasswordResetChallengeDTO) (domain.PasswordResetChallenge, error) {
	r.replaced = true
	r.challenge = domain.PasswordResetChallenge{PasswordResetID: in.PasswordResetID, UserID: in.UserID, CodeHash: in.CodeHash, CodeExpiresAt: in.CodeExpiresAt}
	return r.challenge, nil
}
func (r *resetRepoStub) FindByID(context.Context, uuid.UUID) (domain.PasswordResetChallenge, error) {
	return r.challenge, r.findErr
}
func (r *resetRepoStub) RecordFailedAttempt(context.Context, domain.RecordPasswordResetFailureDTO) (domain.PasswordResetChallenge, error) {
	r.failed = true
	return r.challenge, nil
}
func (r *resetRepoStub) Confirm(_ context.Context, in domain.ConfirmPasswordResetChallengeDTO) (domain.PasswordResetChallenge, error) {
	r.confirmed = in
	return r.challenge, nil
}
func (r *resetRepoStub) Invalidate(context.Context, domain.InvalidatePasswordResetChallengeDTO) error {
	r.invalidated = true
	return nil
}
func (r *resetRepoStub) DeleteStale(context.Context, time.Time) (int64, error) { return 0, nil }
func (r *resetRepoStub) Complete(_ context.Context, input domain.CompletePasswordResetDTO) error {
	r.completed = input
	return r.completeErr
}

type resetCodeStub struct{ value string }

func (s resetCodeStub) GenerateCode() (string, error) { return s.value, nil }

type resetTokenStub struct{ value string }

func (s resetTokenStub) GenerateToken() (string, error) { return s.value, nil }

type resetHashStub struct{}

func (resetHashStub) HashCode(_ uuid.UUID, _ uuid.UUID, v string) []byte  { return []byte(v) }
func (resetHashStub) HashToken(_ uuid.UUID, _ uuid.UUID, v string) []byte { return []byte(v) }
func (resetHashStub) MatchesCode(hash []byte, _ uuid.UUID, _ uuid.UUID, v string) bool {
	return string(hash) == v
}
func (resetHashStub) MatchesToken(hash []byte, _ uuid.UUID, _ uuid.UUID, v string) bool {
	return string(hash) == v
}

type resetSenderStub struct {
	sent bool
	err  error
}

func (s *resetSenderStub) SendPasswordResetCode(context.Context, domain.SendPasswordResetCodeDTO) error {
	s.sent = true
	return s.err
}

type resetLimiterStub struct{ reserved bool }

func (s *resetLimiterStub) Reserve(context.Context, domain.ReservePasswordResetRequestDTO) (uuid.UUID, error) {
	s.reserved = true
	return uuid.New(), nil
}
func (s *resetLimiterStub) Release(context.Context, uuid.UUID) error { return nil }

type resetClock struct{ now time.Time }

func (c resetClock) Now() time.Time { return c.now }

func newResetService(users domain.UserRepository, repo *resetRepoStub, sender *resetSenderStub) *PasswordResetService {
	return NewPasswordResetService(users, repo, resetCodeStub{"123456"}, resetTokenStub{"token"}, resetHashStub{}, sender, resetClock{time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)}, PasswordResetServiceConfig{CodeTTL: 15 * time.Minute, TokenTTL: 10 * time.Minute, ResendCooldown: time.Minute, MaxConfirmAttempts: 5}, &resetLimiterStub{}, nil)
}

func TestPasswordResetRequestDoesNotRevealMissingUser(t *testing.T) {
	repo := &resetRepoStub{}
	sender := &resetSenderStub{}
	users := &stubUserRepository{findByEmailErr: domain.ErrUserNotFound}
	svc := newResetService(users, repo, sender)
	result, err := svc.Request(context.Background(), RequestPasswordResetInput{Email: " USER@example.com ", ClientIP: "203.0.113.1"})
	if err != nil {
		t.Fatal(err)
	}
	if result.PasswordResetID == uuid.Nil || repo.replaced || sender.sent {
		t.Fatalf("unexpected missing-user behavior: %#v replaced=%v sent=%v", result, repo.replaced, sender.sent)
	}
}
func TestPasswordResetRequestCreatesAndSendsChallenge(t *testing.T) {
	userID := uuid.New()
	repo := &resetRepoStub{}
	sender := &resetSenderStub{}
	users := &stubUserRepository{findByEmailResult: domain.User{ID: userID}}
	svc := newResetService(users, repo, sender)
	if _, err := svc.Request(context.Background(), RequestPasswordResetInput{Email: "user@example.com", ClientIP: "203.0.113.1"}); err != nil {
		t.Fatal(err)
	}
	if !repo.replaced || !sender.sent {
		t.Fatalf("replaced=%v sent=%v", repo.replaced, sender.sent)
	}
}
func TestPasswordResetConfirmIssuesTokenAndRejectsWrongCode(t *testing.T) {
	id, userID := uuid.New(), uuid.New()
	repo := &resetRepoStub{challenge: domain.PasswordResetChallenge{PasswordResetID: id, UserID: userID, CodeHash: []byte("123456"), CodeExpiresAt: time.Date(2026, 9, 23, 12, 15, 0, 0, time.UTC)}}
	svc := newResetService(&stubUserRepository{}, repo, &resetSenderStub{})
	result, err := svc.Confirm(context.Background(), ConfirmPasswordResetInput{PasswordResetID: id, Code: "123456"})
	if err != nil || result.PasswordResetToken != "token" || string(repo.confirmed.PasswordResetTokenHash) != "token" {
		t.Fatalf("result=%#v err=%v", result, err)
	}
	_, err = svc.Confirm(context.Background(), ConfirmPasswordResetInput{PasswordResetID: id, Code: "000000"})
	var appErr *sharederrors.AppError
	if !errors.As(err, &appErr) || appErr.Code != sharederrors.ErrCodeInvalidPasswordReset || !repo.failed {
		t.Fatalf("err=%v failed=%v", err, repo.failed)
	}
}

func TestPasswordResetCompleteHashesPasswordAndDelegatesTransaction(t *testing.T) {
	now := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	id, userID := uuid.New(), uuid.New()
	expiresAt := now.Add(time.Minute)
	repo := &resetRepoStub{challenge: domain.PasswordResetChallenge{PasswordResetID: id, UserID: userID, ConfirmedAt: &now, PasswordResetTokenExpiresAt: &expiresAt}}
	passwords := &stubPasswordHasher{hash: "new-hash"}
	svc := newResetService(&stubUserRepository{}, repo, &resetSenderStub{})
	svc.completer, svc.passwords = repo, passwords
	if err := svc.Complete(context.Background(), CompletePasswordResetInput{PasswordResetID: id, PasswordResetToken: "proof", NewPassword: "new-password"}); err != nil {
		t.Fatal(err)
	}
	if passwords.hashInput != "new-password" || repo.completed.PasswordHash != "new-hash" || string(repo.completed.PasswordResetTokenHash) != "proof" {
		t.Fatalf("hashInput=%q completed=%#v", passwords.hashInput, repo.completed)
	}
	if err := svc.Complete(context.Background(), CompletePasswordResetInput{PasswordResetID: id, PasswordResetToken: "proof", NewPassword: "short"}); err == nil {
		t.Fatal("expected validation error")
	}
}
