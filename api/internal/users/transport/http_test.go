package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	httphandler "github.com/rudsonalves/foundry-stack/api/internal/shared/http/handler"
	"github.com/rudsonalves/foundry-stack/api/internal/users/application"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

func TestHandlerRegister(t *testing.T) {
	t.Run("responds 201 with data envelope", func(t *testing.T) {
		handler := NewHandler(application.NewService(
			&stubUserRepository{
				createResult: domain.User{
					ID:           uuid.MustParse("9764ef1d-ae2e-4ba2-a92a-d81a7d4894aa"),
					Name:         "Ada Lovelace",
					Email:        "ada@example.com",
					PasswordHash: "hashed-password",
					CreatedAt:    time.Date(2026, time.July, 20, 10, 0, 0, 0, time.UTC),
					UpdatedAt:    time.Date(2026, time.July, 20, 10, 0, 0, 0, time.UTC),
				},
			},
			&stubPasswordHasher{hash: "hashed-password"},
			stubEmailVerificationHasher{},
			stubClock{},
		))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"name":"  Ada Lovelace  ","email":"  ADA@EXAMPLE.COM ","password":"secure123","email_verification_token":"verification-token"}`))

		httphandler.Handle(handler.Register).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusCreated {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
		}
		if got := recorder.Header().Get("Content-Type"); got != "application/json" {
			t.Fatalf("Content-Type = %q, want application/json", got)
		}

		var body map[string]map[string]any
		if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}

		data, ok := body["data"]
		if !ok {
			t.Fatal("response does not contain data envelope")
		}
		if len(data) != 3 {
			t.Fatalf("data = %#v, want only id, name, and email", data)
		}
		if data["id"] != "9764ef1d-ae2e-4ba2-a92a-d81a7d4894aa" {
			t.Fatalf("data.id = %v", data["id"])
		}
		if data["name"] != "Ada Lovelace" {
			t.Fatalf("data.name = %v", data["name"])
		}
		if data["email"] != "ada@example.com" {
			t.Fatalf("data.email = %v", data["email"])
		}
		if _, ok := data["password"]; ok {
			t.Fatal("response should not expose password")
		}
		if _, ok := data["password_hash"]; ok {
			t.Fatal("response should not expose password_hash")
		}
	})

	t.Run("converts email already registered to 409", func(t *testing.T) {
		handler := NewHandler(application.NewService(
			&stubUserRepository{createErr: domain.ErrEmailAlreadyExists},
			&stubPasswordHasher{hash: "hashed-password"},
			stubEmailVerificationHasher{},
			stubClock{},
		))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"name":"Ada","email":"ada@example.com","password":"secure123","email_verification_token":"verification-token"}`))

		httphandler.Handle(handler.Register).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusConflict {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusConflict)
		}

		var body struct {
			Error struct {
				Code string `json:"code"`
			} `json:"error"`
		}
		if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
			t.Fatalf("decode response: %v", err)
		}
		if body.Error.Code != sharederrors.ErrCodeEmailAlreadyRegistered {
			t.Fatalf("error.code = %q, want %q", body.Error.Code, sharederrors.ErrCodeEmailAlreadyRegistered)
		}
	})

	t.Run("converts rejected proof to invalid email verification", func(t *testing.T) {
		handler := NewHandler(application.NewService(
			&stubUserRepository{createErr: domain.ErrInvalidEmailVerification},
			&stubPasswordHasher{hash: "hashed-password"},
			stubEmailVerificationHasher{},
			stubClock{},
		))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"name":"Ada","email":"ada@example.com","password":"secure123","email_verification_token":"invalid-token"}`))

		httphandler.Handle(handler.Register).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
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
		if body.Error.Code != sharederrors.ErrCodeInvalidEmailVerification {
			t.Fatalf("error.code = %q, want %q", body.Error.Code, sharederrors.ErrCodeInvalidEmailVerification)
		}
		if body.Error.Message != "invalid email verification" {
			t.Fatalf("error.message = %q", body.Error.Message)
		}
	})

	t.Run("converts absent proof to invalid email verification", func(t *testing.T) {
		repo := &stubUserRepository{}
		handler := NewHandler(application.NewService(
			repo,
			&stubPasswordHasher{hash: "hashed-password"},
			stubEmailVerificationHasher{},
			stubClock{},
		))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"name":"Ada","email":"ada@example.com","password":"secure123"}`))

		httphandler.Handle(handler.Register).ServeHTTP(recorder, request)

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
		if body.Error.Code != sharederrors.ErrCodeInvalidEmailVerification {
			t.Fatalf("error.code = %q, want %q", body.Error.Code, sharederrors.ErrCodeInvalidEmailVerification)
		}
		if repo.createCalled {
			t.Fatal("repository create should not be called without a verification token")
		}
	})

	t.Run("converts unexpected failure to 500", func(t *testing.T) {
		repoErr := errors.New("database unavailable")
		handler := NewHandler(application.NewService(
			&stubUserRepository{createErr: repoErr},
			&stubPasswordHasher{hash: "hashed-password"},
			stubEmailVerificationHasher{},
			stubClock{},
		))

		recorder := httptest.NewRecorder()
		request := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBufferString(`{"name":"Ada","email":"ada@example.com","password":"secure123","email_verification_token":"verification-token"}`))
		request = request.WithContext(context.Background())

		httphandler.Handle(handler.Register).ServeHTTP(recorder, request)

		if recorder.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
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
		if body.Error.Code != sharederrors.ErrCodeInternalError {
			t.Fatalf("error.code = %q, want %q", body.Error.Code, sharederrors.ErrCodeInternalError)
		}
		if body.Error.Message != "Erro interno do servidor." {
			t.Fatalf("error.message = %q", body.Error.Message)
		}
	})
}

