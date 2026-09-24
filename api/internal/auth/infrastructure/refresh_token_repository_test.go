package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	authdomain "github.com/rudsonalves/foundry-stack/api/internal/auth/domain"
)

const (
	createRefreshTokenQuery = `
		INSERT INTO refresh_tokens (
			token_hash,
			user_id,
			client_id,
			expires_at
		)
		VALUES ($1, $2, $3, $4)
	`
	findActiveRefreshTokenQuery = `
		SELECT
			user_id,
			client_id,
			created_at,
			expires_at,
			revoked_at
		FROM refresh_tokens
		WHERE token_hash = $1
			AND revoked_at IS NULL
			AND expires_at > $2
	`
	revokeRefreshTokenQuery = `
		UPDATE refresh_tokens
		SET revoked_at = $2
		WHERE token_hash = $1
			AND revoked_at IS NULL
	`
)

func TestSQLRefreshTokenRepositoryCreate(t *testing.T) {
	userID := uuid.MustParse("f00ef324-0117-45c3-b574-b36849e6e7e4")
	expiresAt := time.Date(2026, time.July, 24, 12, 0, 0, 0, time.UTC)
	tokenHash := []byte("refresh-token-hash")
	input := authdomain.CreateRefreshTokenDTO{
		TokenHash: tokenHash,
		UserID:    userID,
		ClientID:  "foundry-stack-mobile",
		ExpiresAt: expiresAt,
	}

	t.Run("executes insert with expected arguments", func(t *testing.T) {
		repository, mock := newMockRefreshTokenRepository(t)
		mock.ExpectExec(regexp.QuoteMeta(createRefreshTokenQuery)).
			WithArgs(tokenHash, userID, "foundry-stack-mobile", expiresAt).
			WillReturnResult(sqlmock.NewResult(0, 1))

		if err := repository.Create(context.Background(), input); err != nil {
			t.Fatalf("Create() unexpected error: %v", err)
		}
	})

	t.Run("wraps insert error", func(t *testing.T) {
		repository, mock := newMockRefreshTokenRepository(t)
		dbErr := errors.New("insert failed")
		mock.ExpectExec(regexp.QuoteMeta(createRefreshTokenQuery)).
			WithArgs(tokenHash, userID, "foundry-stack-mobile", expiresAt).
			WillReturnError(dbErr)

		err := repository.Create(context.Background(), input)
		assertRefreshTokenRepositoryError(t, err, dbErr, "create refresh token")
	})
}

