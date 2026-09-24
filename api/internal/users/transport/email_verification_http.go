package transport

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/rudsonalves/foundry-stack/api/internal/users/application"

	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	httpresponse "github.com/rudsonalves/foundry-stack/api/internal/shared/http/response"
)

type RequestEmailVerificationRequest struct {
	Email string `json:"email"`
}

type RequestEmailVerificationResponse struct {
	VerificationID    string    `json:"verification_id"`
	CodeExpiresAt     time.Time `json:"code_expires_at"`
	ResendAvailableAt time.Time `json:"resend_available_at"`
}

type EmailVerificationHandler struct {
	service EmailVerificationRequester
}

type EmailVerificationRequester interface {
	Request(
		ctx context.Context,
		input application.RequestEmailVerificationInput,
	) (application.RequestEmailVerificationResult, error)

	Confirm(
		ctx context.Context,
		input application.ConfirmEmailVerificationInput,
	) (application.ConfirmEmailVerificationResult, error)
}

type ConfirmEmailVerificationRequest struct {
	VerificationID string `json:"verification_id"`
	Code           string `json:"code"`
}

type ConfirmEmailVerificationResponse struct {
	EmailVerificationToken string    `json:"email_verification_token"`
	ExpiresAt              time.Time `json:"expires_at"`
}

func NewEmailVerificationHandler(
	service EmailVerificationRequester,
) *EmailVerificationHandler {
	return &EmailVerificationHandler{
		service: service,
	}
}

func (h *EmailVerificationHandler) Request(
	w http.ResponseWriter,
	r *http.Request,
) error {
	var request RequestEmailVerificationRequest

	if err := decodeJSONBody(w, r, &request); err != nil {
		return sharederrors.NewBadRequest(
			"invalid request body",
			err,
		)
	}

	result, err := h.service.Request(
		r.Context(),
		application.RequestEmailVerificationInput{
			Email:    request.Email,
			ClientIP: requestClientIP(r),
		},
	)
	if err != nil {
		return err
	}

	return httpresponse.WriteData(
		w,
		http.StatusAccepted,
		RequestEmailVerificationResponse{
			VerificationID:    result.VerificationID.String(),
			CodeExpiresAt:     result.CodeExpiresAt,
			ResendAvailableAt: result.ResendAvailableAt,
		},
	)
}

func (h *EmailVerificationHandler) Confirm(
	w http.ResponseWriter,
	r *http.Request,
) error {
	var request ConfirmEmailVerificationRequest

	if err := decodeJSONBody(w, r, &request); err != nil {
		return sharederrors.NewBadRequest(
			"invalid request body",
			err,
		)
	}

	verificationID, err := uuid.Parse(request.VerificationID)
	if err != nil {
		return sharederrors.NewValidation(
			"invalid input",
			[]sharederrors.FieldViolation{
				{
					Field:   "verification_id",
					Message: "invalid verification ID",
				},
			},
		)
	}

	result, err := h.service.Confirm(
		r.Context(),
		application.ConfirmEmailVerificationInput{
			VerificationID: verificationID,
			Code:           request.Code,
		},
	)
	if err != nil {
		return err
	}

	return httpresponse.WriteData(
		w,
		http.StatusOK,
		ConfirmEmailVerificationResponse{
			EmailVerificationToken: result.EmailVerificationToken,
			ExpiresAt:              result.ExpiresAt,
		},
	)
}

// requestClientIP extracts the client's IP address from the HTTP request. If
// the IP address cannot be determined, it returns "unknown".
func requestClientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	if r.RemoteAddr == "" {
		return "unknown"
	}

	return r.RemoteAddr
}
