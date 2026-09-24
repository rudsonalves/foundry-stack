package middleware

import (
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewRequestID(t *testing.T) {
	requestID, err := newRequestID()
	if err != nil {
		t.Fatalf("newRequestID() unexpected error: %v", err)
	}

	if len(requestID) != 32 {
		t.Errorf("request ID length = %d, want 32", len(requestID))
	}
	if _, err := hex.DecodeString(requestID); err != nil {
		t.Errorf("request ID %q is not hexadecimal: %v", requestID, err)
	}
}

func TestNewRequestIDReturnsRandomSourceError(t *testing.T) {
	wantErr := errors.New("random source unavailable")
	previous := readRandom
	readRandom = func([]byte) (int, error) {
		return 0, wantErr
	}
	t.Cleanup(func() { readRandom = previous })

	requestID, err := newRequestID()
	if requestID != "" {
		t.Errorf("newRequestID() ID = %q, want empty", requestID)
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("newRequestID() error = %v, want %v", err, wantErr)
	}
}

func TestGetRequestIDWithoutContextValue(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	if got := GetRequestID(request); got != "" {
		t.Errorf("GetRequestID() = %q, want empty string", got)
	}
}

func TestRequestIDPreservesProvidedID(t *testing.T) {
	const providedID = "request-provided"

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		if got := GetRequestID(r); got != providedID {
			t.Errorf("GetRequestID() = %q, want %q", got, providedID)
		}
		w.WriteHeader(http.StatusNoContent)
	})
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-Request-ID", providedID)
	recorder := httptest.NewRecorder()

	RequestID(next).ServeHTTP(recorder, request)

	if !nextCalled {
		t.Error("next handler was not called")
	}
	if got := recorder.Header().Get("X-Request-ID"); got != providedID {
		t.Errorf("response X-Request-ID = %q, want %q", got, providedID)
	}
	if recorder.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusNoContent)
	}
}

func TestRequestIDGeneratesMissingID(t *testing.T) {
	var contextID string
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextID = GetRequestID(r)
		w.WriteHeader(http.StatusNoContent)
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	RequestID(next).ServeHTTP(recorder, request)

	responseID := recorder.Header().Get("X-Request-ID")
	if responseID == "" {
		t.Fatal("response X-Request-ID is empty")
	}
	if contextID != responseID {
		t.Errorf("context request ID = %q, response request ID = %q", contextID, responseID)
	}
	if len(responseID) != 32 {
		t.Errorf("generated request ID length = %d, want 32", len(responseID))
	}
	if _, err := hex.DecodeString(responseID); err != nil {
		t.Errorf("generated request ID %q is not hexadecimal: %v", responseID, err)
	}
}

func TestRequestIDHandlesGenerationFailure(t *testing.T) {
	previous := readRandom
	readRandom = func([]byte) (int, error) {
		return 0, io.ErrUnexpectedEOF
	}
	t.Cleanup(func() { readRandom = previous })

	nextCalled := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)

	RequestID(next).ServeHTTP(recorder, request)

	if nextCalled {
		t.Error("next handler was called after request ID generation failure")
	}
	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if got := recorder.Header().Get("X-Request-ID"); got != "" {
		t.Errorf("X-Request-ID = %q, want empty", got)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
}
