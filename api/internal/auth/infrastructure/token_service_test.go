package infrastructure

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	authdomain "github.com/rudsonalves/foundry-stack/api/internal/auth/domain"
)

func TestNewTokenService(t *testing.T) {
	validSecret := []byte("0123456789abcdef0123456789abcdef")

	t.Run("accepts valid config", func(t *testing.T) {
		svc, err := NewTokenService(TokenServiceConfig{
			Secret:    validSecret,
			Issuer:    "foundry-stack",
			Audience:  "clients",
			ClientIDs: []string{"foundry-stack-mobile", "foundry-stack-swagger"},
			Lifetime:  15 * time.Minute,
		})
		if err != nil {
			t.Fatalf("NewTokenService() unexpected error: %v", err)
		}
		if svc == nil {
			t.Fatal("NewTokenService() returned nil service")
		}
	})

	t.Run("rejects secret shorter than 32 bytes", func(t *testing.T) {
		_, err := NewTokenService(TokenServiceConfig{
			Secret:    []byte("short-secret"),
			Issuer:    "foundry-stack",
			Audience:  "clients",
			ClientIDs: []string{"foundry-stack-mobile"},
			Lifetime:  15 * time.Minute,
		})
		if err == nil {
			t.Fatal("NewTokenService() expected error, got nil")
		}
	})

	t.Run("rejects empty issuer", func(t *testing.T) {
		_, err := NewTokenService(TokenServiceConfig{
			Secret:    validSecret,
			Issuer:    "",
			Audience:  "clients",
			ClientIDs: []string{"foundry-stack-mobile"},
			Lifetime:  15 * time.Minute,
		})
		if err == nil {
			t.Fatal("NewTokenService() expected error, got nil")
		}
	})

	t.Run("rejects empty audience", func(t *testing.T) {
		_, err := NewTokenService(TokenServiceConfig{
			Secret:    validSecret,
			Issuer:    "foundry-stack",
			Audience:  "",
			ClientIDs: []string{"foundry-stack-mobile"},
			Lifetime:  15 * time.Minute,
		})
		if err == nil {
			t.Fatal("NewTokenService() expected error, got nil")
		}
	})

	t.Run("rejects zero lifetime", func(t *testing.T) {
		_, err := NewTokenService(TokenServiceConfig{
			Secret:    validSecret,
			Issuer:    "foundry-stack",
			Audience:  "clients",
			ClientIDs: []string{"foundry-stack-mobile"},
			Lifetime:  0,
		})
		if err == nil {
			t.Fatal("NewTokenService() expected error, got nil")
		}
	})

	t.Run("rejects negative lifetime", func(t *testing.T) {
		_, err := NewTokenService(TokenServiceConfig{
			Secret:    validSecret,
			Issuer:    "foundry-stack",
			Audience:  "clients",
			ClientIDs: []string{"foundry-stack-mobile"},
			Lifetime:  -1 * time.Minute,
		})
		if err == nil {
			t.Fatal("NewTokenService() expected error, got nil")
		}
	})

	t.Run("copies provided secret", func(t *testing.T) {
		provided := []byte("0123456789abcdef0123456789abcdef")
		svc, err := NewTokenService(TokenServiceConfig{
			Secret:    provided,
			Issuer:    "foundry-stack",
			Audience:  "clients",
			ClientIDs: []string{"foundry-stack-mobile"},
			Lifetime:  15 * time.Minute,
		})
		if err != nil {
			t.Fatalf("NewTokenService() unexpected error: %v", err)
		}

		originalFirstByte := svc.secret[0]
		provided[0] = 'X'
		if svc.secret[0] != originalFirstByte {
			t.Fatal("service secret changed after mutating input slice")
		}
	})

	t.Run("copies provided client IDs", func(t *testing.T) {
		provided := []string{"foundry-stack-mobile", "foundry-stack-swagger"}
		svc, err := NewTokenService(TokenServiceConfig{
			Secret:    validSecret,
			Issuer:    "foundry-stack",
			Audience:  "clients",
			ClientIDs: provided,
			Lifetime:  15 * time.Minute,
		})
		if err != nil {
			t.Fatalf("NewTokenService() unexpected error: %v", err)
		}

		provided[0] = "changed"
		if svc.clientIDs[0] != "foundry-stack-mobile" {
			t.Fatalf("service clientIDs[0] = %q, want %q", svc.clientIDs[0], "foundry-stack-mobile")
		}
	})
}

