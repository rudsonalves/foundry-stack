package sharederrors

import "fmt"

type FieldViolation struct {
	Field   string
	Message string
}

type AppError struct {
	Code       string
	Message    string
	Violations []FieldViolation
	Err        error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Code, e.Err)
	}

	return e.Message
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewBadRequest(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeBadRequest,
		Message: message,
		Err:     err,
	}
}

func NewValidation(
	message string,
	violations []FieldViolation,
) *AppError {
	return &AppError{
		Code:       ErrCodeBadRequest,
		Message:    message,
		Violations: violations,
	}
}

func NewNotFound(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeNotFound,
		Message: message,
		Err:     err,
	}
}

func NewConflict(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeConflict,
		Message: message,
		Err:     err,
	}
}

func NewUnauthorized(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeUnauthorized,
		Message: message,
		Err:     err,
	}
}

func NewForbidden(message string, err error) *AppError {
	return &AppError{
		Code:    ErrCodeForbidden,
		Message: message,
		Err:     err,
	}
}
