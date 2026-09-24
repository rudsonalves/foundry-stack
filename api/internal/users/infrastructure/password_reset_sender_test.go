package infrastructure

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

func TestSMTPPasswordResetSenderBuildsSafeMultipartMessage(t *testing.T) {
	sender := NewSMTPPasswordResetSender(SMTPEmailVerificationSenderConfig{FromAddress: "no-reply@example.test", FromName: "FoundryStack"})
	message, err := sender.buildMessage(domain.SendPasswordResetCodeDTO{Email: "user@example.com", Code: "012345", ExpiresAt: time.Date(2026, 9, 23, 12, 15, 0, 0, time.UTC)})
	if err != nil {
		t.Fatal(err)
	}
	content := string(message)
	for _, expected := range []string{"Recupere sua senha do FoundryStack", "012345", "2026-09-23T12:15:00Z", "ignore esta mensagem", `text/plain`, `text/html`} {
		if !strings.Contains(content, expected) {
			t.Errorf("message does not contain %q", expected)
		}
	}
	if strings.Contains(content, "http://") || strings.Contains(content, "https://") || strings.Contains(content, "<a ") {
		t.Fatal("password reset message must not contain links")
	}
}

func TestMemoryPasswordResetSenderStoresMessage(t *testing.T) {
	sender := NewMemoryPasswordResetSender()
	input := domain.SendPasswordResetCodeDTO{Email: "user@example.com", Code: "123456"}
	if err := sender.SendPasswordResetCode(context.Background(), input); err != nil {
		t.Fatal(err)
	}
	messages := sender.Messages()
	if len(messages) != 1 || messages[0].Code != "123456" {
		t.Fatalf("messages=%#v", messages)
	}
}
