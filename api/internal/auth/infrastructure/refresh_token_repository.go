package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rudsonalves/foundry-stack/api/internal/auth/domain"
)

type SQLRefreshTokenRepository struct {
	db *sql.DB
}

func NewSQLRefreshTokenRepository(
	db *sql.DB,
) *SQLRefreshTokenRepository {
	return &SQLRefreshTokenRepository{db: db}
}

var _ domain.RefreshTokenRepository = (*SQLRefreshTokenRepository)(nil)

func (r *SQLRefreshTokenRepository) Create(
	ctx context.Context,
	input domain.CreateRefreshTokenDTO,
) error {
	const query = `
		INSERT INTO refresh_tokens (
			token_hash,
			user_id,
			client_id,
			expires_at
		)
		VALUES ($1, $2, $3, $4)
	`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		input.TokenHash,
		input.UserID,
		input.ClientID,
		input.ExpiresAt,
	); err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}

	return nil
}

func (r *SQLRefreshTokenRepository) FindActive(
	ctx context.Context,
	tokenHash []byte,
	now time.Time,
) (domain.RefreshToken, error) {
	const query = `
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

	var token domain.RefreshToken
	err := r.db.QueryRowContext(
		ctx,
		query,
		tokenHash,
		now,
	).Scan(
		&token.UserID,
		&token.ClientID,
		&token.CreatedAt,
		&token.ExpiresAt,
		&token.RevokedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.RefreshToken{},
			domain.ErrInvalidRefreshToken
	}
	if err != nil {
		return domain.RefreshToken{}, fmt.Errorf(
			"find active refresh token: %w",
			err,
		)
	}

	return token, nil
}

func (r *SQLRefreshTokenRepository) Revoke(
	ctx context.Context,
	tokenHash []byte,
	revokedAt time.Time,
) error {
	const query = `
		UPDATE refresh_tokens
		SET revoked_at = $2
		WHERE token_hash = $1
			AND revoked_at IS NULL
	`

	if _, err := r.db.ExecContext(
		ctx,
		query,
		tokenHash,
		revokedAt,
	); err != nil {
		return fmt.Errorf("revoke refresh token: %w", err)
	}

	return nil
}
