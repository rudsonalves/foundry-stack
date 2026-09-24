package infrastructure

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

const (
	existingUserQuery = `
		SELECT 1
		FROM users
		WHERE email = $1
	`
	findVerificationForUserQuery = `
		SELECT confirmed_at
		FROM email_verification_challenges
		WHERE verification_token_hash = $1
			AND email = $2
			AND confirmed_at IS NOT NULL
			AND consumed_at IS NULL
			AND invalidated_at IS NULL
			AND verification_token_expires_at > $3
		FOR UPDATE
	`
	createUserQuery = `
		INSERT INTO users (
			id,
			name,
			email,
			password_hash,
			email_verified_at
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at, updated_at
	`
	consumeVerificationQuery = `
		UPDATE email_verification_challenges
		SET
			consumed_at = $2,
			updated_at = $2
		WHERE verification_token_hash = $1
			AND consumed_at IS NULL
	`
	findUserByIDQuery = `
		SELECT id, name, email, COALESCE(password_hash, '') AS password_hash, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	findUserByEmailQuery = `
		SELECT id, name, email, COALESCE(password_hash, '') AS password_hash, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	listUsersQuery = `
		SELECT id, name, email, created_at, updated_at
		FROM users
		ORDER BY created_at, id
	`
	updateUserQuery = `
		UPDATE users
		SET name = $2,
			email = $3,
			updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, email, created_at, updated_at
	`
	deleteUserQuery = `
		DELETE FROM users
		WHERE id = $1
	`
)

func TestUserStoreCreate(t *testing.T) {
	now := time.Date(2026, time.September, 11, 12, 0, 0, 0, time.UTC)
	confirmedAt := now.Add(-time.Hour)
	createdAt := now
	updatedAt := now
	tokenHash := []byte("verification-token-hash")

	input := domain.CreateUserDTO{
		Name:                  "Ada Lovelace",
		Email:                 "ada@example.com",
		PasswordHash:          "hashedpassword",
		VerificationTokenHash: tokenHash,
		Now:                   now,
	}

	t.Run("creates verified user and consumes verification", func(t *testing.T) {
		store, mock := newMockUserStore(t)

		mock.ExpectBegin()

		mock.ExpectQuery(regexp.QuoteMeta(existingUserQuery)).
			WithArgs(input.Email).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}))

		mock.ExpectQuery(regexp.QuoteMeta(findVerificationForUserQuery)).
			WithArgs(tokenHash, input.Email, now).
			WillReturnRows(
				sqlmock.NewRows([]string{"confirmed_at"}).
					AddRow(confirmedAt),
			)

		mock.ExpectQuery(regexp.QuoteMeta(createUserQuery)).
			WithArgs(
				sqlmock.AnyArg(),
				input.Name,
				input.Email,
				input.PasswordHash,
				confirmedAt,
			).
			WillReturnRows(
				sqlmock.NewRows([]string{"created_at", "updated_at"}).
					AddRow(createdAt, updatedAt),
			)

		mock.ExpectExec(regexp.QuoteMeta(consumeVerificationQuery)).
			WithArgs(tokenHash, now).
			WillReturnResult(sqlmock.NewResult(0, 1))

		mock.ExpectCommit()

		user, err := store.Create(context.Background(), input)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		if user.ID == uuid.Nil {
			t.Fatal("Create() returned nil UUID")
		}
		if user.Name != input.Name || user.Email != input.Email {
			t.Fatalf("Create() user = %#v", user)
		}
		if user.PasswordHash != input.PasswordHash {
			t.Fatalf("Create() password hash = %q", user.PasswordHash)
		}
		if !user.EmailVerifiedAt.Equal(confirmedAt) {
			t.Fatalf(
				"Create() email verified at = %v, want %v",
				user.EmailVerifiedAt,
				confirmedAt,
			)
		}
	})

	t.Run("rolls back user creation when verification consumption fails", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		consumeErr := errors.New("consume failed")

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(existingUserQuery)).
			WithArgs(input.Email).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}))
		mock.ExpectQuery(regexp.QuoteMeta(findVerificationForUserQuery)).
			WithArgs(tokenHash, input.Email, now).
			WillReturnRows(
				sqlmock.NewRows([]string{"confirmed_at"}).
					AddRow(confirmedAt),
			)
		mock.ExpectQuery(regexp.QuoteMeta(createUserQuery)).
			WithArgs(
				sqlmock.AnyArg(),
				input.Name,
				input.Email,
				input.PasswordHash,
				confirmedAt,
			).
			WillReturnRows(
				sqlmock.NewRows([]string{"created_at", "updated_at"}).
					AddRow(createdAt, updatedAt),
			)
		mock.ExpectExec(regexp.QuoteMeta(consumeVerificationQuery)).
			WithArgs(tokenHash, now).
			WillReturnError(consumeErr)
		mock.ExpectRollback()

		_, err := store.Create(context.Background(), input)
		if !errors.Is(err, consumeErr) {
			t.Fatalf("Create() error = %v, want wrapped %v", err, consumeErr)
		}
	})

	t.Run("rejects existing email", func(t *testing.T) {
		store, mock := newMockUserStore(t)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(existingUserQuery)).
			WithArgs(input.Email).
			WillReturnRows(
				sqlmock.NewRows([]string{"exists"}).AddRow(1),
			)
		mock.ExpectRollback()

		_, err := store.Create(context.Background(), input)
		assertError(
			t,
			err,
			domain.ErrEmailAlreadyExists,
			"create user",
		)
	})

	t.Run("rejects invalid verification", func(t *testing.T) {
		store, mock := newMockUserStore(t)

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(existingUserQuery)).
			WithArgs(input.Email).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}))
		mock.ExpectQuery(regexp.QuoteMeta(findVerificationForUserQuery)).
			WithArgs(tokenHash, input.Email, now).
			WillReturnError(sql.ErrNoRows)
		mock.ExpectRollback()

		_, err := store.Create(context.Background(), input)
		assertError(
			t,
			err,
			domain.ErrInvalidEmailVerification,
			"validate email verification",
		)
	})
}

