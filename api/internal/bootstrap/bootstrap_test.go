package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testEmailVerificationCodeSecretBase64 = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="
const testEmailVerificationTokenSecretBase64 = "ZmVkY2JhOTg3NjU0MzIxMGZlZGNiYTk4NzY1NDMyMTA="
const testPasswordResetCodeSecretBase64 = "cGFzc3dvcmQtcmVzZXQtY29kZS1zZWNyZXQtMzJieXRlcw=="
const testPasswordResetTokenSecretBase64 = "cGFzc3dvcmQtcmVzZXQtdG9rZW4tc2VjcmV0LTMyYnl0ZXM="

// --- envOrDefault ---

func TestEnvOrDefault(t *testing.T) {
	const key = "TEST_ENV_OR_DEFAULT"

	t.Run("returns env value when set", func(t *testing.T) {
		t.Setenv(key, "value-from-env")
		got := envOrDefault(key, "fallback")
		if got != "value-from-env" {
			t.Errorf("envOrDefault() = %q, want %q", got, "value-from-env")
		}
	})

	t.Run("returns fallback when not set", func(t *testing.T) {
		os.Unsetenv(key)
		got := envOrDefault(key, "fallback")
		if got != "fallback" {
			t.Errorf("envOrDefault() = %q, want %q", got, "fallback")
		}
	})

	t.Run("returns empty string when set to empty", func(t *testing.T) {
		t.Setenv(key, "")
		got := envOrDefault(key, "fallback")
		if got != "" {
			t.Errorf("envOrDefault() = %q, want empty string", got)
		}
	})
}

// --- envIntOrDefault ---

func TestEnvIntOrDefault(t *testing.T) {
	const key = "TEST_ENV_INT_OR_DEFAULT"

	t.Run("returns int when valid", func(t *testing.T) {
		t.Setenv(key, "42")
		got, err := envInt(key, 0)
		if err != nil {
			t.Fatalf("envInt() unexpected error: %v", err)
		}
		if got != 42 {
			t.Errorf("envIntOrDefault() = %d, want 42", got)
		}
	})

	t.Run("returns fallback when not set", func(t *testing.T) {
		os.Unsetenv(key)
		got, err := envInt(key, 99)
		if err != nil {
			t.Fatalf("envInt() unexpected error: %v", err)
		}
		if got != 99 {
			t.Errorf("envIntOrDefault() = %d, want 99", got)
		}
	})

	t.Run("returns error when value is not an integer", func(t *testing.T) {
		t.Setenv(key, "not-a-number")
		if _, err := envInt(key, 10); err == nil {
			t.Error("envInt() expected error, got nil")
		}
	})

	t.Run("returns error when value is float", func(t *testing.T) {
		t.Setenv(key, "3.14")
		if _, err := envInt(key, 5); err == nil {
			t.Error("envInt() expected error, got nil")
		}
	})

	t.Run("handles zero correctly", func(t *testing.T) {
		t.Setenv(key, "0")
		got, err := envInt(key, 99)
		if err != nil {
			t.Fatalf("envInt() unexpected error: %v", err)
		}
		if got != 0 {
			t.Errorf("envIntOrDefault() = %d, want 0", got)
		}
	})

	t.Run("handles negative value", func(t *testing.T) {
		t.Setenv(key, "-5")
		got, err := envInt(key, 99)
		if err != nil {
			t.Fatalf("envInt() unexpected error: %v", err)
		}
		if got != -5 {
			t.Errorf("envIntOrDefault() = %d, want -5", got)
		}
	})
}

// --- envBoolOrDefault ---

func TestEnvBoolOrDefault(t *testing.T) {
	const key = "TEST_ENV_BOOL_OR_DEFAULT"

	trueCases := []string{"true", "True", "TRUE", "1", "t", "T"}
	falseCases := []string{"false", "False", "FALSE", "0", "f", "F"}

	for _, val := range trueCases {
		val := val
		t.Run("parses true: "+val, func(t *testing.T) {
			t.Setenv(key, val)
			got, err := envBool(key, false)
			if err != nil {
				t.Fatalf("envBool() unexpected error: %v", err)
			}
			if !got {
				t.Errorf("envBoolOrDefault() = false, want true for %q", val)
			}
		})
	}

	for _, val := range falseCases {
		val := val
		t.Run("parses false: "+val, func(t *testing.T) {
			t.Setenv(key, val)
			got, err := envBool(key, true)
			if err != nil {
				t.Fatalf("envBool() unexpected error: %v", err)
			}
			if got {
				t.Errorf("envBoolOrDefault() = true, want false for %q", val)
			}
		})
	}

	t.Run("returns fallback when not set", func(t *testing.T) {
		os.Unsetenv(key)
		got, err := envBool(key, true)
		if err != nil {
			t.Fatalf("envBool() unexpected error: %v", err)
		}
		if !got {
			t.Errorf("envBoolOrDefault() = false, want true (fallback)")
		}
	})

	t.Run("returns error when invalid value", func(t *testing.T) {
		t.Setenv(key, "yes") // not a valid strconv.ParseBool value
		if _, err := envBool(key, true); err == nil {
			t.Error("envBool() expected error, got nil")
		}
	})
}

