package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

const (
	findPasswordResetPattern       = `(?s)SELECT.*FROM password_reset_challenges.*WHERE password_reset_id = \$1`
	lockPasswordResetPattern       = `(?s)SELECT pg_advisory_xact_lock.*hashtextextended`
	invalidatePreviousResetPattern = `(?s)UPDATE password_reset_challenges.*invalidated_at = \$2.*WHERE user_id = \$1`
	insertPasswordResetPattern     = `(?s)INSERT INTO password_reset_challenges.*RETURNING`
	recordResetFailurePattern      = `(?s)UPDATE password_reset_challenges.*failed_attempts = failed_attempts \+ 1`
	confirmPasswordResetPattern    = `(?s)UPDATE password_reset_challenges.*password_reset_token_hash = \$2`
	invalidatePasswordResetPattern = `(?s)UPDATE password_reset_challenges.*invalidated_at = \$2.*WHERE password_reset_id = \$1`
	deleteStaleResetsPattern       = `(?s)DELETE FROM password_reset_challenges.*consumed_at IS NOT NULL`
	lockResetCompletionPattern     = `(?s)SELECT user_id.*password_reset_token_hash = \$2.*FOR UPDATE`
	updateResetPasswordPattern     = `UPDATE users SET password_hash = \$2, updated_at = \$3 WHERE id = \$1`
	consumeResetPattern            = `UPDATE password_reset_challenges SET consumed_at = \$2, updated_at = \$2 WHERE password_reset_id = \$1 AND consumed_at IS NULL`
	revokeUserRefreshPattern       = `UPDATE refresh_tokens SET revoked_at = \$2 WHERE user_id = \$1 AND revoked_at IS NULL`
)

func TestSQLPasswordResetRepositoryComplete(t *testing.T) {
	id, userID := uuid.New(), uuid.New()
	now := time.Date(2026, time.September, 23, 12, 0, 0, 0, time.UTC)
	input := domain.CompletePasswordResetDTO{PasswordResetID: id, PasswordResetTokenHash: []byte("proof-hash"), PasswordHash: "new-password-hash", CompletedAt: now}

	t.Run("updates password consumes proof and revokes all sessions atomically", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		mock.ExpectBegin()
		mock.ExpectQuery(lockResetCompletionPattern).WithArgs(id, input.PasswordResetTokenHash, now).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(userID))
		mock.ExpectExec(updateResetPasswordPattern).WithArgs(userID, input.PasswordHash, now).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(consumeResetPattern).WithArgs(id, now).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(revokeUserRefreshPattern).WithArgs(userID, now).WillReturnResult(sqlmock.NewResult(0, 3))
		mock.ExpectCommit()
		if err := repository.Complete(context.Background(), input); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("rejects consumed or invalid proof", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		mock.ExpectBegin()
		mock.ExpectQuery(lockResetCompletionPattern).WithArgs(id, input.PasswordResetTokenHash, now).WillReturnError(sql.ErrNoRows)
		mock.ExpectRollback()
		err := repository.Complete(context.Background(), input)
		if !errors.Is(err, domain.ErrInvalidPasswordReset) {
			t.Fatalf("error=%v", err)
		}
	})

	t.Run("rolls back every change when session revocation fails", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		mock.ExpectBegin()
		mock.ExpectQuery(lockResetCompletionPattern).WithArgs(id, input.PasswordResetTokenHash, now).WillReturnRows(sqlmock.NewRows([]string{"user_id"}).AddRow(userID))
		mock.ExpectExec(updateResetPasswordPattern).WithArgs(userID, input.PasswordHash, now).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(consumeResetPattern).WithArgs(id, now).WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(revokeUserRefreshPattern).WithArgs(userID, now).WillReturnError(errors.New("revoke failed"))
		mock.ExpectRollback()
		if err := repository.Complete(context.Background(), input); err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestSQLPasswordResetRepositoryFindByID(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	userID := uuid.New()
	now := time.Date(2026, time.September, 23, 12, 0, 0, 0, time.UTC)

	t.Run("finds pending challenge", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		mock.ExpectQuery(findPasswordResetPattern).
			WithArgs(id).
			WillReturnRows(passwordResetRows().AddRow(
				id, userID, []byte("code-hash"), nil, 0,
				now, now, now.Add(15*time.Minute), nil, nil, nil, nil,
			))

		challenge, err := repository.FindByID(context.Background(), id)
		if err != nil {
			t.Fatalf("FindByID() error = %v", err)
		}
		if challenge.PasswordResetID != id || challenge.UserID != userID {
			t.Fatalf("FindByID() challenge = %#v", challenge)
		}
		if challenge.Status() != domain.PasswordResetChallengePending {
			t.Fatalf("Status() = %q", challenge.Status())
		}
	})

	t.Run("reports missing challenge", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		mock.ExpectQuery(findPasswordResetPattern).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		_, err := repository.FindByID(context.Background(), id)
		assertError(t, err, domain.ErrPasswordResetNotFound, "password reset")
	})

	t.Run("wraps database error", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		dbErr := errors.New("database unavailable")
		mock.ExpectQuery(findPasswordResetPattern).
			WithArgs(id).
			WillReturnError(dbErr)

		_, err := repository.FindByID(context.Background(), id)
		assertError(t, err, dbErr, "query password reset")
	})
}

