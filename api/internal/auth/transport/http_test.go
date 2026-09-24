package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rudsonalves/foundry-stack/api/internal/auth/application"
	authdomain "github.com/rudsonalves/foundry-stack/api/internal/auth/domain"
	authinfra "github.com/rudsonalves/foundry-stack/api/internal/auth/infrastructure"
	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	httphandler "github.com/rudsonalves/foundry-stack/api/internal/shared/http/handler"
	usersapp "github.com/rudsonalves/foundry-stack/api/internal/users/application"
	userdomain "github.com/rudsonalves/foundry-stack/api/internal/users/domain"
	usersinfra "github.com/rudsonalves/foundry-stack/api/internal/users/infrastructure"
	"golang.org/x/crypto/bcrypt"
)

func TestHandlerLogin(t *testing.T) {
	t.Run("invalid JSON returns BAD_REQUEST", func(t *testing.T) {
		h := newLoginTestHandler(
			&stubAuthUserAuthenticator{},
			&stubAuthTokenIssuer{},
		)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/auth/login",
			bytes.NewBufferString(`{"email":`),
		)

		httphandler.Handle(h.Login).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
		}

		var body struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body.Error.Code != sharederrors.ErrCodeBadRequest {
			t.Fatalf("error.code = %q, want %q", body.Error.Code, sharederrors.ErrCodeBadRequest)
		}
	})

	t.Run("valid login forwards credentials and returns data envelope with bearer token", func(t *testing.T) {
		ttl := 15 * time.Minute
		authenticator := &stubAuthUserAuthenticator{
			result: userdomain.User{ID: uuid.MustParse("6e0732bc-e94d-44ad-96a1-5816856122f4")},
		}
		issuer := &stubAuthTokenIssuer{
			token:     "access.jwt.token",
			expiresAt: time.Now().Add(ttl),
		}
		h := newLoginTestHandler(authenticator, issuer)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/auth/login",
			bytes.NewBufferString(`{"email":"ada@example.com","password":"secure123","client_id":"foundry-stack-mobile"}`),
		)

		httphandler.Handle(h.Login).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
		}
		if authenticator.email != "ada@example.com" {
			t.Fatalf("Authenticate email = %q, want %q", authenticator.email, "ada@example.com")
		}
		if authenticator.password != "secure123" {
			t.Fatalf("Authenticate password = %q, want %q", authenticator.password, "secure123")
		}
		if issuer.clientID != "foundry-stack-mobile" {
			t.Fatalf("Issue clientID = %q, want %q", issuer.clientID, "foundry-stack-mobile")
		}

		var body map[string]map[string]any
		if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		data, ok := body["data"]
		if !ok {
			t.Fatal("response should contain data envelope")
		}
		if data["access_token"] != "access.jwt.token" {
			t.Fatalf("data.access_token = %v, want %q", data["access_token"], "access.jwt.token")
		}
		if data["refresh_token"] != "refresh-token" {
			t.Fatalf("data.refresh_token = %v, want %q", data["refresh_token"], "refresh-token")
		}
		if data["token_type"] != "Bearer" {
			t.Fatalf("data.token_type = %v, want %q", data["token_type"], "Bearer")
		}

		expiresIn, ok := data["expires_in"].(float64)
		if !ok {
			t.Fatalf("data.expires_in type = %T, want number", data["expires_in"])
		}
		if expiresIn <= 0 {
			t.Fatalf("data.expires_in = %v, want > 0", expiresIn)
		}
		if expiresIn > ttl.Seconds() {
			t.Fatalf("data.expires_in = %v, want <= %v", expiresIn, ttl.Seconds())
		}

		if _, ok := data["password"]; ok {
			t.Fatal("response should not contain password")
		}
		if _, ok := data["password_hash"]; ok {
			t.Fatal("response should not contain password_hash")
		}
	})

	t.Run("unexpected service error propagates as internal error", func(t *testing.T) {
		authenticator := &stubAuthUserAuthenticator{err: errors.New("database unavailable")}
		h := newLoginTestHandler(authenticator, &stubAuthTokenIssuer{})

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/auth/login",
			bytes.NewBufferString(`{"email":"ada@example.com","password":"secure123","client_id":"foundry-stack-mobile"}`),
		)

		httphandler.Handle(h.Login).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
		}
	})

	t.Run("expired result clamps expires_in to zero", func(t *testing.T) {
		authenticator := &stubAuthUserAuthenticator{
			result: userdomain.User{ID: uuid.New()},
		}
		issuer := &stubAuthTokenIssuer{
			token:     "expired.jwt.token",
			expiresAt: time.Now().Add(-time.Minute),
		}
		h := newLoginTestHandler(authenticator, issuer)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/auth/login",
			bytes.NewBufferString(`{"email":"ada@example.com","password":"secure123","client_id":"foundry-stack-mobile"}`),
		)

		httphandler.Handle(h.Login).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
		}

		var body struct {
			Data LoginResponse `json:"data"`
		}
		if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body.Data.ExpiresIn != 0 {
			t.Fatalf("data.expires_in = %d, want 0", body.Data.ExpiresIn)
		}
	})

	t.Run("invalid credentials become 401 with UNAUTHORIZED code and bearer challenge", func(t *testing.T) {
		authenticator := &stubAuthUserAuthenticator{err: usersapp.ErrInvalidCredentials}
		h := newLoginTestHandler(authenticator, &stubAuthTokenIssuer{})

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/auth/login",
			bytes.NewBufferString(`{"email":"ada@example.com","password":"wrong","client_id":"foundry-stack-mobile"}`),
		)

		httphandler.Handle(h.Login).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
		}
		if got := recorder.Header().Get("WWW-Authenticate"); got != "Bearer" {
			t.Fatalf("WWW-Authenticate = %q, want %q", got, "Bearer")
		}

		var body struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body.Error.Code != sharederrors.ErrCodeUnauthorized {
			t.Fatalf("error.code = %q, want %q", body.Error.Code, sharederrors.ErrCodeUnauthorized)
		}
		if body.Error.Message != "invalid email or password" {
			t.Fatalf(
				"error.message = %q, want %q",
				body.Error.Message,
				"invalid email or password",
			)
		}
	})

	t.Run("client not allowed becomes 401 without exposing allowed clients", func(t *testing.T) {
		authenticator := &stubAuthUserAuthenticator{
			result: userdomain.User{ID: uuid.New()},
		}
		issuer := &stubAuthTokenIssuer{err: authdomain.ErrClientNotAllowed}
		h := newLoginTestHandler(authenticator, issuer)

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/auth/login",
			bytes.NewBufferString(`{"email":"ada@example.com","password":"secure123","client_id":"unknown-client"}`),
		)

		httphandler.Handle(h.Login).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
		}

		var body struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body.Error.Code != sharederrors.ErrCodeUnauthorized {
			t.Fatalf("error.code = %q, want %q", body.Error.Code, sharederrors.ErrCodeUnauthorized)
		}
		if body.Error.Message != "client not allowed" {
			t.Fatalf("error.message = %q, want %q", body.Error.Message, "client not allowed")
		}
		if strings.Contains(recorder.Body.String(), "foundry-stack-mobile") ||
			strings.Contains(recorder.Body.String(), "foundry-stack-swagger") {
			t.Fatal("response should not expose allowed client IDs")
		}
	})
}

