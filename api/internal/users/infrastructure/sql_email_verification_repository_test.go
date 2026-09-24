package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

const (
	findVerificationPattern        = `(?s)SELECT.*FROM email_verification_challenges.*WHERE verification_id = \$1`
	lockVerificationPattern        = `(?s)SELECT pg_advisory_xact_lock.*hashtextextended`
	invalidateVerificationPattern  = `(?s)UPDATE email_verification_challenges.*invalidated_at = \$2.*WHERE email = \$1`
	insertVerificationPattern      = `(?s)INSERT INTO email_verification_challenges.*RETURNING`
	recordFailurePattern           = `(?s)UPDATE email_verification_challenges.*failed_attempts = failed_attempts \+ 1`
	confirmVerificationPattern     = `(?s)UPDATE email_verification_challenges.*verification_token_hash = \$2`
	deleteStaleVerificationPattern = `(?s)DELETE FROM email_verification_challenges.*consumed_at IS NOT NULL`
)

func TestSQLEmailVerificationRepositoryFindByID(t *testing.T) {
	id := uuid.New()
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)

	t.Run("finds pending challenge", func(t *testing.T) {
		repository, mock := newMockEmailVerificationRepository(t)

		mock.ExpectQuery(findVerificationPattern).
			WithArgs(id).
			WillReturnRows(
				emailVerificationRows().AddRow(
					id,
					"ada@example.com",
					[]byte("code-hash"),
					nil,
					0,
					now,
					now,
					now.Add(15*time.Minute),
					nil,
					nil,
					nil,
					nil,
				),
			)

		challenge, err := repository.FindByID(
			context.Background(),
			id,
		)
		if err != nil {
			t.Fatalf("FindByID() error = %v", err)
		}

		if challenge.VerificationID != id {
			t.Fatalf(
				"VerificationID = %s, want %s",
				challenge.VerificationID,
				id,
			)
		}
		if challenge.Email != "ada@example.com" {
			t.Fatalf("Email = %q", challenge.Email)
		}
		if challenge.Status() != domain.EmailVerificationChallengePending {
			t.Fatalf("Status() = %q", challenge.Status())
		}
	})

	t.Run("reports missing challenge", func(t *testing.T) {
		repository, mock := newMockEmailVerificationRepository(t)

		mock.ExpectQuery(findVerificationPattern).
			WithArgs(id).
			WillReturnError(sql.ErrNoRows)

		_, err := repository.FindByID(context.Background(), id)
		assertError(
			t,
			err,
			domain.ErrEmailVerificationNotFound,
			"email verification",
		)
	})

	t.Run("wraps database error", func(t *testing.T) {
		repository, mock := newMockEmailVerificationRepository(t)
		dbErr := errors.New("database unavailable")

		mock.ExpectQuery(findVerificationPattern).
			WithArgs(id).
			WillReturnError(dbErr)

		_, err := repository.FindByID(context.Background(), id)
		assertError(t, err, dbErr, "query email verification")
	})
}

func TestSQLEmailVerificationRepositoryReplace(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	input := domain.ReplaceEmailVerificationChallengeDTO{
		VerificationID: uuid.New(),
		Email:          "ada@example.com",
		CodeHash:       []byte("code-hash"),
		CreatedAt:      now,
		CodeExpiresAt:  now.Add(15 * time.Minute),
	}

	t.Run("invalidates previous challenge and creates another", func(t *testing.T) {
		repository, mock := newMockEmailVerificationRepository(t)

		mock.ExpectBegin()
		mock.ExpectExec(lockVerificationPattern).
			WithArgs(input.Email).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(invalidateVerificationPattern).
			WithArgs(input.Email, input.CreatedAt).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(insertVerificationPattern).
			WithArgs(
				input.VerificationID,
				input.Email,
				input.CodeHash,
				input.CreatedAt,
				input.CodeExpiresAt,
			).
			WillReturnRows(
				emailVerificationRows().AddRow(
					input.VerificationID,
					input.Email,
					input.CodeHash,
					nil,
					0,
					input.CreatedAt,
					input.CreatedAt,
					input.CodeExpiresAt,
					nil,
					nil,
					nil,
					nil,
				),
			)
		mock.ExpectCommit()

		challenge, err := repository.Replace(
			context.Background(),
			input,
		)
		if err != nil {
			t.Fatalf("Replace() error = %v", err)
		}
		if challenge.VerificationID != input.VerificationID {
			t.Fatalf("Replace() challenge = %#v", challenge)
		}
	})

	t.Run("rolls back when insertion fails", func(t *testing.T) {
		repository, mock := newMockEmailVerificationRepository(t)
		dbErr := errors.New("insert failed")

		mock.ExpectBegin()
		mock.ExpectExec(lockVerificationPattern).
			WithArgs(input.Email).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec(invalidateVerificationPattern).
			WithArgs(input.Email, input.CreatedAt).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectQuery(insertVerificationPattern).
			WithArgs(
				input.VerificationID,
				input.Email,
				input.CodeHash,
				input.CreatedAt,
				input.CodeExpiresAt,
			).
			WillReturnError(dbErr)
		mock.ExpectRollback()

		_, err := repository.Replace(context.Background(), input)
		assertError(t, err, dbErr, "insert email verification")
	})
}

