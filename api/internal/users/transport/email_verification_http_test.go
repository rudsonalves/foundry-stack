package transport

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	sharederrors "github.com/rudsonalves/foundry-stack/api/internal/shared/errors"
	httphandler "github.com/rudsonalves/foundry-stack/api/internal/shared/http/handler"
	"github.com/rudsonalves/foundry-stack/api/internal/users/application"
)

func TestEmailVerificationHandlerRequest(t *testing.T) {
	verificationID := uuid.MustParse("7fd1ec52-6f58-4fe2-b482-e31c06adfba5")
	codeExpiresAt := time.Date(2026, time.September, 11, 15, 15, 0, 0, time.UTC)
	resendAvailableAt := time.Date(2026, time.September, 11, 15, 1, 0, 0, time.UTC)
	requester := &emailVerificationRequesterStub{
		result: application.RequestEmailVerificationResult{
			VerificationID:    verificationID,
			CodeExpiresAt:     codeExpiresAt,
			ResendAvailableAt: resendAvailableAt,
		},
	}
	handler := NewEmailVerificationHandler(requester)
	request := httptest.NewRequest(
		http.MethodPost,
		"/email-verifications",
		bytes.NewBufferString(`{"email":"user@example.com"}`),
	)
	request.RemoteAddr = "203.0.113.10:4321"
	recorder := httptest.NewRecorder()

	if err := handler.Request(recorder, request); err != nil {
		t.Fatalf("Request() error = %v", err)
	}

	if recorder.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusAccepted)
	}
	if requester.input.Email != "user@example.com" {
		t.Errorf("application email = %q", requester.input.Email)
	}
	if requester.input.ClientIP != "203.0.113.10" {
		t.Errorf("application client IP = %q", requester.input.ClientIP)
	}

	body := recorder.Body.Bytes()
	if strings.Contains(string(body), "123456") {
		t.Fatal("response body exposed verification code")
	}

	var response struct {
		Data map[string]any `json:"data"`
	}
	if err := json.NewDecoder(bytes.NewReader(body)).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Data["verification_id"] != verificationID.String() {
		t.Errorf("verification_id = %v", response.Data["verification_id"])
	}
	if _, exists := response.Data["code"]; exists {
		t.Fatal("response exposed verification code")
	}
}

func TestEmailVerificationHandlerRequestRejectsMalformedBody(t *testing.T) {
	handler := NewEmailVerificationHandler(&emailVerificationRequesterStub{})
	request := httptest.NewRequest(
		http.MethodPost,
		"/email-verifications",
		bytes.NewBufferString(`{"email":`),
	)
	recorder := httptest.NewRecorder()

	err := handler.Request(recorder, request)
	if err == nil {
		t.Fatal("Request() error = nil")
	}
}

func TestEmailVerificationHandlerRequestRejectsOversizedBody(t *testing.T) {
	requester := &emailVerificationRequesterStub{}
	handler := NewEmailVerificationHandler(requester)
	request := httptest.NewRequest(
		http.MethodPost,
		"/email-verifications",
		strings.NewReader(
			`{"email":"user@example.com","padding":"`+
				strings.Repeat("a", int(maxRequestBodyBytes))+
				`"}`,
		),
	)
	recorder := httptest.NewRecorder()

	httphandler.Handle(handler.Request).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
	if requester.input.Email != "" {
		t.Fatal("application should not be called for an oversized body")
	}
}

func TestEmailVerificationHandlerRequestReturnsRateLimitEnvelope(t *testing.T) {
	requester := &emailVerificationRequesterStub{
		err: &sharederrors.AppError{
			Code:    sharederrors.ErrCodeRateLimitExceeded,
			Message: "too many email verification requests",
		},
	}
	handler := NewEmailVerificationHandler(requester)
	request := httptest.NewRequest(
		http.MethodPost,
		"/email-verifications",
		bytes.NewBufferString(`{"email":"user@example.com"}`),
	)
	recorder := httptest.NewRecorder()

	httphandler.Handle(handler.Request).ServeHTTP(recorder, request)

	if recorder.Code != http.StatusTooManyRequests {
		t.Fatalf(
			"status = %d, want %d",
			recorder.Code,
			http.StatusTooManyRequests,
		)
	}

	var response struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.Error.Code != sharederrors.ErrCodeRateLimitExceeded {
		t.Fatalf(
			"error.code = %q, want %q",
			response.Error.Code,
			sharederrors.ErrCodeRateLimitExceeded,
		)
	}
}

func TestEmailVerificationHandlerConfirm(t *testing.T) {
	verificationID := uuid.New()
	expiresAt := time.Date(
		2026,
		time.September,
		12,
		12,
		0,
		0,
		0,
		time.UTC,
	)
	requester := &emailVerificationRequesterStub{
		confirmResult: application.ConfirmEmailVerificationResult{
			EmailVerificationToken: "opaque-token",
			ExpiresAt:              expiresAt,
		},
	}
	handler := NewEmailVerificationHandler(requester)
	request := httptest.NewRequest(
		http.MethodPost,
		"/email-verifications/confirm",
		bytes.NewBufferString(
			`{"verification_id":"`+
				verificationID.String()+
				`","code":"000042"}`,
		),
	)
	recorder := httptest.NewRecorder()

	if err := handler.Confirm(recorder, request); err != nil {
		t.Fatalf("Confirm() error = %v", err)
	}

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	if requester.confirmInput.VerificationID != verificationID {
		t.Errorf(
			"application verification ID = %s",
			requester.confirmInput.VerificationID,
		)
	}

	if requester.confirmInput.Code != "000042" {
		t.Errorf(
			"application code = %q",
			requester.confirmInput.Code,
		)
	}

	var response struct {
		Data ConfirmEmailVerificationResponse `json:"data"`
	}
	if err := json.NewDecoder(recorder.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if response.Data.EmailVerificationToken != "opaque-token" {
		t.Errorf(
			"email_verification_token = %q",
			response.Data.EmailVerificationToken,
		)
	}

	if response.Data.ExpiresAt != expiresAt {
		t.Errorf("expires_at = %v", response.Data.ExpiresAt)
	}
}

func TestEmailVerificationHandlerConfirmRejectsInvalidID(
	t *testing.T,
) {
	requester := &emailVerificationRequesterStub{}
	handler := NewEmailVerificationHandler(requester)
	request := httptest.NewRequest(
		http.MethodPost,
		"/email-verifications/confirm",
		bytes.NewBufferString(
			`{"verification_id":"invalid","code":"123456"}`,
		),
	)
	recorder := httptest.NewRecorder()

	err := handler.Confirm(recorder, request)
	if err == nil {
		t.Fatal("Confirm() error = nil")
	}

	if requester.confirmInput.VerificationID != uuid.Nil {
		t.Fatal("application called for invalid verification ID")
	}
}

type emailVerificationRequesterStub struct {
	input         application.RequestEmailVerificationInput
	result        application.RequestEmailVerificationResult
	err           error
	confirmInput  application.ConfirmEmailVerificationInput
	confirmResult application.ConfirmEmailVerificationResult
	confirmErr    error
}

func (s *emailVerificationRequesterStub) Request(
	_ context.Context,
	input application.RequestEmailVerificationInput,
) (application.RequestEmailVerificationResult, error) {
	s.input = input
	return s.result, s.err
}

func (s *emailVerificationRequesterStub) Confirm(
	_ context.Context,
	input application.ConfirmEmailVerificationInput,
) (application.ConfirmEmailVerificationResult, error) {
	s.confirmInput = input
	return s.confirmResult, s.confirmErr
}
