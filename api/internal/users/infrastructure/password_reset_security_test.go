package infrastructure

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

func TestSecurePasswordResetCodeGenerator(t *testing.T) {
	t.Parallel()

	generator := NewSecurePasswordResetCodeGenerator()

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

func TestFormatPasswordResetCodePreservesLeadingZeros(t *testing.T) {
	t.Parallel()

	tests := []struct {
		value int64
		want  string
	}{
		{value: 0, want: "000000"},
		{value: 42, want: "000042"},
		{value: 999999, want: "999999"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()

			if got := formatPasswordResetCode(tt.value); got != tt.want {
				t.Errorf("formatPasswordResetCode(%d) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}

func TestSecurePasswordResetTokenGenerator(t *testing.T) {
	t.Parallel()

	generator := NewSecurePasswordResetTokenGenerator()

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

	if len(rawToken) != passwordResetTokenSize {
		t.Errorf("decoded token length = %d, want %d", len(rawToken), passwordResetTokenSize)
	}
}

func TestHMACPasswordResetHasherBindsValues(t *testing.T) {
	t.Parallel()

	hasher := newTestPasswordResetHasher(t)
	passwordResetID := uuid.New()
	userID := uuid.New()
	code := "000042"
	token := "opaque-token"

	codeHash := hasher.HashCode(passwordResetID, userID, code)
	tokenHash := hasher.HashToken(passwordResetID, userID, token)

	if !hasher.MatchesCode(codeHash, passwordResetID, userID, code) {
		t.Fatal("MatchesCode() = false for original values")
	}

	codeChanges := []struct {
		name            string
		passwordResetID uuid.UUID
		userID          uuid.UUID
		code            string
	}{
		{name: "reset ID", passwordResetID: uuid.New(), userID: userID, code: code},
		{name: "user ID", passwordResetID: passwordResetID, userID: uuid.New(), code: code},
		{name: "code", passwordResetID: passwordResetID, userID: userID, code: "000043"},
	}

	for _, tt := range codeChanges {
		if hasher.MatchesCode(codeHash, tt.passwordResetID, tt.userID, tt.code) {
			t.Errorf("MatchesCode() = true after changing %s", tt.name)
		}
	}

	if !hasher.MatchesToken(tokenHash, passwordResetID, userID, token) {
		t.Fatal("MatchesToken() = false for original values")
	}

	tokenChanges := []struct {
		name            string
		passwordResetID uuid.UUID
		userID          uuid.UUID
		token           string
	}{
		{name: "reset ID", passwordResetID: uuid.New(), userID: userID, token: token},
		{name: "user ID", passwordResetID: passwordResetID, userID: uuid.New(), token: token},
		{name: "token", passwordResetID: passwordResetID, userID: userID, token: "other-token"},
	}

	for _, tt := range tokenChanges {
		if hasher.MatchesToken(tokenHash, tt.passwordResetID, tt.userID, tt.token) {
			t.Errorf("MatchesToken() = true after changing %s", tt.name)
		}
	}

	if bytes.Equal(
		hasher.HashCode(passwordResetID, userID, token),
		hasher.HashToken(passwordResetID, userID, token),
	) {
		t.Fatal("code and token hashes must use distinct secrets and contexts")
	}
}

func TestHMACPasswordResetHasherCopiesAndSeparatesSecrets(t *testing.T) {
	t.Parallel()

	codeSecret := bytes.Repeat([]byte{1}, 32)
	tokenSecret := bytes.Repeat([]byte{2}, 32)
	hasher, err := NewHMACPasswordResetHasher(codeSecret, tokenSecret)
	if err != nil {
		t.Fatalf("NewHMACPasswordResetHasher() error = %v", err)
	}

	passwordResetID := uuid.New()
	userID := uuid.New()
	before := hasher.HashCode(passwordResetID, userID, "123456")

	codeSecret[0] = 9
	tokenSecret[0] = 9

	after := hasher.HashCode(passwordResetID, userID, "123456")
	if !bytes.Equal(before, after) {
		t.Fatal("hash changed after mutating the supplied secrets")
	}

	sharedSecret := bytes.Repeat([]byte{3}, 32)
	if _, err := NewHMACPasswordResetHasher(sharedSecret, sharedSecret); err == nil {
		t.Fatal("NewHMACPasswordResetHasher() error = nil for shared secret")
	}
}

func TestPasswordResetTokenRotationProducesAnotherHash(t *testing.T) {
	t.Parallel()

	generator := NewSecurePasswordResetTokenGenerator()
	hasher := newTestPasswordResetHasher(t)
	passwordResetID := uuid.New()
	userID := uuid.New()

	firstToken, err := generator.GenerateToken()
	if err != nil {
		t.Fatalf("first GenerateToken() error = %v", err)
	}

	secondToken, err := generator.GenerateToken()
	if err != nil {
		t.Fatalf("second GenerateToken() error = %v", err)
	}

	firstHash := hasher.HashToken(passwordResetID, userID, firstToken)
	secondHash := hasher.HashToken(passwordResetID, userID, secondToken)
	if bytes.Equal(firstHash, secondHash) {
		t.Fatal("rotated token produced the same hash")
	}
}

func TestPasswordResetSensitiveFieldsAreNotSerialized(t *testing.T) {
	t.Parallel()

	values := []any{
		domain.PasswordResetChallenge{
			CodeHash:               []byte("code-hash-secret"),
			PasswordResetTokenHash: []byte("token-hash-secret"),
		},
		domain.SendPasswordResetCodeDTO{Code: "123456"},
		domain.CompletePasswordResetDTO{
			PasswordResetTokenHash: []byte("token-hash-secret"),
			PasswordHash:           "password-hash-secret",
		},
	}

	for _, value := range values {
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatalf("json.Marshal(%T) error = %v", value, err)
		}

		serialized := string(encoded)
		for _, secret := range []string{
			"123456",
			"code-hash-secret",
			"token-hash-secret",
			"password-hash-secret",
		} {
			if strings.Contains(serialized, secret) {
				t.Errorf("json.Marshal(%T) exposed sensitive value %q", value, secret)
			}
		}
	}
}

func newTestPasswordResetHasher(t *testing.T) *HMACPasswordResetHasher {
	t.Helper()

	hasher, err := NewHMACPasswordResetHasher(
		bytes.Repeat([]byte{1}, 32),
		bytes.Repeat([]byte{2}, 32),
	)
	if err != nil {
		t.Fatalf("NewHMACPasswordResetHasher() error = %v", err)
	}

	return hasher
}
