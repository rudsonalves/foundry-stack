package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	authapp "github.com/rudsonalves/foundry-stack/api/internal/auth/application"
	authtransport "github.com/rudsonalves/foundry-stack/api/internal/auth/transport"
	"github.com/rudsonalves/foundry-stack/api/internal/bootstrap"
	userapp "github.com/rudsonalves/foundry-stack/api/internal/users/application"
	userdomain "github.com/rudsonalves/foundry-stack/api/internal/users/domain"
	usertransport "github.com/rudsonalves/foundry-stack/api/internal/users/transport"
)

func TestAuthLoginRouteIsPublicAndReturnsExpectedEnvelope(t *testing.T) {
	authenticator := &routeAuthAuthenticator{
		result: userdomain.User{ID: uuid.MustParse("b7a9c62d-eeef-4f87-bce0-3f4ea7458e8f")},
	}
	issuer := &routeTokenIssuer{
		token:     "route.jwt.token",
		expiresAt: time.Now().Add(15 * time.Minute),
	}
	authHandler := authtransport.NewHandler(authapp.NewAuthService(
		authenticator,
		issuer,
		&routeRefreshTokens{issueResult: authapp.RefreshTokenResult{Token: "route.refresh.token"}},
	),
	)

	userHandler := usertransport.NewHandler(userapp.NewService(
		&routeUserRepo{},
		&routePasswordHasher{},
		nil,
		nil,
	))

	handler := newMiddleware(
		bootstrap.Config{AllowedOrigin: "http://localhost:5173"},
		userHandler,
		newRouteEmailVerificationHandler(&routeEmailVerificationService{}),
		authHandler,
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		bytes.NewBufferString(`{"email":"ada@example.com","password":"secure123","client_id":"foundry-stack-swagger"}`),
	)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("WWW-Authenticate"); got != "" {
		t.Fatalf("WWW-Authenticate = %q, want empty for public route", got)
	}
	if authenticator.calls != 1 {
		t.Fatalf("Authenticate call count = %d, want 1", authenticator.calls)
	}
	if authenticator.email != "ada@example.com" {
		t.Fatalf("Authenticate email = %q, want %q", authenticator.email, "ada@example.com")
	}
	if authenticator.password != "secure123" {
		t.Fatalf("Authenticate password = %q, want %q", authenticator.password, "secure123")
	}
	if issuer.clientID != "foundry-stack-swagger" {
		t.Fatalf("Issue clientID = %q, want %q", issuer.clientID, "foundry-stack-swagger")
	}

	var body map[string]map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, ok := body["data"]
	if !ok {
		t.Fatal("response should contain data envelope")
	}
	if data["access_token"] != "route.jwt.token" {
		t.Fatalf("data.access_token = %v, want %q", data["access_token"], "route.jwt.token")
	}
	if data["refresh_token"] != "route.refresh.token" {
		t.Fatalf("data.refresh_token = %v, want %q", data["refresh_token"], "route.refresh.token")
	}
	if data["token_type"] != "Bearer" {
		t.Fatalf("data.token_type = %v, want %q", data["token_type"], "Bearer")
	}
}

