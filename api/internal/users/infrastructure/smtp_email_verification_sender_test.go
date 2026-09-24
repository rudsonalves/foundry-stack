package infrastructure

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

func TestSMTPEmailVerificationSenderBuildMessage(t *testing.T) {
	sender := NewSMTPEmailVerificationSender(testSMTPConfig())

	input := domain.SendEmailVerificationCodeDTO{
		VerificationID: uuid.New(),
		Email:          "user@example.com",
		Code:           "000042",
		ExpiresAt: time.Date(
			2026,
			time.September,
			11,
			18,
			30,
			0,
			0,
			time.UTC,
		),
	}

	rawMessage, err := sender.buildMessage(input)
	if err != nil {
		t.Fatalf("buildMessage() error = %v", err)
	}

	message, err := mail.ReadMessage(bytes.NewReader(rawMessage))
	if err != nil {
		t.Fatalf("mail.ReadMessage() error = %v", err)
	}

	subject, err := new(mime.WordDecoder).DecodeHeader(
		message.Header.Get("Subject"),
	)
	if err != nil {
		t.Fatalf("decode Subject: %v", err)
	}

	if subject != "Confirme seu e-mail no FoundryStack" {
		t.Errorf("Subject = %q", subject)
	}

	if !strings.Contains(message.Header.Get("From"), "no-reply@example.com") {
		t.Errorf("From = %q", message.Header.Get("From"))
	}

	recipient, err := mail.ParseAddress(message.Header.Get("To"))
	if err != nil {
		t.Fatalf("parse To header: %v", err)
	}

	if recipient.Address != input.Email {
		t.Errorf(
			"To address = %q, want %q",
			recipient.Address,
			input.Email,
		)
	}

	mediaType, parameters, err := mime.ParseMediaType(
		message.Header.Get("Content-Type"),
	)
	if err != nil {
		t.Fatalf("parse Content-Type: %v", err)
	}

	if mediaType != "multipart/alternative" {
		t.Fatalf("Content-Type = %q", mediaType)
	}

	reader := multipart.NewReader(
		message.Body,
		parameters["boundary"],
	)

	parts := make(map[string]string)

	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("NextPart() error = %v", err)
		}

		partType, _, err := mime.ParseMediaType(
			part.Header.Get("Content-Type"),
		)
		if err != nil {
			t.Fatalf("parse part Content-Type: %v", err)
		}

		content, err := io.ReadAll(
			quotedprintable.NewReader(part),
		)
		if err != nil {
			t.Fatalf("read %s part: %v", partType, err)
		}

		parts[partType] = string(content)
	}

	for _, partType := range []string{"text/plain", "text/html"} {
		content, exists := parts[partType]
		if !exists {
			t.Errorf("missing %s part", partType)
			continue
		}

		if !strings.Contains(content, input.Code) {
			t.Errorf("%s part does not contain verification code", partType)
		}

		if !strings.Contains(
			content,
			input.ExpiresAt.Format(time.RFC3339),
		) {
			t.Errorf("%s part does not contain expiration", partType)
		}
	}
}

func TestSMTPEmailVerificationSenderRejectsInvalidRecipient(
	t *testing.T,
) {
	sender := NewSMTPEmailVerificationSender(testSMTPConfig())

	_, err := sender.buildMessage(
		domain.SendEmailVerificationCodeDTO{
			Email: "User <user@example.com>",
			Code:  "123456",
		},
	)

	if err == nil {
		t.Fatal("buildMessage() error = nil")
	}
}

func TestSMTPEmailVerificationSenderRetriesLimitedTimes(
	t *testing.T,
) {
	sender := NewSMTPEmailVerificationSender(testSMTPConfig())

	attempts := 0
	sender.dialContext = func(
		context.Context,
		string,
		string,
	) (net.Conn, error) {
		attempts++
		return nil, errors.New("dial failure")
	}

	input := testEmailVerificationMessage()
	err := sender.SendVerificationCode(context.Background(), input)

	if err == nil {
		t.Fatal("SendVerificationCode() error = nil")
	}

	if attempts != smtpSendMaxAttempts {
		t.Errorf(
			"dial attempts = %d, want %d",
			attempts,
			smtpSendMaxAttempts,
		)
	}

	errorMessage := err.Error()
	if strings.Contains(errorMessage, input.Email) {
		t.Fatal("error exposed recipient email")
	}
	if strings.Contains(errorMessage, input.Code) {
		t.Fatal("error exposed verification code")
	}
}