func TestSQLEmailVerificationRepositoryRecordFailedAttempt(
	t *testing.T,
) {
	id := uuid.New()
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	input := domain.RecordEmailVerificationFailureDTO{
		VerificationID: id,
		FailedAt:       now,
		MaxAttempts:    5,
	}

	t.Run("increments and returns invalidated fifth attempt", func(t *testing.T) {
		repository, mock := newMockEmailVerificationRepository(t)

		mock.ExpectQuery(recordFailurePattern).
			WithArgs(id, now, 5).
			WillReturnRows(
				emailVerificationRows().AddRow(
					id,
					"ada@example.com",
					[]byte("code-hash"),
					nil,
					5,
					now.Add(-time.Minute),
					now,
					now.Add(time.Minute),
					nil,
					nil,
					nil,
					now,
				),
			)

		challenge, err := repository.RecordFailedAttempt(
			context.Background(),
			input,
		)
		if err != nil {
			t.Fatalf("RecordFailedAttempt() error = %v", err)
		}
		if challenge.FailedAttempts != 5 {
			t.Fatalf(
				"FailedAttempts = %d, want 5",
				challenge.FailedAttempts,
			)
		}
		if challenge.Status() != domain.EmailVerificationChallengeInvalidated {
			t.Fatalf("Status() = %q", challenge.Status())
		}
	})

	t.Run("maps unavailable challenge to invalid verification", func(t *testing.T) {
		repository, mock := newMockEmailVerificationRepository(t)

		mock.ExpectQuery(recordFailurePattern).
			WithArgs(id, now, 5).
			WillReturnError(sql.ErrNoRows)

		_, err := repository.RecordFailedAttempt(
			context.Background(),
			input,
		)
		assertError(
			t,
			err,
			domain.ErrInvalidEmailVerification,
			"record failed email verification attempt",
		)
	})
}

func TestSQLEmailVerificationRepositoryConfirm(t *testing.T) {
	id := uuid.New()
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	tokenExpiresAt := now.Add(24 * time.Hour)
	input := domain.ConfirmEmailVerificationChallengeDTO{
		VerificationID:             id,
		VerificationTokenHash:      []byte("token-hash"),
		VerificationTokenExpiresAt: tokenExpiresAt,
		ConfirmedAt:                now,
		MaxAttempts:                5,
	}

	t.Run("confirms and stores token hash", func(t *testing.T) {
		repository, mock := newMockEmailVerificationRepository(t)

		mock.ExpectQuery(confirmVerificationPattern).
			WithArgs(
				id,
				input.VerificationTokenHash,
				tokenExpiresAt,
				now,
				5,
			).
			WillReturnRows(
				emailVerificationRows().AddRow(
					id,
					"ada@example.com",
					[]byte("code-hash"),
					input.VerificationTokenHash,
					0,
					now.Add(-time.Minute),
					now,
					now.Add(15*time.Minute),
					tokenExpiresAt,
					now,
					nil,
					nil,
				),
			)

		challenge, err := repository.Confirm(
			context.Background(),
			input,
		)
		if err != nil {
			t.Fatalf("Confirm() error = %v", err)
		}
		if challenge.Status() != domain.EmailVerificationChallengeConfirmed {
			t.Fatalf("Status() = %q", challenge.Status())
		}
		if string(challenge.VerificationTokenHash) != "token-hash" {
			t.Fatalf(
				"VerificationTokenHash = %q",
				challenge.VerificationTokenHash,
			)
		}
	})

	t.Run("maps unavailable challenge to invalid verification", func(t *testing.T) {
		repository, mock := newMockEmailVerificationRepository(t)

		mock.ExpectQuery(confirmVerificationPattern).
			WithArgs(
				id,
				input.VerificationTokenHash,
				tokenExpiresAt,
				now,
				5,
			).
			WillReturnError(sql.ErrNoRows)

		_, err := repository.Confirm(context.Background(), input)
		assertError(
			t,
			err,
			domain.ErrInvalidEmailVerification,
			"confirm email verification",
		)
	})
}

func TestSQLEmailVerificationRepositoryDeleteStale(t *testing.T) {
	before := time.Date(2026, time.September, 10, 12, 0, 0, 0, time.UTC)

	t.Run("returns deleted row count", func(t *testing.T) {
		repository, mock := newMockEmailVerificationRepository(t)

		mock.ExpectExec(deleteStaleVerificationPattern).
			WithArgs(before).
			WillReturnResult(sqlmock.NewResult(0, 4))

		deleted, err := repository.DeleteStale(
			context.Background(),
			before,
		)
		if err != nil {
			t.Fatalf("DeleteStale() error = %v", err)
		}
		if deleted != 4 {
			t.Fatalf("DeleteStale() = %d, want 4", deleted)
		}
	})

	t.Run("wraps execution error", func(t *testing.T) {
		repository, mock := newMockEmailVerificationRepository(t)
		dbErr := errors.New("delete failed")

		mock.ExpectExec(deleteStaleVerificationPattern).
			WithArgs(before).
			WillReturnError(dbErr)

		_, err := repository.DeleteStale(
			context.Background(),
			before,
		)
		assertError(
			t,
			err,
			dbErr,
			"delete stale email verifications",
		)
	})

	t.Run("wraps rows affected error", func(t *testing.T) {
		repository, mock := newMockEmailVerificationRepository(t)
		resultErr := errors.New("result unavailable")

		mock.ExpectExec(deleteStaleVerificationPattern).
			WithArgs(before).
			WillReturnResult(sqlmock.NewErrorResult(resultErr))

		_, err := repository.DeleteStale(
			context.Background(),
			before,
		)
		assertError(
			t,
			err,
			resultErr,
			"read deleted email verification count",
		)
	})
}

func newMockEmailVerificationRepository(
	t *testing.T,
) (*SQLEmailVerificationRepository, sqlmock.Sqlmock) {
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

	return NewSQLEmailVerificationRepository(db), mock
}

func emailVerificationRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"verification_id",
		"email",
		"code_hash",
		"verification_token_hash",
		"failed_attempts",
		"created_at",
		"updated_at",
		"code_expires_at",
		"verification_token_expires_at",
		"confirmed_at",
		"consumed_at",
		"invalidated_at",
	})
}
