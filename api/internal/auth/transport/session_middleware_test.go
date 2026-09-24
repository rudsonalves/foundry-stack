package transport

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestAuthenticatedUserContext(t *testing.T) {
	want := AuthenticatedUser{ID: uuid.New()}
	ctx := withAuthenticatedUser(context.Background(), want)
	got, ok := ctx.Value(authenticatedUserKey{}).(AuthenticatedUser)
	if !ok || got != want {
		t.Fatalf("authenticated user = %#v, %v; want %#v, true", got, ok, want)
	}
	if _, ok := context.Background().Value(authenticatedUserKey{}).(AuthenticatedUser); ok {
		t.Fatal("empty context unexpectedly contains authenticated user")
	}
}