func TestTokenServiceIssue(t *testing.T) {
	now := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	lifetime := 15 * time.Minute
	svc := mustNewTokenService(t, TokenServiceConfig{
		Secret:    []byte("0123456789abcdef0123456789abcdef"),
		Issuer:    "foundry-stack",
		Audience:  "foundry-stack-api",
		ClientIDs: []string{"foundry-stack-mobile", "foundry-stack-swagger"},
		Lifetime:  lifetime,
	})
	svc.now = func() time.Time { return now }

	userID := uuid.MustParse("7f8ec6fe-1a0c-498d-a985-ca22f7735814")
	tokenString, expiresAt, err := svc.Issue(userID, "foundry-stack-mobile")
	if err != nil {
		t.Fatalf("Issue() unexpected error: %v", err)
	}

	claims := new(AccessClaims)
	parsed, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (any, error) {
			return svc.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithoutClaimsValidation(),
	)
	if err != nil {
		t.Fatalf("ParseWithClaims() unexpected error: %v", err)
	}
	if !parsed.Valid {
		t.Fatal("token should be valid")
	}

	t.Run("contains expected issuer", func(t *testing.T) {
		if claims.Issuer != "foundry-stack" {
			t.Fatalf("claims.Issuer = %q, want %q", claims.Issuer, "foundry-stack")
		}
	})

	t.Run("contains expected subject UUID", func(t *testing.T) {
		if claims.Subject != userID.String() {
			t.Fatalf("claims.Subject = %q, want %q", claims.Subject, userID.String())
		}
	})

	t.Run("contains expected audience", func(t *testing.T) {
		if len(claims.Audience) != 1 || claims.Audience[0] != "foundry-stack-api" {
			t.Fatalf("claims.Audience = %v, want [%q]", claims.Audience, "foundry-stack-api")
		}
	})

	t.Run("contains expected client ID", func(t *testing.T) {
		if claims.ClientID != "foundry-stack-mobile" {
			t.Fatalf("claims.ClientID = %q, want %q", claims.ClientID, "foundry-stack-mobile")
		}
	})

	t.Run("contains issued at", func(t *testing.T) {
		if claims.IssuedAt == nil {
			t.Fatal("claims.IssuedAt should be present")
		}
		if !claims.IssuedAt.Time.Equal(now) {
			t.Fatalf(
				"claims.IssuedAt = %v, want %v",
				claims.IssuedAt.Time,
				now,
			)
		}
	})

	t.Run("contains expiration", func(t *testing.T) {
		if claims.ExpiresAt == nil {
			t.Fatal("claims.ExpiresAt should be present")
		}
		want := now.Add(lifetime)
		if !claims.ExpiresAt.Time.Equal(want) {
			t.Fatalf(
				"claims.ExpiresAt = %v, want %v",
				claims.ExpiresAt.Time,
				want,
			)
		}
	})

	t.Run("expiresAt equals now plus lifetime", func(t *testing.T) {
		want := now.Add(lifetime)
		if !expiresAt.Equal(want) {
			t.Fatalf("expiresAt = %v, want %v", expiresAt, want)
		}
	})

	t.Run("issue then parse returns same user", func(t *testing.T) {
		gotUserID, err := svc.Parse(tokenString)
		if err != nil {
			t.Fatalf("Parse() unexpected error: %v", err)
		}
		if gotUserID != userID {
			t.Fatalf("Parse() userID = %s, want %s", gotUserID, userID)
		}
	})

	t.Run("rejects unknown client", func(t *testing.T) {
		_, _, err := svc.Issue(userID, "unknown-client")
		if !errors.Is(err, authdomain.ErrClientNotAllowed) {
			t.Fatalf("Issue() error = %v, want ErrClientNotAllowed", err)
		}
	})

	t.Run("rejects empty client", func(t *testing.T) {
		_, _, err := svc.Issue(userID, "")
		if !errors.Is(err, authdomain.ErrClientNotAllowed) {
			t.Fatalf("Issue() error = %v, want ErrClientNotAllowed", err)
		}
	})
}