// --- LoadConfig() ---

// loadConfigEnvKeys lists all env vars read by LoadConfig.
var loadConfigEnvKeys = []string{
	"APP_ENV",
	"PORT",
	"READ_TIMEOUT_SECONDS",
	"SHUTDOWN_TIMEOUT_SECONDS",
	"ALLOWED_ORIGIN",
	"DEBUG",
	"DATABASE_URL",
	"DB_MAX_OPEN_CONNS",
	"DB_MAX_IDLE_CONNS",
	"DB_CONN_MAX_IDLE_TIME_SECONDS",
	"DB_CONN_MAX_LIFETIME_SECONDS",
	"JWT_SECRET_BASE64",
	"JWT_ISSUER",
	"JWT_AUDIENCE",
	"JWT_CLIENT_IDS",
	"JWT_ACCESS_TTL_SECONDS",
	"JWT_REFRESH_TTL_SECONDS",
	"EMAIL_PROVIDER",
	"SMTP_HOST",
	"SMTP_PORT",
	"SMTP_USERNAME",
	"SMTP_PASSWORD",
	"SMTP_FROM_ADDRESS",
	"SMTP_FROM_NAME",
	"SMTP_TLS_MODE",
	"SMTP_TIMEOUT_SECONDS",
	"EMAIL_VERIFICATION_CODE_SECRET_BASE64",
	"EMAIL_VERIFICATION_CODE_TTL_SECONDS",
	"EMAIL_VERIFICATION_TOKEN_SECRET_BASE64",
	"EMAIL_VERIFICATION_TOKEN_TTL_SECONDS",
	"EMAIL_VERIFICATION_MAX_CONFIRM_ATTEMPTS",
	"EMAIL_VERIFICATION_RESEND_COOLDOWN_SECONDS",
	"EMAIL_VERIFICATION_RESEND_EMAIL_LIMIT_PER_HOUR",
	"EMAIL_VERIFICATION_RESEND_IP_LIMIT_PER_HOUR",
	"EMAIL_VERIFICATION_RETENTION_SECONDS",
	"PASSWORD_RESET_CODE_SECRET_BASE64",
	"PASSWORD_RESET_CODE_TTL_SECONDS",
	"PASSWORD_RESET_TOKEN_SECRET_BASE64",
	"PASSWORD_RESET_TOKEN_TTL_SECONDS",
	"PASSWORD_RESET_MAX_CONFIRM_ATTEMPTS",
	"PASSWORD_RESET_RESEND_COOLDOWN_SECONDS",
	"PASSWORD_RESET_EMAIL_LIMIT_PER_HOUR",
	"PASSWORD_RESET_IP_LIMIT_PER_HOUR",
	"PASSWORD_RESET_RETENTION_SECONDS",
	"TRUSTED_PROXY_CIDRS",
}

const (
	testJWTSecretBase64 = "MDEyMzQ1Njc4OWFiY2RlZjAxMjM0NTY3ODlhYmNkZWY="
	testJWTIssuer       = "foundry-stack-test"
	testJWTAudience     = "foundry-stack-api"
	testJWTClientIDs    = "foundry-stack-mobile,foundry-stack-swagger"
)

// clearLoadConfigEnv unsets all env vars that LoadConfig reads and restores
// them when the test finishes. This prevents godotenv state leakage between
// subtests, since godotenv.Load sets OS-level variables.
func clearLoadConfigEnv(t *testing.T) {
	t.Helper()
	for _, k := range loadConfigEnvKeys {
		old, existed := os.LookupEnv(k)
		os.Unsetenv(k)
		if existed {
			k := k
			old := old
			t.Cleanup(func() { os.Setenv(k, old) })
		} else {
			k := k
			t.Cleanup(func() { os.Unsetenv(k) })
		}
	}
}