func TestAuthRefreshRouteIsPublicAndReturnsExpectedEnvelope(t *testing.T) {
	issuer := &routeTokenIssuer{
		token:     "refreshed.jwt.token",
		expiresAt: time.Now().Add(15 * time.Minute),
	}
	refreshTokens := &routeRefreshTokens{
		authenticateResult: authapp.RefreshIdentity{
			UserID:   uuid.MustParse("137c1c13-d81e-4dc5-8118-a41b4a99d253"),
			ClientID: "foundry-stack-swagger",
		},
	}
	authHandler := authtransport.NewHandler(authapp.NewAuthService(
		&routeAuthAuthenticator{},
		issuer,
		refreshTokens,
	))
	userHandler := usertransport.NewHandler(userapp.NewService(
		&routeUserRepo{},
		&routePasswordHasher{},
		nil,
		nil,
	))
	handler := newMiddleware(
		bootstrap.Config{AllowedOrigin: "http://localhost:5173"},
		userHandler,
		newRouteEmailVerificationHandler(&routeEmailVerificationService{}),
		authHandler,
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/refresh",
		bytes.NewBufferString(`{"refresh_token":"route.refresh.token"}`),
	)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("WWW-Authenticate"); got != "" {
		t.Fatalf("WWW-Authenticate = %q, want empty for public route", got)
	}
	if refreshTokens.authenticateCalls != 1 {
		t.Fatalf("Authenticate() call count = %d, want 1", refreshTokens.authenticateCalls)
	}
	if refreshTokens.authenticateToken != "route.refresh.token" {
		t.Fatalf("Authenticate() token = %q, want %q", refreshTokens.authenticateToken, "route.refresh.token")
	}
	if issuer.userID != refreshTokens.authenticateResult.UserID {
		t.Fatalf("Issue() userID = %s, want %s", issuer.userID, refreshTokens.authenticateResult.UserID)
	}
	if issuer.clientID != "foundry-stack-swagger" {
		t.Fatalf("Issue() clientID = %q, want %q", issuer.clientID, "foundry-stack-swagger")
	}

	var body map[string]map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data := body["data"]
	if data["access_token"] != "refreshed.jwt.token" {
		t.Fatalf("data.access_token = %v, want %q", data["access_token"], "refreshed.jwt.token")
	}
	if data["token_type"] != "Bearer" {
		t.Fatalf("data.token_type = %v, want %q", data["token_type"], "Bearer")
	}
}

func TestAuthLogoutRouteIsPublicAndReturnsNoContent(t *testing.T) {
	refreshTokens := &routeRefreshTokens{}
	authHandler := authtransport.NewHandler(authapp.NewAuthService(
		&routeAuthAuthenticator{},
		&routeTokenIssuer{},
		refreshTokens,
	))
	userHandler := usertransport.NewHandler(userapp.NewService(
		&routeUserRepo{},
		&routePasswordHasher{},
		nil,
		nil,
	))
	handler := newMiddleware(
		bootstrap.Config{AllowedOrigin: "http://localhost:5173"},
		userHandler,
		newRouteEmailVerificationHandler(&routeEmailVerificationService{}),
		authHandler,
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/logout",
		bytes.NewBufferString(`{"refresh_token":"route.refresh.token"}`),
	)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
	if got := recorder.Header().Get("WWW-Authenticate"); got != "" {
		t.Fatalf("WWW-Authenticate = %q, want empty for public route", got)
	}
	if refreshTokens.revokeCalls != 1 {
		t.Fatalf("Revoke() call count = %d, want 1", refreshTokens.revokeCalls)
	}
	if refreshTokens.revokeToken != "route.refresh.token" {
		t.Fatalf("Revoke() token = %q, want %q", refreshTokens.revokeToken, "route.refresh.token")
	}
	if recorder.Body.Len() != 0 {
		t.Fatalf("response body length = %d, want 0", recorder.Body.Len())
	}
}

func TestUserRegistrationRouteIsPublicWithVerificationToken(t *testing.T) {
	repository := &routeUserRepo{}
	userHandler := usertransport.NewHandler(userapp.NewService(
		repository,
		&routePasswordHasher{},
		routeEmailVerificationHasher{},
		routeClock{},
	))
	handler := newMiddleware(
		bootstrap.Config{AllowedOrigin: "http://localhost:5173"},
		userHandler,
		newRouteEmailVerificationHandler(&routeEmailVerificationService{}),
		newRouteAuthHandler(),
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/users",
		bytes.NewBufferString(
			`{"name":"Ada Lovelace","email":"ada@example.com","password":"secure123","email_verification_token":"verification-token"}`,
		),
	)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if got := recorder.Header().Get("WWW-Authenticate"); got != "" {
		t.Fatalf("WWW-Authenticate = %q, want empty for public route", got)
	}
	if string(repository.createInput.VerificationTokenHash) != "ada@example.com:verification-token" {
		t.Fatalf(
			"verification token hash = %q",
			repository.createInput.VerificationTokenHash,
		)
	}

	var body map[string]map[string]any
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data := body["data"]
	if len(data) != 3 {
		t.Fatalf("data = %#v, want only id, name, and email", data)
	}
	if _, exists := data["access_token"]; exists {
		t.Fatal("registration response must not contain access_token")
	}
	if _, exists := data["refresh_token"]; exists {
		t.Fatal("registration response must not contain refresh_token")
	}
}

