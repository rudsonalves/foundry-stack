package infrastructure

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestNewAccessClaims(t *testing.T) {
	userID := uuid.MustParse("d0f424f7-6654-4f4e-a0f5-b70dc73a36be")
	issuedAt := time.Date(2026, time.July, 21, 12, 0, 0, 0, time.UTC)
	expiresAt := issuedAt.Add(15 * time.Minute)

	claims := newAccessClaims(
		userID,
		"foundry-stack-mobile",
		"foundry-stack",
		"foundry-stack-api",
		issuedAt,
		expiresAt,
	)

	if claims.ClientID != "foundry-stack-mobile" {
		t.Fatalf("ClientID = %q, want %q", claims.ClientID, "foundry-stack-mobile")
	}
	if claims.Issuer != "foundry-stack" {
		t.Fatalf("Issuer = %q, want %q", claims.Issuer, "foundry-stack")
	}
	if claims.Subject != userID.String() {
		t.Fatalf("Subject = %q, want %q", claims.Subject, userID.String())
	}
	if len(claims.Audience) != 1 || claims.Audience[0] != "foundry-stack-api" {
		t.Fatalf("Audience = %v, want [%q]", claims.Audience, "foundry-stack-api")
	}
	if claims.IssuedAt == nil || !claims.IssuedAt.Time.Equal(issuedAt) {
		t.Fatalf("IssuedAt = %v, want %v", claims.IssuedAt, issuedAt)
	}
	if claims.ExpiresAt == nil || !claims.ExpiresAt.Time.Equal(expiresAt) {
		t.Fatalf("ExpiresAt = %v, want %v", claims.ExpiresAt, expiresAt)
	}
}
