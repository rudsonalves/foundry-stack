package bootstrap

import (
	"testing"
	"time"
)

// --- Config.validate() ---

func TestConfig_validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid config",
			cfg: Config{
				Environment:     "dev",
				Port:            "8080",
				AllowedOrigin:   "http://localhost:5173",
				ReadTimeout:     5 * time.Second,
				ShutdownTimeout: 10 * time.Second,
				JWT: JWTConfig{
					Secret:     []byte("0123456789abcdef0123456789abcdef"),
					Issuer:     "foundry-stack-test",
					Audience:   "foundry-stack-api",
					ClientIDs:  []string{"foundry-stack-mobile", "foundry-stack-swagger"},
					AccessTTL:  15 * time.Minute,
					RefreshTTL: 48 * time.Hour,
				},
				Database: DBConfig{
					URL:             "postgres://foundry_stack:secret@localhost:5432/foundry_stack",
					MaxOpenConns:    10,
					MaxIdleConns:    5,
					ConnMaxIdleTime: 5 * time.Minute,
					ConnMaxLifetime: 30 * time.Minute,
				},
				Email: EmailConfig{
					Provider: EmailProviderMemory,
				},
				EmailVerification: validEmailVerificationConfig(),
				PasswordReset:     validPasswordResetConfig(),
			},
			wantErr: false,
		},
		{
			name: "empty port",
			cfg: Config{
				Environment:   "dev",
				Port:          "",
				AllowedOrigin: "http://localhost:5173",
				ReadTimeout:   5 * time.Second,
			},
			wantErr: true,
			errMsg:  "port can't be empty",
		},
		{
			name: "non-numeric port",
			cfg: Config{
				Environment:   "dev",
				Port:          "http",
				AllowedOrigin: "http://localhost:5173",
				ReadTimeout:   5 * time.Second,
			},
			wantErr: true,
			errMsg:  "invalid port \"http\": expected a number between 1 and 65535",
		},
		{
			name: "port below range",
			cfg: Config{
				Environment:   "dev",
				Port:          "0",
				AllowedOrigin: "http://localhost:5173",
				ReadTimeout:   5 * time.Second,
			},
			wantErr: true,
			errMsg:  "invalid port \"0\": expected a number between 1 and 65535",
		},
		{
			name: "port above range",
			cfg: Config{
				Environment:   "dev",
				Port:          "65536",
				AllowedOrigin: "http://localhost:5173",
				ReadTimeout:   5 * time.Second,
			},
			wantErr: true,
			errMsg:  "invalid port \"65536\": expected a number between 1 and 65535",
		},
		{
			name: "invalid environment",
			cfg: Config{
				Environment:   "production",
				Port:          "8080",
				AllowedOrigin: "https://example.com",
				ReadTimeout:   5 * time.Second,
			},
			wantErr: true,
			errMsg:  "invalid environment \"production\"",
		},
		{
			name: "zero read timeout",
			cfg: Config{
				Environment:   "dev",
				Port:          "8080",
				AllowedOrigin: "http://localhost:5173",
				ReadTimeout:   0,
			},
			wantErr: true,
			errMsg:  "read timeout must be greater than zero",
		},
		{
			name: "negative read timeout",
			cfg: Config{
				Environment:   "dev",
				Port:          "8080",
				AllowedOrigin: "http://localhost:5173",
				ReadTimeout:   -1 * time.Second,
			},
			wantErr: true,
			errMsg:  "read timeout must be greater than zero",
		},
		{
			name: "empty allowed origin",
			cfg: Config{
				Environment:   "prod",
				Port:          "8080",
				AllowedOrigin: "",
				ReadTimeout:   5 * time.Second,
			},
			wantErr: true,
			errMsg:  "allowed origin is required",
		},
		{
			name: "prod with local allowed origin",
			cfg: Config{
				Environment:   "prod",
				Port:          "8080",
				AllowedOrigin: "http://localhost:5173",
				ReadTimeout:   5 * time.Second,
			},
			wantErr: true,
			errMsg:  "allowed origin \"http://localhost:5173\" is invalid for environment \"prod\"",
		},
		{
			name: "staging with local allowed origin",
			cfg: Config{
				Environment:   "stag",
				Port:          "8080",
				AllowedOrigin: "http://localhost:5173",
				ReadTimeout:   5 * time.Second,
			},
			wantErr: true,
			errMsg:  "allowed origin \"http://localhost:5173\" is invalid for environment \"stag\"",
		},
		{
			name: "zero refresh ttl",
			cfg: Config{
				Environment:   "dev",
				Port:          "8080",
				AllowedOrigin: "http://localhost:5173",
				ReadTimeout:   5 * time.Second,
				JWT: JWTConfig{
					Secret:     []byte("0123456789abcdef0123456789abcdef"),
					Issuer:     "foundry-stack-test",
					Audience:   "foundry-stack-api",
					ClientIDs:  []string{"foundry-stack-mobile"},
					AccessTTL:  15 * time.Minute,
					RefreshTTL: 0,
				},
				Database: DBConfig{
					URL:             "postgres://foundry_stack:secret@localhost:5432/foundry_stack",
					MaxOpenConns:    10,
					MaxIdleConns:    5,
					ConnMaxIdleTime: 5 * time.Minute,
					ConnMaxLifetime: 30 * time.Minute,
				},
			},
			wantErr: true,
			errMsg:  "JWT refresh TTL must be greater than zero",
		},
		{
			name: "negative refresh ttl",
			cfg: Config{
				Environment:   "dev",
				Port:          "8080",
				AllowedOrigin: "http://localhost:5173",
				ReadTimeout:   5 * time.Second,
				JWT: JWTConfig{
					Secret:     []byte("0123456789abcdef0123456789abcdef"),
					Issuer:     "foundry-stack-test",
					Audience:   "foundry-stack-api",
					ClientIDs:  []string{"foundry-stack-mobile"},
					AccessTTL:  15 * time.Minute,
					RefreshTTL: -1 * time.Second,
				},
				Database: DBConfig{
					URL:             "postgres://foundry_stack:secret@localhost:5432/foundry_stack",
					MaxOpenConns:    10,
					MaxIdleConns:    5,
					ConnMaxIdleTime: 5 * time.Minute,
					ConnMaxLifetime: 30 * time.Minute,
				},
			},
			wantErr: true,
			errMsg:  "JWT refresh TTL must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err.Error() != tt.errMsg {
				t.Errorf("validate() error message = %q, want %q", err.Error(), tt.errMsg)
			}
		})
	}
}