func TestTokenServiceParse(t *testing.T) {
	now := time.Date(2026, time.July, 20, 12, 0, 0, 0, time.UTC)
	secret := []byte("0123456789abcdef0123456789abcdef")
	svc := mustNewTokenService(t, TokenServiceConfig{
		Secret:    secret,
		Issuer:    "foundry-stack",
		Audience:  "foundry-stack-api",
		ClientIDs: []string{"foundry-stack-mobile", "foundry-stack-swagger"},
		Lifetime:  15 * time.Minute,
	})
	svc.now = func() time.Time { return now }
	userID := uuid.MustParse("2be274be-dcd5-4b57-b9a7-173311ea8715")

	validClaims := AccessClaims{
		ClientID: "foundry-stack-mobile",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "foundry-stack",
			Subject:   userID.String(),
			Audience:  jwt.ClaimStrings{"foundry-stack-api"},
			IssuedAt:  jwt.NewNumericDate(now.Add(-1 * time.Minute)),
			ExpiresAt: jwt.NewNumericDate(now.Add(5 * time.Minute)),
		},
	}

	t.Run("accepts valid token", func(t *testing.T) {
		tokenString := mustSignToken(t, jwt.SigningMethodHS256, secret, validClaims)

		got, err := svc.Parse(tokenString)
		if err != nil {
			t.Fatalf("Parse() unexpected error: %v", err)
		}
		if got != userID {
			t.Fatalf("Parse() userID = %s, want %s", got, userID)
		}
	})

	t.Run("rejects token without client ID", func(t *testing.T) {
		claims := validClaims
		claims.ClientID = ""
		tokenString := mustSignToken(t, jwt.SigningMethodHS256, secret, claims)

		_, err := svc.Parse(tokenString)
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("Parse() error = %v, want ErrInvalidToken", err)
		}
	})

	t.Run("rejects token for unknown client", func(t *testing.T) {
		claims := validClaims
		claims.ClientID = "unknown-client"
		tokenString := mustSignToken(t, jwt.SigningMethodHS256, secret, claims)

		_, err := svc.Parse(tokenString)
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("Parse() error = %v, want ErrInvalidToken", err)
		}
	})

	t.Run("rejects expired token", func(t *testing.T) {
		claims := validClaims
		claims.ExpiresAt = jwt.NewNumericDate(now.Add(-1 * time.Minute))
		tokenString := mustSignToken(t, jwt.SigningMethodHS256, secret, claims)

		_, err := svc.Parse(tokenString)
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("Parse() error = %v, want ErrInvalidToken", err)
		}
	})

	t.Run("rejects signature signed with different key", func(t *testing.T) {
		tokenString := mustSignToken(
			t,
			jwt.SigningMethodHS256,
			[]byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
			validClaims,
		)

		_, err := svc.Parse(tokenString)
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("Parse() error = %v, want ErrInvalidToken", err)
		}
	})

	t.Run("rejects different issuer", func(t *testing.T) {
		claims := validClaims
		claims.Issuer = "other-issuer"
		tokenString := mustSignToken(t, jwt.SigningMethodHS256, secret, claims)

		_, err := svc.Parse(tokenString)
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("Parse() error = %v, want ErrInvalidToken", err)
		}
	})

	t.Run("rejects different audience", func(t *testing.T) {
		claims := validClaims
		claims.Audience = jwt.ClaimStrings{"other-audience"}
		tokenString := mustSignToken(t, jwt.SigningMethodHS256, secret, claims)

		_, err := svc.Parse(tokenString)
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("Parse() error = %v, want ErrInvalidToken", err)
		}
	})

	t.Run("rejects token without expiration", func(t *testing.T) {
		claims := validClaims
		claims.ExpiresAt = nil
		tokenString := mustSignToken(t, jwt.SigningMethodHS256, secret, claims)

		_, err := svc.Parse(tokenString)
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("Parse() error = %v, want ErrInvalidToken", err)
		}
	})

	t.Run("rejects empty subject", func(t *testing.T) {
		claims := validClaims
		claims.Subject = ""
		tokenString := mustSignToken(t, jwt.SigningMethodHS256, secret, claims)

		_, err := svc.Parse(tokenString)
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("Parse() error = %v, want ErrInvalidToken", err)
		}
	})

	t.Run("rejects non UUID subject", func(t *testing.T) {
		claims := validClaims
		claims.Subject = "not-a-uuid"
		tokenString := mustSignToken(t, jwt.SigningMethodHS256, secret, claims)

		_, err := svc.Parse(tokenString)
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("Parse() error = %v, want ErrInvalidToken", err)
		}
	})

	t.Run("rejects malformed token text", func(t *testing.T) {
		_, err := svc.Parse("this is not a token")
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("Parse() error = %v, want ErrInvalidToken", err)
		}
	})

	t.Run("rejects token signed with algorithm other than HS256", func(t *testing.T) {
		tokenString := mustSignToken(t, jwt.SigningMethodHS384, secret, validClaims)

		_, err := svc.Parse(tokenString)
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("Parse() error = %v, want ErrInvalidToken", err)
		}
	})

	t.Run("accepts token expired within 30 second leeway", func(t *testing.T) {
		claims := validClaims
		claims.ExpiresAt = jwt.NewNumericDate(now.Add(-29 * time.Second))
		tokenString := mustSignToken(t, jwt.SigningMethodHS256, secret, claims)

		got, err := svc.Parse(tokenString)
		if err != nil {
			t.Fatalf("Parse() unexpected error: %v", err)
		}
		if got != userID {
			t.Fatalf("Parse() userID = %s, want %s", got, userID)
		}
	})

	t.Run("rejects token expired more than 30 seconds ago", func(t *testing.T) {
		claims := validClaims
		claims.ExpiresAt = jwt.NewNumericDate(now.Add(-31 * time.Second))
		tokenString := mustSignToken(t, jwt.SigningMethodHS256, secret, claims)

		_, err := svc.Parse(tokenString)
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("Parse() error = %v, want ErrInvalidToken", err)
		}
	})

	t.Run("accepts issued at within 30 second leeway", func(t *testing.T) {
		claims := validClaims
		claims.IssuedAt = jwt.NewNumericDate(now.Add(29 * time.Second))
		tokenString := mustSignToken(t, jwt.SigningMethodHS256, secret, claims)

		got, err := svc.Parse(tokenString)
		if err != nil {
			t.Fatalf("Parse() unexpected error: %v", err)
		}
		if got != userID {
			t.Fatalf("Parse() userID = %s, want %s", got, userID)
		}
	})

	t.Run("rejects issued at beyond 30 second leeway", func(t *testing.T) {
		claims := validClaims
		claims.IssuedAt = jwt.NewNumericDate(now.Add(31 * time.Second))
		tokenString := mustSignToken(t, jwt.SigningMethodHS256, secret, claims)

		_, err := svc.Parse(tokenString)
		if !errors.Is(err, ErrInvalidToken) {
			t.Fatalf("Parse() error = %v, want ErrInvalidToken", err)
		}
	})
}

func mustNewTokenService(t *testing.T, cfg TokenServiceConfig) *TokenService {
	t.Helper()

	svc, err := NewTokenService(cfg)
	if err != nil {
		t.Fatalf("NewTokenService() unexpected error: %v", err)
	}

	return svc
}

func mustSignToken(
	t *testing.T,
	method jwt.SigningMethod,
	secret []byte,
	claims AccessClaims,
) string {
	t.Helper()

	token := jwt.NewWithClaims(method, claims)
	signed, err := token.SignedString(secret)
	if err != nil {
		t.Fatalf("SignedString() unexpected error: %v", err)
	}

	return signed
}
