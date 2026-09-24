package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

type SQLPasswordResetRepository struct {
	db *sql.DB
}

func NewSQLPasswordResetRepository(db *sql.DB) *SQLPasswordResetRepository {
	return &SQLPasswordResetRepository{db: db}
}

var _ domain.PasswordResetRepository = (*SQLPasswordResetRepository)(nil)
var _ domain.PasswordResetCompleter = (*SQLPasswordResetRepository)(nil)

func (r *SQLPasswordResetRepository) Complete(
	ctx context.Context,
	input domain.CompletePasswordResetDTO,
) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin complete password reset: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const lockQuery = `
		SELECT user_id
		FROM password_reset_challenges
		WHERE password_reset_id = $1
			AND password_reset_token_hash = $2
			AND confirmed_at IS NOT NULL
			AND consumed_at IS NULL
			AND invalidated_at IS NULL
			AND password_reset_token_expires_at > $3
		FOR UPDATE
	`
	var userID uuid.UUID
	err = tx.QueryRowContext(
		ctx,
		lockQuery,
		input.PasswordResetID,
		input.PasswordResetTokenHash,
		input.CompletedAt,
	).Scan(&userID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("lock password reset completion: %w", domain.ErrInvalidPasswordReset)
	}
	if err != nil {
		return fmt.Errorf("lock password reset completion: %w", err)
	}

	if _, err = tx.ExecContext(
		ctx,
		`UPDATE users SET password_hash = $2, updated_at = $3 WHERE id = $1`,
		userID,
		input.PasswordHash,
		input.CompletedAt,
	); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	result, err := tx.ExecContext(
		ctx,
		`UPDATE password_reset_challenges SET consumed_at = $2, updated_at = $2 WHERE password_reset_id = $1 AND consumed_at IS NULL`,
		input.PasswordResetID,
		input.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("consume password reset: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read consumed password reset count: %w", err)
	}
	if affected != 1 {
		return fmt.Errorf("consume password reset: %w", domain.ErrInvalidPasswordReset)
	}

	if _, err = tx.ExecContext(
		ctx,
		`UPDATE refresh_tokens SET revoked_at = $2 WHERE user_id = $1 AND revoked_at IS NULL`,
		userID,
		input.CompletedAt,
	); err != nil {
		return fmt.Errorf("revoke refresh tokens: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit password reset completion: %w", err)
	}
	return nil
}

func (r *SQLPasswordResetRepository) FindByID(
	ctx context.Context,
	passwordResetID uuid.UUID,
) (domain.PasswordResetChallenge, error) {
	const query = `
		SELECT
			password_reset_id,
			user_id,
			code_hash,
			password_reset_token_hash,
			failed_attempts,
			created_at,
			updated_at,
			code_expires_at,
			password_reset_token_expires_at,
			confirmed_at,
			consumed_at,
			invalidated_at
		FROM password_reset_challenges
		WHERE password_reset_id = $1
	`

	challenge, err := scanPasswordResetChallenge(
		r.db.QueryRowContext(ctx, query, passwordResetID),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PasswordResetChallenge{}, fmt.Errorf(
			"password reset %s not found: %w",
			passwordResetID,
			domain.ErrPasswordResetNotFound,
		)
	}
	if err != nil {
		return domain.PasswordResetChallenge{}, fmt.Errorf(
			"query password reset %s: %w",
			passwordResetID,
			err,
		)
	}

	return challenge, nil
}

