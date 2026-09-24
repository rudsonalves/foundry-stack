package httphandler

import (
	"errors"
	"log"
	"net/http"

	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	httpresponse "github.com/rudsonalves/foundry-stack/api/internal/shared/http/response"
	"github.com/rudsonalves/foundry-stack/api/internal/shared/middleware"
)

func handleError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := middleware.GetRequestID(r)

	if appErr, ok := errors.AsType[*sharederrors.AppError](err); ok {
		status := statusFromCode(appErr.Code)
		if status == http.StatusUnauthorized {
			w.Header().Set("WWW-Authenticate", "Bearer")
		}

		httpresponse.WriteError(
			w,
			status,
			appErr.Code,
			appErr.Message,
			requestID,
			fieldErrors(appErr.Violations),
		)
		return
	}

	log.Printf(
		"erro inesperado: request_id=%s error=%v",
		requestID,
		err,
	)

	httpresponse.WriteError(
		w,
		http.StatusInternalServerError,
		sharederrors.ErrCodeInternalError,
		"Erro interno do servidor.",
		requestID,
		nil,
	)
}

func fieldErrors(violations []sharederrors.FieldViolation,
) []httpresponse.FieldError {
	details := make([]httpresponse.FieldError, len(violations))

	for i, violation := range violations {
		details[i] = httpresponse.FieldError{
			Field:   violation.Field,
			Message: violation.Message,
		}
	}

	return details
}
