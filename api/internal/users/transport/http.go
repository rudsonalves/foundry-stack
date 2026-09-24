package transport

import (
	"net/http"

	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	httpresponse "github.com/rudsonalves/foundry-stack/api/internal/shared/http/response"
	"github.com/rudsonalves/foundry-stack/api/internal/users/application"
)

type RegisterRequest struct {
	Name                   string `json:"name"`
	Email                  string `json:"email"`
	Password               string `json:"password"`
	EmailVerificationToken string `json:"email_verification_token"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(
	w http.ResponseWriter,
	r *http.Request,
) error {
	var request RegisterRequest
	if err := decodeJSONBody(w, r, &request); err != nil {
		return sharederrors.NewBadRequest("invalid request body", err)
	}

	user, err := h.service.Register(
		r.Context(),
		application.RegisterInput{
			Name:                   request.Name,
			Email:                  request.Email,
			Password:               request.Password,
			EmailVerificationToken: request.EmailVerificationToken,
		},
	)
	if err != nil {
		return err
	}

	return httpresponse.WriteData(
		w,
		http.StatusCreated,
		UserResponse{
			ID:    user.ID.String(),
			Name:  user.Name,
			Email: user.Email,
		},
	)
}