func TestUserStoreFindByID(t *testing.T) {
	id := uuid.MustParse("9764ef1d-ae2e-4ba2-a92a-d81a7d4894aa")
	createdAt := time.Date(2026, time.July, 17, 10, 0, 0, 0, time.UTC)

	t.Run("finds user", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		mock.ExpectQuery(regexp.QuoteMeta(findUserByIDQuery)).
			WithArgs(id).
			WillReturnRows(userRowsWithPassword().AddRow(id, "Ada Lovelace", "ada@example.com", "hashedpassword", createdAt, createdAt))

		user, err := store.FindByID(context.Background(), id)
		if err != nil {
			t.Fatalf("FindByID() error = %v", err)
		}
		if user.ID != id || user.Name != "Ada Lovelace" || user.Email != "ada@example.com" {
			t.Fatalf("FindByID() user = %#v", user)
		}
		if user.PasswordHash != "hashedpassword" {
			t.Fatalf("FindByID() password hash = %q", user.PasswordHash)
		}
	})

	t.Run("reports missing user", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		mock.ExpectQuery(regexp.QuoteMeta(findUserByIDQuery)).WithArgs(id).WillReturnError(sql.ErrNoRows)

		_, err := store.FindByID(context.Background(), id)
		assertError(t, err, domain.ErrUserNotFound, "user "+id.String()+" not found")
	})

	t.Run("wraps database error", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		dbErr := errors.New("query failed")
		mock.ExpectQuery(regexp.QuoteMeta(findUserByIDQuery)).WithArgs(id).WillReturnError(dbErr)

		_, err := store.FindByID(context.Background(), id)
		assertError(t, err, dbErr, "query user by id "+id.String())
	})
}

