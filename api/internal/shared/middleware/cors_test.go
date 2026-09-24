package middleware

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
)

func TestCORSForRegularRequests(t *testing.T) {
	const allowedOrigin = "https://app.example.com"
	tests := []struct {
		name            string
		origin          string
		wantAllowOrigin string
		wantVary        bool
	}{
		{
			name:            "allows configured origin",
			origin:          allowedOrigin,
			wantAllowOrigin: allowedOrigin,
			wantVary:        true,
		},
		{
			name:     "does not allow different origin",
			origin:   "https://untrusted.example.com",
			wantVary: true,
		},
		{
			name: "request without origin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusAccepted)
			})
			request := httptest.NewRequest(http.MethodGet, "/resource", nil)
			if tt.origin != "" {
				request.Header.Set("Origin", tt.origin)
			}
			recorder := httptest.NewRecorder()

			CORS(allowedOrigin)(next).ServeHTTP(recorder, request)

			if !nextCalled {
				t.Error("next handler was not called for regular request")
			}
			if recorder.Code != http.StatusAccepted {
				t.Errorf("status = %d, want %d", recorder.Code, http.StatusAccepted)
			}
			if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != tt.wantAllowOrigin {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantAllowOrigin)
			}
			if got := slices.Contains(recorder.Header().Values("Vary"), "Origin"); got != tt.wantVary {
				t.Errorf("Vary contains Origin = %v, want %v; values=%v", got, tt.wantVary, recorder.Header().Values("Vary"))
			}
		})
	}
}

func TestCORSPreflight(t *testing.T) {
	const allowedOrigin = "https://app.example.com"
	tests := []struct {
		name            string
		origin          string
		wantAllowOrigin string
	}{
		{name: "allowed origin", origin: allowedOrigin, wantAllowOrigin: allowedOrigin},
		{name: "rejected origin", origin: "https://untrusted.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
				nextCalled = true
			})
			request := httptest.NewRequest(http.MethodOptions, "/resource", nil)
			request.Header.Set("Origin", tt.origin)
			recorder := httptest.NewRecorder()

			CORS(allowedOrigin)(next).ServeHTTP(recorder, request)

			if nextCalled {
				t.Error("next handler was called for preflight request")
			}
			if recorder.Code != http.StatusNoContent {
				t.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
			}
			if got := recorder.Header().Get("Access-Control-Allow-Origin"); got != tt.wantAllowOrigin {
				t.Errorf("Access-Control-Allow-Origin = %q, want %q", got, tt.wantAllowOrigin)
			}
			if got := recorder.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, PUT, DELETE, PATCH, OPTIONS" {
				t.Errorf("Access-Control-Allow-Methods = %q", got)
			}
			if got := recorder.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type, Authorization, X-Request-ID" {
				t.Errorf("Access-Control-Allow-Headers = %q", got)
			}
			if !slices.Contains(recorder.Header().Values("Vary"), "Origin") {
				t.Errorf("Vary does not contain Origin; values=%v", recorder.Header().Values("Vary"))
			}
		})
	}
}
