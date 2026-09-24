package transport

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rudsonalves/foundry-stack/api/internal/users/application"
)

type passwordResetRequesterStub struct {
	request  application.RequestPasswordResetInput
	confirm  application.ConfirmPasswordResetInput
	complete application.CompletePasswordResetInput
}

func (s *passwordResetRequesterStub) Request(_ context.Context, in application.RequestPasswordResetInput) (application.RequestPasswordResetResult, error) {
	s.request = in
	return application.RequestPasswordResetResult{PasswordResetID: uuid.New(), CodeExpiresAt: time.Now(), ResendAvailableAt: time.Now()}, nil
}
func (s *passwordResetRequesterStub) Confirm(_ context.Context, in application.ConfirmPasswordResetInput) (application.ConfirmPasswordResetResult, error) {
	s.confirm = in
	return application.ConfirmPasswordResetResult{PasswordResetToken: "secret-token", ExpiresAt: time.Now()}, nil
}
func (s *passwordResetRequesterStub) Complete(_ context.Context, in application.CompletePasswordResetInput) error {
	s.complete = in
	return nil
}

func TestPasswordResetHandlerRequestUsesTrustedProxyAndHidesCode(t *testing.T) {
	resolver, _ := NewClientIPResolver([]string{"10.0.0.0/8"})
	service := &passwordResetRequesterStub{}
	handler := NewPasswordResetHandler(service, resolver)
	req := httptest.NewRequest(http.MethodPost, "/password-resets", bytes.NewBufferString(`{"email":"user@example.com"}`))
	req.RemoteAddr = "10.0.0.2:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.9")
	rec := httptest.NewRecorder()
	if err := handler.Request(rec, req); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusAccepted || service.request.ClientIP != "203.0.113.9" || strings.Contains(rec.Body.String(), `"code"`) {
		t.Fatalf("status=%d ip=%q body=%s", rec.Code, service.request.ClientIP, rec.Body.String())
	}
}
func TestPasswordResetHandlerConfirmReturnsToken(t *testing.T) {
	service := &passwordResetRequesterStub{}
	handler := NewPasswordResetHandler(service, nil)
	id := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/password-resets/confirm", bytes.NewBufferString(`{"password_reset_id":"`+id.String()+`","code":"123456"}`))
	rec := httptest.NewRecorder()
	if err := handler.Confirm(rec, req); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusOK || service.confirm.PasswordResetID != id || !strings.Contains(rec.Body.String(), "secret-token") {
		t.Fatalf("status=%d input=%#v body=%s", rec.Code, service.confirm, rec.Body.String())
	}
}

func TestPasswordResetHandlerCompleteReturnsNoContent(t *testing.T) {
	service := &passwordResetRequesterStub{}
	handler := NewPasswordResetHandler(service, nil)
	id := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/password-resets/complete", bytes.NewBufferString(`{"password_reset_id":"`+id.String()+`","password_reset_token":"proof","new_password":"new-password"}`))
	rec := httptest.NewRecorder()
	if err := handler.Complete(rec, req); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 || service.complete.NewPassword != "new-password" {
		t.Fatalf("status=%d body=%q input=%#v", rec.Code, rec.Body.String(), service.complete)
	}
}