func TestUserStoreFindByEmail(t *testing.T) {
	email := "ada@example.com"
	id := uuid.MustParse("9764ef1d-ae2e-4ba2-a92a-d81a7d4894aa")
	createdAt := time.Date(2026, time.July, 17, 10, 0, 0, 0, time.UTC)

	t.Run("finds user", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		mock.ExpectQuery(regexp.QuoteMeta(findUserByEmailQuery)).
			WithArgs(email).
			WillReturnRows(userRowsWithPassword().AddRow(id, "Ada Lovelace", email, "hashedpassword", createdAt, createdAt))

		user, err := store.FindByEmail(context.Background(), email)
		if err != nil {
			t.Fatalf("FindByEmail() error = %v", err)
		}
		if user.ID != id || user.Name != "Ada Lovelace" || user.Email != email {
			t.Fatalf("FindByEmail() user = %#v", user)
		}
		if user.PasswordHash != "hashedpassword" {
			t.Fatalf("FindByEmail() password hash = %q", user.PasswordHash)
		}
	})

	t.Run("reports missing user", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		mock.ExpectQuery(regexp.QuoteMeta(findUserByEmailQuery)).WithArgs(email).WillReturnError(sql.ErrNoRows)

		_, err := store.FindByEmail(context.Background(), email)
		assertError(t, err, domain.ErrUserNotFound, "user "+email+" not found")
	})

	t.Run("wraps database error", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		dbErr := errors.New("query failed")
		mock.ExpectQuery(regexp.QuoteMeta(findUserByEmailQuery)).WithArgs(email).WillReturnError(dbErr)

		_, err := store.FindByEmail(context.Background(), email)
		assertError(t, err, dbErr, "query user by email "+email)
	})
}

func TestUserStoreList(t *testing.T) {
	firstID := uuid.MustParse("9764ef1d-ae2e-4ba2-a92a-d81a7d4894aa")
	secondID := uuid.MustParse("bfa30984-a5c2-4e70-999c-0e1014993418")
	createdAt := time.Date(2026, time.July, 17, 10, 0, 0, 0, time.UTC)

	t.Run("lists users in database order", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		mock.ExpectQuery(regexp.QuoteMeta(listUsersQuery)).WillReturnRows(userRows().
			AddRow(firstID, "Ada Lovelace", "ada@example.com", createdAt, createdAt).
			AddRow(secondID, "Grace Hopper", "grace@example.com", createdAt.Add(time.Minute), createdAt.Add(time.Minute)))

		users, err := store.List(context.Background())
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(users) != 2 || users[0].ID != firstID || users[1].ID != secondID {
			t.Fatalf("List() users = %#v", users)
		}
	})

	t.Run("returns an empty non-nil slice representation", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		mock.ExpectQuery(regexp.QuoteMeta(listUsersQuery)).WillReturnRows(userRows())

		users, err := store.List(context.Background())
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}
		if len(users) != 0 {
			t.Fatalf("List() users = %#v, want empty", users)
		}
	})

	t.Run("wraps query error", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		dbErr := errors.New("query failed")
		mock.ExpectQuery(regexp.QuoteMeta(listUsersQuery)).WillReturnError(dbErr)

		_, err := store.List(context.Background())
		assertError(t, err, dbErr, "query users")
	})

	t.Run("wraps scan error", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		mock.ExpectQuery(regexp.QuoteMeta(listUsersQuery)).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(firstID))

		_, err := store.List(context.Background())
		assertError(t, err, nil, "scan user")
	})

	t.Run("wraps iteration error", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		iterationErr := errors.New("connection interrupted")
		mock.ExpectQuery(regexp.QuoteMeta(listUsersQuery)).WillReturnRows(userRows().
			AddRow(firstID, "Ada Lovelace", "ada@example.com", createdAt, createdAt).
			AddRow(secondID, "Grace Hopper", "grace@example.com", createdAt, createdAt).
			RowError(1, iterationErr))

		_, err := store.List(context.Background())
		assertError(t, err, iterationErr, "iterate users")
	})
}

