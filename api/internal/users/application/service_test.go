package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

func TestServiceRegister(t *testing.T) {
	t.Run("registers user with normalized input", func(t *testing.T) {
		repo := &stubUserRepository{}
		hasher := &stubPasswordHasher{hash: "hashed-password"}
		svc := newTestUserService(repo, hasher)

		created := domain.User{
			ID:           uuid.MustParse("9764ef1d-ae2e-4ba2-a92a-d81a7d4894aa"),
			Name:         "Ada Lovelace",
			Email:        "ada@example.com",
			PasswordHash: "hashed-password",
			CreatedAt:    time.Date(2026, time.July, 20, 10, 0, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2026, time.July, 20, 10, 0, 0, 0, time.UTC),
		}
		repo.createResult = created

		user, err := svc.Register(context.Background(), RegisterInput{
			Name:                   "  Ada Lovelace  ",
			Email:                  "  ADA@EXAMPLE.COM ",
			Password:               "secure123",
			EmailVerificationToken: "verification-token",
		})
		if err != nil {
			t.Fatalf("Register() error = %v", err)
		}
		if user != created {
			t.Fatalf("Register() user = %#v", user)
		}
		if hasher.hashInput != "secure123" {
			t.Fatalf("Hash() input = %q", hasher.hashInput)
		}
		if repo.createInput.Name != "Ada Lovelace" {
			t.Fatalf("Create() name = %q", repo.createInput.Name)
		}
		if repo.createInput.Email != "ada@example.com" {
			t.Fatalf("Create() email = %q", repo.createInput.Email)
		}
		if repo.createInput.PasswordHash != "hashed-password" {
			t.Fatalf("Create() password hash = %q", repo.createInput.PasswordHash)
		}
		if string(repo.createInput.VerificationTokenHash) != "ada@example.com:verification-token" {
			t.Fatalf("Create() verification token hash = %q", repo.createInput.VerificationTokenHash)
		}
		if !repo.createInput.Now.Equal(testUserServiceNow) {
			t.Fatalf("Create() now = %v, want %v", repo.createInput.Now, testUserServiceNow)
		}
	})

	t.Run("returns validation error for invalid input", func(t *testing.T) {
		repo := &stubUserRepository{}
		hasher := &stubPasswordHasher{}
		svc := newTestUserService(repo, hasher)

		_, err := svc.Register(context.Background(), RegisterInput{
			Name:                   " ",
			Email:                  "not-an-email",
			Password:               "short",
			EmailVerificationToken: "verification-token",
		})
		if err == nil {
			t.Fatal("expected validation error, got nil")
		}

		var appErr *sharederrors.AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("error type = %T, want *AppError", err)
		}
		if appErr.Code != sharederrors.ErrCodeBadRequest {
			t.Fatalf("validation code = %q", appErr.Code)
		}
		if len(appErr.Violations) != 3 {
			t.Fatalf("violations = %#v", appErr.Violations)
		}
		if hasher.hashInput != "" {
			t.Fatal("hash should not be called on invalid input")
		}
		if repo.createCalled {
			t.Fatal("repository create should not be called on invalid input")
		}
	})

	t.Run("returns hasher error", func(t *testing.T) {
		repo := &stubUserRepository{}
		hashErr := errors.New("hash failed")
		hasher := &stubPasswordHasher{err: hashErr}
		svc := newTestUserService(repo, hasher)

		_, err := svc.Register(context.Background(), RegisterInput{
			Name:                   "Ada Lovelace",
			Email:                  "ada@example.com",
			Password:               "secure123",
			EmailVerificationToken: "verification-token",
		})
		if !errors.Is(err, hashErr) {
			t.Fatalf("error %v does not wrap %v", err, hashErr)
		}
		if repo.createCalled {
			t.Fatal("repository create should not be called when hash fails")
		}
	})

	t.Run("maps duplicate email to email already registered", func(t *testing.T) {
		repo := &stubUserRepository{createErr: domain.ErrEmailAlreadyExists}
		hasher := &stubPasswordHasher{hash: "hashed-password"}
		svc := newTestUserService(repo, hasher)

		_, err := svc.Register(context.Background(), RegisterInput{
			Name:                   "Ada Lovelace",
			Email:                  "ada@example.com",
			Password:               "secure123",
			EmailVerificationToken: "verification-token",
		})

		var appErr *sharederrors.AppError
		if !errors.As(err, &appErr) {
			t.Fatalf("error type = %T, want *AppError", err)
		}
		if appErr.Code != sharederrors.ErrCodeEmailAlreadyRegistered {
			t.Fatalf("error code = %q, want %q", appErr.Code, sharederrors.ErrCodeEmailAlreadyRegistered)
		}
		if !errors.Is(err, domain.ErrEmailAlreadyExists) {
			t.Fatalf("error %v does not wrap %v", err, domain.ErrEmailAlreadyExists)
		}
	})

	t.Run("maps every rejected proof to invalid email verification", func(t *testing.T) {
		for _, name := range []string{
			"invalid",
			"expired",
			"consumed",
			"belongs to another email",
		} {
			t.Run(name, func(t *testing.T) {
				repo := &stubUserRepository{createErr: domain.ErrInvalidEmailVerification}
				hasher := &stubPasswordHasher{hash: "hashed-password"}
				svc := newTestUserService(repo, hasher)

				_, err := svc.Register(context.Background(), RegisterInput{
					Name:                   "Ada Lovelace",
					Email:                  "ada@example.com",
					Password:               "secure123",
					EmailVerificationToken: "verification-token",
				})

				assertRegistrationInvalidEmailVerification(t, err)
			})
		}
	})

	t.Run("maps absent proof to invalid email verification", func(t *testing.T) {
		repo := &stubUserRepository{}
		hasher := &stubPasswordHasher{}
		svc := newTestUserService(repo, hasher)

		_, err := svc.Register(context.Background(), RegisterInput{
			Name:     "Ada Lovelace",
			Email:    "ada@example.com",
			Password: "secure123",
		})

		assertRegistrationInvalidEmailVerification(t, err)
		if hasher.hashInput != "" {
			t.Fatal("password hash should not be called without a verification token")
		}
		if repo.createCalled {
			t.Fatal("repository create should not be called without a verification token")
		}
	})

	t.Run("returns repository error", func(t *testing.T) {
		repoErr := errors.New("database unavailable")
		repo := &stubUserRepository{createErr: repoErr}
		hasher := &stubPasswordHasher{hash: "hashed-password"}
		svc := newTestUserService(repo, hasher)

		_, err := svc.Register(context.Background(), RegisterInput{
			Name:                   "Ada Lovelace",
			Email:                  "ada@example.com",
			Password:               "secure123",
			EmailVerificationToken: "verification-token",
		})
		if !errors.Is(err, repoErr) {
			t.Fatalf("error %v does not wrap %v", err, repoErr)
		}
	})
}