func TestSQLPasswordResetRepositoryReplace(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.September, 23, 12, 0, 0, 0, time.UTC)
	input := domain.ReplacePasswordResetChallengeDTO{
		PasswordResetID: uuid.New(),
		UserID:          uuid.New(),
		CodeHash:        []byte("code-hash"),
		CreatedAt:       now,
		CodeExpiresAt:   now.Add(15 * time.Minute),
	}

	t.Run("invalidates previous challenge and inserts another", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		mock.ExpectBegin()
		mock.ExpectExec(lockPasswordResetPattern).
			WithArgs(input.UserID.String()).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(invalidatePreviousResetPattern).
			WithArgs(input.UserID, input.CreatedAt).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(insertPasswordResetPattern).
			WithArgs(
				input.PasswordResetID,
				input.UserID,
				input.CodeHash,
				input.CreatedAt,
				input.CodeExpiresAt,
			).
			WillReturnRows(passwordResetRows().AddRow(
				input.PasswordResetID, input.UserID, input.CodeHash, nil, 0,
				input.CreatedAt, input.CreatedAt, input.CodeExpiresAt,
				nil, nil, nil, nil,
			))
		mock.ExpectCommit()

		challenge, err := repository.Replace(context.Background(), input)
		if err != nil {
			t.Fatalf("Replace() error = %v", err)
		}
		if challenge.PasswordResetID != input.PasswordResetID {
			t.Fatalf("Replace() challenge = %#v", challenge)
		}
	})

	t.Run("rolls back when insertion fails", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		dbErr := errors.New("insert failed")
		mock.ExpectBegin()
		mock.ExpectExec(lockPasswordResetPattern).
			WithArgs(input.UserID.String()).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(invalidatePreviousResetPattern).
			WithArgs(input.UserID, input.CreatedAt).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(insertPasswordResetPattern).
			WithArgs(
				input.PasswordResetID,
				input.UserID,
				input.CodeHash,
				input.CreatedAt,
				input.CodeExpiresAt,
			).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		_, err := repository.Replace(context.Background(), input)
		assertError(t, err, dbErr, "insert password reset")
	})

	t.Run("translates active challenge conflict", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		conflict := &pgconn.PgError{
			Code:           "23505",
			ConstraintName: "password_reset_challenges_active_user_idx",
		}
		mock.ExpectBegin()
		mock.ExpectExec(lockPasswordResetPattern).
			WithArgs(input.UserID.String()).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(invalidatePreviousResetPattern).
			WithArgs(input.UserID, input.CreatedAt).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(insertPasswordResetPattern).
			WithArgs(
				input.PasswordResetID,
				input.UserID,
				input.CodeHash,
				input.CreatedAt,
				input.CodeExpiresAt,
			).
			WillReturnError(conflict)
		mock.ExpectRollback()

		_, err := repository.Replace(context.Background(), input)
		assertError(t, err, domain.ErrPasswordResetConflict, "insert password reset")
	})
}

func TestSQLPasswordResetRepositoryRecordFailedAttempt(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	userID := uuid.New()
	now := time.Date(2026, time.September, 23, 12, 0, 0, 0, time.UTC)
	input := domain.RecordPasswordResetFailureDTO{
		PasswordResetID: id,
		FailedAt:        now,
		MaxAttempts:     5,
	}

	t.Run("invalidates fifth failed attempt", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		mock.ExpectQuery(recordResetFailurePattern).
			WithArgs(id, now, 5).
			WillReturnRows(passwordResetRows().AddRow(
				id, userID, []byte("code-hash"), nil, 5,
				now.Add(-time.Minute), now, now.Add(time.Minute),
				nil, nil, nil, now,
			))

		challenge, err := repository.RecordFailedAttempt(context.Background(), input)
		if err != nil {
			t.Fatalf("RecordFailedAttempt() error = %v", err)
		}
		if challenge.FailedAttempts != 5 ||
			challenge.Status() != domain.PasswordResetChallengeInvalidated {
			t.Fatalf("RecordFailedAttempt() challenge = %#v", challenge)
		}
	})

	t.Run("maps unavailable challenge to invalid reset", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		mock.ExpectQuery(recordResetFailurePattern).
			WithArgs(id, now, 5).
			WillReturnError(sql.ErrNoRows)

		_, err := repository.RecordFailedAttempt(context.Background(), input)
		assertError(t, err, domain.ErrInvalidPasswordReset, "record failed password reset attempt")
	})
}

