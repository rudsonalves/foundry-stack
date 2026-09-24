package infrastructure

import (
	"bytes"
	"encoding/base64"
	"strconv"
	"testing"

	"github.com/google/uuid"
)

func TestSecureEmailVerificationCodeGenerator(t *testing.T) {
	generator := NewSecureEmailVerificationCodeGenerator()

	for range 100 {
		code, err := generator.GenerateCode()
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}

		if len(code) != 6 {
			t.Fatalf("GenerateCode() = %q, want 6 digits", code)
		}

		if _, err := strconv.Atoi(code); err != nil {
			t.Fatalf("GenerateCode() = %q, want only digits", code)
		}
	}
}

func TestFormatEmailVerificationCodePreservesLeadingZeros(t *testing.T) {
	tests := []struct {
		value int64
		want  string
	}{
		{value: 0, want: "000000"},
		{value: 42, want: "000042"},
		{value: 999999, want: "999999"},
	}

	for _, test := range tests {
		got := formatEmailVerificationCode(test.value)
		if got != test.want {
			t.Errorf(
				"formatEmailVerificationCode(%d) = %q, want %q",
				test.value,
				got,
				test.want,
			)
		}
	}
}

func TestSecureEmailVerificationTokenGenerator(t *testing.T) {
	generator := NewSecureEmailVerificationTokenGenerator()

	first, err := generator.GenerateToken()
	if err != nil {
		t.Fatalf("first GenerateToken() error = %v", err)
	}

	second, err := generator.GenerateToken()
	if err != nil {
		t.Fatalf("second GenerateToken() error = %v", err)
	}

	if first == second {
		t.Fatal("GenerateToken() returned the same token twice")
	}

	rawToken, err := base64.RawURLEncoding.DecodeString(first)
	if err != nil {
		t.Fatalf("decode generated token: %v", err)
	}

	if len(rawToken) != emailVerificationTokenSize {
		t.Errorf(
			"decoded token length = %d, want %d",
			len(rawToken),
			emailVerificationTokenSize,
		)
	}
}

func TestHMACEmailVerificationHasher(t *testing.T) {
	codeSecret := bytes.Repeat([]byte{1}, 32)
	tokenSecret := bytes.Repeat([]byte{2}, 32)

	hasher, err := NewHMACEmailVerificationHasher(
		codeSecret,
		tokenSecret,
	)
	if err != nil {
		t.Fatalf("NewHMACEmailVerificationHasher() error = %v", err)
	}

	email := "user@example.com"
	code := "000042"
	token := "opaque-token"

	verificationID := uuid.New()

	codeHash := hasher.HashCode(verificationID, email, code)
	tokenHash := hasher.HashToken(email, token)

	if !hasher.MatchesCode(
		codeHash,
		verificationID,
		email,
		code,
	) {
		t.Fatal("MatchesCode() = false for original values")
	}

	if hasher.MatchesCode(
		codeHash,
		verificationID,
		"other@example.com",
		code,
	) {
		t.Fatal("MatchesCode() = true for a different email")
	}

	if hasher.MatchesCode(
		codeHash,
		verificationID,
		email,
		"000043",
	) {
		t.Fatal("MatchesCode() = true for a different code")
	}

	if hasher.MatchesCode(
		codeHash,
		uuid.New(),
		email,
		code,
	) {
		t.Fatal("MatchesCode() = true for a different verification ID")
	}

	if !hasher.MatchesToken(tokenHash, email, token) {
		t.Fatal("MatchesToken() = false for the original email and token")
	}

	if hasher.MatchesToken(tokenHash, "other@example.com", token) {
		t.Fatal("MatchesToken() = true for a different email")
	}

	if hasher.MatchesToken(tokenHash, email, "other-token") {
		t.Fatal("MatchesToken() = true for a different token")
	}

	if bytes.Equal(codeHash, tokenHash) {
		t.Fatal("code and token hashes must use distinct contexts")
	}
}

func TestHMACEmailVerificationHasherCopiesSecrets(t *testing.T) {
	codeSecret := bytes.Repeat([]byte{1}, 32)
	tokenSecret := bytes.Repeat([]byte{2}, 32)

	hasher, err := NewHMACEmailVerificationHasher(
		codeSecret,
		tokenSecret,
	)
	if err != nil {
		t.Fatalf("NewHMACEmailVerificationHasher() error = %v", err)
	}

	verificationID := uuid.New()

	before := hasher.HashCode(
		verificationID,
		"user@example.com",
		"123456",
	)

	codeSecret[0] = 9
	tokenSecret[0] = 9

	after := hasher.HashCode(
		verificationID,
		"user@example.com",
		"123456",
	)

	if !bytes.Equal(before, after) {
		t.Fatal("hash changed after mutating the supplied secrets")
	}
}

func TestHMACEmailVerificationHasherRejectsSharedSecret(t *testing.T) {
	secret := bytes.Repeat([]byte{1}, 32)

	_, err := NewHMACEmailVerificationHasher(secret, secret)
	if err == nil {
		t.Fatal("NewHMACEmailVerificationHasher() error = nil")
	}
}

func TestEmailVerificationTokenRotationProducesAnotherHash(t *testing.T) {
	generator := NewSecureEmailVerificationTokenGenerator()
	hasher, err := NewHMACEmailVerificationHasher(
		bytes.Repeat([]byte{1}, 32),
		bytes.Repeat([]byte{2}, 32),
	)
	if err != nil {
		t.Fatalf("NewHMACEmailVerificationHasher() error = %v", err)
	}

	firstToken, err := generator.GenerateToken()
	if err != nil {
		t.Fatalf("first GenerateToken() error = %v", err)
	}

	secondToken, err := generator.GenerateToken()
	if err != nil {
		t.Fatalf("second GenerateToken() error = %v", err)
	}

	firstHash := hasher.HashToken("user@example.com", firstToken)
	secondHash := hasher.HashToken("user@example.com", secondToken)

	if bytes.Equal(firstHash, secondHash) {
		t.Fatal("rotated token produced the same hash")
	}
}