func TestUserStoreUpdate(t *testing.T) {
	id := uuid.MustParse("9764ef1d-ae2e-4ba2-a92a-d81a7d4894aa")
	createdAt := time.Date(2026, time.July, 17, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)

	t.Run("updates user", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		mock.ExpectQuery(regexp.QuoteMeta(updateUserQuery)).
			WithArgs(id, "Ada Byron", "ada.byron@example.com").
			WillReturnRows(userRows().AddRow(id, "Ada Byron", "ada.byron@example.com", createdAt, updatedAt))

		user, err := store.Update(context.Background(), id, "Ada Byron", "ada.byron@example.com")
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}
		if user.ID != id || user.Name != "Ada Byron" || user.Email != "ada.byron@example.com" || !user.UpdatedAt.Equal(updatedAt) {
			t.Fatalf("Update() user = %#v", user)
		}
	})

	t.Run("reports missing user", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		mock.ExpectQuery(regexp.QuoteMeta(updateUserQuery)).
			WithArgs(id, "Ada Byron", "ada.byron@example.com").
			WillReturnError(sql.ErrNoRows)

		_, err := store.Update(context.Background(), id, "Ada Byron", "ada.byron@example.com")
		assertError(t, err, domain.ErrUserNotFound, "user "+id.String()+" not found")
	})

	t.Run("wraps database error", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		dbErr := errors.New("update failed")
		mock.ExpectQuery(regexp.QuoteMeta(updateUserQuery)).
			WithArgs(id, "Ada Byron", "ada.byron@example.com").
			WillReturnError(dbErr)

		_, err := store.Update(context.Background(), id, "Ada Byron", "ada.byron@example.com")
		assertError(t, err, dbErr, "update user "+id.String())
	})
}

func TestUserStoreDelete(t *testing.T) {
	id := uuid.MustParse("9764ef1d-ae2e-4ba2-a92a-d81a7d4894aa")

	t.Run("deletes user", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		mock.ExpectExec(regexp.QuoteMeta(deleteUserQuery)).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 1))

		if err := store.Delete(context.Background(), id); err != nil {
			t.Fatalf("Delete() error = %v", err)
		}
	})

	t.Run("reports missing user", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		mock.ExpectExec(regexp.QuoteMeta(deleteUserQuery)).WithArgs(id).WillReturnResult(sqlmock.NewResult(0, 0))

		err := store.Delete(context.Background(), id)
		assertError(t, err, domain.ErrUserNotFound, "user "+id.String()+" not found")
	})

	t.Run("wraps execution error", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		dbErr := errors.New("delete failed")
		mock.ExpectExec(regexp.QuoteMeta(deleteUserQuery)).WithArgs(id).WillReturnError(dbErr)

		err := store.Delete(context.Background(), id)
		assertError(t, err, dbErr, "delete user "+id.String())
	})

	t.Run("wraps rows affected error", func(t *testing.T) {
		store, mock := newMockUserStore(t)
		resultErr := errors.New("result unavailable")
		mock.ExpectExec(regexp.QuoteMeta(deleteUserQuery)).WithArgs(id).WillReturnResult(sqlmock.NewErrorResult(resultErr))

		err := store.Delete(context.Background(), id)
		assertError(t, err, resultErr, "delete user "+id.String())
	})
}

func TestTranslateUserError(t *testing.T) {
	t.Run("translates missing row", func(t *testing.T) {
		if err := translateUserError(sql.ErrNoRows); !errors.Is(err, domain.ErrUserNotFound) {
			t.Fatalf("translateUserError() error = %v", err)
		}
	})

	t.Run("translates duplicate email", func(t *testing.T) {
		err := &pgconn.PgError{Code: "23505", ConstraintName: "users_email_unique"}
		if err := translateUserError(err); !errors.Is(err, domain.ErrEmailAlreadyExists) {
			t.Fatalf("translateUserError() error = %v", err)
		}
	})

	t.Run("preserves unexpected cause", func(t *testing.T) {
		cause := errors.New("database unavailable")
		if err := translateUserError(cause); !errors.Is(err, cause) {
			t.Fatalf("translateUserError() error = %v", err)
		}
	})
}

func newMockUserStore(t *testing.T) (*SQLUserRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New() error = %v", err)
	}
	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unmet database expectations: %v", err)
		}
		mock.ExpectClose()
		if err := db.Close(); err != nil {
			t.Errorf("close mock database: %v", err)
		}
	})

	return NewSQLUserRepository(db), mock
}

func userRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "name", "email", "created_at", "updated_at"})
}

func userRowsWithPassword() *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "name", "email", "password_hash", "created_at", "updated_at"})
}

func assertError(t *testing.T, err error, cause error, message string) {
	t.Helper()

	if err == nil {
		t.Fatal("expected an error, got nil")
	}
	if cause != nil && !errors.Is(err, cause) {
		t.Fatalf("error %v does not wrap %v", err, cause)
	}
	if !strings.Contains(err.Error(), message) {
		t.Fatalf("error %q does not contain %q", err, message)
	}
}
