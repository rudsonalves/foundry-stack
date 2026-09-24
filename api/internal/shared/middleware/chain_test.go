package middleware

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestChainAppliesMiddlewaresInDeclaredOrder(t *testing.T) {
	var calls []string

	middlewareA := recordingMiddleware("a", &calls)
	middlewareB := recordingMiddleware("b", &calls)
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		calls = append(calls, "handler")
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	Chain(handler, middlewareA, middlewareB).ServeHTTP(recorder, request)

	want := []string{"a before", "b before", "handler", "b after", "a after"}
	if !slices.Equal(calls, want) {
		t.Errorf("calls = %v, want %v", calls, want)
	}
}

func TestChainWithoutMiddlewareCallsHandler(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	})

	Chain(handler).ServeHTTP(
		httptest.NewRecorder(),
		httptest.NewRequest(http.MethodGet, "/", nil),
	)

	if !called {
		t.Error("handler was not called")
	}
}

func recordingMiddleware(name string, calls *[]string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			*calls = append(*calls, name+" before")
			next.ServeHTTP(w, r)
			*calls = append(*calls, name+" after")
		})
	}
}