func (r *SQLPasswordResetRepository) Replace(
	ctx context.Context,
	input domain.ReplacePasswordResetChallengeDTO,
) (domain.PasswordResetChallenge, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.PasswordResetChallenge{}, fmt.Errorf(
			"begin replace password reset: %w",
			err,
		)
	}
	defer func() { _ = tx.Rollback() }()

	const lockQuery = `
		SELECT pg_advisory_xact_lock(hashtextextended($1, 0))
	`
	if _, err := tx.ExecContext(
		ctx,
		lockQuery,
		input.UserID.String(),
	); err != nil {
		return domain.PasswordResetChallenge{}, fmt.Errorf(
			"lock password reset replacement: %w",
			err,
		)
	}

	const invalidateQuery = `
		UPDATE password_reset_challenges
		SET
			invalidated_at = $2,
			updated_at = $2
		WHERE user_id = $1
			AND consumed_at IS NULL
			AND invalidated_at IS NULL
	`
	if _, err := tx.ExecContext(
		ctx,
		invalidateQuery,
		input.UserID,
		input.CreatedAt,
	); err != nil {
		return domain.PasswordResetChallenge{}, fmt.Errorf(
			"invalidate previous password reset: %w",
			err,
		)
	}

	const insertQuery = `
		INSERT INTO password_reset_challenges (
			password_reset_id,
			user_id,
			code_hash,
			failed_attempts,
			created_at,
			updated_at,
			code_expires_at
		)
		VALUES ($1, $2, $3, 0, $4, $4, $5)
		RETURNING
			password_reset_id,
			user_id,
			code_hash,
			password_reset_token_hash,
			failed_attempts,
			created_at,
			updated_at,
			code_expires_at,
			password_reset_token_expires_at,
			confirmed_at,
			consumed_at,
			invalidated_at
	`

	challenge, err := scanPasswordResetChallenge(
		tx.QueryRowContext(
			ctx,
			insertQuery,
			input.PasswordResetID,
			input.UserID,
			input.CodeHash,
			input.CreatedAt,
			input.CodeExpiresAt,
		),
	)
	if err != nil {
		return domain.PasswordResetChallenge{}, fmt.Errorf(
			"insert password reset: %w",
			translatePasswordResetConflict(err),
		)
	}

	if err := tx.Commit(); err != nil {
		return domain.PasswordResetChallenge{}, fmt.Errorf(
			"commit password reset replacement: %w",
			err,
		)
	}

	return challenge, nil
}

func (r *SQLPasswordResetRepository) RecordFailedAttempt(
	ctx context.Context,
	input domain.RecordPasswordResetFailureDTO,
) (domain.PasswordResetChallenge, error) {
	const query = `
		UPDATE password_reset_challenges
		SET
			failed_attempts = failed_attempts + 1,
			invalidated_at = CASE
				WHEN failed_attempts + 1 >= $3 THEN $2
				ELSE invalidated_at
			END,
			updated_at = $2
		WHERE password_reset_id = $1
			AND confirmed_at IS NULL
			AND consumed_at IS NULL
			AND invalidated_at IS NULL
			AND code_expires_at > $2
			AND failed_attempts < $3
		RETURNING
			password_reset_id,
			user_id,
			code_hash,
			password_reset_token_hash,
			failed_attempts,
			created_at,
			updated_at,
			code_expires_at,
			password_reset_token_expires_at,
			confirmed_at,
			consumed_at,
			invalidated_at
	`

	challenge, err := scanPasswordResetChallenge(
		r.db.QueryRowContext(
			ctx,
			query,
			input.PasswordResetID,
			input.FailedAt,
			input.MaxAttempts,
		),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PasswordResetChallenge{}, fmt.Errorf(
			"record failed password reset attempt: %w",
			domain.ErrInvalidPasswordReset,
		)
	}
	if err != nil {
		return domain.PasswordResetChallenge{}, fmt.Errorf(
			"record failed password reset attempt: %w",
			err,
		)
	}

	return challenge, nil
}