func TestServiceAuthenticate(t *testing.T) {
	t.Run("authenticates correct credentials", func(t *testing.T) {
		storedUser := domain.User{
			ID:           uuid.MustParse("9764ef1d-ae2e-4ba2-a92a-d81a7d4894aa"),
			Name:         "Ada Lovelace",
			Email:        "ada@example.com",
			PasswordHash: "hashed-password",
			CreatedAt:    time.Date(2026, time.July, 20, 10, 0, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2026, time.July, 20, 10, 0, 0, 0, time.UTC),
		}
		repo := &stubUserRepository{findByEmailResult: storedUser}
		hasher := &stubPasswordHasher{}
		svc := newTestUserService(repo, hasher)

		user, err := svc.Authenticate(context.Background(), "  ADA@EXAMPLE.COM ", "secure123")
		if err != nil {
			t.Fatalf("Authenticate() error = %v", err)
		}
		if user != storedUser {
			t.Fatalf("Authenticate() user = %#v", user)
		}
		if repo.findByEmailInput != "ada@example.com" {
			t.Fatalf("FindByEmail() email = %q", repo.findByEmailInput)
		}
		if hasher.compareHash != "hashed-password" {
			t.Fatalf("Compare() hash = %q", hasher.compareHash)
		}
		if hasher.comparePassword != "secure123" {
			t.Fatalf("Compare() password = %q", hasher.comparePassword)
		}
	})

	t.Run("returns invalid credentials for wrong password", func(t *testing.T) {
		repo := &stubUserRepository{findByEmailResult: domain.User{PasswordHash: "hashed-password"}}
		hasher := &stubPasswordHasher{compareErr: errors.New("bcrypt: password mismatch")}
		svc := newTestUserService(repo, hasher)

		_, err := svc.Authenticate(context.Background(), "ada@example.com", "wrong-password")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("error %v does not wrap %v", err, ErrInvalidCredentials)
		}
	})

	t.Run("returns invalid credentials for missing user", func(t *testing.T) {
		repo := &stubUserRepository{findByEmailErr: domain.ErrUserNotFound}
		hasher := &stubPasswordHasher{}
		svc := newTestUserService(repo, hasher)

		_, err := svc.Authenticate(context.Background(), "ada@example.com", "secure123")
		if !errors.Is(err, ErrInvalidCredentials) {
			t.Fatalf("error %v does not wrap %v", err, ErrInvalidCredentials)
		}
		if hasher.compareCalled {
			t.Fatal("Compare() should not be called when user is missing")
		}
	})

	t.Run("preserves repository error", func(t *testing.T) {
		repoErr := errors.New("database unavailable")
		repo := &stubUserRepository{findByEmailErr: repoErr}
		hasher := &stubPasswordHasher{}
		svc := newTestUserService(repo, hasher)

		_, err := svc.Authenticate(context.Background(), "ada@example.com", "secure123")
		if !errors.Is(err, repoErr) {
			t.Fatalf("error %v does not wrap %v", err, repoErr)
		}
	})
}