func TestSQLPasswordResetRepositoryConfirm(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	userID := uuid.New()
	now := time.Date(2026, time.September, 23, 12, 0, 0, 0, time.UTC)
	input := domain.ConfirmPasswordResetChallengeDTO{
		PasswordResetID:             id,
		PasswordResetTokenHash:      []byte("token-hash"),
		PasswordResetTokenExpiresAt: now.Add(15 * time.Minute),
		ConfirmedAt:                 now,
		MaxAttempts:                 5,
	}

	t.Run("confirms and rotates token hash", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		mock.ExpectQuery(confirmPasswordResetPattern).
			WithArgs(
				id,
				input.PasswordResetTokenHash,
				input.PasswordResetTokenExpiresAt,
				now,
				5,
			).
			WillReturnRows(passwordResetRows().AddRow(
				id, userID, []byte("code-hash"), input.PasswordResetTokenHash, 0,
				now.Add(-time.Minute), now, now.Add(time.Minute),
				input.PasswordResetTokenExpiresAt, now, nil, nil,
			))

		challenge, err := repository.Confirm(context.Background(), input)
		if err != nil {
			t.Fatalf("Confirm() error = %v", err)
		}
		if challenge.Status() != domain.PasswordResetChallengeConfirmed ||
			!bytesEqual(challenge.PasswordResetTokenHash, input.PasswordResetTokenHash) {
			t.Fatalf("Confirm() challenge = %#v", challenge)
		}
	})

	t.Run("maps unavailable challenge to invalid reset", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		mock.ExpectQuery(confirmPasswordResetPattern).
			WithArgs(
				id,
				input.PasswordResetTokenHash,
				input.PasswordResetTokenExpiresAt,
				now,
				5,
			).
			WillReturnError(sql.ErrNoRows)

		_, err := repository.Confirm(context.Background(), input)
		assertError(t, err, domain.ErrInvalidPasswordReset, "confirm password reset")
	})
}

func TestSQLPasswordResetRepositoryInvalidate(t *testing.T) {
	t.Parallel()

	input := domain.InvalidatePasswordResetChallengeDTO{
		PasswordResetID: uuid.New(),
		InvalidatedAt:   time.Now().UTC(),
	}

	t.Run("invalidates active challenge", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		mock.ExpectExec(invalidatePasswordResetPattern).
			WithArgs(input.PasswordResetID, input.InvalidatedAt).
			WillReturnResult(sqlmock.NewResult(0, 1))

		if err := repository.Invalidate(context.Background(), input); err != nil {
			t.Fatalf("Invalidate() error = %v", err)
		}
	})

	t.Run("maps unavailable challenge to invalid reset", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		mock.ExpectExec(invalidatePasswordResetPattern).
			WithArgs(input.PasswordResetID, input.InvalidatedAt).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repository.Invalidate(context.Background(), input)
		assertError(t, err, domain.ErrInvalidPasswordReset, "invalidate password reset")
	})
}

func TestSQLPasswordResetRepositoryDeleteStale(t *testing.T) {
	t.Parallel()

	before := time.Date(2026, time.September, 23, 12, 0, 0, 0, time.UTC)

	t.Run("returns deleted row count", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		mock.ExpectExec(deleteStaleResetsPattern).
			WithArgs(before).
			WillReturnResult(sqlmock.NewResult(0, 4))

		deleted, err := repository.DeleteStale(context.Background(), before)
		if err != nil {
			t.Fatalf("DeleteStale() error = %v", err)
		}
		if deleted != 4 {
			t.Fatalf("DeleteStale() = %d, want 4", deleted)
		}
	})

	t.Run("wraps execution error", func(t *testing.T) {
		repository, mock := newMockPasswordResetRepository(t)
		dbErr := errors.New("delete failed")
		mock.ExpectExec(deleteStaleResetsPattern).
			WithArgs(before).
			WillReturnError(dbErr)

		_, err := repository.DeleteStale(context.Background(), before)
		assertError(t, err, dbErr, "delete stale password resets")
	})
}

func newMockPasswordResetRepository(
	t *testing.T,
) (*SQLPasswordResetRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}

	t.Cleanup(func() {
		mock.ExpectClose()
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() error = %v", err)
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet database expectations: %v", err)
		}
	})

	return NewSQLPasswordResetRepository(db), mock
}

func passwordResetRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"password_reset_id",
		"user_id",
		"code_hash",
		"password_reset_token_hash",
		"failed_attempts",
		"created_at",
		"updated_at",
		"code_expires_at",
		"password_reset_token_expires_at",
		"confirmed_at",
		"consumed_at",
		"invalidated_at",
	})
}

func bytesEqual(left, right []byte) bool {
	return string(left) == string(right)
}