func TestLoadConfig(t *testing.T) {
	t.Run("requires environment option", func(t *testing.T) {
		_, err := LoadConfig(Options{EnvFile: ".env.dev"})
		if err == nil {
			t.Fatal("LoadConfig() expected error, got nil")
		}
		if err.Error() != "environment option is required" {
			t.Errorf("error = %q, want %q", err, "environment option is required")
		}
	})

	t.Run("requires environment file option", func(t *testing.T) {
		_, err := LoadConfig(Options{Environment: EnvDevelopment})
		if err == nil {
			t.Fatal("LoadConfig() expected error, got nil")
		}
		if err.Error() != "environment file option is required" {
			t.Errorf("error = %q, want %q", err, "environment file option is required")
		}
	})

	t.Run("returns error when selected environment file is missing", func(t *testing.T) {
		clearLoadConfigEnv(t)
		original, _ := os.Getwd()
		tmp := t.TempDir()
		if err := os.Chdir(tmp); err != nil {
			t.Fatalf("failed to chdir to tempdir: %v", err)
		}
		defer os.Chdir(original)

		_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: ".env.dev"})
		if err == nil {
			t.Fatal("LoadConfig() expected error, got nil")
		}
	})

	t.Run("requires APP_ENV in selected file", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(t, "PORT=8080\n")

		_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err == nil {
			t.Fatal("LoadConfig() expected error, got nil")
		}
		if err.Error() != "APP_ENV is required" {
			t.Errorf("error = %q, want %q", err, "APP_ENV is required")
		}
	})

	t.Run("rejects environment mismatch", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(t, "APP_ENV=prod\n")

		_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err == nil {
			t.Fatal("LoadConfig() expected error, got nil")
		}
		want := "environment mismatch: selected \"dev\", APP_ENV is \"prod\""
		if err.Error() != want {
			t.Errorf("error = %q, want %q", err, want)
		}
	})

	t.Run("rejects invalid DEBUG", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(t, "APP_ENV=dev\nDEBUG=yes\n")

		_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err == nil {
			t.Fatal("LoadConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "DEBUG must be a boolean") {
			t.Errorf("error = %q, want DEBUG boolean validation", err)
		}
	})

	t.Run("rejects invalid READ_TIMEOUT_SECONDS", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(t, "APP_ENV=dev\nREAD_TIMEOUT_SECONDS=slow\n")

		_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err == nil {
			t.Fatal("LoadConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "READ_TIMEOUT_SECONDS must be an integer") {
			t.Errorf("error = %q, want read timeout integer validation", err)
		}
	})

	t.Run("loads explicit shutdown timeout", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(
			t,
			"APP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\nSHUTDOWN_TIMEOUT_SECONDS=25\n",
		)

		cfg, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err != nil {
			t.Fatalf("LoadConfig() unexpected error: %v", err)
		}
		if cfg.ShutdownTimeout != 25*time.Second {
			t.Errorf(
				"ShutdownTimeout = %v, want %v",
				cfg.ShutdownTimeout,
				25*time.Second,
			)
		}
	})

	t.Run("defaults shutdown timeout to 10 seconds", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(
			t,
			"APP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\n",
		)

		cfg, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err != nil {
			t.Fatalf("LoadConfig() unexpected error: %v", err)
		}
		if cfg.ShutdownTimeout != 10*time.Second {
			t.Errorf(
				"ShutdownTimeout = %v, want %v",
				cfg.ShutdownTimeout,
				10*time.Second,
			)
		}
	})

	t.Run("rejects non numeric shutdown timeout", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(
			t,
			"APP_ENV=dev\nSHUTDOWN_TIMEOUT_SECONDS=slow\n",
		)

		_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err == nil {
			t.Fatal("LoadConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "SHUTDOWN_TIMEOUT_SECONDS must be an integer") {
			t.Errorf("error = %q, want shutdown timeout integer validation", err)
		}
	})

	t.Run("rejects zero or negative shutdown timeout", func(t *testing.T) {
		tests := []struct {
			name  string
			value string
		}{
			{name: "zero", value: "0"},
			{name: "negative", value: "-10"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				clearLoadConfigEnv(t)
				envFile := writeEnvironmentFile(
					t,
					"APP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\nSHUTDOWN_TIMEOUT_SECONDS="+tt.value+"\n",
				)

				_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
				if err == nil {
					t.Fatal("LoadConfig() expected error, got nil")
				}
				if !strings.Contains(err.Error(), "shutdown timeout must be greater than zero") {
					t.Errorf("error = %q, want shutdown timeout validation", err)
				}
			})
		}
	})

	t.Run("rejects invalid DB numeric settings", func(t *testing.T) {
		tests := []struct {
			name    string
			env     string
			wantErr string
		}{
			{
				name:    "invalid DB_MAX_OPEN_CONNS",
				env:     "APP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\nDB_MAX_OPEN_CONNS=abc\n",
				wantErr: "DB_MAX_OPEN_CONNS must be an integer",
			},
			{
				name:    "invalid DB_MAX_IDLE_CONNS",
				env:     "APP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\nDB_MAX_IDLE_CONNS=abc\n",
				wantErr: "DB_MAX_IDLE_CONNS must be an integer",
			},
			{
				name:    "invalid DB_CONN_MAX_IDLE_TIME_SECONDS",
				env:     "APP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\nDB_CONN_MAX_IDLE_TIME_SECONDS=abc\n",
				wantErr: "DB_CONN_MAX_IDLE_TIME_SECONDS must be an integer",
			},
			{
				name:    "invalid DB_CONN_MAX_LIFETIME_SECONDS",
				env:     "APP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\nDB_CONN_MAX_LIFETIME_SECONDS=abc\n",
				wantErr: "DB_CONN_MAX_LIFETIME_SECONDS must be an integer",
			},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				clearLoadConfigEnv(t)
				envFile := writeEnvironmentFile(t, tt.env)

				_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
				if err == nil {
					t.Fatal("LoadConfig() expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error = %q, want %q", err, tt.wantErr)
				}
			})
		}
	})

	t.Run("rejects invalid JWT secret base64", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(t, "APP_ENV=dev\nJWT_SECRET_BASE64=not-base64\n")

		_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err == nil {
			t.Fatal("LoadConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "decode JWT secret") {
			t.Errorf("error = %q, want JWT decode failure", err)
		}
	})

	t.Run("rejects decoded JWT secret shorter than 32 bytes", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(
			t,
			"APP_ENV=dev\nJWT_SECRET_BASE64=YWFhYQ==\n",
		)

		_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err == nil {
			t.Fatal("LoadConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "JWT secret must be at least 32 bytes long") {
			t.Errorf("error = %q, want JWT secret length validation", err)
		}
	})

	t.Run("rejects non numeric JWT access TTL", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(
			t,
			"APP_ENV=dev\nJWT_ACCESS_TTL_SECONDS=abc\n",
		)

		_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err == nil {
			t.Fatal("LoadConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "JWT_ACCESS_TTL_SECONDS must be an integer") {
			t.Errorf("error = %q, want JWT access TTL integer validation", err)
		}
	})

	t.Run("rejects zero or negative JWT access TTL", func(t *testing.T) {
		tests := []struct {
			name  string
			value string
		}{
			{name: "zero", value: "0"},
			{name: "negative", value: "-10"},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				clearLoadConfigEnv(t)
				envFile := writeEnvironmentFile(
					t,
					"APP_ENV=dev\nJWT_ACCESS_TTL_SECONDS="+tt.value+"\n",
				)

				_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
				if err == nil {
					t.Fatal("LoadConfig() expected error, got nil")
				}
				if !strings.Contains(err.Error(), "JWT access TTL must be greater than zero") {
					t.Errorf("error = %q, want JWT access TTL validation", err)
				}
			})
		}
	})

	t.Run("loads JWT secret issuer audience and configured TTL", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(
			t,
			"APP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\nJWT_ACCESS_TTL_SECONDS=1200\nJWT_REFRESH_TTL_SECONDS=7200\n",
		)

		cfg, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err != nil {
			t.Fatalf("LoadConfig() unexpected error: %v", err)
		}

		if string(cfg.JWT.Secret) != "0123456789abcdef0123456789abcdef" {
			t.Errorf("JWT.Secret decoded value = %q, want expected secret", string(cfg.JWT.Secret))
		}
		if cfg.JWT.Issuer != testJWTIssuer {
			t.Errorf("JWT.Issuer = %q, want %q", cfg.JWT.Issuer, testJWTIssuer)
		}
		if cfg.JWT.Audience != testJWTAudience {
			t.Errorf("JWT.Audience = %q, want %q", cfg.JWT.Audience, testJWTAudience)
		}
		if got, want := strings.Join(cfg.JWT.ClientIDs, ","), testJWTClientIDs; got != want {
			t.Errorf("JWT.ClientIDs = %q, want %q", got, want)
		}
		if cfg.JWT.AccessTTL != 1200*time.Second {
			t.Errorf("JWT.AccessTTL = %v, want %v", cfg.JWT.AccessTTL, 1200*time.Second)
		}
		if cfg.JWT.RefreshTTL != 7200*time.Second {
			t.Errorf("JWT.RefreshTTL = %v, want %v", cfg.JWT.RefreshTTL, 7200*time.Second)
		}
	})

	t.Run("loads and normalizes JWT client IDs", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(
			t,
			"APP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\nJWT_CLIENT_IDS= foundry-stack-mobile , foundry-stack-swagger \n",
		)

		cfg, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err != nil {
			t.Fatalf("LoadConfig() unexpected error: %v", err)
		}

		want := []string{"foundry-stack-mobile", "foundry-stack-swagger"}
		if len(cfg.JWT.ClientIDs) != len(want) {
			t.Fatalf("JWT.ClientIDs = %v, want %v", cfg.JWT.ClientIDs, want)
		}
		for i := range want {
			if cfg.JWT.ClientIDs[i] != want[i] {
				t.Fatalf("JWT.ClientIDs[%d] = %q, want %q", i, cfg.JWT.ClientIDs[i], want[i])
			}
		}
	})

	t.Run("rejects empty JWT client IDs", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(
			t,
			"APP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\nJWT_CLIENT_IDS=\n",
		)

		_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err == nil {
			t.Fatal("LoadConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "JWT client IDs can't contain empty values") {
			t.Fatalf("LoadConfig() error = %q, want client ID validation", err)
		}
	})

	t.Run("defaults JWT access TTL to 900 seconds when variable is missing", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(t, "APP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\n")

		cfg, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err != nil {
			t.Fatalf("LoadConfig() unexpected error: %v", err)
		}

		if cfg.JWT.AccessTTL != 900*time.Second {
			t.Errorf("JWT.AccessTTL = %v, want %v", cfg.JWT.AccessTTL, 900*time.Second)
		}
	})

	t.Run("defaults JWT refresh TTL to 172800 seconds when variable is missing", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(t, "APP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\n")

		cfg, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err != nil {
			t.Fatalf("LoadConfig() unexpected error: %v", err)
		}

		if cfg.JWT.RefreshTTL != 172800*time.Second {
			t.Errorf("JWT.RefreshTTL = %v, want %v", cfg.JWT.RefreshTTL, 172800*time.Second)
		}
	})

	t.Run("rejects non numeric JWT refresh TTL", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(
			t,
			"APP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\nJWT_REFRESH_TTL_SECONDS=abc\n",
		)

		_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err == nil {
			t.Fatal("LoadConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "JWT_REFRESH_TTL_SECONDS must be an integer") {
			t.Errorf("error = %q, want JWT refresh TTL integer validation", err)
		}
	})

	t.Run("rejects zero or negative JWT refresh TTL", func(t *testing.T) {
		tests := []struct {
			name  string
			value string
		}{
			{name: "zero", value: "0"},
			{name: "negative", value: "-10"},
		}

		for _, tt := range tests {
			tt := tt
			t.Run(tt.name, func(t *testing.T) {
				clearLoadConfigEnv(t)
				envFile := writeEnvironmentFile(
					t,
					"APP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\nJWT_REFRESH_TTL_SECONDS="+tt.value+"\n",
				)

				_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
				if err == nil {
					t.Fatal("LoadConfig() expected error, got nil")
				}
				if !strings.Contains(err.Error(), "JWT refresh TTL must be greater than zero") {
					t.Errorf("error = %q, want JWT refresh TTL validation", err)
				}
			})
		}
	})

	t.Run("rejects DB_MAX_IDLE_CONNS greater than DB_MAX_OPEN_CONNS", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(
			t,
			"APP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\nDB_MAX_OPEN_CONNS=5\nDB_MAX_IDLE_CONNS=6\n",
		)

		_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: envFile})
		if err == nil {
			t.Fatal("LoadConfig() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "database max idle connections can't exceed max open connections") {
			t.Errorf("error = %q, want max idle validation", err)
		}
	})

	t.Run("loads config from env file", func(t *testing.T) {
		clearLoadConfigEnv(t)
		tmp := t.TempDir()
		original, _ := os.Getwd()
		if err := os.Chdir(tmp); err != nil {
			t.Fatalf("failed to chdir to tempdir: %v", err)
		}
		defer os.Chdir(original)

		envContent := "APP_ENV=prod\nPORT=9090\nREAD_TIMEOUT_SECONDS=10\nALLOWED_ORIGIN=https://example.com\nDEBUG=false\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\n"
		envContent += "JWT_SECRET_BASE64=" + testJWTSecretBase64 + "\n"
		envContent += "JWT_ISSUER=" + testJWTIssuer + "\n"
		envContent += "JWT_AUDIENCE=" + testJWTAudience + "\n"
		envContent += "JWT_CLIENT_IDS=" + testJWTClientIDs + "\n"
		envContent += "EMAIL_PROVIDER=smtp\n"
		envContent += "EMAIL_VERIFICATION_CODE_SECRET_BASE64=" +
			testEmailVerificationCodeSecretBase64 + "\n"
		envContent += "EMAIL_VERIFICATION_TOKEN_SECRET_BASE64=" +
			testEmailVerificationTokenSecretBase64 + "\n"
		envContent += passwordResetEnvironment()
		envContent += "SMTP_HOST=smtp.example.com\n"
		envContent += "SMTP_PORT=587\n"
		envContent += "SMTP_USERNAME=foundry-stack\n"
		envContent += "SMTP_PASSWORD=test-password\n"
		envContent += "SMTP_FROM_ADDRESS=no-reply@example.com\n"
		envContent += "SMTP_FROM_NAME=FoundryStack\n"
		envContent += "SMTP_TLS_MODE=starttls\n"
		envContent += "SMTP_TIMEOUT_SECONDS=10\n"
		if err := os.WriteFile(".env.prod", []byte(envContent), 0600); err != nil {
			t.Fatalf("failed to write .env.prod: %v", err)
		}

		cfg, err := LoadConfig(Options{Environment: EnvProduction, EnvFile: ".env.prod"})
		if err != nil {
			t.Fatalf("LoadConfig() unexpected error: %v", err)
		}

		if cfg.Environment != "prod" {
			t.Errorf("Environment = %q, want %q", cfg.Environment, "prod")
		}
		if cfg.Port != "9090" {
			t.Errorf("Port = %q, want %q", cfg.Port, "9090")
		}
		if cfg.ReadTimeout != 10*time.Second {
			t.Errorf("ReadTimeout = %v, want %v", cfg.ReadTimeout, 10*time.Second)
		}
		if cfg.AllowedOrigin != "https://example.com" {
			t.Errorf("AllowedOrigin = %q, want %q", cfg.AllowedOrigin, "https://example.com")
		}
		if cfg.Debug {
			t.Errorf("Debug = true, want false")
		}
		if cfg.Database.MaxOpenConns != 10 {
			t.Errorf("Database.MaxOpenConns = %d, want 10", cfg.Database.MaxOpenConns)
		}
		if cfg.Database.MaxIdleConns != 5 {
			t.Errorf("Database.MaxIdleConns = %d, want 5", cfg.Database.MaxIdleConns)
		}
		if cfg.Database.ConnMaxIdleTime != 5*time.Minute {
			t.Errorf("Database.ConnMaxIdleTime = %v, want 5m", cfg.Database.ConnMaxIdleTime)
		}
		if cfg.Database.ConnMaxLifetime != 30*time.Minute {
			t.Errorf("Database.ConnMaxLifetime = %v, want 30m", cfg.Database.ConnMaxLifetime)
		}
	})

	t.Run("uses defaults for missing optional keys", func(t *testing.T) {
		clearLoadConfigEnv(t)
		tmp := t.TempDir()
		original, _ := os.Getwd()
		if err := os.Chdir(tmp); err != nil {
			t.Fatalf("failed to chdir to tempdir: %v", err)
		}
		defer os.Chdir(original)

		// Only the required keys
		envContent := "PORT=8080\nAPP_ENV=dev\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\n"
		envContent += "JWT_SECRET_BASE64=" + testJWTSecretBase64 + "\n"
		envContent += "JWT_ISSUER=" + testJWTIssuer + "\n"
		envContent += "JWT_AUDIENCE=" + testJWTAudience + "\n"
		envContent += "JWT_CLIENT_IDS=" + testJWTClientIDs + "\n"
		envContent += "EMAIL_PROVIDER=memory\n"
		envContent += "EMAIL_VERIFICATION_CODE_SECRET_BASE64=" +
			testEmailVerificationCodeSecretBase64 + "\n"
		envContent += "EMAIL_VERIFICATION_TOKEN_SECRET_BASE64=" +
			testEmailVerificationTokenSecretBase64 + "\n"
		envContent += passwordResetEnvironment()
		if err := os.WriteFile(".env.dev", []byte(envContent), 0600); err != nil {
			t.Fatalf("failed to write .env.dev: %v", err)
		}

		cfg, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: ".env.dev"})
		if err != nil {
			t.Fatalf("LoadConfig() unexpected error: %v", err)
		}

		if cfg.ReadTimeout != 5*time.Second {
			t.Errorf("ReadTimeout = %v, want default 5s", cfg.ReadTimeout)
		}
		if cfg.AllowedOrigin != "http://localhost:5173" {
			t.Errorf("AllowedOrigin = %q, want default", cfg.AllowedOrigin)
		}
		// APP_ENV=dev => Debug defaults to true
		if !cfg.Debug {
			t.Errorf("Debug = false, want true for dev environment")
		}
		if cfg.PasswordReset.CodeTTL != 15*time.Minute ||
			cfg.PasswordReset.TokenTTL != 15*time.Minute ||
			cfg.PasswordReset.MaxConfirmAttempts != 5 ||
			cfg.PasswordReset.ResendCooldown != time.Minute ||
			cfg.PasswordReset.EmailLimitPerHour != 5 ||
			cfg.PasswordReset.IPLimitPerHour != 10 ||
			cfg.PasswordReset.Retention != 24*time.Hour {
			t.Errorf("PasswordReset defaults = %#v", cfg.PasswordReset)
		}
	})

	t.Run("uses local allowed origin default for test environment", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(
			t,
			"APP_ENV=test\nPORT=8080\nDATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\n",
		)

		cfg, err := LoadConfig(Options{Environment: EnvTest, EnvFile: envFile})
		if err != nil {
			t.Fatalf("LoadConfig() unexpected error: %v", err)
		}

		if cfg.AllowedOrigin != "http://localhost:5173" {
			t.Errorf(
				"AllowedOrigin = %q, want %q",
				cfg.AllowedOrigin,
				"http://localhost:5173",
			)
		}
	})

	t.Run("loads password reset overrides and trusted proxies", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(
			t,
			"APP_ENV=dev\n"+
				"DATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\n"+
				"PASSWORD_RESET_CODE_TTL_SECONDS=600\n"+
				"PASSWORD_RESET_TOKEN_TTL_SECONDS=300\n"+
				"PASSWORD_RESET_MAX_CONFIRM_ATTEMPTS=4\n"+
				"PASSWORD_RESET_RESEND_COOLDOWN_SECONDS=90\n"+
				"PASSWORD_RESET_EMAIL_LIMIT_PER_HOUR=3\n"+
				"PASSWORD_RESET_IP_LIMIT_PER_HOUR=8\n"+
				"PASSWORD_RESET_RETENTION_SECONDS=43200\n"+
				"TRUSTED_PROXY_CIDRS=10.0.0.0/8, 2001:db8::/32\n",
		)

		cfg, err := LoadConfig(Options{
			Environment: EnvDevelopment,
			EnvFile:     envFile,
		})
		if err != nil {
			t.Fatalf("LoadConfig() unexpected error: %v", err)
		}

		if cfg.PasswordReset.CodeTTL != 10*time.Minute ||
			cfg.PasswordReset.TokenTTL != 5*time.Minute ||
			cfg.PasswordReset.MaxConfirmAttempts != 4 ||
			cfg.PasswordReset.ResendCooldown != 90*time.Second ||
			cfg.PasswordReset.EmailLimitPerHour != 3 ||
			cfg.PasswordReset.IPLimitPerHour != 8 ||
			cfg.PasswordReset.Retention != 12*time.Hour {
			t.Errorf("PasswordReset overrides = %#v", cfg.PasswordReset)
		}
		if len(cfg.TrustedProxyCIDRs) != 2 ||
			cfg.TrustedProxyCIDRs[0] != "10.0.0.0/8" ||
			cfg.TrustedProxyCIDRs[1] != "2001:db8::/32" {
			t.Errorf("TrustedProxyCIDRs = %#v", cfg.TrustedProxyCIDRs)
		}
	})

	t.Run("rejects invalid password reset secret encoding", func(t *testing.T) {
		clearLoadConfigEnv(t)
		envFile := writeEnvironmentFile(
			t,
			"APP_ENV=dev\n"+
				"DATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\n"+
				"PASSWORD_RESET_CODE_SECRET_BASE64=not-base64!\n",
		)

		_, err := LoadConfig(Options{
			Environment: EnvDevelopment,
			EnvFile:     envFile,
		})
		if err == nil || !strings.Contains(
			err.Error(),
			"decode password reset code secret",
		) {
			t.Fatalf("LoadConfig() error = %v", err)
		}
	})

	t.Run("rejects missing password reset secret", func(t *testing.T) {
		clearLoadConfigEnv(t)
		content := "APP_ENV=dev\n" +
			"DATABASE_URL=postgres://foundry_stack:secret@localhost:5432/foundry_stack\n" +
			"PASSWORD_RESET_CODE_SECRET_BASE64=\n"
		envFile := writeEnvironmentFile(t, content)

		_, err := LoadConfig(Options{
			Environment: EnvDevelopment,
			EnvFile:     envFile,
		})
		if err == nil || !strings.Contains(
			err.Error(),
			"password reset code secret must be at least 32 bytes long",
		) {
			t.Fatalf("LoadConfig() error = %v", err)
		}
	})

	t.Run("returns error when port is missing from env file", func(t *testing.T) {
		clearLoadConfigEnv(t)
		tmp := t.TempDir()
		original, _ := os.Getwd()
		if err := os.Chdir(tmp); err != nil {
			t.Fatalf("failed to chdir to tempdir: %v", err)
		}
		defer os.Chdir(original)

		// PORT and APP_ENV explicitly empty
		envContent := "PORT=\nAPP_ENV=\n"
		if err := os.WriteFile(".env.dev", []byte(envContent), 0600); err != nil {
			t.Fatalf("failed to write .env.dev: %v", err)
		}

		_, err := LoadConfig(Options{Environment: EnvDevelopment, EnvFile: ".env.dev"})
		if err == nil {
			t.Error("LoadConfig() expected validation error, got nil")
		}
	})

	t.Run("enables debug when option is set", func(t *testing.T) {
		clearLoadConfigEnv(t)
		t.Setenv("APP_ENV", "prod")
		t.Setenv("ALLOWED_ORIGIN", "https://example.com")
		t.Setenv("DATABASE_URL", "postgres://foundry_stack:secret@localhost:5432/foundry_stack")
		t.Setenv("JWT_SECRET_BASE64", testJWTSecretBase64)
		t.Setenv("JWT_ISSUER", testJWTIssuer)
		t.Setenv("JWT_AUDIENCE", testJWTAudience)
		t.Setenv("JWT_CLIENT_IDS", testJWTClientIDs)
		envFile := filepath.Join(t.TempDir(), ".env.prod")
		envContent := "APP_ENV=prod\n"
		envContent += "EMAIL_PROVIDER=smtp\n"
		envContent += "EMAIL_VERIFICATION_CODE_SECRET_BASE64=" +
			testEmailVerificationCodeSecretBase64 + "\n"
		envContent += "EMAIL_VERIFICATION_TOKEN_SECRET_BASE64=" +
			testEmailVerificationTokenSecretBase64 + "\n"
		envContent += passwordResetEnvironment()
		envContent += "SMTP_HOST=smtp.example.com\n"
		envContent += "SMTP_PORT=587\n"
		envContent += "SMTP_USERNAME=foundry-stack\n"
		envContent += "SMTP_PASSWORD=test-password\n"
		envContent += "SMTP_FROM_ADDRESS=no-reply@example.com\n"
		envContent += "SMTP_FROM_NAME=FoundryStack\n"
		envContent += "SMTP_TLS_MODE=starttls\n"
		envContent += "SMTP_TIMEOUT_SECONDS=10\n"
		if err := os.WriteFile(envFile, []byte(envContent), 0600); err != nil {
			t.Fatalf("failed to write environment file: %v", err)
		}

		cfg, err := LoadConfig(Options{
			Environment: EnvProduction,
			EnvFile:     envFile,
			Debug:       true,
		})
		if err != nil {
			t.Fatalf("LoadConfig() unexpected error: %v", err)
		}

		if !cfg.Debug {
			t.Errorf("Debug = false, want true when debug option is enabled")
		}
	})

	t.Run("prod requires non-local allowed origin", func(t *testing.T) {
		clearLoadConfigEnv(t)
		t.Setenv("APP_ENV", "prod")
		envFile := filepath.Join(t.TempDir(), ".env.prod")
		if err := os.WriteFile(envFile, []byte("APP_ENV=prod\n"), 0600); err != nil {
			t.Fatalf("failed to write environment file: %v", err)
		}

		_, err := LoadConfig(Options{Environment: EnvProduction, EnvFile: envFile})
		if err == nil {
			t.Error("LoadConfig() expected validation error, got nil")
		}
	})
}

