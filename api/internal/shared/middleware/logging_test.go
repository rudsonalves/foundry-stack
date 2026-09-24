package middleware

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStatusResponseWriterWriteHeaderUsesFirstStatus(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := &statusResponseWriter{
		ResponseWriter: recorder,
		status:         http.StatusOK,
	}

	writer.WriteHeader(http.StatusCreated)
	writer.WriteHeader(http.StatusInternalServerError)

	if writer.status != http.StatusCreated {
		t.Errorf("tracked status = %d, want %d", writer.status, http.StatusCreated)
	}
	if !writer.wroteHeader {
		t.Error("wroteHeader = false, want true")
	}
	if recorder.Code != http.StatusCreated {
		t.Errorf("response status = %d, want %d", recorder.Code, http.StatusCreated)
	}
}

func TestStatusResponseWriterWriteImplicitlyUsesOK(t *testing.T) {
	recorder := httptest.NewRecorder()
	writer := &statusResponseWriter{
		ResponseWriter: recorder,
		status:         http.StatusOK,
	}

	written, err := writer.Write([]byte("response body"))
	if err != nil {
		t.Fatalf("Write() unexpected error: %v", err)
	}

	if written != len("response body") {
		t.Errorf("written bytes = %d, want %d", written, len("response body"))
	}
	if writer.status != http.StatusOK {
		t.Errorf("tracked status = %d, want %d", writer.status, http.StatusOK)
	}
	if !writer.wroteHeader {
		t.Error("wroteHeader = false, want true")
	}
	if recorder.Code != http.StatusOK {
		t.Errorf("response status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if recorder.Body.String() != "response body" {
		t.Errorf("response body = %q, want %q", recorder.Body.String(), "response body")
	}
}

func TestLoggingRecordsRequestDataAndStatus(t *testing.T) {
	var output bytes.Buffer
	restoreLog := captureDefaultLogger(&output)
	defer restoreLog()

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
	request := httptest.NewRequest(http.MethodPatch, "/items/item-123", nil)
	request.RemoteAddr = "203.0.113.10:4321"
	request = request.WithContext(context.WithValue(
		request.Context(),
		requestIDKey,
		"request-123",
	))
	recorder := httptest.NewRecorder()

	Logging(next).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusAccepted)
	}

	got := output.String()
	wantParts := []string{
		"request_id=request-123",
		"method=PATCH",
		"path=/items/item-123",
		"status=202",
		"duration=",
		"remote_addr=203.0.113.10:4321",
	}
	for _, part := range wantParts {
		if !strings.Contains(got, part) {
			t.Errorf("log %q does not contain %q", got, part)
		}
	}
}

func TestLoggingUsesOKWhenHandlerDoesNotWrite(t *testing.T) {
	var output bytes.Buffer
	restoreLog := captureDefaultLogger(&output)
	defer restoreLog()

	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	Logging(next).ServeHTTP(recorder, request)

	if !strings.Contains(output.String(), "status=200") {
		t.Errorf("log %q does not contain status=200", output.String())
	}
}

func captureDefaultLogger(output *bytes.Buffer) func() {
	previousOutput := log.Writer()
	previousFlags := log.Flags()
	previousPrefix := log.Prefix()

	log.SetOutput(output)
	log.SetFlags(0)
	log.SetPrefix("")

	return func() {
		log.SetOutput(previousOutput)
		log.SetFlags(previousFlags)
		log.SetPrefix(previousPrefix)
	}
}
