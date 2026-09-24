package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/rudsonalves/foundry-stack/api/internal/bootstrap"

	authapp "github.com/rudsonalves/foundry-stack/api/internal/auth/application"
	userapp "github.com/rudsonalves/foundry-stack/api/internal/users/application"

	userdomain "github.com/rudsonalves/foundry-stack/api/internal/users/domain"

	authinfra "github.com/rudsonalves/foundry-stack/api/internal/auth/infrastructure"
	usersinfra "github.com/rudsonalves/foundry-stack/api/internal/users/infrastructure"

	authtransport "github.com/rudsonalves/foundry-stack/api/internal/auth/transport"
	usertransport "github.com/rudsonalves/foundry-stack/api/internal/users/transport"
)

// closeDB closes the database connection and logs any errors that occur during closure.
func closeDB(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("failed to close database: %v", err)
	}
}

func run(
	ctx context.Context,
	args []string,
) error {
	// Parse command-line options
	options, err := bootstrap.ParseOptions(args)
	if err != nil {
		return fmt.Errorf("invalid flags: %w", err)
	}

	// Load configuration from environment variables and files
	cfg, err := bootstrap.LoadConfig(options)
	if err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	return runConfigured(ctx, cfg)
}

var buildServer = func(
	cfg bootstrap.Config,
	handler http.Handler,
) httpServer {
	return newServer(cfg, handler)
}

