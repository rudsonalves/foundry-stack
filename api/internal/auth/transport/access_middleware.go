package transport

import (
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	httphandler "github.com/rudsonalves/foundry-stack/api/internal/shared/http/handler"
)

var ErrMissingBearerToken = errors.New("missing bearer token")

type TokenParser interface {
	Parse(token string) (uuid.UUID, error)
}

type Authentication struct {
	tokens TokenParser
}

func NewAuthentication(
	tokens TokenParser,
) *Authentication {
	return &Authentication{
		tokens: tokens,
	}
}

func (a *Authentication) RequireUser(
	next httphandler.AppHandler,
) httphandler.AppHandler {
	return func(
		w http.ResponseWriter,
		r *http.Request,
	) error {
		token, err := bearerToken(r)
		if err != nil {
			w.Header().Set("WWW-Authenticate", "Bearer")
			return sharederrors.NewUnauthorized(
				"Unauthorized.",
				err,
			)
		}

		userID, err := a.tokens.Parse(token)
		if err != nil {
			w.Header().Set("WWW-Authenticate", "Bearer")
			return sharederrors.NewUnauthorized(
				"Unauthorized.",
				err,
			)
		}

		ctx := withAuthenticatedUser(
			r.Context(),
			AuthenticatedUser{ID: userID},
		)

		return next(w, r.WithContext(ctx))
	}
}

func bearerToken(r *http.Request) (string, error) {
	values := r.Header.Values("Authorization")
	if len(values) != 1 {
		return "", ErrMissingBearerToken
	}

	parts := strings.Fields(values[0])
	if len(parts) != 2 ||
		!strings.EqualFold(parts[0], "Bearer") ||
		parts[1] == "" {
		return "", ErrMissingBearerToken
	}

	return parts[1], nil
}
