package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	httpresponse "github.com/rudsonalves/foundry-stack/api/internal/shared/http/response"
)

func TestRecoverPassesThroughWithoutPanic(t *testing.T) {
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthy", nil)

	Recover(next).ServeHTTP(recorder, request)

	if !nextCalled {
		t.Error("next handler was not called")
	}
	if recorder.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestRecoverConvertsPanicToInternalError(t *testing.T) {
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("simulated failure")
	})
	handler := RequestID(Recover(next))
	request := httptest.NewRequest(http.MethodGet, "/panic", nil)
	request.Header.Set("X-Request-ID", "request-123")
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}

	var response httpresponse.ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Error.Code != sharederrors.ErrCodeInternalError {
		t.Errorf("error.code = %q, want %q", response.Error.Code, sharederrors.ErrCodeInternalError)
	}
	if response.Error.Message != "Erro inesperado do servidor." {
		t.Errorf("error.message = %q, want %q", response.Error.Message, "Erro inesperado do servidor.")
	}
	if response.Error.RequestID != "request-123" {
		t.Errorf("error.request_id = %q, want %q", response.Error.RequestID, "request-123")
	}
}
