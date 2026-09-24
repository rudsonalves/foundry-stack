package application

import (
	"errors"
	"testing"

	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

func TestInvalidPasswordResetErrorHidesSemanticCause(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		cause error
	}{
		{
			name:  "challenge not found",
			cause: domain.ErrPasswordResetNotFound,
		},
		{
			name:  "invalid reset",
			cause: domain.ErrInvalidPasswordReset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := invalidPasswordResetError(tt.cause)

			if err.Code != sharederrors.ErrCodeInvalidPasswordReset {
				t.Errorf("code = %q, want %q", err.Code, sharederrors.ErrCodeInvalidPasswordReset)
			}
			if err.Message != "invalid password reset" {
				t.Errorf("message = %q, want %q", err.Message, "invalid password reset")
			}
			if !errors.Is(err, tt.cause) {
				t.Errorf("error does not preserve cause %v", tt.cause)
			}
		})
	}
}