func TestSMTPEmailVerificationSenderHonorsCanceledContext(
	t *testing.T,
) {
	sender := NewSMTPEmailVerificationSender(testSMTPConfig())

	attempts := 0
	sender.dialContext = func(
		context.Context,
		string,
		string,
	) (net.Conn, error) {
		attempts++
		return nil, errors.New("unexpected dial")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := sender.SendVerificationCode(
		ctx,
		testEmailVerificationMessage(),
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("SendVerificationCode() error = %v, want context.Canceled", err)
	}

	if attempts != 0 {
		t.Errorf("dial attempts = %d, want 0", attempts)
	}
}

func TestSMTPEmailVerificationSenderRequiresSTARTTLS(
	t *testing.T,
) {
	sender := NewSMTPEmailVerificationSender(testSMTPConfig())

	attempts := 0
	sender.dialContext = func(
		context.Context,
		string,
		string,
	) (net.Conn, error) {
		attempts++

		clientConnection, serverConnection := net.Pipe()

		go serveSMTPWithoutSTARTTLS(serverConnection)

		return clientConnection, nil
	}

	err := sender.SendVerificationCode(
		context.Background(),
		testEmailVerificationMessage(),
	)

	if err == nil {
		t.Fatal("SendVerificationCode() error = nil")
	}

	if attempts != smtpSendMaxAttempts {
		t.Errorf(
			"dial attempts = %d, want %d",
			attempts,
			smtpSendMaxAttempts,
		)
	}

	if !strings.Contains(err.Error(), "STARTTLS") {
		t.Fatalf("error = %q, want STARTTLS failure", err)
	}
}

func TestSMTPEmailVerificationSenderAppliesTimeout(
	t *testing.T,
) {
	config := testSMTPConfig()
	config.Timeout = 20 * time.Millisecond

	sender := NewSMTPEmailVerificationSender(config)

	sender.dialContext = func(
		context.Context,
		string,
		string,
	) (net.Conn, error) {
		clientConnection, serverConnection := net.Pipe()

		go func() {
			defer serverConnection.Close()
			_, _ = io.Copy(io.Discard, serverConnection)
		}()

		return clientConnection, nil
	}

	startedAt := time.Now()

	err := sender.SendVerificationCode(
		context.Background(),
		testEmailVerificationMessage(),
	)

	if err == nil {
		t.Fatal("SendVerificationCode() error = nil")
	}

	maximumExpectedDuration :=
		time.Duration(smtpSendMaxAttempts+1) * config.Timeout

	if elapsed := time.Since(startedAt); elapsed > maximumExpectedDuration {
		t.Errorf(
			"SendVerificationCode() took %v, want at most %v",
			elapsed,
			maximumExpectedDuration,
		)
	}
}

func testSMTPConfig() SMTPEmailVerificationSenderConfig {
	return SMTPEmailVerificationSenderConfig{
		Host:        "smtp.example.com",
		Port:        587,
		Username:    "foundry-stack",
		Password:    "smtp-password",
		FromAddress: "no-reply@example.com",
		FromName:    "FoundryStack",
		TLSMode:     smtpTLSModeSTARTTLS,
		Timeout:     time.Second,
	}
}

func testEmailVerificationMessage() domain.SendEmailVerificationCodeDTO {
	return domain.SendEmailVerificationCodeDTO{
		VerificationID: uuid.New(),
		Email:          "user@example.com",
		Code:           "123456",
		ExpiresAt:      time.Now().UTC().Add(15 * time.Minute),
	}
}

func serveSMTPWithoutSTARTTLS(connection net.Conn) {
	defer connection.Close()

	reader := bufio.NewReader(connection)

	_, _ = fmt.Fprint(
		connection,
		"220 smtp.example.com ESMTP ready\r\n",
	)

	_, err := reader.ReadString('\n')
	if err != nil {
		return
	}

	_, _ = fmt.Fprint(
		connection,
		"250 smtp.example.com\r\n",
	)
}