func TestRequestEmailVerificationRouteIsPublic(t *testing.T) {
	service := &routeEmailVerificationService{
		requestResult: userapp.RequestEmailVerificationResult{
			VerificationID: uuid.MustParse(
				"9764ef1d-ae2e-4ba2-a92a-d81a7d4894aa",
			),
			CodeExpiresAt: time.Date(
				2026, time.September, 11, 12, 10, 0, 0, time.UTC,
			),
			ResendAvailableAt: time.Date(
				2026, time.September, 11, 12, 1, 0, 0, time.UTC,
			),
		},
	}
	handler := newMiddleware(
		bootstrap.Config{AllowedOrigin: "http://localhost:5173"},
		newRouteUserHandler(),
		newRouteEmailVerificationHandler(service),
		newRouteAuthHandler(),
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/email-verifications",
		bytes.NewBufferString(`{"email":"ada@example.com"}`),
	)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusAccepted)
	}
	if got := recorder.Header().Get("WWW-Authenticate"); got != "" {
		t.Fatalf("WWW-Authenticate = %q, want empty for public route", got)
	}
	if service.requestCalls != 1 {
		t.Fatalf("Request() calls = %d, want 1", service.requestCalls)
	}
}

func TestConfirmEmailVerificationRouteIsPublic(t *testing.T) {
	service := &routeEmailVerificationService{
		confirmResult: userapp.ConfirmEmailVerificationResult{
			EmailVerificationToken: "verification-token",
			ExpiresAt: time.Date(
				2026, time.September, 12, 12, 0, 0, 0, time.UTC,
			),
		},
	}
	handler := newMiddleware(
		bootstrap.Config{AllowedOrigin: "http://localhost:5173"},
		newRouteUserHandler(),
		newRouteEmailVerificationHandler(service),
		newRouteAuthHandler(),
	)

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/email-verifications/confirm",
		bytes.NewBufferString(
			`{"verification_id":"9764ef1d-ae2e-4ba2-a92a-d81a7d4894aa","code":"012345"}`,
		),
	)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if got := recorder.Header().Get("WWW-Authenticate"); got != "" {
		t.Fatalf("WWW-Authenticate = %q, want empty for public route", got)
	}
	if service.confirmCalls != 1 {
		t.Fatalf("Confirm() calls = %d, want 1", service.confirmCalls)
	}
}

type routeAuthAuthenticator struct {
	result   userdomain.User
	err      error
	email    string
	password string
	calls    int
}

func (s *routeAuthAuthenticator) Authenticate(
	_ context.Context,
	email string,
	password string,
) (userdomain.User, error) {
	s.calls++
	s.email = email
	s.password = password
	if s.err != nil {
		return userdomain.User{}, s.err
	}
	return s.result, nil
}

type routeTokenIssuer struct {
	token     string
	expiresAt time.Time
	err       error
	userID    uuid.UUID
	clientID  string
}

func (s *routeTokenIssuer) Issue(
	userID uuid.UUID,
	clientID string,
) (string, time.Time, error) {
	s.userID = userID
	s.clientID = clientID
	if s.err != nil {
		return "", time.Time{}, s.err
	}
	return s.token, s.expiresAt, nil
}

type routeUserRepo struct {
	createInput userdomain.CreateUserDTO
}

func (s *routeUserRepo) Create(
	_ context.Context,
	input userdomain.CreateUserDTO,
) (userdomain.User, error) {
	s.createInput = input
	return userdomain.User{
		ID:    uuid.MustParse("9764ef1d-ae2e-4ba2-a92a-d81a7d4894aa"),
		Name:  input.Name,
		Email: input.Email,
	}, nil
}

