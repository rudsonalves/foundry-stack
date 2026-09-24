package transport

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type AuthenticatedUser struct {
	ID uuid.UUID
}

type authenticatedUserKey struct{}

func withAuthenticatedUser(
	ctx context.Context,
	user AuthenticatedUser,
) context.Context {
	return context.WithValue(
		ctx,
		authenticatedUserKey{},
		user,
	)
}

func CurrentUser(
	ctx context.Context,
) (AuthenticatedUser, bool) {
	user, ok := ctx.Value(authenticatedUserKey{}).(AuthenticatedUser)
	return user, ok
}

func RequireCurrentUser(
	ctx context.Context,
) (AuthenticatedUser, error) {
	user, ok := CurrentUser(ctx)
	if !ok {
		return AuthenticatedUser{}, errors.New(
			"handler called without authenticated user",
		)
	}

	return user, nil
}