func runConfigured(
	ctx context.Context,
	cfg bootstrap.Config,
) error {

	// Initialize the database connection
	db, err := openDatabase(cfg.Database)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer closeDB(db)

	if cfg.Debug {
		log.Println("debug mode enabled")
	}

	// Initialize token service
	tokenService, err := authinfra.NewTokenService(
		authinfra.TokenServiceConfig{
			Secret:    cfg.JWT.Secret,
			Issuer:    cfg.JWT.Issuer,
			Audience:  cfg.JWT.Audience,
			ClientIDs: cfg.JWT.ClientIDs,
			Lifetime:  cfg.JWT.AccessTTL,
		},
	)
	if err != nil {
		return fmt.Errorf("initialize token service: %w", err)
	}

	// Initialize repositories and services
	passwordHasher := usersinfra.NewBcryptHasher(bcrypt.DefaultCost)
	userRepository := usersinfra.NewSQLUserRepository(db)
	emailVerificationRepository :=
		usersinfra.NewSQLEmailVerificationRepository(db)
	passwordResetRepository := usersinfra.NewSQLPasswordResetRepository(db)
	refreshTokenRepository :=
		authinfra.NewSQLRefreshTokenRepository(db)

	emailVerificationHasher, err :=
		usersinfra.NewHMACEmailVerificationHasher(
			cfg.EmailVerification.CodeSecret,
			cfg.EmailVerification.TokenSecret,
		)
	if err != nil {
		return fmt.Errorf(
			"initialize email verification hasher: %w",
			err,
		)
	}

	clock := usersinfra.SystemClock{}
	passwordResetHasher, err := usersinfra.NewHMACPasswordResetHasher(cfg.PasswordReset.CodeSecret, cfg.PasswordReset.TokenSecret)
	if err != nil {
		return fmt.Errorf("initialize password reset hasher: %w", err)
	}

	var emailVerificationSender userdomain.EmailVerificationSender
	var passwordResetSender userdomain.PasswordResetSender
	switch cfg.Email.Provider {
	case bootstrap.EmailProviderMemory:
		emailVerificationSender =
			usersinfra.NewMemoryEmailVerificationSender()
		passwordResetSender = usersinfra.NewMemoryPasswordResetSender()

	case bootstrap.EmailProviderSMTP:
		emailVerificationSender =
			usersinfra.NewSMTPEmailVerificationSender(
				usersinfra.SMTPEmailVerificationSenderConfig{
					Host:        cfg.Email.SMTP.Host,
					Port:        cfg.Email.SMTP.Port,
					Username:    cfg.Email.SMTP.Username,
					Password:    cfg.Email.SMTP.Password,
					FromAddress: cfg.Email.SMTP.FromAddress,
					FromName:    cfg.Email.SMTP.FromName,
					TLSMode:     cfg.Email.SMTP.TLSMode,
					Timeout:     cfg.Email.SMTP.Timeout,
				},
			)
		passwordResetSender = usersinfra.NewSMTPPasswordResetSender(usersinfra.SMTPEmailVerificationSenderConfig{Host: cfg.Email.SMTP.Host, Port: cfg.Email.SMTP.Port, Username: cfg.Email.SMTP.Username, Password: cfg.Email.SMTP.Password, FromAddress: cfg.Email.SMTP.FromAddress, FromName: cfg.Email.SMTP.FromName, TLSMode: cfg.Email.SMTP.TLSMode, Timeout: cfg.Email.SMTP.Timeout})
	}

	if shouldLogEmailVerificationCode(cfg.Environment) {
		emailVerificationSender =
			usersinfra.NewLoggingEmailVerificationSender(
				emailVerificationSender,
				log.Default(),
			)
	}

	emailVerificationRateLimiter :=
		usersinfra.NewMemoryEmailVerificationRateLimiter(
			usersinfra.MemoryEmailVerificationRateLimiterConfig{
				Cooldown: cfg.EmailVerification.ResendCooldown,
				EmailLimitPerHour: cfg.EmailVerification.
					ResendEmailLimitPerHour,
				IPLimitPerHour: cfg.EmailVerification.
					ResendIPLimitPerHour,
			},
		)
	passwordResetRateLimiter := usersinfra.NewMemoryPasswordResetRateLimiter(usersinfra.MemoryPasswordResetRateLimiterConfig{Cooldown: cfg.PasswordReset.ResendCooldown, EmailLimitPerHour: cfg.PasswordReset.EmailLimitPerHour, IPLimitPerHour: cfg.PasswordReset.IPLimitPerHour})

	// Initialize application services
	userService := userapp.NewService(
		userRepository,
		passwordHasher,
		emailVerificationHasher,
		clock,
	)

	emailVerificationService :=
		userapp.NewEmailVerificationService(
			userRepository,
			emailVerificationRepository,
			usersinfra.NewSecureEmailVerificationCodeGenerator(),
			usersinfra.NewSecureEmailVerificationTokenGenerator(),
			emailVerificationHasher,
			emailVerificationSender,
			clock,
			userapp.EmailVerificationServiceConfig{
				CodeTTL:  cfg.EmailVerification.CodeTTL,
				TokenTTL: cfg.EmailVerification.TokenTTL,
				ResendCooldown: cfg.EmailVerification.
					ResendCooldown,
				MaxConfirmAttempts: cfg.EmailVerification.
					MaxConfirmAttempts,
			},
			emailVerificationRateLimiter,
		)
	passwordResetService := userapp.NewPasswordResetService(userRepository, passwordResetRepository, usersinfra.NewSecurePasswordResetCodeGenerator(), usersinfra.NewSecurePasswordResetTokenGenerator(), passwordResetHasher, passwordResetSender, clock, userapp.PasswordResetServiceConfig{CodeTTL: cfg.PasswordReset.CodeTTL, TokenTTL: cfg.PasswordReset.TokenTTL, ResendCooldown: cfg.PasswordReset.ResendCooldown, MaxConfirmAttempts: cfg.PasswordReset.MaxConfirmAttempts}, passwordResetRateLimiter, log.Default(), userapp.PasswordResetCompletionDependencies{Completer: passwordResetRepository, Passwords: passwordHasher})
	passwordResetCleaner := userapp.NewPasswordResetCleaner(passwordResetRepository, clock, cfg.PasswordReset.Retention, time.Hour, log.Default())
	cleanerCtx, stopPasswordResetCleaner := context.WithCancel(ctx)
	passwordResetCleanerStopped := make(chan struct{})
	go func() {
		defer close(passwordResetCleanerStopped)
		passwordResetCleaner.Run(cleanerCtx)
	}()
	defer func() {
		stopPasswordResetCleaner()
		<-passwordResetCleanerStopped
	}()

	refreshTokenService := authapp.NewRefreshTokenService(
		refreshTokenRepository,
		cfg.JWT.RefreshTTL,
	)
	authService := authapp.NewAuthService(
		userService,
		tokenService,
		refreshTokenService,
	)

	// Initialize HTTP handlers
	userHandler := usertransport.NewHandler(userService)
	emailVerificationHandler :=
		usertransport.NewEmailVerificationHandler(
			emailVerificationService,
		)
	authHandler := authtransport.NewHandler(authService)
	clientIPResolver, err := usertransport.NewClientIPResolver(cfg.TrustedProxyCIDRs)
	if err != nil {
		return fmt.Errorf("initialize client IP resolver: %w", err)
	}
	passwordResetHandler := usertransport.NewPasswordResetHandler(passwordResetService, clientIPResolver)

	// Initialize middleware and server
	handler := newMiddleware(
		cfg,
		userHandler,
		emailVerificationHandler,
		authHandler,
		passwordResetHandler,
	)

	// Initialize and start the HTTP server
	server := buildServer(cfg, handler)
	log.Printf("server initialized: addr=%s env=%s",
		cfg.Addr(),
		cfg.Environment,
	)

	// Start the server and handle errors
	return serve(
		ctx,
		server,
		cfg.ShutdownTimeout,
	)
}

func shouldLogEmailVerificationCode(environment string) bool {
	return environment == bootstrap.EnvDevelopment ||
		environment == bootstrap.EnvStaging
}
