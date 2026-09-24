package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
	"golang.org/x/crypto/bcrypt"
)

func TestSQLPasswordResetRepositoryConcurrentReplacementAndRotation(
	t *testing.T,
) {
	databaseURL := os.Getenv("FOUNDRY_STACK_TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("FOUNDRY_STACK_TEST_DATABASE_URL is not configured")
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Errorf("db.Close() error = %v", err)
		}
	})
	db.SetMaxOpenConns(4)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("database unavailable: %v", err)
	}

	userID := uuid.New()
	email := "password-reset-" + uuid.NewString() + "@example.test"
	now := time.Now().UTC().Truncate(time.Microsecond)

	oldPasswordHash, err := bcrypt.GenerateFromPassword([]byte("old-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash old password: %v", err)
	}
	_, err = db.ExecContext(
		ctx,
		`INSERT INTO users (
			id, name, email, password_hash, email_verified_at
		) VALUES ($1, $2, $3, $4, $5)`,
		userID,
		"Password Reset Test",
		email,
		string(oldPasswordHash),
		now,
	)
	if err != nil {
		t.Fatalf("insert test user: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cleanupCancel()

		if _, err := db.ExecContext(
			cleanupCtx,
			`DELETE FROM users WHERE id = $1`,
			userID,
		); err != nil {
			t.Errorf("delete test user: %v", err)
		}
	})

	repository := NewSQLPasswordResetRepository(db)
	inputs := []domain.ReplacePasswordResetChallengeDTO{
		{
			PasswordResetID: uuid.New(),
			UserID:          userID,
			CodeHash:        []byte("code-" + uuid.NewString()),
			CreatedAt:       now.Add(time.Second),
			CodeExpiresAt:   now.Add(16 * time.Minute),
		},
		{
			PasswordResetID: uuid.New(),
			UserID:          userID,
			CodeHash:        []byte("code-" + uuid.NewString()),
			CreatedAt:       now.Add(time.Second),
			CodeExpiresAt:   now.Add(16 * time.Minute),
		},
	}

	start := make(chan struct{})
	errorsCh := make(chan error, len(inputs))
	var workers sync.WaitGroup
	workers.Add(len(inputs))

	for _, input := range inputs {
		go func() {
			defer workers.Done()
			<-start
			_, err := repository.Replace(ctx, input)
			errorsCh <- err
		}()
	}

	close(start)
	workers.Wait()
	close(errorsCh)

	for err := range errorsCh {
		if err != nil {
			t.Fatalf("concurrent Replace() error = %v", err)
		}
	}

	var total int
	var active int
	var activeID uuid.UUID
	err = db.QueryRowContext(
		ctx,
		`SELECT
			COUNT(*),
			COUNT(*) FILTER (
				WHERE consumed_at IS NULL AND invalidated_at IS NULL
			),
			MAX(password_reset_id::text) FILTER (
				WHERE consumed_at IS NULL AND invalidated_at IS NULL
			)::uuid
		FROM password_reset_challenges
		WHERE user_id = $1`,
		userID,
	).Scan(&total, &active, &activeID)
	if err != nil {
		t.Fatalf("inspect concurrent replacements: %v", err)
	}
	if total != 2 || active != 1 {
		t.Fatalf("replacement counts = total %d, active %d; want 2, 1", total, active)
	}

	firstConfirmedAt := now.Add(3 * time.Second)
	first, err := repository.Confirm(
		ctx,
		domain.ConfirmPasswordResetChallengeDTO{
			PasswordResetID:             activeID,
			PasswordResetTokenHash:      []byte("token-" + uuid.NewString()),
			PasswordResetTokenExpiresAt: now.Add(18 * time.Minute),
			ConfirmedAt:                 firstConfirmedAt,
			MaxAttempts:                 5,
		},
	)
	if err != nil {
		t.Fatalf("first Confirm() error = %v", err)
	}

	rotatedTokenHash := []byte("token-" + uuid.NewString())
	second, err := repository.Confirm(
		ctx,
		domain.ConfirmPasswordResetChallengeDTO{
			PasswordResetID:             activeID,
			PasswordResetTokenHash:      rotatedTokenHash,
			PasswordResetTokenExpiresAt: now.Add(19 * time.Minute),
			ConfirmedAt:                 now.Add(4 * time.Second),
			MaxAttempts:                 5,
		},
	)
	if err != nil {
		t.Fatalf("rotating Confirm() error = %v", err)
	}
	if second.ConfirmedAt == nil || !second.ConfirmedAt.Equal(*first.ConfirmedAt) {
		t.Fatalf("rotating Confirm() changed confirmed_at: first=%v second=%v", first.ConfirmedAt, second.ConfirmedAt)
	}
	if string(second.PasswordResetTokenHash) != string(rotatedTokenHash) {
		t.Fatal("rotating Confirm() did not replace token hash")
	}

	_, err = repository.RecordFailedAttempt(
		ctx,
		domain.RecordPasswordResetFailureDTO{
			PasswordResetID: activeID,
			FailedAt:        now.Add(5 * time.Second),
			MaxAttempts:     5,
		},
	)
	if !errors.Is(err, domain.ErrInvalidPasswordReset) {
		t.Fatalf("RecordFailedAttempt() error = %v, want invalid reset", err)
	}

	for _, clientID := range []string{"mobile-ios", "mobile-android"} {
		_, err = db.ExecContext(ctx, `INSERT INTO refresh_tokens (token_hash, user_id, client_id, expires_at) VALUES ($1, $2, $3, $4)`, []byte(uuid.NewString()), userID, clientID, now.Add(24*time.Hour))
		if err != nil {
			t.Fatalf("insert refresh token: %v", err)
		}
	}
	newPasswordHash, err := bcrypt.GenerateFromPassword([]byte("new-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("hash new password: %v", err)
	}
	if err := repository.Complete(ctx, domain.CompletePasswordResetDTO{PasswordResetID: activeID, PasswordResetTokenHash: rotatedTokenHash, PasswordHash: string(newPasswordHash), CompletedAt: now.Add(6 * time.Second)}); err != nil {
		t.Fatalf("Complete() error = %v", err)
	}
	var storedHash string
	if err := db.QueryRowContext(ctx, `SELECT password_hash FROM users WHERE id = $1`, userID).Scan(&storedHash); err != nil {
		t.Fatalf("read password hash: %v", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("old-password")) == nil {
		t.Fatal("old password still authenticates")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte("new-password")); err != nil {
		t.Fatalf("new password does not authenticate: %v", err)
	}
	var activeRefreshTokens int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1 AND revoked_at IS NULL`, userID).Scan(&activeRefreshTokens); err != nil {
		t.Fatalf("count refresh tokens: %v", err)
	}
	if activeRefreshTokens != 0 {
		t.Fatalf("active refresh tokens = %d, want 0", activeRefreshTokens)
	}
	if err := repository.Complete(ctx, domain.CompletePasswordResetDTO{PasswordResetID: activeID, PasswordResetTokenHash: rotatedTokenHash, PasswordHash: string(newPasswordHash), CompletedAt: now.Add(7 * time.Second)}); !errors.Is(err, domain.ErrInvalidPasswordReset) {
		t.Fatalf("repeated Complete() error = %v", err)
	}
}
