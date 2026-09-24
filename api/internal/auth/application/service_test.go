package application

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	authdomain "github.com/rudsonalves/foundry-stack/api/internal/auth/domain"
	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	usersapp "github.com/rudsonalves/foundry-stack/api/internal/users/application"
	userdomain "github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

func TestAuthServiceLogin(t *testing.T) {
	t.Run("successful authentication issues token and returns issuer output", func(t *testing.T) {
		userID := uuid.MustParse("fd6ea0d1-4d34-4f4f-a9cb-2f137f775e7d")
		expiresAt := time.Date(2026, time.July, 20, 13, 0, 0, 0, time.UTC)
		authenticator := &stubCredentialAuthenticator{
			result: userdomain.User{ID: userID},
		}
		issuer := &stubTokenIssuer{
			token:     "jwt-token",
			expiresAt: expiresAt,
		}
		refreshTokens := &stubRefreshTokens{issueResult: RefreshTokenResult{Token: "refresh-token"}}
		svc := NewAuthService(authenticator, issuer, refreshTokens)

		result, err := svc.Login(
			context.Background(),
			"ada@example.com",
			"secure123",
			"foundry-stack-mobile",
		)
		if err != nil {
			t.Fatalf("Login() unexpected error: %v", err)
		}

		if authenticator.email != "ada@example.com" {
			t.Fatalf("Authenticate email = %q, want %q", authenticator.email, "ada@example.com")
		}
		if authenticator.password != "secure123" {
			t.Fatalf("Authenticate password = %q, want %q", authenticator.password, "secure123")
		}
		if issuer.issueCalls != 1 {
			t.Fatalf("Issue call count = %d, want 1", issuer.issueCalls)
		}
		if issuer.userID != userID {
			t.Fatalf("Issue userID = %s, want %s", issuer.userID, userID)
		}
		if issuer.clientID != "foundry-stack-mobile" {
			t.Fatalf("Issue clientID = %q, want %q", issuer.clientID, "foundry-stack-mobile")
		}
		if refreshTokens.issueCalls != 1 {
			t.Fatalf("refresh Issue call count = %d, want 1", refreshTokens.issueCalls)
		}
		if refreshTokens.issueUserID != userID {
			t.Fatalf("refresh Issue userID = %s, want %s", refreshTokens.issueUserID, userID)
		}
		if refreshTokens.issueClientID != "foundry-stack-mobile" {
			t.Fatalf("refresh Issue clientID = %q, want %q", refreshTokens.issueClientID, "foundry-stack-mobile")
		}
		if result.AccessToken != "jwt-token" {
			t.Fatalf("AccessToken = %q, want %q", result.AccessToken, "jwt-token")
		}
		if result.RefreshToken != "refresh-token" {
			t.Fatalf("RefreshToken = %q, want %q", result.RefreshToken, "refresh-token")
		}
		if !result.ExpiresAt.Equal(expiresAt) {
			t.Fatalf("ExpiresAt = %v, want %v", result.ExpiresAt, expiresAt)
		}
	})

	t.Run("invalid credentials become unauthorized app error and keep wrapping original error", func(t *testing.T) {
		authenticator := &stubCredentialAuthenticator{err: usersapp.ErrInvalidCredentials}
		issuer := &stubTokenIssuer{}
		svc := NewAuthService(authenticator, issuer, &stubRefreshTokens{})

		_, err := svc.Login(
			context.Background(),
			"ada@example.com",
			"wrong-password",
			"foundry-stack-mobile",
		)
		if err == nil {
			t.Fatal("Login() expected error, got nil")
		}

		appErr, ok := err.(*sharederrors.AppError)
		if !ok {
			t.Fatalf("error type = %T, want *AppError", err)
		}
		if appErr.Code != sharederrors.ErrCodeUnauthorized {
			t.Fatalf("AppError.Code = %q, want %q", appErr.Code, sharederrors.ErrCodeUnauthorized)
		}
		if appErr.Message != "invalid email or password" {
			t.Fatalf(
				"AppError.Message = %q, want %q",
				appErr.Message,
				"invalid email or password",
			)
		}
		if !errors.Is(err, usersapp.ErrInvalidCredentials) {
			t.Fatal("error should wrap usersapp.ErrInvalidCredentials")
		}
		if issuer.issueCalls != 0 {
			t.Fatalf("Issue call count = %d, want 0", issuer.issueCalls)
		}
		if svc.refreshTokens.(*stubRefreshTokens).issueCalls != 0 {
			t.Fatalf("refresh Issue call count = %d, want 0", svc.refreshTokens.(*stubRefreshTokens).issueCalls)
		}
	})

	t.Run("unexpected authenticator error is preserved and issuer is not called", func(t *testing.T) {
		authErr := errors.New("database unavailable")
		authenticator := &stubCredentialAuthenticator{err: authErr}
		issuer := &stubTokenIssuer{}
		svc := NewAuthService(authenticator, issuer, &stubRefreshTokens{})

		_, err := svc.Login(
			context.Background(),
			"ada@example.com",
			"secure123",
			"foundry-stack-mobile",
		)
		if !errors.Is(err, authErr) {
			t.Fatalf("Login() error = %v, want wrapped %v", err, authErr)
		}
		if issuer.issueCalls != 0 {
			t.Fatalf("Issue call count = %d, want 0", issuer.issueCalls)
		}
		if svc.refreshTokens.(*stubRefreshTokens).issueCalls != 0 {
			t.Fatalf("refresh Issue call count = %d, want 0", svc.refreshTokens.(*stubRefreshTokens).issueCalls)
		}
	})

	t.Run("issuer error is wrapped and remains discoverable with errors.Is", func(t *testing.T) {
		issueErr := errors.New("sign failed")
		authenticator := &stubCredentialAuthenticator{
			result: userdomain.User{ID: uuid.MustParse("8f8da709-275f-4e30-a2bf-c0132db45b90")},
		}
		issuer := &stubTokenIssuer{err: issueErr}
		svc := NewAuthService(authenticator, issuer, &stubRefreshTokens{})

		_, err := svc.Login(
			context.Background(),
			"ada@example.com",
			"secure123",
			"foundry-stack-mobile",
		)
		if err == nil {
			t.Fatal("Login() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "issue access token") {
			t.Fatalf("error = %q, want to contain %q", err.Error(), "issue access token")
		}
		if !errors.Is(err, issueErr) {
			t.Fatal("error should wrap issuer error")
		}
		if svc.refreshTokens.(*stubRefreshTokens).issueCalls != 0 {
			t.Fatalf("refresh Issue call count = %d, want 0", svc.refreshTokens.(*stubRefreshTokens).issueCalls)
		}
	})

	t.Run("client not allowed becomes unauthorized app error", func(t *testing.T) {
		authenticator := &stubCredentialAuthenticator{
			result: userdomain.User{ID: uuid.New()},
		}
		issuer := &stubTokenIssuer{err: authdomain.ErrClientNotAllowed}
		svc := NewAuthService(authenticator, issuer, &stubRefreshTokens{})

		_, err := svc.Login(
			context.Background(),
			"ada@example.com",
			"secure123",
			"unknown-client",
		)
		if err == nil {
			t.Fatal("Login() expected error, got nil")
		}

		appErr, ok := err.(*sharederrors.AppError)
		if !ok {
			t.Fatalf("error type = %T, want *AppError", err)
		}
		if appErr.Code != sharederrors.ErrCodeUnauthorized {
			t.Fatalf("AppError.Code = %q, want %q", appErr.Code, sharederrors.ErrCodeUnauthorized)
		}
		if appErr.Message != "client not allowed" {
			t.Fatalf("AppError.Message = %q, want %q", appErr.Message, "client not allowed")
		}
		if !errors.Is(err, authdomain.ErrClientNotAllowed) {
			t.Fatal("error should wrap authdomain.ErrClientNotAllowed")
		}
		if svc.refreshTokens.(*stubRefreshTokens).issueCalls != 0 {
			t.Fatalf("refresh Issue call count = %d, want 0", svc.refreshTokens.(*stubRefreshTokens).issueCalls)
		}
	})

	t.Run("refresh token issuance error is wrapped and returned", func(t *testing.T) {
		issueErr := errors.New("persist refresh token failed")
		authenticator := &stubCredentialAuthenticator{
			result: userdomain.User{ID: uuid.MustParse("f0f9d265-05de-46ff-abcc-f7256dbfef64")},
		}
		issuer := &stubTokenIssuer{token: "jwt-token", expiresAt: time.Now().Add(15 * time.Minute)}
		refreshTokens := &stubRefreshTokens{issueErr: issueErr}
		svc := NewAuthService(authenticator, issuer, refreshTokens)

		_, err := svc.Login(context.Background(), "ada@example.com", "secure123", "foundry-stack-mobile")
		if err == nil {
			t.Fatal("Login() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "issue refresh token") {
			t.Fatalf("error = %q, want to contain %q", err.Error(), "issue refresh token")
		}
		if !errors.Is(err, issueErr) {
			t.Fatal("error should wrap refresh token issuer error")
		}
		if refreshTokens.issueCalls != 1 {
			t.Fatalf("refresh Issue call count = %d, want 1", refreshTokens.issueCalls)
		}
	})
}

func TestAuthServiceRefresh(t *testing.T) {
	t.Run("successful refresh authenticates token and issues new access token", func(t *testing.T) {
		userID := uuid.MustParse("74356d10-064a-4109-9be7-00b08a32d48f")
		expiresAt := time.Date(2026, time.July, 22, 15, 0, 0, 0, time.UTC)
		issuer := &stubTokenIssuer{token: "new-access-token", expiresAt: expiresAt}
		refreshTokens := &stubRefreshTokens{
			authenticateResult: RefreshIdentity{UserID: userID, ClientID: "foundry-stack-mobile"},
		}
		svc := NewAuthService(&stubCredentialAuthenticator{}, issuer, refreshTokens)

		result, err := svc.Refresh(context.Background(), "refresh-token")
		if err != nil {
			t.Fatalf("Refresh() unexpected error: %v", err)
		}
		if refreshTokens.authenticateCalls != 1 {
			t.Fatalf("Authenticate() call count = %d, want 1", refreshTokens.authenticateCalls)
		}
		if refreshTokens.authenticateToken != "refresh-token" {
			t.Fatalf("Authenticate() token = %q, want %q", refreshTokens.authenticateToken, "refresh-token")
		}
		if issuer.userID != userID {
			t.Fatalf("Issue() userID = %s, want %s", issuer.userID, userID)
		}
		if issuer.clientID != "foundry-stack-mobile" {
			t.Fatalf("Issue() clientID = %q, want %q", issuer.clientID, "foundry-stack-mobile")
		}
		if result.AccessToken != "new-access-token" {
			t.Fatalf("AccessToken = %q, want %q", result.AccessToken, "new-access-token")
		}
		if !result.ExpiresAt.Equal(expiresAt) {
			t.Fatalf("ExpiresAt = %v, want %v", result.ExpiresAt, expiresAt)
		}
	})

	t.Run("invalid refresh token becomes unauthorized app error", func(t *testing.T) {
		refreshTokens := &stubRefreshTokens{authenticateErr: authdomain.ErrInvalidRefreshToken}
		issuer := &stubTokenIssuer{}
		svc := NewAuthService(&stubCredentialAuthenticator{}, issuer, refreshTokens)

		_, err := svc.Refresh(context.Background(), "refresh-token")
		if err == nil {
			t.Fatal("Refresh() expected error, got nil")
		}
		appErr, ok := err.(*sharederrors.AppError)
		if !ok {
			t.Fatalf("error type = %T, want *AppError", err)
		}
		if appErr.Code != sharederrors.ErrCodeUnauthorized {
			t.Fatalf("AppError.Code = %q, want %q", appErr.Code, sharederrors.ErrCodeUnauthorized)
		}
		if appErr.Message != "invalid refresh token" {
			t.Fatalf("AppError.Message = %q, want %q", appErr.Message, "invalid refresh token")
		}
		if issuer.issueCalls != 0 {
			t.Fatalf("Issue() call count = %d, want 0", issuer.issueCalls)
		}
	})

	t.Run("unexpected refresh authentication error is wrapped", func(t *testing.T) {
		authErr := errors.New("repository unavailable")
		svc := NewAuthService(
			&stubCredentialAuthenticator{},
			&stubTokenIssuer{},
			&stubRefreshTokens{authenticateErr: authErr},
		)

		_, err := svc.Refresh(context.Background(), "refresh-token")
		if err == nil {
			t.Fatal("Refresh() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "authenticate refresh token") {
			t.Fatalf("error = %q, want to contain %q", err.Error(), "authenticate refresh token")
		}
		if !errors.Is(err, authErr) {
			t.Fatal("error should wrap refresh authentication error")
		}
	})

	t.Run("access token issuance error is wrapped", func(t *testing.T) {
		issueErr := errors.New("sign failed")
		refreshTokens := &stubRefreshTokens{
			authenticateResult: RefreshIdentity{UserID: uuid.New(), ClientID: "foundry-stack-mobile"},
		}
		issuer := &stubTokenIssuer{err: issueErr}
		svc := NewAuthService(&stubCredentialAuthenticator{}, issuer, refreshTokens)

		_, err := svc.Refresh(context.Background(), "refresh-token")
		if err == nil {
			t.Fatal("Refresh() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "issue access token") {
			t.Fatalf("error = %q, want to contain %q", err.Error(), "issue access token")
		}
		if !errors.Is(err, issueErr) {
			t.Fatal("error should wrap access token issuer error")
		}
	})
}

func TestAuthServiceLogout(t *testing.T) {
	t.Run("successful logout revokes refresh token", func(t *testing.T) {
		refreshTokens := &stubRefreshTokens{}
		svc := NewAuthService(&stubCredentialAuthenticator{}, &stubTokenIssuer{}, refreshTokens)

		if err := svc.Logout(context.Background(), "refresh-token"); err != nil {
			t.Fatalf("Logout() unexpected error: %v", err)
		}
		if refreshTokens.revokeCalls != 1 {
			t.Fatalf("Revoke() call count = %d, want 1", refreshTokens.revokeCalls)
		}
		if refreshTokens.revokeToken != "refresh-token" {
			t.Fatalf("Revoke() token = %q, want %q", refreshTokens.revokeToken, "refresh-token")
		}
	})

	t.Run("revoke error is wrapped", func(t *testing.T) {
		revokeErr := errors.New("update failed")
		refreshTokens := &stubRefreshTokens{revokeErr: revokeErr}
		svc := NewAuthService(&stubCredentialAuthenticator{}, &stubTokenIssuer{}, refreshTokens)

		err := svc.Logout(context.Background(), "refresh-token")
		if err == nil {
			t.Fatal("Logout() expected error, got nil")
		}
		if !strings.Contains(err.Error(), "revoke refresh token") {
			t.Fatalf("error = %q, want to contain %q", err.Error(), "revoke refresh token")
		}
		if !errors.Is(err, revokeErr) {
			t.Fatal("error should wrap revoke error")
		}
	})
}

type stubCredentialAuthenticator struct {
	result   userdomain.User
	err      error
	email    string
	password string
}

func (s *stubCredentialAuthenticator) Authenticate(
	_ context.Context,
	email string,
	password string,
) (userdomain.User, error) {
	s.email = email
	s.password = password
	if s.err != nil {
		return userdomain.User{}, s.err
	}
	return s.result, nil
}

type stubTokenIssuer struct {
	token      string
	expiresAt  time.Time
	err        error
	userID     uuid.UUID
	clientID   string
	issueCalls int
}

func (s *stubTokenIssuer) Issue(
	userID uuid.UUID,
	clientID string,
) (string, time.Time, error) {
	s.userID = userID
	s.clientID = clientID
	s.issueCalls++
	if s.err != nil {
		return "", time.Time{}, s.err
	}
	return s.token, s.expiresAt, nil
}

type stubRefreshTokens struct {
	issueResult        RefreshTokenResult
	issueErr           error
	issueUserID        uuid.UUID
	issueClientID      string
	issueCalls         int
	authenticateResult RefreshIdentity
	authenticateErr    error
	authenticateToken  string
	authenticateCalls  int
	revokeErr          error
	revokeToken        string
	revokeCalls        int
}

func (s *stubRefreshTokens) Issue(
	_ context.Context,
	userID uuid.UUID,
	clientID string,
) (RefreshTokenResult, error) {
	s.issueCalls++
	s.issueUserID = userID
	s.issueClientID = clientID
	if s.issueErr != nil {
		return RefreshTokenResult{}, s.issueErr
	}
	if s.issueResult.Token == "" && !s.issueResult.ExpiresAt.IsZero() {
		return s.issueResult, nil
	}
	if s.issueResult.Token == "" {
		return RefreshTokenResult{}, nil
	}
	return s.issueResult, nil
}

func (s *stubRefreshTokens) Authenticate(
	_ context.Context,
	token string,
) (RefreshIdentity, error) {
	s.authenticateCalls++
	s.authenticateToken = token
	if s.authenticateErr != nil {
		return RefreshIdentity{}, s.authenticateErr
	}
	return s.authenticateResult, nil
}

func (s *stubRefreshTokens) Revoke(
	_ context.Context,
	token string,
) error {
	s.revokeCalls++
	s.revokeToken = token
	return s.revokeErr
}
