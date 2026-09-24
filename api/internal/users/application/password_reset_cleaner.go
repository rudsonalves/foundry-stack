package application

import (
	"context"
	"log"
	"time"
)

type PasswordResetStaleDeleter interface {
	DeleteStale(context.Context, time.Time) (int64, error)
}

type PasswordResetCleaner struct {
	resets    PasswordResetStaleDeleter
	clock     domainClock
	retention time.Duration
	interval  time.Duration
	logger    *log.Logger
}

type domainClock interface{ Now() time.Time }

func NewPasswordResetCleaner(resets PasswordResetStaleDeleter, clock domainClock, retention, interval time.Duration, logger *log.Logger) *PasswordResetCleaner {
	return &PasswordResetCleaner{resets: resets, clock: clock, retention: retention, interval: interval, logger: logger}
}

func (c *PasswordResetCleaner) Cleanup(ctx context.Context) {
	_, err := c.resets.DeleteStale(ctx, c.clock.Now().UTC().Add(-c.retention))
	if err != nil && c.logger != nil {
		c.logger.Printf("password reset cleanup failed")
	}
}

func (c *PasswordResetCleaner) Run(ctx context.Context) {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.Cleanup(ctx)
		}
	}
}
