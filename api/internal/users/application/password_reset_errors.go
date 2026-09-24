package application

import (
	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
)

func invalidPasswordResetError(err error) *sharederrors.AppError {
	return &sharederrors.AppError{
		Code:    sharederrors.ErrCodeInvalidPasswordReset,
		Message: "invalid password reset",
		Err:     err,
	}
}