func TestDBConfig_validate(t *testing.T) {
	valid := DBConfig{
		URL:             "postgres://foundry_stack:secret@localhost:5432/foundry_stack",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxIdleTime: 5 * time.Minute,
		ConnMaxLifetime: 30 * time.Minute,
	}

	tests := []struct {
		name   string
		change func(*DBConfig)
		errMsg string
	}{
		{"empty URL", func(c *DBConfig) { c.URL = "" }, "database URL is required"},
		{"zero max open", func(c *DBConfig) { c.MaxOpenConns = 0 }, "database max open connections must be greater than zero"},
		{"negative max idle", func(c *DBConfig) { c.MaxIdleConns = -1 }, "database max idle connections can't be negative"},
		{"max idle exceeds max open", func(c *DBConfig) { c.MaxIdleConns = 11 }, "database max idle connections can't exceed max open connections"},
		{"zero max idle time", func(c *DBConfig) { c.ConnMaxIdleTime = 0 }, "database max idle time must be greater than zero"},
		{"zero max lifetime", func(c *DBConfig) { c.ConnMaxLifetime = 0 }, "database max lifetime must be greater than zero"},
	}

	if err := valid.validate(); err != nil {
		t.Fatalf("valid DBConfig returned error: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid
			tt.change(&cfg)

			err := cfg.validate()
			if err == nil {
				t.Fatal("validate() expected error, got nil")
			}
			if err.Error() != tt.errMsg {
				t.Errorf("validate() error = %q, want %q", err, tt.errMsg)
			}
		})
	}
}