type stubEmailVerificationHasher struct{}

func (stubEmailVerificationHasher) HashCode(uuid.UUID, string, string) []byte {
	return nil
}

func (stubEmailVerificationHasher) HashToken(email string, token string) []byte {
	return []byte(email + ":" + token)
}

func (stubEmailVerificationHasher) MatchesCode([]byte, uuid.UUID, string, string) bool {
	return false
}

func (stubEmailVerificationHasher) MatchesToken([]byte, string, string) bool {
	return false
}

type stubClock struct{}

func (stubClock) Now() time.Time {
	return time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
}

type stubUserRepository struct {
	createInput       domain.CreateUserDTO
	createCalled      bool
	createResult      domain.User
	createErr         error
	findByEmailInput  string
	findByEmailResult domain.User
	findByEmailErr    error
}

func (s *stubUserRepository) Create(_ context.Context, input domain.CreateUserDTO) (domain.User, error) {
	s.createCalled = true
	s.createInput = input
	if s.createErr != nil {
		return domain.User{}, s.createErr
	}
	return s.createResult, nil
}

func (s *stubUserRepository) FindByID(_ context.Context, _ uuid.UUID) (domain.User, error) {
	return domain.User{}, errors.New("not implemented")
}

func (s *stubUserRepository) FindByEmail(_ context.Context, email string) (domain.User, error) {
	s.findByEmailInput = email
	if s.findByEmailErr != nil {
		return domain.User{}, s.findByEmailErr
	}
	return s.findByEmailResult, nil
}

type stubPasswordHasher struct {
	hash            string
	compareErr      error
	hashInput       string
	compareHash     string
	comparePassword string
	compareCalled   bool
}

func (s *stubPasswordHasher) Hash(password string) (string, error) {
	s.hashInput = password
	return s.hash, nil
}

func (s *stubPasswordHasher) Compare(hash string, password string) error {
	s.compareCalled = true
	s.compareHash = hash
	s.comparePassword = password
	return s.compareErr
}
