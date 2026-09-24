package infrastructure

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

type MemoryPasswordResetRateLimiterConfig struct {
	Cooldown          time.Duration
	EmailLimitPerHour int
	IPLimitPerHour    int
}

type passwordResetRateLimitEvent struct {
	ID        uuid.UUID
	Email     string
	ClientIP  string
	CreatedAt time.Time
}

type MemoryPasswordResetRateLimiter struct {
	mu     sync.Mutex
	config MemoryPasswordResetRateLimiterConfig
	events []passwordResetRateLimitEvent
}

var _ domain.PasswordResetRateLimiter = (*MemoryPasswordResetRateLimiter)(nil)

func NewMemoryPasswordResetRateLimiter(
	config MemoryPasswordResetRateLimiterConfig,
) *MemoryPasswordResetRateLimiter {
	return &MemoryPasswordResetRateLimiter{
		config: config,
		events: make([]passwordResetRateLimitEvent, 0),
	}
}

func (l *MemoryPasswordResetRateLimiter) Reserve(
	ctx context.Context,
	input domain.ReservePasswordResetRequestDTO,
) (uuid.UUID, error) {
	if err := ctx.Err(); err != nil {
		return uuid.Nil, err
	}

	normalizedEmail := strings.ToLower(strings.TrimSpace(input.Email))
	now := input.CreatedAt.UTC()

	l.mu.Lock()
	defer l.mu.Unlock()

	l.removeExpiredEvents(now)

	emailCount := 0
	ipCount := 0
	var latestEmailRequest time.Time

	for _, event := range l.events {
		if event.ClientIP == input.ClientIP {
			ipCount++
		}

		if event.Email != normalizedEmail {
			continue
		}

		emailCount++
		if event.CreatedAt.After(latestEmailRequest) {
			latestEmailRequest = event.CreatedAt
		}
	}

	if !latestEmailRequest.IsZero() &&
		now.Before(latestEmailRequest.Add(l.config.Cooldown)) {
		return uuid.Nil, domain.ErrPasswordResetCooldown
	}

	if emailCount >= l.config.EmailLimitPerHour ||
		ipCount >= l.config.IPLimitPerHour {
		return uuid.Nil, domain.ErrPasswordResetRateLimit
	}

	reservationID := uuid.New()
	l.events = append(l.events, passwordResetRateLimitEvent{
		ID:        reservationID,
		Email:     normalizedEmail,
		ClientIP:  input.ClientIP,
		CreatedAt: now,
	})

	return reservationID, nil
}

func (l *MemoryPasswordResetRateLimiter) Release(
	ctx context.Context,
	reservationID uuid.UUID,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for index, event := range l.events {
		if event.ID != reservationID {
			continue
		}

		l.events = append(l.events[:index], l.events[index+1:]...)
		return nil
	}

	return errors.New("password reset rate limit reservation not found")
}

func (l *MemoryPasswordResetRateLimiter) removeExpiredEvents(now time.Time) {
	cutoff := now.Add(-time.Hour)
	active := l.events[:0]

	for _, event := range l.events {
		if event.CreatedAt.After(cutoff) {
			active = append(active, event)
		}
	}

	l.events = active
}
