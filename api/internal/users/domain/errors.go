package domain

import (
	"errors"
)

var (
	ErrUserNotFound               = errors.New("user not found")
	ErrEmailAlreadyExists         = errors.New("email already exists")
	ErrEmailVerificationNotFound  = errors.New("email verification not found")
	ErrInvalidEmailVerification   = errors.New("invalid email verification")
	ErrEmailVerificationCooldown  = errors.New("email verification resend cooldown")
	ErrEmailVerificationRateLimit = errors.New("email verification rate limit exceeded")
	ErrPasswordResetNotFound      = errors.New("password reset not found")
	ErrInvalidPasswordReset       = errors.New("invalid password reset")
	ErrPasswordResetConflict      = errors.New("password reset conflict")
	ErrPasswordResetCooldown      = errors.New("password reset resend cooldown")
	ErrPasswordResetRateLimit     = errors.New("password reset rate limit exceeded")
)
