package transport

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	httpresponse "github.com/rudsonalves/foundry-stack/api/internal/shared/http/response"
	"github.com/rudsonalves/foundry-stack/api/internal/users/application"
)

type PasswordResetRequester interface {
	Request(context.Context, application.RequestPasswordResetInput) (application.RequestPasswordResetResult, error)
	Confirm(context.Context, application.ConfirmPasswordResetInput) (application.ConfirmPasswordResetResult, error)
	Complete(context.Context, application.CompletePasswordResetInput) error
}
type PasswordResetHandler struct {
	service   PasswordResetRequester
	clientIPs *ClientIPResolver
}

func NewPasswordResetHandler(service PasswordResetRequester, clientIPs *ClientIPResolver) *PasswordResetHandler {
	return &PasswordResetHandler{service: service, clientIPs: clientIPs}
}

type RequestPasswordResetRequest struct {
	Email string `json:"email"`
}
type RequestPasswordResetResponse struct {
	PasswordResetID   string    `json:"password_reset_id"`
	CodeExpiresAt     time.Time `json:"code_expires_at"`
	ResendAvailableAt time.Time `json:"resend_available_at"`
}
type ConfirmPasswordResetRequest struct {
	PasswordResetID string `json:"password_reset_id"`
	Code            string `json:"code"`
}
type ConfirmPasswordResetResponse struct {
	PasswordResetToken string    `json:"password_reset_token"`
	ExpiresAt          time.Time `json:"expires_at"`
}
type CompletePasswordResetRequest struct {
	PasswordResetID    string `json:"password_reset_id"`
	PasswordResetToken string `json:"password_reset_token"`
	NewPassword        string `json:"new_password"`
}

func (h *PasswordResetHandler) Request(w http.ResponseWriter, r *http.Request) error {
	var request RequestPasswordResetRequest
	if err := decodeJSONBody(w, r, &request); err != nil {
		return sharederrors.NewBadRequest("invalid request body", err)
	}
	clientIP := "unknown"
	if h.clientIPs != nil {
		clientIP = h.clientIPs.Resolve(r)
	}
	result, err := h.service.Request(r.Context(), application.RequestPasswordResetInput{Email: request.Email, ClientIP: clientIP})
	if err != nil {
		return err
	}
	return httpresponse.WriteData(w, http.StatusAccepted, RequestPasswordResetResponse{PasswordResetID: result.PasswordResetID.String(), CodeExpiresAt: result.CodeExpiresAt, ResendAvailableAt: result.ResendAvailableAt})
}
func (h *PasswordResetHandler) Confirm(w http.ResponseWriter, r *http.Request) error {
	var request ConfirmPasswordResetRequest
	if err := decodeJSONBody(w, r, &request); err != nil {
		return sharederrors.NewBadRequest("invalid request body", err)
	}
	id, err := uuid.Parse(request.PasswordResetID)
	if err != nil {
		return sharederrors.NewValidation("invalid input", []sharederrors.FieldViolation{{Field: "password_reset_id", Message: "invalid password reset ID"}})
	}
	result, err := h.service.Confirm(r.Context(), application.ConfirmPasswordResetInput{PasswordResetID: id, Code: request.Code})
	if err != nil {
		return err
	}
	return httpresponse.WriteData(w, http.StatusOK, ConfirmPasswordResetResponse{PasswordResetToken: result.PasswordResetToken, ExpiresAt: result.ExpiresAt})
}

func (h *PasswordResetHandler) Complete(w http.ResponseWriter, r *http.Request) error {
	var request CompletePasswordResetRequest
	if err := decodeJSONBody(w, r, &request); err != nil {
		return sharederrors.NewBadRequest("invalid request body", err)
	}
	id, err := uuid.Parse(request.PasswordResetID)
	if err != nil {
		return sharederrors.NewValidation("invalid input", []sharederrors.FieldViolation{{Field: "password_reset_id", Message: "invalid password reset ID"}})
	}
	if err := h.service.Complete(r.Context(), application.CompletePasswordResetInput{PasswordResetID: id, PasswordResetToken: request.PasswordResetToken, NewPassword: request.NewPassword}); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