func TestSQLRefreshTokenRepositoryFindActive(t *testing.T) {
	tokenHash := []byte("refresh-token-hash")
	now := time.Date(2026, time.July, 22, 12, 0, 0, 0, time.UTC)
	userID := uuid.MustParse("5042e46d-4599-4d26-b541-2f7300efdc05")
	createdAt := now.Add(-time.Hour)
	expiresAt := now.Add(24 * time.Hour)

	t.Run("returns active refresh token and queries only non revoked non expired rows", func(t *testing.T) {
		repository, mock := newMockRefreshTokenRepository(t)
		mock.ExpectQuery(regexp.QuoteMeta(findActiveRefreshTokenQuery)).
			WithArgs(tokenHash, now).
			WillReturnRows(sqlmock.NewRows([]string{"user_id", "client_id", "created_at", "expires_at", "revoked_at"}).
				AddRow(userID, "foundry-stack-mobile", createdAt, expiresAt, nil))

		token, err := repository.FindActive(context.Background(), tokenHash, now)
		if err != nil {
			t.Fatalf("FindActive() unexpected error: %v", err)
		}
		if token.UserID != userID {
			t.Fatalf("FindActive() userID = %s, want %s", token.UserID, userID)
		}
		if token.ClientID != "foundry-stack-mobile" {
			t.Fatalf("FindActive() clientID = %q, want %q", token.ClientID, "foundry-stack-mobile")
		}
		if !token.CreatedAt.Equal(createdAt) {
			t.Fatalf("FindActive() createdAt = %v, want %v", token.CreatedAt, createdAt)
		}
		if !token.ExpiresAt.Equal(expiresAt) {
			t.Fatalf("FindActive() expiresAt = %v, want %v", token.ExpiresAt, expiresAt)
		}
		if token.RevokedAt != nil {
			t.Fatalf("FindActive() revokedAt = %v, want nil", token.RevokedAt)
		}
	})

	t.Run("translates missing row to invalid refresh token", func(t *testing.T) {
		repository, mock := newMockRefreshTokenRepository(t)
		mock.ExpectQuery(regexp.QuoteMeta(findActiveRefreshTokenQuery)).
			WithArgs(tokenHash, now).
			WillReturnError(sql.ErrNoRows)

		_, err := repository.FindActive(context.Background(), tokenHash, now)
		if !errors.Is(err, authdomain.ErrInvalidRefreshToken) {
			t.Fatalf("FindActive() error = %v, want %v", err, authdomain.ErrInvalidRefreshToken)
		}
	})

	t.Run("wraps unexpected query error", func(t *testing.T) {
		repository, mock := newMockRefreshTokenRepository(t)
		dbErr := errors.New("query failed")
		mock.ExpectQuery(regexp.QuoteMeta(findActiveRefreshTokenQuery)).
			WithArgs(tokenHash, now).
			WillReturnError(dbErr)

		_, err := repository.FindActive(context.Background(), tokenHash, now)
		assertRefreshTokenRepositoryError(t, err, dbErr, "find active refresh token")
	})
}

func TestSQLRefreshTokenRepositoryRevoke(t *testing.T) {
	tokenHash := []byte("refresh-token-hash")
	revokedAt := time.Date(2026, time.July, 22, 13, 0, 0, 0, time.UTC)

	t.Run("fills revoked_at timestamp", func(t *testing.T) {
		repository, mock := newMockRefreshTokenRepository(t)
		mock.ExpectExec(regexp.QuoteMeta(revokeRefreshTokenQuery)).
			WithArgs(tokenHash, revokedAt).
			WillReturnResult(sqlmock.NewResult(0, 1))

		if err := repository.Revoke(context.Background(), tokenHash, revokedAt); err != nil {
			t.Fatalf("Revoke() unexpected error: %v", err)
		}
	})

	t.Run("second revocation remains idempotent", func(t *testing.T) {
		repository, mock := newMockRefreshTokenRepository(t)
		mock.ExpectExec(regexp.QuoteMeta(revokeRefreshTokenQuery)).
			WithArgs(tokenHash, revokedAt).
			WillReturnResult(sqlmock.NewResult(0, 0))

		if err := repository.Revoke(context.Background(), tokenHash, revokedAt); err != nil {
			t.Fatalf("Revoke() unexpected error: %v", err)
		}
	})

	t.Run("wraps update error", func(t *testing.T) {
		repository, mock := newMockRefreshTokenRepository(t)
		dbErr := errors.New("update failed")
		mock.ExpectExec(regexp.QuoteMeta(revokeRefreshTokenQuery)).
			WithArgs(tokenHash, revokedAt).
			WillReturnError(dbErr)

		err := repository.Revoke(context.Background(), tokenHash, revokedAt)
		assertRefreshTokenRepositoryError(t, err, dbErr, "revoke refresh token")
	})
}

func newMockRefreshTokenRepository(t *testing.T) (*SQLRefreshTokenRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet database expectations: %v", err)
		}
		mock.ExpectClose()
		if err := db.Close(); err != nil {
			t.Errorf("close mock database: %v", err)
		}
	})

	return NewSQLRefreshTokenRepository(db), mock
}

func assertRefreshTokenRepositoryError(t *testing.T, err error, cause error, message string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if cause != nil && !errors.Is(err, cause) {
		t.Fatalf("error %v does not wrap %v", err, cause)
	}
	if !strings.Contains(err.Error(), message) {
		t.Fatalf("error %q does not contain %q", err, message)
	}
}
