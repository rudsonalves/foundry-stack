package infrastructure

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

func TestMemoryEmailVerificationRateLimiterCooldown(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	limiter := newTestEmailVerificationRateLimiter()

	reserveEmailVerificationRequest(t, limiter, "user@example.com", "203.0.113.1", now)

	_, err := limiter.Reserve(context.Background(), domain.ReserveEmailVerificationRequestDTO{
		Email:     "user@example.com",
		ClientIP:  "203.0.113.2",
		CreatedAt: now.Add(59 * time.Second),
	})
	if !errors.Is(err, domain.ErrEmailVerificationCooldown) {
		t.Fatalf("Reserve() error = %v, want cooldown", err)
	}

	_, err = limiter.Reserve(context.Background(), domain.ReserveEmailVerificationRequestDTO{
		Email:     "user@example.com",
		ClientIP:  "203.0.113.2",
		CreatedAt: now.Add(time.Minute),
	})
	if err != nil {
		t.Fatalf("Reserve() at cooldown boundary error = %v", err)
	}
}

func TestMemoryEmailVerificationRateLimiterEmailLimit(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	limiter := newTestEmailVerificationRateLimiter()

	for index := 0; index < 5; index++ {
		reserveEmailVerificationRequest(
			t,
			limiter,
			"user@example.com",
			"203.0.113."+string(rune('1'+index)),
			now.Add(time.Duration(index)*time.Minute),
		)
	}

	_, err := limiter.Reserve(context.Background(), domain.ReserveEmailVerificationRequestDTO{
		Email:     "user@example.com",
		ClientIP:  "203.0.113.20",
		CreatedAt: now.Add(5 * time.Minute),
	})
	if !errors.Is(err, domain.ErrEmailVerificationRateLimit) {
		t.Fatalf("Reserve() error = %v, want email rate limit", err)
	}
}

func TestMemoryEmailVerificationRateLimiterIPLimit(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	limiter := newTestEmailVerificationRateLimiter()

	for index := 0; index < 10; index++ {
		reserveEmailVerificationRequest(
			t,
			limiter,
			"user"+string(rune('a'+index))+"@example.com",
			"203.0.113.1",
			now,
		)
	}

	_, err := limiter.Reserve(context.Background(), domain.ReserveEmailVerificationRequestDTO{
		Email:     "another@example.com",
		ClientIP:  "203.0.113.1",
		CreatedAt: now,
	})
	if !errors.Is(err, domain.ErrEmailVerificationRateLimit) {
		t.Fatalf("Reserve() error = %v, want IP rate limit", err)
	}
}

func TestMemoryEmailVerificationRateLimiterReleaseEmail(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	limiter := newTestEmailVerificationRateLimiter()

	reservationID := reserveEmailVerificationRequest(
		t,
		limiter,
		"user@example.com",
		"203.0.113.1",
		now,
	)

	if err := limiter.ReleaseEmail(context.Background(), reservationID); err != nil {
		t.Fatalf("ReleaseEmail() error = %v", err)
	}

	_, err := limiter.Reserve(context.Background(), domain.ReserveEmailVerificationRequestDTO{
		Email:     "user@example.com",
		ClientIP:  "203.0.113.2",
		CreatedAt: now,
	})
	if err != nil {
		t.Fatalf("Reserve() after release error = %v", err)
	}
}

func TestMemoryEmailVerificationRateLimiterConcurrentCooldown(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	limiter := newTestEmailVerificationRateLimiter()

	const requests = 20
	errorsByRequest := make(chan error, requests)
	var waitGroup sync.WaitGroup

	for index := 0; index < requests; index++ {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			_, err := limiter.Reserve(context.Background(), domain.ReserveEmailVerificationRequestDTO{
				Email:     "user@example.com",
				ClientIP:  uuid.NewString(),
				CreatedAt: now,
			})
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
		case errors.Is(err, domain.ErrEmailVerificationCooldown):
			cooldowns++
		default:
			t.Fatalf("unexpected Reserve() error = %v", err)
		}
	}

	if successes != 1 || cooldowns != requests-1 {
		t.Fatalf("successes = %d, cooldowns = %d", successes, cooldowns)
	}
}

func newTestEmailVerificationRateLimiter() *MemoryEmailVerificationRateLimiter {
	return NewMemoryEmailVerificationRateLimiter(
		MemoryEmailVerificationRateLimiterConfig{
			Cooldown:          time.Minute,
			EmailLimitPerHour: 5,
			IPLimitPerHour:    10,
		},
	)
}

func reserveEmailVerificationRequest(
	t *testing.T,
	limiter *MemoryEmailVerificationRateLimiter,
	email string,
	clientIP string,
	createdAt time.Time,
) uuid.UUID {
	t.Helper()

	reservationID, err := limiter.Reserve(
		context.Background(),
		domain.ReserveEmailVerificationRequestDTO{
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
