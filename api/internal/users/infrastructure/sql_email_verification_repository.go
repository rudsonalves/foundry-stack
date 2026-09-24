package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

type SQLEmailVerificationRepository struct {
	db *sql.DB
}

func NewSQLEmailVerificationRepository(
	db *sql.DB,
) *SQLEmailVerificationRepository {
	return &SQLEmailVerificationRepository{
		db: db,
	}
}

var _ domain.EmailVerificationRepository = (*SQLEmailVerificationRepository)(nil)

func (s *SQLEmailVerificationRepository) FindByID(
	ctx context.Context,
	verificationID uuid.UUID,
) (domain.EmailVerificationChallenge, error) {
	const query = `
		SELECT
			verification_id,
			email,
			code_hash,
			verification_token_hash,
			failed_attempts,
			created_at,
			updated_at,
			code_expires_at,
			verification_token_expires_at,
			confirmed_at,
			consumed_at,
			invalidated_at
		FROM email_verification_challenges
		WHERE verification_id = $1
	`

	challenge, err := scanEmailVerificationChallenge(
		s.db.QueryRowContext(
			ctx,
			query,
			verificationID,
		),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.EmailVerificationChallenge{}, fmt.Errorf(
			"email verification %s not found: %w",
			verificationID,
			domain.ErrEmailVerificationNotFound,
		)
	}
	if err != nil {
		return domain.EmailVerificationChallenge{}, fmt.Errorf(
			"query email verification %s: %w",
			verificationID,
			err,
		)
	}

	return challenge, nil
}

func (s *SQLEmailVerificationRepository) Replace(
	ctx context.Context,
	input domain.ReplaceEmailVerificationChallengeDTO,
) (domain.EmailVerificationChallenge, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.EmailVerificationChallenge{}, fmt.Errorf(
			"begin replace email verification: %w",
			err,
		)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const lockQuery = `
		SELECT pg_advisory_xact_lock(
			hashtextextended($1, 0)
		)
	`

	if _, err := tx.ExecContext(
		ctx,
		lockQuery,
		input.Email,
	); err != nil {
		return domain.EmailVerificationChallenge{}, fmt.Errorf(
			"lock email verification replacement: %w",
			err,
		)
	}

	const invalidateQuery = `
		UPDATE email_verification_challenges
			SET
				invalidated_at = $2,
				updated_at = $2
			WHERE email = $1
				AND consumed_at IS NULL
				AND invalidated_at IS NULL
	`

	if _, err := tx.ExecContext(
		ctx,
		invalidateQuery,
		input.Email,
		input.CreatedAt,
	); err != nil {
		return domain.EmailVerificationChallenge{}, fmt.Errorf(
			"invalidate previous email verification: %w",
			err,
		)
	}

	const insertQuery = `
		INSERT INTO email_verification_challenges (
			verification_id,
			email,
			code_hash,
			failed_attempts,
			created_at,
			updated_at,
			code_expires_at
		)
		VALUES ($1, $2, $3, 0, $4, $4, $5)
		RETURNING
			verification_id,
			email,
			code_hash,
			verification_token_hash,
			failed_attempts,
			created_at,
			updated_at,
			code_expires_at,
			verification_token_expires_at,
			confirmed_at,
			consumed_at,
			invalidated_at
	`

	challenge, err := scanEmailVerificationChallenge(
		tx.QueryRowContext(
			ctx,
			insertQuery,
			input.VerificationID,
			input.Email,
			input.CodeHash,
			input.CreatedAt,
			input.CodeExpiresAt,
		),
	)
	if err != nil {
		return domain.EmailVerificationChallenge{}, fmt.Errorf(
			"insert email verification: %w",
			err,
		)
	}

	if err := tx.Commit(); err != nil {
		return domain.EmailVerificationChallenge{}, fmt.Errorf(
			"commit email verification replacement: %w",
			err,
		)
	}

	return challenge, nil
}

func (s *SQLEmailVerificationRepository) RecordFailedAttempt(
	ctx context.Context,
	input domain.RecordEmailVerificationFailureDTO,
) (domain.EmailVerificationChallenge, error) {
	const query = `
		UPDATE email_verification_challenges
		SET
			failed_attempts = failed_attempts + 1,
			invalidated_at = CASE
				WHEN failed_attempts + 1 >= $3 THEN $2
				ELSE invalidated_at
			END,
			updated_at = $2
		WHERE verification_id = $1
			AND confirmed_at IS NULL
			AND consumed_at IS NULL
			AND invalidated_at IS NULL
			AND code_expires_at > $2
			AND failed_attempts < $3
		RETURNING
			verification_id,
			email,
			code_hash,
			verification_token_hash,
			failed_attempts,
			created_at,
			updated_at,
			code_expires_at,
			verification_token_expires_at,
			confirmed_at,
			consumed_at,
			invalidated_at
	`

	challenge, err := scanEmailVerificationChallenge(
		s.db.QueryRowContext(
			ctx,
			query,
			input.VerificationID,
			input.FailedAt,
			input.MaxAttempts,
		),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.EmailVerificationChallenge{}, fmt.Errorf(
			"record failed email verification attempt: %w",
			domain.ErrInvalidEmailVerification,
		)
	}
	if err != nil {
		return domain.EmailVerificationChallenge{}, fmt.Errorf(
			"record failed email verification attempt: %w",
			err,
		)
	}

	return challenge, nil
}

