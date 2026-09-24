package domain

import (
	"testing"
	"time"
)

func TestEmailVerificationChallengeStatus(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.September, 10, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		challenge EmailVerificationChallenge
		want      EmailVerificationChallengeStatus
	}{
		{
			name:      "pending without terminal timestamps",
			challenge: EmailVerificationChallenge{},
			want:      EmailVerificationChallengePending,
		},
		{
			name: "confirmed",
			challenge: EmailVerificationChallenge{
				ConfirmedAt: timePointer(now),
			},
			want: EmailVerificationChallengeConfirmed,
		},
		{
			name: "consumed after confirmation",
			challenge: EmailVerificationChallenge{
				ConfirmedAt: timePointer(now),
				ConsumedAt:  timePointer(now.Add(time.Minute)),
			},
			want: EmailVerificationChallengeConsumed,
		},
		{
			name: "invalidated",
			challenge: EmailVerificationChallenge{
				InvalidatedAt: timePointer(now),
			},
			want: EmailVerificationChallengeInvalidated,
		},
		{
			name: "invalidation takes precedence over confirmation",
			challenge: EmailVerificationChallenge{
				ConfirmedAt:   timePointer(now),
				InvalidatedAt: timePointer(now.Add(time.Minute)),
			},
			want: EmailVerificationChallengeInvalidated,
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

func TestEmailVerificationChallengeIsCodeExpired(t *testing.T) {
	t.Parallel()

	expiresAt := time.Date(
		2026,
		time.September,
		10,
		12,
		15,
		0,
		0,
		time.UTC,
	)

	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{
			name: "valid before expiration",
			now:  expiresAt.Add(-time.Nanosecond),
			want: false,
		},
		{
			name: "expired at expiration",
			now:  expiresAt,
			want: true,
		},
		{
			name: "expired after expiration",
			now:  expiresAt.Add(time.Nanosecond),
			want: true,
		},
		{
			name: "compares equivalent instants in different locations",
			now:  expiresAt.In(time.FixedZone("UTC-3", -3*60*60)),
			want: true,
		},
	}

	challenge := EmailVerificationChallenge{
		CodeExpiresAt: expiresAt,
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := challenge.IsCodeExpired(tt.now); got != tt.want {
				t.Fatalf(
					"IsCodeExpired(%v) = %v, want %v",
					tt.now,
					got,
					tt.want,
				)
			}
		})
	}
}

func TestEmailVerificationChallengeIsVerificationTokenExpired(
	t *testing.T,
) {
	t.Parallel()

	expiresAt := time.Date(
		2026,
		time.September,
		11,
		12,
		0,
		0,
		0,
		time.UTC,
	)

	tests := []struct {
		name      string
		expiresAt *time.Time
		now       time.Time
		want      bool
	}{
		{
			name:      "token not emitted",
			expiresAt: nil,
			now:       expiresAt,
			want:      false,
		},
		{
			name:      "valid before expiration",
			expiresAt: timePointer(expiresAt),
			now:       expiresAt.Add(-time.Nanosecond),
			want:      false,
		},
		{
			name:      "expired at expiration",
			expiresAt: timePointer(expiresAt),
			now:       expiresAt,
			want:      true,
		},
		{
			name:      "expired after expiration",
			expiresAt: timePointer(expiresAt),
			now:       expiresAt.Add(time.Nanosecond),
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			challenge := EmailVerificationChallenge{
				VerificationTokenExpiresAt: tt.expiresAt,
			}

			if got := challenge.IsVerificationTokenExpired(tt.now); got != tt.want {
				t.Fatalf(
					"IsVerificationTokenExpired(%v) = %v, want %v",
					tt.now,
					got,
					tt.want,
				)
			}
		})
	}
}

func timePointer(value time.Time) *time.Time {
	return &value
}