func writeEnvironmentFile(t *testing.T, content string) string {
	t.Helper()

	if !strings.Contains(content, "JWT_SECRET_BASE64=") {
		content += "JWT_SECRET_BASE64=" + testJWTSecretBase64 + "\n"
	}
	if !strings.Contains(content, "JWT_ISSUER=") {
		content += "JWT_ISSUER=" + testJWTIssuer + "\n"
	}
	if !strings.Contains(content, "JWT_AUDIENCE=") {
		content += "JWT_AUDIENCE=" + testJWTAudience + "\n"
	}
	if !strings.Contains(content, "JWT_CLIENT_IDS=") {
		content += "JWT_CLIENT_IDS=" + testJWTClientIDs + "\n"
	}
	if !strings.Contains(content, "EMAIL_PROVIDER=") {
		content += "EMAIL_PROVIDER=memory\n"
	}

	if !strings.Contains(
		content,
		"EMAIL_VERIFICATION_CODE_SECRET_BASE64=",
	) {
		content += "EMAIL_VERIFICATION_CODE_SECRET_BASE64=" +
			testEmailVerificationCodeSecretBase64 + "\n"
	}

	if !strings.Contains(content, "PASSWORD_RESET_CODE_SECRET_BASE64=") {
		content += "PASSWORD_RESET_CODE_SECRET_BASE64=" +
			testPasswordResetCodeSecretBase64 + "\n"
	}
	if !strings.Contains(content, "PASSWORD_RESET_TOKEN_SECRET_BASE64=") {
		content += "PASSWORD_RESET_TOKEN_SECRET_BASE64=" +
			testPasswordResetTokenSecretBase64 + "\n"
	}

	if !strings.Contains(
		content,
		"EMAIL_VERIFICATION_TOKEN_SECRET_BASE64=",
	) {
		content += "EMAIL_VERIFICATION_TOKEN_SECRET_BASE64=" +
			testEmailVerificationTokenSecretBase64 + "\n"
	}

	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write environment file: %v", err)
	}

	return path
}

func passwordResetEnvironment() string {
	return "PASSWORD_RESET_CODE_SECRET_BASE64=" +
		testPasswordResetCodeSecretBase64 + "\n" +
		"PASSWORD_RESET_TOKEN_SECRET_BASE64=" +
		testPasswordResetTokenSecretBase64 + "\n"
}
