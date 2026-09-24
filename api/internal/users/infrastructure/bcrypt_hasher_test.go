package infrastructure

import (
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestBcryptHasherHashAndCompare(t *testing.T) {
	hasher := NewBcryptHasher(bcrypt.MinCost)

	hash, err := hasher.Hash("secure123")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}
	if hash == "" {
		t.Fatal("Hash() returned empty hash")
	}

	if err := hasher.Compare(hash, "secure123"); err != nil {
		t.Fatalf("Compare() error = %v", err)
	}
}

func TestBcryptHasherCompareWrongPassword(t *testing.T) {
	hasher := NewBcryptHasher(bcrypt.MinCost)
	hash, err := hasher.Hash("secure123")
	if err != nil {
		t.Fatalf("Hash() error = %v", err)
	}

	err = hasher.Compare(hash, "other-password")
	if err == nil {
		t.Fatal("expected compare error for wrong password")
	}
}

func TestBcryptHasherHashInvalidCost(t *testing.T) {
	hasher := NewBcryptHasher(bcrypt.MaxCost + 1)

	_, err := hasher.Hash("secure123")
	if err == nil {
		t.Fatal("expected hash error for invalid cost")
	}
	if !strings.Contains(err.Error(), "hash password") {
		t.Fatalf("error %q does not contain %q", err, "hash password")
	}
}
