package domain

import (
	"testing"
	"time"
)

func TestPasswordResetChallengeStatus(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.September, 23, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		challenge PasswordResetChallenge
		want      PasswordResetChallengeStatus
	}{
		{
			name: "pending",
			want: PasswordResetChallengePending,
		},
		{
			name: "confirmed",
			challenge: PasswordResetChallenge{
				ConfirmedAt: passwordResetTimePointer(now),
			},
			want: PasswordResetChallengeConfirmed,
		},
		{
			name: "consumed",
			challenge: PasswordResetChallenge{
				ConfirmedAt: passwordResetTimePointer(now),
				ConsumedAt:  passwordResetTimePointer(now.Add(time.Minute)),
			},
			want: PasswordResetChallengeConsumed,
		},
		{
			name: "invalidated",
			challenge: PasswordResetChallenge{
				InvalidatedAt: passwordResetTimePointer(now),
			},
			want: PasswordResetChallengeInvalidated,
		},
		{
			name: "invalidation takes precedence",
			challenge: PasswordResetChallenge{
				ConfirmedAt:   passwordResetTimePointer(now),
				InvalidatedAt: passwordResetTimePointer(now.Add(time.Minute)),
			},
			want: PasswordResetChallengeInvalidated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.challenge.Status(); got != tt.want {
				t.Fatalf("Status() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPasswordResetChallengeExpiration(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2026, time.September, 23, 12, 15, 0, 0, time.UTC)

	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{name: "valid before expiration", now: expiresAt.Add(-time.Nanosecond)},
		{name: "expired at expiration", now: expiresAt, want: true},
		{name: "expired after expiration", now: expiresAt.Add(time.Nanosecond), want: true},
		{
			name: "same instant in another location",
			now:  expiresAt.In(time.FixedZone("UTC-3", -3*60*60)),
			want: true,
		},
	}

	challenge := PasswordResetChallenge{CodeExpiresAt: expiresAt}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := challenge.IsCodeExpired(tt.now); got != tt.want {
				t.Errorf("IsCodeExpired(%v) = %v, want %v", tt.now, got, tt.want)
			}
		})
	}
}

func TestPasswordResetChallengeTokenExpiration(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(2026, time.September, 23, 12, 15, 0, 0, time.UTC)

	tests := []struct {
		name      string
		expiresAt *time.Time
		now       time.Time
		want      bool
	}{
		{name: "token not emitted", now: expiresAt},
		{
			name:      "valid before expiration",
			expiresAt: passwordResetTimePointer(expiresAt),
			now:       expiresAt.Add(-time.Nanosecond),
		},
		{
			name:      "expired at expiration",
			expiresAt: passwordResetTimePointer(expiresAt),
			now:       expiresAt,
			want:      true,
		},
		{
			name:      "expired after expiration",
			expiresAt: passwordResetTimePointer(expiresAt),
			now:       expiresAt.Add(time.Nanosecond),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			challenge := PasswordResetChallenge{
				PasswordResetTokenExpiresAt: tt.expiresAt,
			}

			if got := challenge.IsPasswordResetTokenExpired(tt.now); got != tt.want {
				t.Errorf(
					"IsPasswordResetTokenExpired(%v) = %v, want %v",
					tt.now,
					got,
					tt.want,
				)
			}
		})
	}
}

func passwordResetTimePointer(value time.Time) *time.Time {
	return &value
}
