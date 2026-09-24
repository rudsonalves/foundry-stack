package httphandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	httpresponse "github.com/rudsonalves/foundry-stack/api/internal/shared/http/response"
	"github.com/rudsonalves/foundry-stack/api/internal/shared/middleware"
)

func TestHandleSuccess(t *testing.T) {
	called := false
	handler := Handle(func(w http.ResponseWriter, _ *http.Request) error {
		called = true
		w.WriteHeader(http.StatusCreated)
		return nil
	})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", nil))

	if !called {
		t.Error("application handler was not called")
	}
	if recorder.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
}

func TestHandleMapsApplicationErrors(t *testing.T) {
	tests := []struct {
		name       string
		appErr     *sharederrors.AppError
		wantStatus int
		wantAuth   bool
	}{
		{
			"bad request",
			&sharederrors.AppError{
				Code:    sharederrors.ErrCodeBadRequest,
				Message: "bad request",
			},
			http.StatusBadRequest,
			false,
		},
		{
			"unauthorized",
			&sharederrors.AppError{
				Code:    sharederrors.ErrCodeUnauthorized,
				Message: "unauthorized",
			},
			http.StatusUnauthorized,
			true,
		},
		{
			"forbidden",
			&sharederrors.AppError{
				Code:    sharederrors.ErrCodeForbidden,
				Message: "forbidden",
			},
			http.StatusForbidden,
			false,
		},
		{
			"not found",
			&sharederrors.AppError{
				Code:    sharederrors.ErrCodeNotFound,
				Message: "not found",
			},
			http.StatusNotFound,
			false,
		},
		{
			"conflict",
			&sharederrors.AppError{
				Code:    sharederrors.ErrCodeConflict,
				Message: "conflict",
			},
			http.StatusConflict,
			false,
		},
		{
			"unknown code",
			&sharederrors.AppError{
				Code:    "UNKNOWN",
				Message: "unknown",
			},
			http.StatusInternalServerError,
			false,
		},
		{
			"email already registered",
			&sharederrors.AppError{
				Code:    sharederrors.ErrCodeEmailAlreadyRegistered,
				Message: "email already registered",
			},
			http.StatusConflict,
			false,
		},
		{
			"email delivery unavailable",
			&sharederrors.AppError{
				Code:    sharederrors.ErrCodeEmailDeliveryUnavailable,
				Message: "email delivery temporarily unavailable",
			},
			http.StatusServiceUnavailable,
			false,
		},
		{
			"rate limit exceeded",
			&sharederrors.AppError{
				Code:    sharederrors.ErrCodeRateLimitExceeded,
				Message: "too many email verification requests",
			},
			http.StatusTooManyRequests,
			false,
		},
		{
			"invalid email verification",
			&sharederrors.AppError{
				Code:    sharederrors.ErrCodeInvalidEmailVerification,
				Message: "invalid email verification",
			},
			http.StatusBadRequest,
			false,
		},
		{
			"invalid password reset",
			&sharederrors.AppError{
				Code:    sharederrors.ErrCodeInvalidPasswordReset,
				Message: "invalid password reset",
			},
			http.StatusBadRequest,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := Handle(func(http.ResponseWriter, *http.Request) error {
				return tt.appErr
			})
			recorder := httptest.NewRecorder()

			handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/", nil))

			if recorder.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", recorder.Code, tt.wantStatus)
			}
			if got := recorder.Header().Get("WWW-Authenticate"); (got == "Bearer") != tt.wantAuth {
				t.Errorf("WWW-Authenticate = %q, want Bearer presence %v", got, tt.wantAuth)
			}

			body := decodeErrorResponse(t, recorder)
			if body.Code != tt.appErr.Code {
				t.Errorf("error.code = %q, want %q", body.Code, tt.appErr.Code)
			}
			if body.Message != tt.appErr.Message {
				t.Errorf("error.message = %q, want %q", body.Message, tt.appErr.Message)
			}
		})
	}
}

func TestHandleReturnsInternalErrorForUnexpectedError(t *testing.T) {
	handler := middleware.RequestID(Handle(func(http.ResponseWriter, *http.Request) error {
		return errors.New("database unavailable")
	}))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Request-ID", "request-123")

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	body := decodeErrorResponse(t, recorder)
	if body.Code != sharederrors.ErrCodeInternalError {
		t.Errorf("error.code = %q, want %q", body.Code, sharederrors.ErrCodeInternalError)
	}
	if body.Message != "Erro interno do servidor." {
		t.Errorf("error.message = %q, want %q", body.Message, "Erro interno do servidor.")
	}
	if body.RequestID != "request-123" {
		t.Errorf("error.request_id = %q, want %q", body.RequestID, "request-123")
	}
}

func TestHandleIncludesValidationDetails(t *testing.T) {
	violations := []sharederrors.FieldViolation{
		{Field: "name", Message: "é obrigatório"},
		{Field: "quantity", Message: "deve ser maior que zero"},
	}
	handler := Handle(func(http.ResponseWriter, *http.Request) error {
		return sharederrors.NewValidation("Dados inválidos.", violations)
	})
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, "/", nil))

	body := decodeErrorResponse(t, recorder)
	if len(body.Details) != len(violations) {
		t.Fatalf("error.details length = %d, want %d", len(body.Details), len(violations))
	}
	for i, violation := range violations {
		if body.Details[i].Field != violation.Field || body.Details[i].Message != violation.Message {
			t.Errorf("error.details[%d] = %#v, want field=%q message=%q", i, body.Details[i], violation.Field, violation.Message)
		}
	}
}

func decodeErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder) httpresponse.ErrorBody {
	t.Helper()

	var response httpresponse.ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return response.Error
}