func (s *routeUserRepo) FindByID(
	_ context.Context,
	_ uuid.UUID,
) (userdomain.User, error) {
	return userdomain.User{}, userdomain.ErrUserNotFound
}

func (s *routeUserRepo) FindByEmail(
	_ context.Context,
	_ string,
) (userdomain.User, error) {
	return userdomain.User{}, userdomain.ErrUserNotFound
}

type routePasswordHasher struct{}

func (s *routePasswordHasher) Hash(password string) (string, error) {
	return password, nil
}

func (s *routePasswordHasher) Compare(_, _ string) error {
	return nil
}

type routeEmailVerificationHasher struct{}

func (routeEmailVerificationHasher) HashCode(
	uuid.UUID,
	string,
	string,
) []byte {
	return nil
}

func (routeEmailVerificationHasher) HashToken(
	email string,
	token string,
) []byte {
	return []byte(email + ":" + token)
}

func (routeEmailVerificationHasher) MatchesCode(
	[]byte,
	uuid.UUID,
	string,
	string,
) bool {
	return false
}

func (routeEmailVerificationHasher) MatchesToken(
	[]byte,
	string,
	string,
) bool {
	return false
}

type routeClock struct{}

func (routeClock) Now() time.Time {
	return time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
}

type routeRefreshTokens struct {
	issueResult        authapp.RefreshTokenResult
	issueErr           error
	issueCalls         int
	authenticateResult authapp.RefreshIdentity
	authenticateErr    error
	authenticateToken  string
	authenticateCalls  int
	revokeErr          error
	revokeToken        string
	revokeCalls        int
}

func (s *routeRefreshTokens) Issue(
	_ context.Context,
	_ uuid.UUID,
	_ string,
) (authapp.RefreshTokenResult, error) {
	s.issueCalls++
	if s.issueErr != nil {
		return authapp.RefreshTokenResult{}, s.issueErr
	}
	if s.issueResult.Token == "" {
		return authapp.RefreshTokenResult{Token: "route.refresh.token"}, nil
	}
	return s.issueResult, nil
}

func (s *routeRefreshTokens) Authenticate(
	_ context.Context,
	token string,
) (authapp.RefreshIdentity, error) {
	s.authenticateCalls++
	s.authenticateToken = token
	if s.authenticateErr != nil {
		return authapp.RefreshIdentity{}, s.authenticateErr
	}
	return s.authenticateResult, nil
}

func (s *routeRefreshTokens) Revoke(
	_ context.Context,
	token string,
) error {
	s.revokeCalls++
	s.revokeToken = token
	return s.revokeErr
}

type routeEmailVerificationService struct {
	requestCalls  int
	confirmCalls  int
	requestResult userapp.RequestEmailVerificationResult
	confirmResult userapp.ConfirmEmailVerificationResult
}

func (s *routeEmailVerificationService) Request(
	_ context.Context,
	_ userapp.RequestEmailVerificationInput,
) (userapp.RequestEmailVerificationResult, error) {
	s.requestCalls++
	return s.requestResult, nil
}

func (s *routeEmailVerificationService) Confirm(
	_ context.Context,
	_ userapp.ConfirmEmailVerificationInput,
) (userapp.ConfirmEmailVerificationResult, error) {
	s.confirmCalls++
	return s.confirmResult, nil
}

func newRouteEmailVerificationHandler(
	service *routeEmailVerificationService,
) *usertransport.EmailVerificationHandler {
	return usertransport.NewEmailVerificationHandler(service)
}

func newRouteUserHandler() *usertransport.Handler {
	return usertransport.NewHandler(userapp.NewService(
		&routeUserRepo{},
		&routePasswordHasher{},
		routeEmailVerificationHasher{},
		routeClock{},
	))
}

func newRouteAuthHandler() *authtransport.Handler {
	return authtransport.NewHandler(authapp.NewAuthService(
		&routeAuthAuthenticator{},
		&routeTokenIssuer{},
		&routeRefreshTokens{},
	))
}
