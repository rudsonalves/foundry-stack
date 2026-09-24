package application

import (
	"context"
	"errors"
	"testing"
	"time"
)

type staleDeleterStub struct {
	calls  int
	before time.Time
}

func (s *staleDeleterStub) DeleteStale(_ context.Context, before time.Time) (int64, error) {
	s.calls++
	s.before = before
	if s.calls == 1 {
		return 0, errors.New("temporary")
	}
	return 1, nil
}

func TestPasswordResetCleanerRetriesAfterFailure(t *testing.T) {
	now := time.Date(2026, 9, 23, 12, 0, 0, 0, time.UTC)
	repo := &staleDeleterStub{}
	cleaner := NewPasswordResetCleaner(repo, resetClock{now}, 24*time.Hour, time.Hour, nil)
	cleaner.Cleanup(context.Background())
	cleaner.Cleanup(context.Background())
	if repo.calls != 2 || !repo.before.Equal(now.Add(-24*time.Hour)) {
		t.Fatalf("calls=%d before=%v", repo.calls, repo.before)
	}
}
