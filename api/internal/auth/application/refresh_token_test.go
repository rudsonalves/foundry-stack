package application

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	authdomain "github.com/rudsonalves/foundry-stack/api/internal/auth/domain"
)

func TestRefreshTokenServiceIssue(t *testing.T) {
	now := time.Date(2026, time.July, 22, 10, 30, 0, 0, time.FixedZone("-0300", -3*60*60))
	lifetime := 48 * time.Hour
	userID := uuid.MustParse("34d90b26-95ce-468b-87d0-f70e88cf8f35")
	repository := &stubRefreshTokenRepository{}
	service := NewRefreshTokenService(repository, lifetime)
	service.now = func() time.Time { return now }

	result, err := service.Issue(context.Background(), userID, "foundry-stack-mobile")
	if err != nil {
		t.Fatalf("Issue() unexpected error: %v", err)
	}

	if result.Token == "" {
		t.Fatal("Issue() token should not be empty")
	}
	wantExpiresAt := now.UTC().Add(lifetime)
	if !result.ExpiresAt.Equal(wantExpiresAt) {
		t.Fatalf("Issue() expiresAt = %v, want %v", result.ExpiresAt, wantExpiresAt)
	}
	if repository.createCalls != 1 {
		t.Fatalf("Create() call count = %d, want 1", repository.createCalls)
	}
	if repository.createInput.UserID != userID {
		t.Fatalf("Create() userID = %s, want %s", repository.createInput.UserID, userID)
	}
	if repository.createInput.ClientID != "foundry-stack-mobile" {
		t.Fatalf("Create() clientID = %q, want %q", repository.createInput.ClientID, "foundry-stack-mobile")
	}
	if !repository.createInput.ExpiresAt.Equal(wantExpiresAt) {
		t.Fatalf("Create() expiresAt = %v, want %v", repository.createInput.ExpiresAt, wantExpiresAt)
	}
	if bytes.Equal(repository.createInput.TokenHash, []byte(result.Token)) {
		t.Fatal("Create() should persist only the token hash, not the original token")
	}
	if !bytes.Equal(repository.createInput.TokenHash, refreshTokenHash(result.Token)) {
		t.Fatal("Create() token hash does not match the issued token")
	}
	if strings.Contains(string(repository.createInput.TokenHash), result.Token) {
		t.Fatal("Create() token hash should not contain the original token")
	}
}

func TestRefreshTokenServiceIssuePropagatesPersistenceError(t *testing.T) {
	repository := &stubRefreshTokenRepository{createErr: errors.New("insert failed")}
	service := NewRefreshTokenService(repository, 48*time.Hour)

	_, err := service.Issue(context.Background(), uuid.New(), "foundry-stack-mobile")
	if err == nil {
		t.Fatal("Issue() expected error, got nil")
	}
	if !strings.Contains(err.Error(), "persist refresh token") {
		t.Fatalf("error = %q, want to contain %q", err.Error(), "persist refresh token")
	}
	if !errors.Is(err, repository.createErr) {
		t.Fatal("Issue() should wrap repository error")
	}
}

func TestRefreshTokenServiceAuthenticate(t *testing.T) {
	t.Run("valid token authenticates and returns refresh identity", func(t *testing.T) {
		now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.FixedZone("-0300", -3*60*60))
		userID := uuid.MustParse("063fdb1f-bd40-4e39-b15b-64d4f809ef61")
		repository := &stubRefreshTokenRepository{
			findResult: authdomain.RefreshToken{
				UserID:   userID,
				ClientID: "foundry-stack-swagger",
			},
		}
		service := NewRefreshTokenService(repository, 48*time.Hour)
		service.now = func() time.Time { return now }

		identity, err := service.Authenticate(context.Background(), "refresh-token")
		if err != nil {
			t.Fatalf("Authenticate() unexpected error: %v", err)
		}

		if repository.findCalls != 1 {
			t.Fatalf("FindActive() call count = %d, want 1", repository.findCalls)
		}
		if !bytes.Equal(repository.findTokenHash, refreshTokenHash("refresh-token")) {
			t.Fatal("FindActive() token hash does not match the provided token")
		}
		wantNow := now.UTC()
		if !repository.findNow.Equal(wantNow) {
			t.Fatalf("FindActive() now = %v, want %v", repository.findNow, wantNow)
		}
		if identity.UserID != userID {
			t.Fatalf("Authenticate() userID = %s, want %s", identity.UserID, userID)
		}
		if identity.ClientID != "foundry-stack-swagger" {
			t.Fatalf("Authenticate() clientID = %q, want %q", identity.ClientID, "foundry-stack-swagger")
		}
	})

	t.Run("empty token returns invalid refresh token without consulting repository", func(t *testing.T) {
		repository := &stubRefreshTokenRepository{}
		service := NewRefreshTokenService(repository, 48*time.Hour)

		_, err := service.Authenticate(context.Background(), "")
		if !errors.Is(err, authdomain.ErrInvalidRefreshToken) {
			t.Fatalf("Authenticate() error = %v, want %v", err, authdomain.ErrInvalidRefreshToken)
		}
		if repository.findCalls != 0 {
			t.Fatalf("FindActive() call count = %d, want 0", repository.findCalls)
		}
	})

	for _, name := range []string{"expired", "revoked", "nonexistent"} {
		name := name
		t.Run("propagates "+name+" refresh token", func(t *testing.T) {
			repository := &stubRefreshTokenRepository{findErr: authdomain.ErrInvalidRefreshToken}
			service := NewRefreshTokenService(repository, 48*time.Hour)

			_, err := service.Authenticate(context.Background(), "refresh-token")
			if !errors.Is(err, authdomain.ErrInvalidRefreshToken) {
				t.Fatalf("Authenticate() error = %v, want %v", err, authdomain.ErrInvalidRefreshToken)
			}
		})
	}
}

