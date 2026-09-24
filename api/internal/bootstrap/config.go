package bootstrap

import (
	"bytes"
	"errors"
	"fmt"
	"net/mail"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

type JWTConfig struct {
	Secret     []byte
	Issuer     string
	Audience   string
	ClientIDs  []string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func (c JWTConfig) validate() error {
	if len(c.Secret) < 32 {
		return errors.New("JWT secret must be at least 32 bytes long")
	}
	if c.Issuer == "" {
		return errors.New("JWT issuer is required")
	}
	if c.Audience == "" {
		return errors.New("JWT audience is required")
	}
	if c.AccessTTL <= 0 {
		return errors.New("JWT access TTL must be greater than zero")
	}
	if len(c.ClientIDs) == 0 {
		return errors.New("JWT client IDs must be specified")
	}

	seen := make(map[string]struct{}, len(c.ClientIDs))
	for _, clientID := range c.ClientIDs {
		normalized := strings.TrimSpace(clientID)
		if normalized == "" {
			return errors.New("JWT client IDs can't contain empty values")
		}
		if _, exists := seen[normalized]; exists {
			return fmt.Errorf("duplicate JWT client ID: %q", normalized)
		}
		seen[normalized] = struct{}{}
	}

	return nil
}

type DBConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
}

type Config struct {
	Environment       string
	Port              string
	AllowedOrigin     string
	ReadTimeout       time.Duration
	ShutdownTimeout   time.Duration
	Debug             bool
	Database          DBConfig
	JWT               JWTConfig
	Email             EmailConfig
	EmailVerification EmailVerificationConfig
	PasswordReset     PasswordResetConfig
	TrustedProxyCIDRs []string
}

func (c Config) Addr() string {
	return ":" + c.Port
}

func (c Config) validate() error {
	if c.Port == "" {
		return errors.New("port can't be empty")
	}

	port, err := strconv.Atoi(c.Port)
	if err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid port %q: expected a number between 1 and 65535", c.Port)
	}

	switch c.Environment {
	case EnvDevelopment, EnvStaging, EnvProduction, EnvTest:
	default:
		return fmt.Errorf("invalid environment %q", c.Environment)
	}

	if c.ReadTimeout <= 0 {
		return errors.New("read timeout must be greater than zero")
	}

	if c.AllowedOrigin == "" {
		return errors.New("allowed origin is required")
	}

	if (c.Environment == EnvStaging || c.Environment == EnvProduction) &&
		c.AllowedOrigin == "http://localhost:5173" {
		return fmt.Errorf(
			"allowed origin %q is invalid for environment %q",
			c.AllowedOrigin,
			c.Environment,
		)
	}

	if err := c.JWT.validate(); err != nil {
		return fmt.Errorf("JWT config validation: %w", err)
	}

	if c.JWT.RefreshTTL <= 0 {
		return errors.New(
			"JWT refresh TTL must be greater than zero",
		)
	}

	if c.ShutdownTimeout <= 0 {
		return errors.New(
			"shutdown timeout must be greater than zero",
		)
	}

	if err := c.EmailVerification.validate(); err != nil {
		return fmt.Errorf(
			"email verification config validation: %w",
			err,
		)
	}

	if err := c.PasswordReset.validate(); err != nil {
		return fmt.Errorf(
			"password reset config validation: %w",
			err,
		)
	}

	if err := validatePasswordResetSecretSeparation(
		c.EmailVerification,
		c.PasswordReset,
	); err != nil {
		return err
	}

	if err := validateTrustedProxyCIDRs(c.TrustedProxyCIDRs); err != nil {
		return err
	}

	if err := c.Email.validate(c.Environment); err != nil {
		return fmt.Errorf("email config validation: %w", err)
	}

	return c.Database.validate()
}

func (c DBConfig) validate() error {
	if c.URL == "" {
		return errors.New("database URL is required")
	}

	if c.MaxOpenConns < 1 {
		return errors.New("database max open connections must be greater than zero")
	}

	if c.MaxIdleConns < 0 {
		return errors.New("database max idle connections can't be negative")
	}

	if c.MaxIdleConns > c.MaxOpenConns {
		return errors.New("database max idle connections can't exceed max open connections")
	}

	if c.ConnMaxIdleTime <= 0 {
		return errors.New("database max idle time must be greater than zero")
	}

	if c.ConnMaxLifetime <= 0 {
		return errors.New("database max lifetime must be greater than zero")
	}

	return nil
}

type EmailVerificationConfig struct {
	CodeSecret              []byte
	CodeTTL                 time.Duration
	TokenSecret             []byte
	TokenTTL                time.Duration
	MaxConfirmAttempts      int
	ResendCooldown          time.Duration
	ResendEmailLimitPerHour int
	ResendIPLimitPerHour    int
	Retention               time.Duration
}

type PasswordResetConfig struct {
	CodeSecret         []byte
	CodeTTL            time.Duration
	TokenSecret        []byte
	TokenTTL           time.Duration
	MaxConfirmAttempts int
	ResendCooldown     time.Duration
	EmailLimitPerHour  int
	IPLimitPerHour     int
	Retention          time.Duration
}

const (
	EmailProviderMemory = "memory"
	EmailProviderSMTP   = "smtp"

	SMTPTLSModeImplicit = "tls"
	SMTPTLSModeSTARTTLS = "starttls"
)

type EmailConfig struct {
	Provider string
	SMTP     SMTPConfig
}

type SMTPConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
	FromName    string
	TLSMode     string
	Timeout     time.Duration
}