func TestHandlerRefresh(t *testing.T) {
	t.Run("invalid JSON returns BAD_REQUEST", func(t *testing.T) {
		h := NewHandler(application.NewAuthService(
			&stubAuthUserAuthenticator{},
			&stubAuthTokenIssuer{},
			&stubAuthRefreshTokens{},
		))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/auth/refresh",
			bytes.NewBufferString(`{"refresh_token":`),
		)

		httphandler.Handle(h.Refresh).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("valid refresh returns data envelope with bearer token", func(t *testing.T) {
		userID := uuid.MustParse("8e4b2f7a-3bc0-4f8c-8fab-58d3cba1375a")
		ttl := 10 * time.Minute
		issuer := &stubAuthTokenIssuer{token: "new.access.token", expiresAt: time.Now().Add(ttl)}
		refreshTokens := &stubAuthRefreshTokens{
			authenticateResult: application.RefreshIdentity{UserID: userID, ClientID: "foundry-stack-mobile"},
		}
		h := NewHandler(application.NewAuthService(
			&stubAuthUserAuthenticator{},
			issuer,
			refreshTokens,
		))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/auth/refresh",
			bytes.NewBufferString(`{"refresh_token":"refresh-token"}`),
		)

		httphandler.Handle(h.Refresh).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
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

		var body map[string]map[string]any
		if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		data := body["data"]
		if data["access_token"] != "new.access.token" {
			t.Fatalf("data.access_token = %v, want %q", data["access_token"], "new.access.token")
		}
		if data["token_type"] != "Bearer" {
			t.Fatalf("data.token_type = %v, want %q", data["token_type"], "Bearer")
		}
		expiresIn, ok := data["expires_in"].(float64)
		if !ok {
			t.Fatalf("data.expires_in type = %T, want number", data["expires_in"])
		}
		if expiresIn <= 0 {
			t.Fatalf("data.expires_in = %v, want > 0", expiresIn)
		}
	})

	t.Run("invalid refresh token returns unauthorized", func(t *testing.T) {
		h := NewHandler(application.NewAuthService(
			&stubAuthUserAuthenticator{},
			&stubAuthTokenIssuer{},
			&stubAuthRefreshTokens{authenticateErr: authdomain.ErrInvalidRefreshToken},
		))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/auth/refresh",
			bytes.NewBufferString(`{"refresh_token":"refresh-token"}`),
		)

		httphandler.Handle(h.Refresh).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusUnauthorized {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusUnauthorized)
		}
	})

	t.Run("unexpected service error returns internal server error", func(t *testing.T) {
		h := NewHandler(application.NewAuthService(
			&stubAuthUserAuthenticator{},
			&stubAuthTokenIssuer{},
			&stubAuthRefreshTokens{authenticateErr: errors.New("repository unavailable")},
		))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/auth/refresh",
			bytes.NewBufferString(`{"refresh_token":"refresh-token"}`),
		)

		httphandler.Handle(h.Refresh).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
		}
	})

	t.Run("expired result clamps expires_in to zero", func(t *testing.T) {
		h := NewHandler(application.NewAuthService(
			&stubAuthUserAuthenticator{},
			&stubAuthTokenIssuer{
				token:     "expired.access.token",
				expiresAt: time.Now().Add(-time.Minute),
			},
			&stubAuthRefreshTokens{
				authenticateResult: application.RefreshIdentity{UserID: uuid.New(), ClientID: "foundry-stack-mobile"},
			},
		))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/auth/refresh",
			bytes.NewBufferString(`{"refresh_token":"refresh-token"}`),
		)

		httphandler.Handle(h.Refresh).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
		}

		var body struct {
			Data RefreshResponse `json:"data"`
		}
		if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body.Data.ExpiresIn != 0 {
			t.Fatalf("data.expires_in = %d, want 0", body.Data.ExpiresIn)
		}
	})
}