func (s *SQLEmailVerificationRepository) Confirm(
	ctx context.Context,
	input domain.ConfirmEmailVerificationChallengeDTO,
) (domain.EmailVerificationChallenge, error) {
	const query = `
		UPDATE email_verification_challenges
		SET
			verification_token_hash = $2,
			verification_token_expires_at = $3,
			confirmed_at = COALESCE(confirmed_at, $4),
			updated_at = $4
		WHERE verification_id = $1
			AND consumed_at IS NULL
			AND invalidated_at IS NULL
			AND code_expires_at > $4
			AND failed_attempts < $5
		RETURNING
			verification_id,
			email,
			code_hash,
			verification_token_hash,
			failed_attempts,
			created_at,
			updated_at,
			code_expires_at,
			verification_token_expires_at,
			confirmed_at,
			consumed_at,
			invalidated_at
	`

	challenge, err := scanEmailVerificationChallenge(
		s.db.QueryRowContext(
			ctx,
			query,
			input.VerificationID,
			input.VerificationTokenHash,
			input.VerificationTokenExpiresAt,
			input.ConfirmedAt,
			input.MaxAttempts,
		),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.EmailVerificationChallenge{}, fmt.Errorf(
			"confirm email verification: %w",
			domain.ErrInvalidEmailVerification,
		)
	}
	if err != nil {
		return domain.EmailVerificationChallenge{}, fmt.Errorf(
			"confirm email verification: %w",
			err,
		)
	}

	return challenge, nil
}

func (s *SQLEmailVerificationRepository) DeleteStale(
	ctx context.Context,
	before time.Time,
) (int64, error) {
	const query = `
		DELETE FROM email_verification_challenges
		WHERE
			(consumed_at IS NOT NULL AND consumed_at <= $1)
			OR
			(invalidated_at IS NOT NULL AND invalidated_at <= $1)
			OR
			(
				confirmed_at IS NULL
				AND code_expires_at <= $1
			)
			OR
			(
				confirmed_at IS NOT NULL
				AND verification_token_expires_at <= $1
			)
	`

	result, err := s.db.ExecContext(
		ctx,
		query,
		before.UTC(),
	)
	if err != nil {
		return 0, fmt.Errorf(
			"delete stale email verifications: %w",
			err,
		)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf(
			"read deleted email verification count: %w",
			err,
		)
	}

	return deleted, nil
}

func (r *SQLEmailVerificationRepository) Invalidate(
	ctx context.Context,
	verificationID uuid.UUID,
	invalidatedAt time.Time,
) error {
	const query = `
		UPDATE email_verification_challenges
		SET
			invalidated_at = $2,
			updated_at = $2
		WHERE verification_id = $1
			AND consumed_at IS NULL
			AND invalidated_at IS NULL
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		verificationID,
		invalidatedAt,
	)
	if err != nil {
		return fmt.Errorf(
			"invalidate email verification: %w",
			err,
		)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"read invalidated email verification count: %w",
			err,
		)
	}

	if affected == 0 {
		return fmt.Errorf(
			"invalidate email verification: %w",
			domain.ErrInvalidEmailVerification,
		)
	}

	return nil
}

// emailVerificationScanner is an interface that abstracts the scanning of
// a single row from the email_verification_challenges table. It is implemented
// by both *sql.Row and *sql.Rows.
type emailVerificationScanner interface {
	Scan(dest ...any) error
}

// scanEmailVerificationChallenge scans a single row from the
// email_verification_challenges table into a domain.EmailVerificationChallenge
// struct. It handles nullable time fields appropriately.
func scanEmailVerificationChallenge(
	scanner emailVerificationScanner,
) (domain.EmailVerificationChallenge, error) {
	var challenge domain.EmailVerificationChallenge
	var verificationTokenExpiresAt sql.NullTime
	var confirmedAt sql.NullTime
	var consumedAt sql.NullTime
	var invalidatedAt sql.NullTime

	err := scanner.Scan(
		&challenge.VerificationID,
		&challenge.Email,
		&challenge.CodeHash,
		&challenge.VerificationTokenHash,
		&challenge.FailedAttempts,
		&challenge.CreatedAt,
		&challenge.UpdatedAt,
		&challenge.CodeExpiresAt,
		&verificationTokenExpiresAt,
		&confirmedAt,
		&consumedAt,
		&invalidatedAt,
	)
	if err != nil {
		return domain.EmailVerificationChallenge{}, err
	}

	challenge.VerificationTokenExpiresAt = nullableTimePointer(
		verificationTokenExpiresAt,
	)
	challenge.ConfirmedAt = nullableTimePointer(confirmedAt)
	challenge.ConsumedAt = nullableTimePointer(consumedAt)
	challenge.InvalidatedAt = nullableTimePointer(invalidatedAt)

	return challenge, nil
}

// nullableTimePointer converts a sql.NullTime to a *time.Time, returning
// nil if the value is not valid.
func nullableTimePointer(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}

	result := value.Time
	return &result
}
