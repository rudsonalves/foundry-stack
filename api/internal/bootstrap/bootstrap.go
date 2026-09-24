package bootstrap

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const (
	EnvDevelopment = "dev"
	EnvStaging     = "stag"
	EnvProduction  = "prod"
	EnvTest        = "test"
)

func LoadConfig(opt Options) (Config, error) {
	if opt.Environment == "" {
		return Config{}, errors.New("environment option is required")
	}

	if opt.EnvFile == "" {
		return Config{}, errors.New("environment file option is required")
	}

	if err := godotenv.Load(opt.EnvFile); err != nil {
		return Config{}, fmt.Errorf(
			"load environment file %q: %w",
			opt.EnvFile,
			err,
		)
	}

	env := os.Getenv("APP_ENV")
	if env == "" {
		return Config{}, errors.New("APP_ENV is required")
	}

	if env != opt.Environment {
		return Config{}, fmt.Errorf(
			"environment mismatch: selected %q, APP_ENV is %q",
			opt.Environment,
			env,
		)
	}

	allowedOriginDefault := ""

	if env == EnvDevelopment || env == EnvTest {
		allowedOriginDefault = "http://localhost:5173"
	}

	debug, err := envBool("DEBUG", env == EnvDevelopment)
	if err != nil {
		return Config{}, err
	}

	if opt.Debug {
		debug = true
	}

	readTimeout, err := envTime("READ_TIMEOUT_SECONDS", 5)
	if err != nil {
		return Config{}, err
	}

	dbURL := os.Getenv("DATABASE_URL")
	dbMaxOpenConns, err := envInt("DB_MAX_OPEN_CONNS", 10)
	if err != nil {
		return Config{}, err
	}
	dbMaxIdleConns, err := envInt("DB_MAX_IDLE_CONNS", 5)
	if err != nil {
		return Config{}, err
	}
	dbConnMaxIdleTime, err := envTime("DB_CONN_MAX_IDLE_TIME_SECONDS", 300)
	if err != nil {
		return Config{}, err
	}
	dbConnMaxLifetime, err := envTime("DB_CONN_MAX_LIFETIME_SECONDS", 1800)
	if err != nil {
		return Config{}, err
	}

	secret, err := base64.StdEncoding.DecodeString(
		os.Getenv("JWT_SECRET_BASE64"),
	)
	if err != nil {
		return Config{}, fmt.Errorf(
			"decode JWT secret: %w",
			err,
		)
	}
	accessTTL, err := envTime("JWT_ACCESS_TTL_SECONDS", 900)
	if err != nil {
		return Config{}, err
	}
	clientIDs := strings.Split(os.Getenv("JWT_CLIENT_IDS"), ",")
	for i := range clientIDs {
		clientIDs[i] = strings.TrimSpace(clientIDs[i])
	}

	refreshTTL, err := envTime("JWT_REFRESH_TTL_SECONDS", 2*24*60*60)
	if err != nil {
		return Config{}, err
	}

	emailVerificationCodeSecret, err := base64.StdEncoding.DecodeString(
		os.Getenv("EMAIL_VERIFICATION_CODE_SECRET_BASE64"),
	)
	if err != nil {
		return Config{}, fmt.Errorf(
			"decode email verification code secret: %w",
			err,
		)
	}

	emailVerificationTokenSecret, err := base64.StdEncoding.DecodeString(
		os.Getenv("EMAIL_VERIFICATION_TOKEN_SECRET_BASE64"),
	)
	if err != nil {
		return Config{}, fmt.Errorf(
			"decode email verification token secret: %w",
			err,
		)
	}

	emailVerificationCodeTTL, err := envTime(
		"EMAIL_VERIFICATION_CODE_TTL_SECONDS",
		15*60,
	)
	if err != nil {
		return Config{}, err
	}

	emailVerificationTokenTTL, err := envTime(
		"EMAIL_VERIFICATION_TOKEN_TTL_SECONDS",
		24*60*60,
	)
	if err != nil {
		return Config{}, err
	}

	emailVerificationMaxAttempts, err := envInt(
		"EMAIL_VERIFICATION_MAX_CONFIRM_ATTEMPTS",
		5,
	)
	if err != nil {
		return Config{}, err
	}

	emailVerificationResendCooldown, err := envTime(
		"EMAIL_VERIFICATION_RESEND_COOLDOWN_SECONDS",
		60,
	)
	if err != nil {
		return Config{}, err
	}

	emailVerificationEmailLimit, err := envInt(
		"EMAIL_VERIFICATION_RESEND_EMAIL_LIMIT_PER_HOUR",
		5,
	)
	if err != nil {
		return Config{}, err
	}

	emailVerificationIPLimit, err := envInt(
		"EMAIL_VERIFICATION_RESEND_IP_LIMIT_PER_HOUR",
		10,
	)
	if err != nil {
		return Config{}, err
	}

	emailVerificationRetention, err := envTime(
		"EMAIL_VERIFICATION_RETENTION_SECONDS",
		24*60*60,
	)
	if err != nil {
		return Config{}, err
	}

	passwordResetCodeSecret, err := base64.StdEncoding.DecodeString(
		os.Getenv("PASSWORD_RESET_CODE_SECRET_BASE64"),
	)
	if err != nil {
		return Config{}, fmt.Errorf(
			"decode password reset code secret: %w",
			err,
		)
	}

	passwordResetTokenSecret, err := base64.StdEncoding.DecodeString(
		os.Getenv("PASSWORD_RESET_TOKEN_SECRET_BASE64"),
	)
	if err != nil {
		return Config{}, fmt.Errorf(
			"decode password reset token secret: %w",
			err,
		)
	}

	passwordResetCodeTTL, err := envTime(
		"PASSWORD_RESET_CODE_TTL_SECONDS",
		15*60,
	)
	if err != nil {
		return Config{}, err
	}

	passwordResetTokenTTL, err := envTime(
		"PASSWORD_RESET_TOKEN_TTL_SECONDS",
		15*60,
	)
	if err != nil {
		return Config{}, err
	}

	passwordResetMaxAttempts, err := envInt(
		"PASSWORD_RESET_MAX_CONFIRM_ATTEMPTS",
		5,
	)
	if err != nil {
		return Config{}, err
	}

	passwordResetResendCooldown, err := envTime(
		"PASSWORD_RESET_RESEND_COOLDOWN_SECONDS",
		60,
	)
	if err != nil {
		return Config{}, err
	}

	passwordResetEmailLimit, err := envInt(
		"PASSWORD_RESET_EMAIL_LIMIT_PER_HOUR",
		5,
	)
	if err != nil {
		return Config{}, err
	}

	passwordResetIPLimit, err := envInt(
		"PASSWORD_RESET_IP_LIMIT_PER_HOUR",
		10,
	)
	if err != nil {
		return Config{}, err
	}

	passwordResetRetention, err := envTime(
		"PASSWORD_RESET_RETENTION_SECONDS",
		24*60*60,
	)
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := envTime(
		"SHUTDOWN_TIMEOUT_SECONDS",
		10,
	)
	if err != nil {
		return Config{}, err
	}

	smtpPort, err := envInt("SMTP_PORT", 587)
	if err != nil {
		return Config{}, err
	}

	smtpTimeout, err := envTime("SMTP_TIMEOUT_SECONDS", 10)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		Environment:     env,
		Port:            envOrDefault("PORT", "8080"),
		ReadTimeout:     readTimeout,
		ShutdownTimeout: shutdownTimeout,
		AllowedOrigin:   envOrDefault("ALLOWED_ORIGIN", allowedOriginDefault),
		Debug:           debug,
		Database: DBConfig{
			URL:             dbURL,
			MaxOpenConns:    dbMaxOpenConns,
			MaxIdleConns:    dbMaxIdleConns,
			ConnMaxIdleTime: dbConnMaxIdleTime,
			ConnMaxLifetime: dbConnMaxLifetime,
		},
		JWT: JWTConfig{
			Secret:     secret,
			Issuer:     os.Getenv("JWT_ISSUER"),
			Audience:   os.Getenv("JWT_AUDIENCE"),
			ClientIDs:  clientIDs,
			AccessTTL:  accessTTL,
			RefreshTTL: refreshTTL,
		},
		Email: EmailConfig{
			Provider: os.Getenv("EMAIL_PROVIDER"),
			SMTP: SMTPConfig{
				Host:        os.Getenv("SMTP_HOST"),
				Port:        smtpPort,
				Username:    os.Getenv("SMTP_USERNAME"),
				Password:    os.Getenv("SMTP_PASSWORD"),
				FromAddress: os.Getenv("SMTP_FROM_ADDRESS"),
				FromName:    os.Getenv("SMTP_FROM_NAME"),
				TLSMode:     os.Getenv("SMTP_TLS_MODE"),
				Timeout:     smtpTimeout,
			},
		},
		EmailVerification: EmailVerificationConfig{
			CodeSecret:              emailVerificationCodeSecret,
			CodeTTL:                 emailVerificationCodeTTL,
			TokenSecret:             emailVerificationTokenSecret,
			TokenTTL:                emailVerificationTokenTTL,
			MaxConfirmAttempts:      emailVerificationMaxAttempts,
			ResendCooldown:          emailVerificationResendCooldown,
			ResendEmailLimitPerHour: emailVerificationEmailLimit,
			ResendIPLimitPerHour:    emailVerificationIPLimit,
			Retention:               emailVerificationRetention,
		},
		PasswordReset: PasswordResetConfig{
			CodeSecret:         passwordResetCodeSecret,
			CodeTTL:            passwordResetCodeTTL,
			TokenSecret:        passwordResetTokenSecret,
			TokenTTL:           passwordResetTokenTTL,
			MaxConfirmAttempts: passwordResetMaxAttempts,
			ResendCooldown:     passwordResetResendCooldown,
			EmailLimitPerHour:  passwordResetEmailLimit,
			IPLimitPerHour:     passwordResetIPLimit,
			Retention:          passwordResetRetention,
		},
		TrustedProxyCIDRs: envCSV("TRUSTED_PROXY_CIDRS"),
	}

	return cfg, cfg.validate()
}

func envTime(key string, fallback int) (time.Duration, error) {
	value, err := envInt(key, fallback)
	if err != nil {
		return 0, err
	}

	return time.Duration(value) * time.Second, nil
}

func envOrDefault(key string, fallback string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	return value
}

func envInt(key string, fallback int) (int, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer: %w", key, err)
	}

	return intValue, nil
}

func envBool(key string, fallback bool) (bool, error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback, nil
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}

	return boolValue, nil
}

func envCSV(key string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}

	values := strings.Split(value, ",")
	for index := range values {
		values[index] = strings.TrimSpace(values[index])
	}

	return values
}