func TestHandlerLogout(t *testing.T) {
	t.Run("valid logout returns no content", func(t *testing.T) {
		refreshTokens := &stubAuthRefreshTokens{}
		h := NewHandler(application.NewAuthService(
			&stubAuthUserAuthenticator{},
			&stubAuthTokenIssuer{},
			refreshTokens,
		))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/auth/logout",
			bytes.NewBufferString(`{"refresh_token":"refresh-token"}`),
		)

		httphandler.Handle(h.Logout).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNoContent)
		}
		if refreshTokens.revokeCalls != 1 {
			t.Fatalf("Revoke() call count = %d, want 1", refreshTokens.revokeCalls)
		}
		if refreshTokens.revokeToken != "refresh-token" {
			t.Fatalf("Revoke() token = %q, want %q", refreshTokens.revokeToken, "refresh-token")
		}
		if recorder.Body.Len() != 0 {
			t.Fatalf("response body length = %d, want 0", recorder.Body.Len())
		}
	})

	t.Run("invalid JSON returns BAD_REQUEST", func(t *testing.T) {
		h := NewHandler(application.NewAuthService(
			&stubAuthUserAuthenticator{},
			&stubAuthTokenIssuer{},
			&stubAuthRefreshTokens{},
		))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/auth/logout",
			bytes.NewBufferString(`{"refresh_token":`),
		)

		httphandler.Handle(h.Logout).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
		}
	})

	t.Run("revoke failure returns internal server error", func(t *testing.T) {
		h := NewHandler(application.NewAuthService(
			&stubAuthUserAuthenticator{},
			&stubAuthTokenIssuer{},
			&stubAuthRefreshTokens{revokeErr: errors.New("update failed")},
		))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(
			http.MethodPost,
			"/auth/logout",
			bytes.NewBufferString(`{"refresh_token":"refresh-token"}`),
		)

		httphandler.Handle(h.Logout).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
		}
	})
}