func TestValidateRegistration(t *testing.T) {
	t.Run("accepts valid registration", func(t *testing.T) {
		violations := validateRegistration("Ada", "ada@example.com", "secure123")
		if len(violations) != 0 {
			t.Fatalf("violations = %#v", violations)
		}
	})

	t.Run("rejects password over 72 chars", func(t *testing.T) {
		password := "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		violations := validateRegistration("Ada", "ada@example.com", password)
		if len(violations) != 1 {
			t.Fatalf("violations = %#v", violations)
		}
		if violations[0].Field != "password" {
			t.Fatalf("field = %q", violations[0].Field)
		}
	})
}

var testUserServiceNow = time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)

func assertRegistrationInvalidEmailVerification(t *testing.T, err error) {
	t.Helper()

	var appErr *sharederrors.AppError
	if !errors.As(err, &appErr) {
		t.Fatalf("error type = %T, want *AppError", err)
	}
	if appErr.Code != sharederrors.ErrCodeInvalidEmailVerification {
		t.Fatalf("error code = %q, want %q", appErr.Code, sharederrors.ErrCodeInvalidEmailVerification)
	}
	if appErr.Message != "invalid email verification" {
		t.Fatalf("error message = %q", appErr.Message)
	}
	if !errors.Is(err, domain.ErrInvalidEmailVerification) {
		t.Fatalf("error %v does not wrap %v", err, domain.ErrInvalidEmailVerification)
	}
}

func newTestUserService(
	users domain.UserRepository,
	passwords domain.PasswordHasher,
) *Service {
	return NewService(
		users,
		passwords,
		requestHasherStub{},
		requestClockStub{now: testUserServiceNow},
	)
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
	hashInput       string
	hash            string
	err             error
	compareHash     string
	comparePassword string
	compareErr      error
	compareCalled   bool
}

func (s *stubPasswordHasher) Hash(password string) (string, error) {
	s.hashInput = password
	if s.err != nil {
		return "", s.err
	}
	return s.hash, nil
}

func (s *stubPasswordHasher) Compare(hash string, password string) error {
	s.compareCalled = true
	s.compareHash = hash
	s.comparePassword = password
	if s.compareErr != nil {
		return s.compareErr
	}
	return nil
}
