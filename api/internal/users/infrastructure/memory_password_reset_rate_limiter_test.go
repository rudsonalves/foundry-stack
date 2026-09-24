package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

func TestMemoryPasswordResetRateLimiterCooldownAndNormalization(t *testing.T) {
	t.Parallel()

	now := passwordResetRateLimitTime()
	limiter := newTestPasswordResetRateLimiter()

	reservePasswordResetRequest(
		t,
		limiter,
		" User@Example.COM ",
		"203.0.113.1",
		now,
	)

	_, err := limiter.Reserve(
		context.Background(),
		domain.ReservePasswordResetRequestDTO{
			Email:     "user@example.com",
			ClientIP:  "203.0.113.2",
			CreatedAt: now.Add(59 * time.Second),
		},
	)
	if !errors.Is(err, domain.ErrPasswordResetCooldown) {
		t.Fatalf("Reserve() error = %v, want cooldown", err)
	}

	_, err = limiter.Reserve(
		context.Background(),
		domain.ReservePasswordResetRequestDTO{
			Email:     "user@example.com",
			ClientIP:  "203.0.113.2",
			CreatedAt: now.Add(time.Minute),
		},
	)
	if err != nil {
		t.Fatalf("Reserve() at cooldown boundary error = %v", err)
	}
}

func TestMemoryPasswordResetRateLimiterEmailLimit(t *testing.T) {
	t.Parallel()

	now := passwordResetRateLimitTime()
	limiter := newTestPasswordResetRateLimiter()

	for index := 0; index < 5; index++ {
		reservePasswordResetRequest(
			t,
			limiter,
			"user@example.com",
			fmt.Sprintf("203.0.113.%d", index+1),
			now.Add(time.Duration(index)*time.Minute),
		)
	}

	_, err := limiter.Reserve(
		context.Background(),
		domain.ReservePasswordResetRequestDTO{
			Email:     "user@example.com",
			ClientIP:  "203.0.113.20",
			CreatedAt: now.Add(5 * time.Minute),
		},
	)
	if !errors.Is(err, domain.ErrPasswordResetRateLimit) {
		t.Fatalf("Reserve() error = %v, want email rate limit", err)
	}
}

func TestMemoryPasswordResetRateLimiterIPLimit(t *testing.T) {
	t.Parallel()

	now := passwordResetRateLimitTime()
	limiter := newTestPasswordResetRateLimiter()

	for index := 0; index < 10; index++ {
		reservePasswordResetRequest(
			t,
			limiter,
			fmt.Sprintf("user%d@example.com", index),
			"203.0.113.1",
			now,
		)
	}

	_, err := limiter.Reserve(
		context.Background(),
		domain.ReservePasswordResetRequestDTO{
			Email:     "another@example.com",
			ClientIP:  "203.0.113.1",
			CreatedAt: now,
		},
	)
	if !errors.Is(err, domain.ErrPasswordResetRateLimit) {
		t.Fatalf("Reserve() error = %v, want IP rate limit", err)
	}
}

func TestMemoryPasswordResetRateLimiterDoesNotDependOnAccountExistence(
	t *testing.T,
) {
	t.Parallel()

	now := passwordResetRateLimitTime()

	for _, name := range []string{"existing account", "missing account"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			limiter := newTestPasswordResetRateLimiter()
			email := "same-public-behavior@example.com"

			reservePasswordResetRequest(t, limiter, email, "203.0.113.1", now)
			_, err := limiter.Reserve(
				context.Background(),
				domain.ReservePasswordResetRequestDTO{
					Email:     email,
					ClientIP:  "203.0.113.2",
					CreatedAt: now.Add(30 * time.Second),
				},
			)
			if !errors.Is(err, domain.ErrPasswordResetCooldown) {
				t.Fatalf("Reserve() error = %v, want cooldown", err)
			}
		})
	}
}

func TestMemoryPasswordResetRateLimiterReleaseAndExpiration(t *testing.T) {
	t.Parallel()

	now := passwordResetRateLimitTime()
	limiter := newTestPasswordResetRateLimiter()
	reservationID := reservePasswordResetRequest(
		t,
		limiter,
		"user@example.com",
		"203.0.113.1",
		now,
	)

	if err := limiter.Release(context.Background(), reservationID); err != nil {
		t.Fatalf("Release() error = %v", err)
	}

	reservePasswordResetRequest(
		t,
		limiter,
		"user@example.com",
		"203.0.113.1",
		now,
	)

	_, err := limiter.Reserve(
		context.Background(),
		domain.ReservePasswordResetRequestDTO{
			Email:     "user@example.com",
			ClientIP:  "203.0.113.1",
			CreatedAt: now.Add(time.Hour),
		},
	)
	if err != nil {
		t.Fatalf("Reserve() at hourly boundary error = %v", err)
	}

	if err := limiter.Release(context.Background(), uuid.New()); err == nil {
		t.Fatal("Release() missing reservation error = nil")
	}
}

func TestMemoryPasswordResetRateLimiterHonorsContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	limiter := newTestPasswordResetRateLimiter()

	_, err := limiter.Reserve(ctx, domain.ReservePasswordResetRequestDTO{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Reserve() error = %v, want context canceled", err)
	}

	err = limiter.Release(ctx, uuid.New())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("Release() error = %v, want context canceled", err)
	}
}

func TestMemoryPasswordResetRateLimiterConcurrentCooldown(t *testing.T) {
	t.Parallel()

	now := passwordResetRateLimitTime()
	limiter := newTestPasswordResetRateLimiter()
	const requests = 20

	errorsByRequest := make(chan error, requests)
	var waitGroup sync.WaitGroup
	waitGroup.Add(requests)

	for range requests {
		go func() {
			defer waitGroup.Done()
			_, err := limiter.Reserve(
				context.Background(),
				domain.ReservePasswordResetRequestDTO{
					Email:     "user@example.com",
					ClientIP:  uuid.NewString(),
					CreatedAt: now,
				},
			)
			errorsByRequest <- err
		}()
	}

	waitGroup.Wait()
	close(errorsByRequest)

	successes := 0
	cooldowns := 0
	for err := range errorsByRequest {
		switch {
		case err == nil:
			successes++
		case errors.Is(err, domain.ErrPasswordResetCooldown):
			cooldowns++
		default:
			t.Fatalf("unexpected Reserve() error = %v", err)
		}
	}

	if successes != 1 || cooldowns != requests-1 {
		t.Fatalf("successes = %d, cooldowns = %d", successes, cooldowns)
	}
}

func newTestPasswordResetRateLimiter() *MemoryPasswordResetRateLimiter {
	return NewMemoryPasswordResetRateLimiter(
		MemoryPasswordResetRateLimiterConfig{
			Cooldown:          time.Minute,
			EmailLimitPerHour: 5,
			IPLimitPerHour:    10,
		},
	)
}

func reservePasswordResetRequest(
	t *testing.T,
	limiter *MemoryPasswordResetRateLimiter,
	email string,
	clientIP string,
	createdAt time.Time,
) uuid.UUID {
	t.Helper()

	reservationID, err := limiter.Reserve(
		context.Background(),
		domain.ReservePasswordResetRequestDTO{
			Email:     email,
			ClientIP:  clientIP,
			CreatedAt: createdAt,
		},
	)
	if err != nil {
		t.Fatalf("Reserve() error = %v", err)
	}

	return reservationID
}

func passwordResetRateLimitTime() time.Time {
	return time.Date(2026, time.September, 23, 12, 0, 0, 0, time.UTC)
}