func TestJWTConfig_validate(t *testing.T) {
	valid := JWTConfig{
		Secret:    []byte("0123456789abcdef0123456789abcdef"),
		Issuer:    "foundry-stack",
		Audience:  "foundry-stack-api",
		ClientIDs: []string{"foundry-stack-mobile", "foundry-stack-swagger"},
		AccessTTL: 15 * time.Minute,
	}

	tests := []struct {
		name   string
		change func(*JWTConfig)
		errMsg string
	}{
		{
			name: "valid config",
			change: func(*JWTConfig) {
			},
		},
		{
			name: "short secret",
			change: func(c *JWTConfig) {
				c.Secret = []byte("short")
			},
			errMsg: "JWT secret must be at least 32 bytes long",
		},
		{
			name: "missing issuer",
			change: func(c *JWTConfig) {
				c.Issuer = ""
			},
			errMsg: "JWT issuer is required",
		},
		{
			name: "missing audience",
			change: func(c *JWTConfig) {
				c.Audience = ""
			},
			errMsg: "JWT audience is required",
		},
		{
			name: "zero ttl",
			change: func(c *JWTConfig) {
				c.AccessTTL = 0
			},
			errMsg: "JWT access TTL must be greater than zero",
		},
		{
			name: "negative ttl",
			change: func(c *JWTConfig) {
				c.AccessTTL = -1 * time.Second
			},
			errMsg: "JWT access TTL must be greater than zero",
		},
		{
			name: "missing client IDs",
			change: func(c *JWTConfig) {
				c.ClientIDs = nil
			},
			errMsg: "JWT client IDs must be specified",
		},
		{
			name: "empty client ID",
			change: func(c *JWTConfig) {
				c.ClientIDs = []string{"foundry-stack-mobile", " "}
			},
			errMsg: "JWT client IDs can't contain empty values",
		},
		{
			name: "duplicate client ID",
			change: func(c *JWTConfig) {
				c.ClientIDs = []string{"foundry-stack-mobile", " foundry-stack-mobile "}
			},
			errMsg: "duplicate JWT client ID: \"foundry-stack-mobile\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := valid
			tt.change(&cfg)

			err := cfg.validate()
			if tt.errMsg == "" {
				if err != nil {
					t.Fatalf("validate() unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("validate() expected error, got nil")
			}
			if err.Error() != tt.errMsg {
				t.Fatalf("validate() error = %q, want %q", err.Error(), tt.errMsg)
			}
		})
	}
}

// --- Config.Addr() ---
func TestEmailConfigValidate(t *testing.T) {
	validSMTP := SMTPConfig{
		Host:        "smtp.example.com",
		Port:        587,
		Username:    "foundry-stack",
		Password:    "secret",
		FromAddress: "no-reply@example.com",
		FromName:    "FoundryStack",
		TLSMode:     SMTPTLSModeSTARTTLS,
		Timeout:     10 * time.Second,
	}

	tests := []struct {
		name        string
		config      EmailConfig
		environment string
		errMsg      string
	}{
		{
			name: "accepts memory in development",
			config: EmailConfig{
				Provider: EmailProviderMemory,
			},
			environment: EnvDevelopment,
		},
		{
			name: "accepts memory in test",
			config: EmailConfig{
				Provider: EmailProviderMemory,
			},
			environment: EnvTest,
		},
		{
			name: "rejects memory in production",
			config: EmailConfig{
				Provider: EmailProviderMemory,
			},
			environment: EnvProduction,
			errMsg:      "memory email provider is only allowed in development and test",
		},
		{
			name: "accepts SMTP",
			config: EmailConfig{
				Provider: EmailProviderSMTP,
				SMTP:     validSMTP,
			},
			environment: EnvProduction,
		},
		{
			name:        "rejects missing provider",
			config:      EmailConfig{},
			environment: EnvDevelopment,
			errMsg:      `invalid email provider ""`,
		},
		{
			name: "rejects unknown provider",
			config: EmailConfig{
				Provider: "unknown",
			},
			environment: EnvDevelopment,
			errMsg:      `invalid email provider "unknown"`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.config.validate(test.environment)

			if test.errMsg == "" {
				if err != nil {
					t.Fatalf("validate() unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("validate() error = nil")
			}

			if err.Error() != test.errMsg {
				t.Fatalf(
					"validate() error = %q, want %q",
					err.Error(),
					test.errMsg,
				)
			}
		})
	}
}

func TestSMTPConfigValidate(t *testing.T) {
	valid := SMTPConfig{
		Host:        "smtp.example.com",
		Port:        587,
		Username:    "foundry-stack",
		Password:    "secret",
		FromAddress: "no-reply@example.com",
		FromName:    "FoundryStack",
		TLSMode:     SMTPTLSModeSTARTTLS,
		Timeout:     10 * time.Second,
	}

	tests := []struct {
		name   string
		change func(*SMTPConfig)
		errMsg string
	}{
		{
			name:   "valid config",
			change: func(*SMTPConfig) {},
		},
		{
			name: "implicit TLS",
			change: func(config *SMTPConfig) {
				config.TLSMode = SMTPTLSModeImplicit
			},
		},
		{
			name: "missing host",
			change: func(config *SMTPConfig) {
				config.Host = ""
			},
			errMsg: "SMTP host is required",
		},
		{
			name: "invalid port",
			change: func(config *SMTPConfig) {
				config.Port = 0
			},
			errMsg: "SMTP port must be between 1 and 65535",
		},
		{
			name: "port above range",
			change: func(config *SMTPConfig) {
				config.Port = 65536
			},
			errMsg: "SMTP port must be between 1 and 65535",
		},
		{
			name: "missing username",
			change: func(config *SMTPConfig) {
				config.Username = ""
			},
			errMsg: "SMTP username is required",
		},
		{
			name: "missing password",
			change: func(config *SMTPConfig) {
				config.Password = ""
			},
			errMsg: "SMTP password is required",
		},
		{
			name: "missing from address",
			change: func(config *SMTPConfig) {
				config.FromAddress = ""
			},
			errMsg: "SMTP from address is required",
		},
		{
			name: "invalid from address",
			change: func(config *SMTPConfig) {
				config.FromAddress = "invalid"
			},
			errMsg: "SMTP from address is invalid",
		},
		{
			name: "missing from name",
			change: func(config *SMTPConfig) {
				config.FromName = ""
			},
			errMsg: "SMTP from name is required",
		},
		{
			name: "invalid TLS mode",
			change: func(config *SMTPConfig) {
				config.TLSMode = "none"
			},
			errMsg: `invalid SMTP TLS mode "none"`,
		},
		{
			name: "invalid timeout",
			change: func(config *SMTPConfig) {
				config.Timeout = 0
			},
			errMsg: "SMTP timeout must be greater than zero",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := valid
			test.change(&config)

			err := config.validate()

			if test.errMsg == "" {
				if err != nil {
					t.Fatalf("validate() unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("validate() error = nil")
			}

			if err.Error() != test.errMsg {
				t.Fatalf(
					"validate() error = %q, want %q",
					err.Error(),
					test.errMsg,
				)
			}
		})
	}
}

func TestEmailVerificationConfigValidate(t *testing.T) {
	tests := []struct {
		name   string
		change func(*EmailVerificationConfig)
		errMsg string
	}{
		{
			name:   "valid config",
			change: func(*EmailVerificationConfig) {},
		},
		{
			name: "short code secret",
			change: func(config *EmailVerificationConfig) {
				config.CodeSecret = []byte("short")
			},
			errMsg: "email verification code secret must be at least 32 bytes long",
		},
		{
			name: "short token secret",
			change: func(config *EmailVerificationConfig) {
				config.TokenSecret = []byte("short")
			},
			errMsg: "email verification token secret must be at least 32 bytes long",
		},
		{
			name: "shared secret",
			change: func(config *EmailVerificationConfig) {
				config.TokenSecret = append(
					[]byte(nil),
					config.CodeSecret...,
				)
			},
			errMsg: "email verification code and token secrets must be distinct",
		},
		{
			name: "zero code TTL",
			change: func(config *EmailVerificationConfig) {
				config.CodeTTL = 0
			},
			errMsg: "email verification code TTL must be greater than zero",
		},
		{
			name: "zero token TTL",
			change: func(config *EmailVerificationConfig) {
				config.TokenTTL = 0
			},
			errMsg: "email verification token TTL must be greater than zero",
		},
		{
			name: "zero max attempts",
			change: func(config *EmailVerificationConfig) {
				config.MaxConfirmAttempts = 0
			},
			errMsg: "email verification max confirm attempts must be greater than zero",
		},
		{
			name: "zero resend cooldown",
			change: func(config *EmailVerificationConfig) {
				config.ResendCooldown = 0
			},
			errMsg: "email verification resend cooldown must be greater than zero",
		},
		{
			name: "zero email limit",
			change: func(config *EmailVerificationConfig) {
				config.ResendEmailLimitPerHour = 0
			},
			errMsg: "email verification resend email limit must be greater than zero",
		},
		{
			name: "zero IP limit",
			change: func(config *EmailVerificationConfig) {
				config.ResendIPLimitPerHour = 0
			},
			errMsg: "email verification resend IP limit must be greater than zero",
		},
		{
			name: "zero retention",
			change: func(config *EmailVerificationConfig) {
				config.Retention = 0
			},
			errMsg: "email verification retention must be greater than zero",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := validEmailVerificationConfig()
			test.change(&config)

			err := config.validate()

			if test.errMsg == "" {
				if err != nil {
					t.Fatalf("validate() unexpected error: %v", err)
				}
				return
			}

			if err == nil {
				t.Fatal("validate() error = nil")
			}

			if err.Error() != test.errMsg {
				t.Fatalf(
					"validate() error = %q, want %q",
					err.Error(),
					test.errMsg,
				)
			}
		})
	}
}

func TestConfig_Addr(t *testing.T) {
	tests := []struct {
		name string
		port string
		want string
	}{
		{"standard port", "8080", ":8080"},
		{"port 80", "80", ":80"},
		{"port 443", "443", ":443"},
		{"empty port", "", ":"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{Port: tt.port}
			if got := cfg.Addr(); got != tt.want {
				t.Errorf("Addr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func validEmailVerificationConfig() EmailVerificationConfig {
	return EmailVerificationConfig{
		CodeSecret:              []byte("0123456789abcdef0123456789abcdef"),
		CodeTTL:                 15 * time.Minute,
		TokenSecret:             []byte("fedcba9876543210fedcba9876543210"),
		TokenTTL:                24 * time.Hour,
		MaxConfirmAttempts:      5,
		ResendCooldown:          time.Minute,
		ResendEmailLimitPerHour: 5,
		ResendIPLimitPerHour:    10,
		Retention:               24 * time.Hour,
	}
}

func validPasswordResetConfig() PasswordResetConfig {
	return PasswordResetConfig{
		CodeSecret:         []byte("password-reset-code-secret-32bytes"),
		CodeTTL:            15 * time.Minute,
		TokenSecret:        []byte("password-reset-token-secret-32byt"),
		TokenTTL:           15 * time.Minute,
		MaxConfirmAttempts: 5,
		ResendCooldown:     time.Minute,
		EmailLimitPerHour:  5,
		IPLimitPerHour:     10,
		Retention:          24 * time.Hour,
	}
}

func TestPasswordResetConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		change func(*PasswordResetConfig)
		errMsg string
	}{
		{name: "valid config", change: func(*PasswordResetConfig) {}},
		{
			name: "short code secret",
			change: func(config *PasswordResetConfig) {
				config.CodeSecret = []byte("short")
			},
			errMsg: "password reset code secret must be at least 32 bytes long",
		},
		{
			name: "short token secret",
			change: func(config *PasswordResetConfig) {
				config.TokenSecret = []byte("short")
			},
			errMsg: "password reset token secret must be at least 32 bytes long",
		},
		{
			name: "shared secret",
			change: func(config *PasswordResetConfig) {
				config.TokenSecret = append([]byte(nil), config.CodeSecret...)
			},
			errMsg: "password reset code and token secrets must be distinct",
		},
		{
			name:   "zero code TTL",
			change: func(config *PasswordResetConfig) { config.CodeTTL = 0 },
			errMsg: "password reset code TTL must be greater than zero",
		},
		{
			name:   "zero token TTL",
			change: func(config *PasswordResetConfig) { config.TokenTTL = 0 },
			errMsg: "password reset token TTL must be greater than zero",
		},
		{
			name:   "zero max attempts",
			change: func(config *PasswordResetConfig) { config.MaxConfirmAttempts = 0 },
			errMsg: "password reset max confirm attempts must be between 1 and 5",
		},
		{
			name:   "max attempts exceeds database constraint",
			change: func(config *PasswordResetConfig) { config.MaxConfirmAttempts = 6 },
			errMsg: "password reset max confirm attempts must be between 1 and 5",
		},
		{
			name:   "zero cooldown",
			change: func(config *PasswordResetConfig) { config.ResendCooldown = 0 },
			errMsg: "password reset resend cooldown must be greater than zero",
		},
		{
			name:   "zero email limit",
			change: func(config *PasswordResetConfig) { config.EmailLimitPerHour = 0 },
			errMsg: "password reset email limit must be greater than zero",
		},
		{
			name:   "zero IP limit",
			change: func(config *PasswordResetConfig) { config.IPLimitPerHour = 0 },
			errMsg: "password reset IP limit must be greater than zero",
		},
		{
			name:   "zero retention",
			change: func(config *PasswordResetConfig) { config.Retention = 0 },
			errMsg: "password reset retention must be greater than zero",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			config := validPasswordResetConfig()
			tt.change(&config)

			err := config.validate()
			if tt.errMsg == "" {
				if err != nil {
					t.Fatalf("validate() unexpected error: %v", err)
				}
				return
			}
			if err == nil || err.Error() != tt.errMsg {
				t.Fatalf("validate() error = %v, want %q", err, tt.errMsg)
			}
		})
	}
}

func TestPasswordResetSecretsMustDifferFromEmailVerification(t *testing.T) {
	t.Parallel()

	emailVerification := validEmailVerificationConfig()
	passwordReset := validPasswordResetConfig()
	passwordReset.CodeSecret = append([]byte(nil), emailVerification.TokenSecret...)

	err := validatePasswordResetSecretSeparation(emailVerification, passwordReset)
	if err == nil || err.Error() !=
		"password reset secrets must be distinct from email verification secrets" {
		t.Fatalf("validation error = %v", err)
	}
}

func TestValidateTrustedProxyCIDRs(t *testing.T) {
	t.Parallel()

	if err := validateTrustedProxyCIDRs([]string{"10.0.0.0/8", "2001:db8::/32"}); err != nil {
		t.Fatalf("valid CIDRs error = %v", err)
	}
	if err := validateTrustedProxyCIDRs([]string{"invalid"}); err == nil {
		t.Fatal("invalid CIDR error = nil")
	}
	if err := validateTrustedProxyCIDRs([]string{"10.0.0.0/8", "10.0.0.1/8"}); err == nil {
		t.Fatal("duplicate normalized CIDR error = nil")
	}
}
