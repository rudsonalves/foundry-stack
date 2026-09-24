package domain

import "testing"

func TestDomainErrors(t *testing.T) {
	if ErrClientNotAllowed == nil {
		t.Fatal("ErrClientNotAllowed should be initialized")
	}
	if ErrInvalidRefreshToken == nil {
		t.Fatal("ErrInvalidRefreshToken should be initialized")
	}
	if ErrClientNotAllowed.Error() != "client not allowed" {
		t.Fatalf("ErrClientNotAllowed = %q, want %q", ErrClientNotAllowed.Error(), "client not allowed")
	}
	if ErrInvalidRefreshToken.Error() != "refresh token inválido" {
		t.Fatalf("ErrInvalidRefreshToken = %q, want %q", ErrInvalidRefreshToken.Error(), "refresh token inválido")
	}
}
