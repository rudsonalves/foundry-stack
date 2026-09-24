package infrastructure

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

func TestLoggingEmailVerificationSender(t *testing.T) {
	buffer := &bytes.Buffer{}
	next := &loggingEmailVerificationSenderStub{}
	sender := NewLoggingEmailVerificationSender(
		next,
		log.New(buffer, "", 0),
	)
	input := domain.SendEmailVerificationCodeDTO{
		VerificationID: uuid.MustParse(
			"9764ef1d-ae2e-4ba2-a92a-d81a7d4894aa",
		),
		Email:     "user@example.com",
		Code:      "000042",
		ExpiresAt: time.Date(2026, time.September, 12, 15, 0, 0, 0, time.UTC),
	}

	if err := sender.SendVerificationCode(context.Background(), input); err != nil {
		t.Fatalf("SendVerificationCode() error = %v", err)
	}

	if next.input != input {
		t.Fatalf("delegated input = %#v, want %#v", next.input, input)
	}
	output := buffer.String()
	for _, expected := range []string{
		input.VerificationID.String(),
		"code=000042",
		"expires_at=2026-09-12T15:00:00Z",
	} {
		if !strings.Contains(output, expected) {
			t.Fatalf("log = %q, want to contain %q", output, expected)
		}
	}
	if strings.Contains(output, input.Email) {
		t.Fatalf("log exposed email address: %q", output)
	}
}

func TestLoggingEmailVerificationSenderPreservesDeliveryError(t *testing.T) {
	deliveryErr := errors.New("delivery failed")
	next := &loggingEmailVerificationSenderStub{err: deliveryErr}
	sender := NewLoggingEmailVerificationSender(
		next,
		log.New(&bytes.Buffer{}, "", 0),
	)

	err := sender.SendVerificationCode(
		context.Background(),
		domain.SendEmailVerificationCodeDTO{Code: "123456"},
	)
	if !errors.Is(err, deliveryErr) {
		t.Fatalf("error = %v, want %v", err, deliveryErr)
	}
}

type loggingEmailVerificationSenderStub struct {
	input domain.SendEmailVerificationCodeDTO
	err   error
}

func (s *loggingEmailVerificationSenderStub) SendVerificationCode(
	_ context.Context,
	input domain.SendEmailVerificationCodeDTO,
) error {
	s.input = input
	return s.err
}
