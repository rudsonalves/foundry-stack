package sharederrors

import (
	"errors"
	"reflect"
	"testing"
)

func TestAppErrorError(t *testing.T) {
	t.Run("returns message without underlying error", func(t *testing.T) {
		err := &AppError{
			Code:    ErrCodeBadRequest,
			Message: "Requisição inválida.",
		}

		if got := err.Error(); got != "Requisição inválida." {
			t.Errorf("Error() = %q, want %q", got, "Requisição inválida.")
		}
	})

	t.Run("returns code and underlying error", func(t *testing.T) {
		cause := errors.New("database unavailable")
		err := &AppError{
			Code:    ErrCodeInternalError,
			Message: "Erro interno.",
			Err:     cause,
		}

		if got := err.Error(); got != "INTERNAL_ERROR: database unavailable" {
			t.Errorf("Error() = %q, want %q", got, "INTERNAL_ERROR: database unavailable")
		}
	})
}

func TestAppErrorUnwrap(t *testing.T) {
	cause := errors.New("original error")
	err := &AppError{
		Code: ErrCodeConflict,
		Err:  cause,
	}

	if got := errors.Unwrap(err); got != cause {
		t.Errorf("errors.Unwrap() = %v, want original cause", got)
	}
	if !errors.Is(err, cause) {
		t.Error("errors.Is() = false, want true for original cause")
	}
}

func TestAppErrorConstructors(t *testing.T) {
	cause := errors.New("cause")
	tests := []struct {
		name    string
		newErr  func() *AppError
		wantErr AppError
	}{
		{
			name: "bad request",
			newErr: func() *AppError {
				return NewBadRequest("Requisição inválida.", cause)
			},
			wantErr: AppError{
				Code:    ErrCodeBadRequest,
				Message: "Requisição inválida.",
				Err:     cause,
			},
		},
		{
			name: "not found",
			newErr: func() *AppError {
				return NewNotFound("Recurso não encontrado.", cause)
			},
			wantErr: AppError{
				Code:    ErrCodeNotFound,
				Message: "Recurso não encontrado.",
				Err:     cause,
			},
		},
		{
			name: "conflict",
			newErr: func() *AppError {
				return NewConflict("Recurso já existente.", cause)
			},
			wantErr: AppError{
				Code:    ErrCodeConflict,
				Message: "Recurso já existente.",
				Err:     cause,
			},
		},
		{
			name: "unauthorized",
			newErr: func() *AppError {
				return NewUnauthorized("Não autorizado.", cause)
			},
			wantErr: AppError{
				Code:    ErrCodeUnauthorized,
				Message: "Não autorizado.",
				Err:     cause,
			},
		},
		{
			name: "forbidden",
			newErr: func() *AppError {
				return NewForbidden("Operação não permitida.", cause)
			},
			wantErr: AppError{
				Code:    ErrCodeForbidden,
				Message: "Operação não permitida.",
				Err:     cause,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.newErr()
			if got.Code != tt.wantErr.Code {
				t.Errorf("Code = %q, want %q", got.Code, tt.wantErr.Code)
			}
			if got.Message != tt.wantErr.Message {
				t.Errorf("Message = %q, want %q", got.Message, tt.wantErr.Message)
			}
			if !errors.Is(got, tt.wantErr.Err) {
				t.Errorf("error %v does not wrap %v", got, tt.wantErr.Err)
			}
			if !errors.Is(got, cause) {
				t.Error("constructed error does not wrap the provided cause")
			}
		})
	}
}

func TestNewValidation(t *testing.T) {
	violations := []FieldViolation{
		{Field: "name", Message: "é obrigatório"},
		{Field: "quantity", Message: "deve ser maior que zero"},
	}

	err := NewValidation("Dados inválidos.", violations)

	if err.Code != ErrCodeBadRequest {
		t.Errorf("Code = %q, want %q", err.Code, ErrCodeBadRequest)
	}
	if err.Message != "Dados inválidos." {
		t.Errorf("Message = %q, want %q", err.Message, "Dados inválidos.")
	}
	if !reflect.DeepEqual(err.Violations, violations) {
		t.Errorf("Violations = %#v, want %#v", err.Violations, violations)
	}
	if err.Err != nil {
		t.Errorf("Err = %v, want nil", err.Err)
	}
}