func TestRefreshTokenServiceRevoke(t *testing.T) {
	t.Run("revokes using token hash and current time", func(t *testing.T) {
		now := time.Date(2026, time.July, 22, 13, 0, 0, 0, time.FixedZone("-0300", -3*60*60))
		repository := &stubRefreshTokenRepository{}
		service := NewRefreshTokenService(repository, 48*time.Hour)
		service.now = func() time.Time { return now }

		if err := service.Revoke(context.Background(), "refresh-token"); err != nil {
			t.Fatalf("Revoke() unexpected error: %v", err)
		}

		if repository.revokeCalls != 1 {
			t.Fatalf("Revoke() call count = %d, want 1", repository.revokeCalls)
		}
		if !bytes.Equal(repository.revokeTokenHash, refreshTokenHash("refresh-token")) {
			t.Fatal("Revoke() token hash does not match the provided token")
		}
		wantRevokedAt := now.UTC()
		if !repository.revokeAt.Equal(wantRevokedAt) {
			t.Fatalf("Revoke() revokedAt = %v, want %v", repository.revokeAt, wantRevokedAt)
		}
	})

	t.Run("empty token is idempotent", func(t *testing.T) {
		repository := &stubRefreshTokenRepository{}
		service := NewRefreshTokenService(repository, 48*time.Hour)

		if err := service.Revoke(context.Background(), ""); err != nil {
			t.Fatalf("Revoke() unexpected error: %v", err)
		}
		if repository.revokeCalls != 0 {
			t.Fatalf("Revoke() call count = %d, want 0", repository.revokeCalls)
		}
	})
}

type stubRefreshTokenRepository struct {
	createInput     authdomain.CreateRefreshTokenDTO
	createErr       error
	createCalls     int
	findResult      authdomain.RefreshToken
	findErr         error
	findTokenHash   []byte
	findNow         time.Time
	findCalls       int
	revokeErr       error
	revokeTokenHash []byte
	revokeAt        time.Time
	revokeCalls     int
}

func (s *stubRefreshTokenRepository) Create(
	_ context.Context,
	input authdomain.CreateRefreshTokenDTO,
) error {
	s.createCalls++
	s.createInput = authdomain.CreateRefreshTokenDTO{
		TokenHash: append([]byte(nil), input.TokenHash...),
		UserID:    input.UserID,
		ClientID:  input.ClientID,
		ExpiresAt: input.ExpiresAt,
	}
	return s.createErr
}

func (s *stubRefreshTokenRepository) FindActive(
	_ context.Context,
	tokenHash []byte,
	now time.Time,
) (authdomain.RefreshToken, error) {
	s.findCalls++
	s.findTokenHash = append([]byte(nil), tokenHash...)
	s.findNow = now
	if s.findErr != nil {
		return authdomain.RefreshToken{}, s.findErr
	}
	return s.findResult, nil
}

func (s *stubRefreshTokenRepository) Revoke(
	_ context.Context,
	tokenHash []byte,
	revokedAt time.Time,
) error {
	s.revokeCalls++
	s.revokeTokenHash = append([]byte(nil), tokenHash...)
	s.revokeAt = revokedAt
	return s.revokeErr
}