func (c EmailConfig) validate(environment string) error {
	switch c.Provider {
	case EmailProviderMemory:
		if environment != EnvDevelopment && environment != EnvTest {
			return errors.New(
				"memory email provider is only allowed in development and test",
			)
		}

		return nil

	case EmailProviderSMTP:
		return c.SMTP.validate()

	default:
		return fmt.Errorf("invalid email provider %q", c.Provider)
	}
}

func (c SMTPConfig) validate() error {
	if c.Host == "" {
		return errors.New("SMTP host is required")
	}

	if c.Port < 1 || c.Port > 65535 {
		return errors.New(
			"SMTP port must be between 1 and 65535",
		)
	}

	if c.Username == "" {
		return errors.New("SMTP username is required")
	}

	if c.Password == "" {
		return errors.New("SMTP password is required")
	}

	if c.FromAddress == "" {
		return errors.New("SMTP from address is required")
	}

	address, err := mail.ParseAddress(c.FromAddress)
	if err != nil || address.Address != c.FromAddress {
		return errors.New("SMTP from address is invalid")
	}

	if c.FromName == "" {
		return errors.New("SMTP from name is required")
	}

	switch c.TLSMode {
	case SMTPTLSModeImplicit, SMTPTLSModeSTARTTLS:
	default:
		return fmt.Errorf("invalid SMTP TLS mode %q", c.TLSMode)
	}

	if c.Timeout <= 0 {
		return errors.New("SMTP timeout must be greater than zero")
	}

	return nil
}

func (c EmailVerificationConfig) validate() error {
	if len(c.CodeSecret) < 32 {
		return errors.New(
			"email verification code secret must be at least 32 bytes long",
		)
	}

	if len(c.TokenSecret) < 32 {
		return errors.New(
			"email verification token secret must be at least 32 bytes long",
		)
	}

	if bytes.Equal(c.CodeSecret, c.TokenSecret) {
		return errors.New(
			"email verification code and token secrets must be distinct",
		)
	}

	if c.CodeTTL <= 0 {
		return errors.New(
			"email verification code TTL must be greater than zero",
		)
	}

	if c.TokenTTL <= 0 {
		return errors.New(
			"email verification token TTL must be greater than zero",
		)
	}

	if c.MaxConfirmAttempts <= 0 {
		return errors.New(
			"email verification max confirm attempts must be greater than zero",
		)
	}

	if c.ResendCooldown <= 0 {
		return errors.New(
			"email verification resend cooldown must be greater than zero",
		)
	}

	if c.ResendEmailLimitPerHour <= 0 {
		return errors.New(
			"email verification resend email limit must be greater than zero",
		)
	}

	if c.ResendIPLimitPerHour <= 0 {
		return errors.New(
			"email verification resend IP limit must be greater than zero",
		)
	}

	if c.Retention <= 0 {
		return errors.New(
			"email verification retention must be greater than zero",
		)
	}

	return nil
}

func (c PasswordResetConfig) validate() error {
	if len(c.CodeSecret) < 32 {
		return errors.New(
			"password reset code secret must be at least 32 bytes long",
		)
	}

	if len(c.TokenSecret) < 32 {
		return errors.New(
			"password reset token secret must be at least 32 bytes long",
		)
	}

	if bytes.Equal(c.CodeSecret, c.TokenSecret) {
		return errors.New(
			"password reset code and token secrets must be distinct",
		)
	}

	if c.CodeTTL <= 0 {
		return errors.New("password reset code TTL must be greater than zero")
	}

	if c.TokenTTL <= 0 {
		return errors.New("password reset token TTL must be greater than zero")
	}

	if c.MaxConfirmAttempts < 1 || c.MaxConfirmAttempts > 5 {
		return errors.New(
			"password reset max confirm attempts must be between 1 and 5",
		)
	}

	if c.ResendCooldown <= 0 {
		return errors.New(
			"password reset resend cooldown must be greater than zero",
		)
	}

	if c.EmailLimitPerHour <= 0 {
		return errors.New(
			"password reset email limit must be greater than zero",
		)
	}

	if c.IPLimitPerHour <= 0 {
		return errors.New(
			"password reset IP limit must be greater than zero",
		)
	}

	if c.Retention <= 0 {
		return errors.New(
			"password reset retention must be greater than zero",
		)
	}

	return nil
}

func validatePasswordResetSecretSeparation(
	emailVerification EmailVerificationConfig,
	passwordReset PasswordResetConfig,
) error {
	passwordSecrets := [][]byte{
		passwordReset.CodeSecret,
		passwordReset.TokenSecret,
	}
	emailSecrets := [][]byte{
		emailVerification.CodeSecret,
		emailVerification.TokenSecret,
	}

	for _, passwordSecret := range passwordSecrets {
		for _, emailSecret := range emailSecrets {
			if bytes.Equal(passwordSecret, emailSecret) {
				return errors.New(
					"password reset secrets must be distinct from email verification secrets",
				)
			}
		}
	}

	return nil
}

func validateTrustedProxyCIDRs(values []string) error {
	seen := make(map[netip.Prefix]struct{}, len(values))
	for _, value := range values {
		prefix, err := netip.ParsePrefix(value)
		if err != nil {
			return fmt.Errorf("invalid trusted proxy CIDR %q: %w", value, err)
		}

		prefix = prefix.Masked()
		if _, exists := seen[prefix]; exists {
			return fmt.Errorf("duplicate trusted proxy CIDR %q", value)
		}
		seen[prefix] = struct{}{}
	}

	return nil
}
