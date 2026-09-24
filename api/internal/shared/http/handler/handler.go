package httphandler

import (
	"net/http"

	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
)

type AppHandler func(
	w http.ResponseWriter,
	r *http.Request,
) error

func Handle(fn AppHandler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := fn(w, r); err != nil {
			handleError(w, r, err)
			return
		}
	})
}

func statusFromCode(code string) int {
	switch code {
	case sharederrors.ErrCodeBadRequest,
		sharederrors.ErrCodeInvalidEmailVerification,
		sharederrors.ErrCodeInvalidPasswordReset:
		return http.StatusBadRequest
	case sharederrors.ErrCodeUnauthorized:
		return http.StatusUnauthorized
	case sharederrors.ErrCodeForbidden:
		return http.StatusForbidden
	case sharederrors.ErrCodeNotFound:
		return http.StatusNotFound
	case sharederrors.ErrCodeConflict,
		sharederrors.ErrCodeEmailAlreadyRegistered:
		return http.StatusConflict
	case sharederrors.ErrCodeEmailDeliveryUnavailable:
		return http.StatusServiceUnavailable
	case sharederrors.ErrCodeRateLimitExceeded:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}
