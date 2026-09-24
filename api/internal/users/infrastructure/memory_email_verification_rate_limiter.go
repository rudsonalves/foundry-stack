package infrastructure

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

type MemoryEmailVerificationRateLimiterConfig struct {
	Cooldown          time.Duration
	EmailLimitPerHour int
	IPLimitPerHour    int
}

type emailVerificationRateLimitEvent struct {
	ID                uuid.UUID
	Email             string
	ClientIP          string
	CreatedAt         time.Time
	CountsTowardEmail bool
}

type MemoryEmailVerificationRateLimiter struct {
	mu     sync.Mutex
	config MemoryEmailVerificationRateLimiterConfig
	events []emailVerificationRateLimitEvent
}

var _ domain.EmailVerificationRateLimiter = (*MemoryEmailVerificationRateLimiter)(nil)

func NewMemoryEmailVerificationRateLimiter(
	config MemoryEmailVerificationRateLimiterConfig,
) *MemoryEmailVerificationRateLimiter {
	return &MemoryEmailVerificationRateLimiter{
		config: config,
		events: make([]emailVerificationRateLimitEvent, 0),
	}
}

func (l *MemoryEmailVerificationRateLimiter) Reserve(
	ctx context.Context,
	input domain.ReserveEmailVerificationRequestDTO,
) (uuid.UUID, error) {
	if err := ctx.Err(); err != nil {
		return uuid.Nil, err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.removeExpiredEvents(input.CreatedAt)

	emailCount := 0
	ipCount := 0
	var latestEmailRequest time.Time

	for _, event := range l.events {
		if event.ClientIP == input.ClientIP {
			ipCount++
		}

		if event.Email != input.Email || !event.CountsTowardEmail {
			continue
		}

		emailCount++

		if event.CreatedAt.After(latestEmailRequest) {
			latestEmailRequest = event.CreatedAt
		}
	}

	if !latestEmailRequest.IsZero() &&
		input.CreatedAt.Before(
			latestEmailRequest.Add(l.config.Cooldown),
		) {
		return uuid.Nil, domain.ErrEmailVerificationCooldown
	}

	if emailCount >= l.config.EmailLimitPerHour {
		return uuid.Nil, domain.ErrEmailVerificationRateLimit
	}

	if ipCount >= l.config.IPLimitPerHour {
		return uuid.Nil, domain.ErrEmailVerificationRateLimit
	}

	reservationID := uuid.New()

	l.events = append(
		l.events,
		emailVerificationRateLimitEvent{
			ID:                reservationID,
			Email:             input.Email,
			ClientIP:          input.ClientIP,
			CreatedAt:         input.CreatedAt,
			CountsTowardEmail: true,
		},
	)

	return reservationID, nil
}

func (l *MemoryEmailVerificationRateLimiter) ReleaseEmail(
	ctx context.Context,
	reservationID uuid.UUID,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	for index := range l.events {
		if l.events[index].ID == reservationID {
			l.events[index].CountsTowardEmail = false
			return nil
		}
	}

	return errors.New(
		"email verification rate limit reservation not found",
	)
}

func (l *MemoryEmailVerificationRateLimiter) removeExpiredEvents(
	now time.Time,
) {
	cutoff := now.Add(-time.Hour)
	active := l.events[:0]

	for _, event := range l.events {
		if event.CreatedAt.After(cutoff) {
			active = append(active, event)
		}
	}

	l.events = active
}
