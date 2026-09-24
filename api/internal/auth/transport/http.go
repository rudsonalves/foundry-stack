package transport

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/rudsonalves/foundry-stack/api/internal/auth/application"
	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	httpresponse "github.com/rudsonalves/foundry-stack/api/internal/shared/http/response"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	ClientID string `json:"client_id"`
}

type LoginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
}

type Handler struct {
	service *application.AuthService
}

func NewHandler(service *application.AuthService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(
	w http.ResponseWriter,
	r *http.Request,
) error {
	var request LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return sharederrors.NewBadRequest(
			"invalid request body",
			err,
		)
	}

	result, err := h.service.Login(
		r.Context(),
		request.Email,
		request.Password,
		request.ClientID,
	)
	if err != nil {
		return err
	}

	expiresIn := max(
		int64(time.Until(result.ExpiresAt).Seconds()),
		0,
	)

	return httpresponse.WriteData(
		w,
		http.StatusOK,
		LoginResponse{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
			TokenType:    "Bearer",
			ExpiresIn:    expiresIn,
		},
	)
}

func (h *Handler) Refresh(
	w http.ResponseWriter,
	r *http.Request,
) error {
	var request RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return sharederrors.NewBadRequest(
			"invalid request body",
			err,
		)
	}

	result, err := h.service.Refresh(
		r.Context(),
		request.RefreshToken,
	)
	if err != nil {
		return err
	}

	return httpresponse.WriteData(
		w,
		http.StatusOK,
		RefreshResponse{
			AccessToken: result.AccessToken,
			TokenType:   "Bearer",
			ExpiresIn: max(
				int64(time.Until(result.ExpiresAt).Seconds()),
				0,
			),
		},
	)
}

func (h *Handler) Logout(
	w http.ResponseWriter,
	r *http.Request,
) error {
	var request RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return sharederrors.NewBadRequest(
			"invalid request body",
			err,
		)
	}

	if err := h.service.Logout(
		r.Context(),
		request.RefreshToken,
	); err != nil {
		return err
	}

	w.WriteHeader(http.StatusNoContent)
	return nil
}
