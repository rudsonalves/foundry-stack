package infrastructure

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

func TestMemoryEmailVerificationSender(t *testing.T) {
	sender := NewMemoryEmailVerificationSender()

	input := domain.SendEmailVerificationCodeDTO{
		VerificationID: uuid.New(),
		Email:          "user@example.com",
		Code:           "000042",
		ExpiresAt:      time.Now().UTC().Add(15 * time.Minute),
	}

	if err := sender.SendVerificationCode(
		context.Background(),
		input,
	); err != nil {
		t.Fatalf("SendVerificationCode() error = %v", err)
	}

	messages := sender.Messages()
	if len(messages) != 1 {
		t.Fatalf("Messages() length = %d, want 1", len(messages))
	}

	if messages[0] != input {
		t.Fatalf("Messages()[0] = %#v, want %#v", messages[0], input)
	}

	messages[0].Code = "changed"

	if sender.Messages()[0].Code != input.Code {
		t.Fatal("Messages() exposed the internal slice")
	}
}

func TestMemoryEmailVerificationSenderHonorsCanceledContext(
	t *testing.T,
) {
	sender := NewMemoryEmailVerificationSender()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sender.SendVerificationCode(
		ctx,
		domain.SendEmailVerificationCodeDTO{},
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("SendVerificationCode() error = %v, want context.Canceled", err)
	}

	if len(sender.Messages()) != 0 {
		t.Fatal("canceled send stored a message")
	}
}
