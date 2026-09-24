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

type SQLUserRepository struct {
	db *sql.DB
}

func NewSQLUserRepository(db *sql.DB) *SQLUserRepository {
	return &SQLUserRepository{db: db}
}

var _ domain.UserRepository = (*SQLUserRepository)(nil)

func (s *SQLUserRepository) Create(
	ctx context.Context,
	input domain.CreateUserDTO,
) (domain.User, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.User{}, fmt.Errorf(
			"begin create user: %w",
			err,
		)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	const existingUserQuery = `
		SELECT 1
		FROM users
		WHERE email = $1
	`

	var existing int
	err = tx.QueryRowContext(
		ctx,
		existingUserQuery,
		input.Email,
	).Scan(&existing)

	switch {
	case err == nil:
		return domain.User{}, fmt.Errorf(
			"create user: %w",
			domain.ErrEmailAlreadyExists,
		)
	case !errors.Is(err, sql.ErrNoRows):
		return domain.User{}, fmt.Errorf(
			"check existing user: %w",
			translateUserError(err),
		)
	}

	const verificationQuery = `
		SELECT confirmed_at
		FROM email_verification_challenges
		WHERE verification_token_hash = $1
			AND email = $2
			AND confirmed_at IS NOT NULL
			AND consumed_at IS NULL
			AND invalidated_at IS NULL
			AND verification_token_expires_at > $3
		FOR UPDATE
	`

	var emailVerifiedAt time.Time

	err = tx.QueryRowContext(
		ctx,
		verificationQuery,
		input.VerificationTokenHash,
		input.Email,
		input.Now.UTC(),
	).Scan(&emailVerifiedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, fmt.Errorf(
			"validate email verification: %w",
			domain.ErrInvalidEmailVerification,
		)
	}
	if err != nil {
		return domain.User{}, fmt.Errorf(
			"validate email verification: %w",
			err,
		)
	}

	user := domain.User{
		ID:              uuid.New(),
		Name:            input.Name,
		Email:           input.Email,
		PasswordHash:    input.PasswordHash,
		EmailVerifiedAt: emailVerifiedAt,
	}

	const insertQuery = `
		INSERT INTO users (
			id,
			name,
			email,
			password_hash,
			email_verified_at
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`

	err = tx.QueryRowContext(
		ctx,
		insertQuery,
		user.ID,
		user.Name,
		user.Email,
		user.PasswordHash,
		user.EmailVerifiedAt,
	).Scan(
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return domain.User{}, fmt.Errorf(
			"insert user: %w",
			translateUserError(err),
		)
	}

	const consumeQuery = `
		UPDATE email_verification_challenges
		SET
			consumed_at = $2,
			updated_at = $2
		WHERE verification_token_hash = $1
			AND consumed_at IS NULL
	`

	result, err := tx.ExecContext(
		ctx,
		consumeQuery,
		input.VerificationTokenHash,
		input.Now.UTC(),
	)
	if err != nil {
		return domain.User{}, fmt.Errorf(
			"consume email verification: %w",
			err,
		)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return domain.User{}, fmt.Errorf(
			"read consumed email verification count: %w",
			err,
		)
	}
	if affected != 1 {
		return domain.User{}, fmt.Errorf(
			"consume email verification: %w",
			domain.ErrInvalidEmailVerification,
		)
	}

	if err := tx.Commit(); err != nil {
		return domain.User{}, fmt.Errorf(
			"commit create user: %w",
			err,
		)
	}

	return user, nil
}

func (s *SQLUserRepository) FindByID(
	ctx context.Context,
	id uuid.UUID,
) (domain.User, error) {
	const query = `
		SELECT
			id,
			name,
			email,
			COALESCE(password_hash, '') AS password_hash,
			created_at,
			updated_at
		FROM users
		WHERE id = $1
	`

	var user domain.User
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, fmt.Errorf(
			"user %s not found: %w",
			id,
			translateUserError(err),
		)
	}
	if err != nil {
		return domain.User{}, fmt.Errorf(
			"query user by id %s: %w",
			id,
			translateUserError(err),
		)
	}

	return user, nil
}

func (s *SQLUserRepository) FindByEmail(
	ctx context.Context,
	email string,
) (domain.User, error) {
	const query = `
		SELECT
			id,
			name,
			email,
			COALESCE(password_hash, '') AS password_hash,
			created_at,
			updated_at
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := s.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, fmt.Errorf(
			"user %s not found: %w",
			email,
			translateUserError(err),
		)
	}
	if err != nil {
		return domain.User{}, fmt.Errorf(
			"query user by email %s: %w",
			email,
			translateUserError(err),
		)
	}

	return user, nil
}

func (s *SQLUserRepository) List(
	ctx context.Context,
) ([]domain.User, error) {
	const query = `
		SELECT id, name, email, created_at, updated_at
		FROM users
		ORDER BY created_at, id
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf(
			"query users: %w",
			translateUserError(err),
		)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID,
			&user.Name,
			&user.Email,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf(
				"scan user: %w",
				translateUserError(err),
			)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"iterate users: %w",
			translateUserError(err),
		)
	}

	return users, nil
}

func (s *SQLUserRepository) Update(
	ctx context.Context,
	id uuid.UUID,
	name string,
	email string,
) (domain.User, error) {
	const query = `
		UPDATE users
		SET name = $2,
			email = $3,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, email, created_at, updated_at
	`

	var user domain.User
	err := s.db.QueryRowContext(
		ctx,
		query,
		id,
		name,
		email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.User{}, fmt.Errorf(
			"user %s not found: %w",
			id,
			translateUserError(err),
		)
	}
	if err != nil {
		return domain.User{}, fmt.Errorf(
			"update user %s: %w",
			id,
			translateUserError(err),
		)
	}

	return user, nil
}

func (s *SQLUserRepository) Delete(
	ctx context.Context,
	id uuid.UUID,
) error {
	const query = `
		DELETE FROM users
		WHERE id = $1
	`

	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf(
			"delete user %s: %w",
			id,
			translateUserError(err),
		)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"delete user %s: %w",
			id,
			translateUserError(err),
		)
	}
	if affected == 0 {
		return fmt.Errorf(
			"user %s not found: %w",
			id,
			domain.ErrUserNotFound,
		)
	}

	return nil
}

func translateUserError(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return domain.ErrUserNotFound
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) &&
		pgErr.Code == "23505" &&
		pgErr.ConstraintName == "users_email_unique" {
		return domain.ErrEmailAlreadyExists
	}

	return fmt.Errorf("access users: %w", err)
}