func (r *SQLPasswordResetRepository) Confirm(
	ctx context.Context,
	input domain.ConfirmPasswordResetChallengeDTO,
) (domain.PasswordResetChallenge, error) {
	const query = `
		UPDATE password_reset_challenges
		SET
			password_reset_token_hash = $2,
			password_reset_token_expires_at = $3,
			confirmed_at = COALESCE(confirmed_at, $4),
			updated_at = $4
		WHERE password_reset_id = $1
			AND consumed_at IS NULL
			AND invalidated_at IS NULL
			AND code_expires_at > $4
			AND failed_attempts < $5
		RETURNING
			password_reset_id,
			user_id,
			code_hash,
			password_reset_token_hash,
			failed_attempts,
			created_at,
			updated_at,
			code_expires_at,
			password_reset_token_expires_at,
			confirmed_at,
			consumed_at,
			invalidated_at
	`

	challenge, err := scanPasswordResetChallenge(
		r.db.QueryRowContext(
			ctx,
			query,
			input.PasswordResetID,
			input.PasswordResetTokenHash,
			input.PasswordResetTokenExpiresAt,
			input.ConfirmedAt,
			input.MaxAttempts,
		),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.PasswordResetChallenge{}, fmt.Errorf(
			"confirm password reset: %w",
			domain.ErrInvalidPasswordReset,
		)
	}
	if err != nil {
		return domain.PasswordResetChallenge{}, fmt.Errorf(
			"confirm password reset: %w",
			translatePasswordResetConflict(err),
		)
	}

	return challenge, nil
}

func (r *SQLPasswordResetRepository) Invalidate(
	ctx context.Context,
	input domain.InvalidatePasswordResetChallengeDTO,
) error {
	const query = `
		UPDATE password_reset_challenges
		SET
			invalidated_at = $2,
			updated_at = $2
		WHERE password_reset_id = $1
			AND consumed_at IS NULL
			AND invalidated_at IS NULL
	`

	result, err := r.db.ExecContext(
		ctx,
		query,
		input.PasswordResetID,
		input.InvalidatedAt,
	)
	if err != nil {
		return fmt.Errorf("invalidate password reset: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read invalidated password reset count: %w", err)
	}
	if affected == 0 {
		return fmt.Errorf(
			"invalidate password reset: %w",
			domain.ErrInvalidPasswordReset,
		)
	}

	return nil
}

func (r *SQLPasswordResetRepository) DeleteStale(
	ctx context.Context,
	before time.Time,
) (int64, error) {
	const query = `
		DELETE FROM password_reset_challenges
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
				AND password_reset_token_expires_at <= $1
			)
	`

	result, err := r.db.ExecContext(ctx, query, before.UTC())
	if err != nil {
		return 0, fmt.Errorf("delete stale password resets: %w", err)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("read deleted password reset count: %w", err)
	}

	return deleted, nil
}

type passwordResetScanner interface {
	Scan(dest ...any) error
}

func scanPasswordResetChallenge(
	scanner passwordResetScanner,
) (domain.PasswordResetChallenge, error) {
	var challenge domain.PasswordResetChallenge
	var tokenExpiresAt sql.NullTime
	var confirmedAt sql.NullTime
	var consumedAt sql.NullTime
	var invalidatedAt sql.NullTime

	err := scanner.Scan(
		&challenge.PasswordResetID,
		&challenge.UserID,
		&challenge.CodeHash,
		&challenge.PasswordResetTokenHash,
		&challenge.FailedAttempts,
		&challenge.CreatedAt,
		&challenge.UpdatedAt,
		&challenge.CodeExpiresAt,
		&tokenExpiresAt,
		&confirmedAt,
		&consumedAt,
		&invalidatedAt,
	)
	if err != nil {
		return domain.PasswordResetChallenge{}, err
	}

	challenge.PasswordResetTokenExpiresAt = nullableTimePointer(tokenExpiresAt)
	challenge.ConfirmedAt = nullableTimePointer(confirmedAt)
	challenge.ConsumedAt = nullableTimePointer(consumedAt)
	challenge.InvalidatedAt = nullableTimePointer(invalidatedAt)

	return challenge, nil
}

func translatePasswordResetConflict(err error) error {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return err
	}

	switch pgErr.ConstraintName {
	case "password_reset_challenges_active_user_idx",
		"password_reset_challenges_code_hash_unique",
		"password_reset_challenges_token_hash_unique":
		return fmt.Errorf("%w: %s", domain.ErrPasswordResetConflict, pgErr.ConstraintName)
	default:
		return err
	}
}
