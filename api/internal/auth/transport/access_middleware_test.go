package transport

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
)

type stubTokenParser struct {
	userID uuid.UUID
	err    error
	calls  int
	token  string
}

func (s *stubTokenParser) Parse(token string) (uuid.UUID, error) {
	s.calls++
	s.token = token
	return s.userID, s.err
}

func TestAuthenticationRequireUser(t *testing.T) {
	userID := uuid.MustParse("30787324-9cb8-4281-a502-c19a006c12cf")
	parserFailure := errors.New("invalid token")
	tests := []struct {
		name       string
		headers    []string
		parser     *stubTokenParser
		wantNext   bool
		wantCalls  int
		wantToken  string
		wantAuth   string
		wantAppErr bool
	}{
		{name: "missing bearer", parser: &stubTokenParser{}, wantAuth: "Bearer", wantAppErr: true},
		{name: "invalid scheme", headers: []string{"Basic abc"}, parser: &stubTokenParser{}, wantAuth: "Bearer", wantAppErr: true},
		{name: "empty token", headers: []string{"Bearer"}, parser: &stubTokenParser{}, wantAuth: "Bearer", wantAppErr: true},
		{name: "repeated header", headers: []string{"Bearer first", "Bearer second"}, parser: &stubTokenParser{}, wantAuth: "Bearer", wantAppErr: true},
		{name: "parser rejects", headers: []string{"Bearer rejected"}, parser: &stubTokenParser{err: parserFailure}, wantCalls: 1, wantToken: "rejected", wantAuth: "Bearer", wantAppErr: true},
		{name: "parser accepts", headers: []string{"bearer accepted"}, parser: &stubTokenParser{userID: userID}, wantCalls: 1, wantToken: "accepted", wantNext: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := func(w http.ResponseWriter, r *http.Request) error {
				nextCalled = true
				got, ok := CurrentUser(r.Context())
				if !ok || got.ID != userID {
					t.Fatalf("CurrentUser() = %#v, %v; want ID %s", got, ok, userID)
				}
				w.WriteHeader(http.StatusNoContent)
				return nil
			}
			request := httptest.NewRequest(http.MethodGet, "/private", nil)
			for _, value := range tt.headers {
				request.Header.Add("Authorization", value)
			}
			recorder := httptest.NewRecorder()

			err := NewAuthentication(tt.parser).RequireUser(next)(recorder, request)

			if nextCalled != tt.wantNext {
				t.Fatalf("next called = %v, want %v", nextCalled, tt.wantNext)
			}
			if tt.parser.calls != tt.wantCalls || tt.parser.token != tt.wantToken {
				t.Fatalf("Parse calls/token = %d/%q, want %d/%q", tt.parser.calls, tt.parser.token, tt.wantCalls, tt.wantToken)
			}
			if got := recorder.Header().Get("WWW-Authenticate"); got != tt.wantAuth {
				t.Fatalf("WWW-Authenticate = %q, want %q", got, tt.wantAuth)
			}
			var appErr *sharederrors.AppError
			if got := errors.As(err, &appErr); got != tt.wantAppErr {
				t.Fatalf("errors.As(AppError) = %v, want %v (err=%v)", got, tt.wantAppErr, err)
			}
			if appErr != nil && appErr.Code != sharederrors.ErrCodeUnauthorized {
				t.Fatalf("error code = %q, want %q", appErr.Code, sharederrors.ErrCodeUnauthorized)
			}
		})
	}
}

func TestRequireCurrentUser(t *testing.T) {
	want := AuthenticatedUser{ID: uuid.MustParse("73b9485c-fcd2-48bd-a736-8825404c0e7d")}

	if _, err := RequireCurrentUser(context.Background()); err == nil {
		t.Fatal("RequireCurrentUser() error = nil without identity")
	}

	got, err := RequireCurrentUser(withAuthenticatedUser(context.Background(), want))
	if err != nil {
		t.Fatalf("RequireCurrentUser() error = %v", err)
	}
	if got != want {
		t.Fatalf("RequireCurrentUser() = %#v, want %#v", got, want)
	}
}
