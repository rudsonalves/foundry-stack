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
)

func TestSQLUserRepositoryCreatePreventsConcurrentDuplicates(
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

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("database unavailable: %v", err)
	}

	now := time.Now().UTC()
	verificationID := uuid.New()
	email := "concurrent-" + uuid.NewString() + "@example.test"
	tokenHash := []byte("token-" + uuid.NewString())
	codeHash := []byte("code-" + uuid.NewString())

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cleanupCancel()

		if _, err := db.ExecContext(
			cleanupCtx,
			`DELETE FROM users WHERE email = $1`,
			email,
		); err != nil {
			t.Errorf("delete test user: %v", err)
		}

		if _, err := db.ExecContext(
			cleanupCtx,
			`
				DELETE FROM email_verification_challenges
				WHERE verification_id = $1
			`,
			verificationID,
		); err != nil {
			t.Errorf("delete test verification: %v", err)
		}
	})

	createdAt := now.Add(-time.Minute)
	confirmedAt := now.Add(-30 * time.Second)

	_, err = db.ExecContext(
		ctx,
		`
			INSERT INTO email_verification_challenges (
				verification_id,
				email,
				code_hash,
				verification_token_hash,
				failed_attempts,
				created_at,
				updated_at,
				code_expires_at,
				verification_token_expires_at,
				confirmed_at
			)
			VALUES ($1, $2, $3, $4, 0, $5, $6, $7, $8, $9)
		`,
		verificationID,
		email,
		codeHash,
		tokenHash,
		createdAt,
		confirmedAt,
		now.Add(15*time.Minute),
		now.Add(24*time.Hour),
		confirmedAt,
	)
	if err != nil {
		t.Fatalf("insert verification challenge: %v", err)
	}

	repository := NewSQLUserRepository(db)
	input := domain.CreateUserDTO{
		Name:                  "Concurrent User",
		Email:                 email,
		PasswordHash:          "hashed-password",
		VerificationTokenHash: tokenHash,
		Now:                   now,
	}

	type createResult struct {
		user domain.User
		err  error
	}

	start := make(chan struct{})
	results := make(chan createResult, 2)

	var workers sync.WaitGroup
	workers.Add(2)

	for range 2 {
		go func() {
			defer workers.Done()

			<-start

			user, err := repository.Create(ctx, input)
			results <- createResult{
				user: user,
				err:  err,
			}
		}()
	}

	close(start)
	workers.Wait()
	close(results)

	successes := 0
	failures := 0

	for result := range results {
		if result.err == nil {
			successes++

			if result.user.ID == uuid.Nil {
				t.Error("successful Create() returned nil UUID")
			}

			continue
		}

		failures++

		if !errors.Is(result.err, domain.ErrEmailAlreadyExists) &&
			!errors.Is(result.err, domain.ErrInvalidEmailVerification) {
			t.Errorf("unexpected concurrent Create() error = %v", result.err)
		}
	}

	if successes != 1 {
		t.Fatalf("successful creations = %d, want 1", successes)
	}
	if failures != 1 {
		t.Fatalf("failed creations = %d, want 1", failures)
	}

	var userCount int
	err = db.QueryRowContext(
		ctx,
		`SELECT COUNT(*) FROM users WHERE email = $1`,
		email,
	).Scan(&userCount)
	if err != nil {
		t.Fatalf("count users: %v", err)
	}
	if userCount != 1 {
		t.Fatalf("persisted users = %d, want 1", userCount)
	}

	var consumed bool
	err = db.QueryRowContext(
		ctx,
		`
			SELECT consumed_at IS NOT NULL
			FROM email_verification_challenges
			WHERE verification_id = $1
		`,
		verificationID,
	).Scan(&consumed)
	if err != nil {
		t.Fatalf("query verification consumption: %v", err)
	}
	if !consumed {
		t.Fatal("verification was not consumed")
	}
}