func TestHandlerLoginWithRealServices(t *testing.T) {
	userID := uuid.MustParse("fab76963-c677-4d93-a6da-06a86afbc783")
	hasher := usersinfra.NewBcryptHasher(bcrypt.MinCost)
	passwordHash, err := hasher.Hash("secure123")
	if err != nil {
		t.Fatalf("Hash() unexpected error: %v", err)
	}

	userService := usersapp.NewService(
		&loginUserRepository{user: userdomain.User{
			ID:           userID,
			Name:         "Ada Lovelace",
			Email:        "ada@example.com",
			PasswordHash: passwordHash,
		}},
		hasher,
		nil,
		nil,
	)
	tokenService, err := authinfra.NewTokenService(
		authinfra.TokenServiceConfig{
			Secret:    []byte("0123456789abcdef0123456789abcdef"),
			Issuer:    "foundry-stack-test",
			Audience:  "foundry-stack-api",
			ClientIDs: []string{"foundry-stack-mobile"},
			Lifetime:  15 * time.Minute,
		},
	)
	if err != nil {
		t.Fatalf("NewTokenService() unexpected error: %v", err)
	}

	handler := NewHandler(application.NewAuthService(
		userService,
		tokenService,
		&stubAuthRefreshTokens{issueResult: application.RefreshTokenResult{Token: "refresh-token"}},
	))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(
		http.MethodPost,
		"/auth/login",
		bytes.NewBufferString(`{"email":"ADA@EXAMPLE.COM","password":"secure123","client_id":"foundry-stack-mobile"}`),
	)

	httphandler.Handle(handler.Login).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var body struct {
		Data LoginResponse `json:"data"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	parsedUserID, err := tokenService.Parse(body.Data.AccessToken)
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if parsedUserID != userID {
		t.Fatalf("Parse() userID = %s, want %s", parsedUserID, userID)
	}
}

func newLoginTestHandler(
	authenticator *stubAuthUserAuthenticator,
	issuer *stubAuthTokenIssuer,
) *Handler {
	service := application.NewAuthService(
		authenticator,
		issuer,
		&stubAuthRefreshTokens{issueResult: application.RefreshTokenResult{Token: "refresh-token"}},
	)
	return NewHandler(service)
}

type stubAuthUserAuthenticator struct {
	result   userdomain.User
	err      error
	email    string
	password string
}

func (s *stubAuthUserAuthenticator) Authenticate(
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

type stubAuthTokenIssuer struct {
	token     string
	expiresAt time.Time
	err       error
	userID    uuid.UUID
	clientID  string
}

type loginUserRepository struct {
	user userdomain.User
}

func (r *loginUserRepository) Create(
	_ context.Context,
	_ userdomain.CreateUserDTO,
) (userdomain.User, error) {
	return userdomain.User{}, errors.New("not implemented")
}

func (r *loginUserRepository) FindByID(
	_ context.Context,
	_ uuid.UUID,
) (userdomain.User, error) {
	return userdomain.User{}, errors.New("not implemented")
}

func (r *loginUserRepository) FindByEmail(
	_ context.Context,
	_ string,
) (userdomain.User, error) {
	return r.user, nil
}

func (s *stubAuthTokenIssuer) Issue(
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

type stubAuthRefreshTokens struct {
	issueResult        application.RefreshTokenResult
	issueErr           error
	issueUserID        uuid.UUID
	issueClientID      string
	issueCalls         int
	authenticateResult application.RefreshIdentity
	authenticateErr    error
	authenticateToken  string
	authenticateCalls  int
	revokeErr          error
	revokeToken        string
	revokeCalls        int
}

func (s *stubAuthRefreshTokens) Issue(
	_ context.Context,
	userID uuid.UUID,
	clientID string,
) (application.RefreshTokenResult, error) {
	s.issueCalls++
	s.issueUserID = userID
	s.issueClientID = clientID
	if s.issueErr != nil {
		return application.RefreshTokenResult{}, s.issueErr
	}
	if s.issueResult.Token == "" {
		return application.RefreshTokenResult{Token: "refresh-token"}, nil
	}
	return s.issueResult, nil
}

func (s *stubAuthRefreshTokens) Authenticate(
	_ context.Context,
	token string,
) (application.RefreshIdentity, error) {
	s.authenticateCalls++
	s.authenticateToken = token
	if s.authenticateErr != nil {
		return application.RefreshIdentity{}, s.authenticateErr
	}
	return s.authenticateResult, nil
}

func (s *stubAuthRefreshTokens) Revoke(
	_ context.Context,
	token string,
) error {
	s.revokeCalls++
	s.revokeToken = token
	return s.revokeErr
}
