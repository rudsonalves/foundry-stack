package httpresponse

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteError(t *testing.T) {
	recorder := httptest.NewRecorder()
	details := []FieldError{
		{Field: "name", Message: "é obrigatório"},
	}

	WriteError(
		recorder,
		http.StatusBadRequest,
		"VALIDATION_ERROR",
		"Dados inválidos.",
		"request-123",
		details,
	)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}

	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want %q", got, "application/json")
	}

	var got ErrorResponse
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if got.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("error.code = %q, want %q", got.Error.Code, "VALIDATION_ERROR")
	}
	if got.Error.Message != "Dados inválidos." {
		t.Errorf("error.message = %q, want %q", got.Error.Message, "Dados inválidos.")
	}
	if got.Error.RequestID != "request-123" {
		t.Errorf("error.request_id = %q, want %q", got.Error.RequestID, "request-123")
	}
	if len(got.Error.Details) != 1 || got.Error.Details[0] != details[0] {
		t.Errorf("error.details = %#v, want %#v", got.Error.Details, details)
	}
}

type failingResponseWriter struct {
	header http.Header
	status int
	err    error
}

func (w *failingResponseWriter) Header() http.Header {
	return w.header
}

func (w *failingResponseWriter) WriteHeader(status int) {
	w.status = status
}

func (w *failingResponseWriter) Write([]byte) (int, error) {
	return 0, w.err
}

func TestWriteErrorHandlesResponseWriteFailure(t *testing.T) {
	wantErr := errors.New("client disconnected")
	writer := &failingResponseWriter{
		header: make(http.Header),
		err:    wantErr,
	}

	WriteError(
		writer,
		http.StatusInternalServerError,
		"INTERNAL_ERROR",
		"Erro interno do servidor.",
		"request-123",
		nil,
	)

	if writer.status != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", writer.status, http.StatusInternalServerError)
	}
	if got := writer.header.Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
}

func TestWriteErrorOmitsEmptyOptionalFields(t *testing.T) {
	recorder := httptest.NewRecorder()

	WriteError(
		recorder,
		http.StatusInternalServerError,
		"INTERNAL_ERROR",
		"Erro interno do servidor.",
		"",
		nil,
	)

	var got map[string]map[string]json.RawMessage
	if err := json.NewDecoder(recorder.Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	errorBody, ok := got["error"]
	if !ok {
		t.Fatal("response does not contain error envelope")
	}
	if _, ok := errorBody["request_id"]; ok {
		t.Error("error.request_id should be omitted when empty")
	}
	if _, ok := errorBody["details"]; ok {
		t.Error("error.details should be omitted when empty")
	}
}

func TestWriteData(t *testing.T) {
	recorder := httptest.NewRecorder()
	want := map[string]any{"id": "item-123", "quantity": float64(2)}
	if err := WriteData(recorder, http.StatusCreated, want); err != nil {
		t.Fatalf("WriteData() unexpected error: %v", err)
	}
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if got := recorder.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want application/json", got)
	}
	var body DataResponse
	if err := json.NewDecoder(recorder.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	got := body.Data.(map[string]any)
	if got["id"] != want["id"] || got["quantity"] != want["quantity"] {
		t.Fatalf("data = %#v, want %#v", got, want)
	}
}

func TestWriteDataReturnsMarshalErrorBeforeWritingHeaders(t *testing.T) {
	recorder := httptest.NewRecorder()
	err := WriteData(recorder, http.StatusOK, func() {})
	if err == nil || !errors.Is(err, &json.UnsupportedTypeError{}) {
		// UnsupportedTypeError is not comparable through errors.Is without the same value.
		if err == nil || !strings.Contains(err.Error(), "marshal response") {
			t.Fatalf("WriteData() error = %v, want marshal response error", err)
		}
	}
	if recorder.Header().Get("Content-Type") != "" || recorder.Body.Len() != 0 {
		t.Fatal("response was written after marshal failure")
	}
}

func TestWriteDataHandlesResponseWriteFailure(t *testing.T) {
	writer := &failingResponseWriter{header: make(http.Header), err: errors.New("client disconnected")}
	if err := WriteData(writer, http.StatusAccepted, "value"); err != nil {
		t.Fatalf("WriteData() unexpected error: %v", err)
	}
	if writer.status != http.StatusAccepted || writer.header.Get("Content-Type") != "application/json" {
		t.Fatalf("status/content type = %d/%q", writer.status, writer.header.Get("Content-Type"))
	}
}
