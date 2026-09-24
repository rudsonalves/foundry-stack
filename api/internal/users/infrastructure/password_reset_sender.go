package infrastructure

import (
	"bytes"
	"context"
	"fmt"
	"html"
	"mime"
	"mime/multipart"
	"net/mail"
	"sync"
	"time"

	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

type MemoryPasswordResetSender struct {
	mu       sync.Mutex
	messages []domain.SendPasswordResetCodeDTO
}

func NewMemoryPasswordResetSender() *MemoryPasswordResetSender { return &MemoryPasswordResetSender{} }
func (s *MemoryPasswordResetSender) SendPasswordResetCode(ctx context.Context, input domain.SendPasswordResetCodeDTO) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.messages = append(s.messages, input)
	return nil
}
func (s *MemoryPasswordResetSender) Messages() []domain.SendPasswordResetCodeDTO {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]domain.SendPasswordResetCodeDTO(nil), s.messages...)
}

type SMTPPasswordResetSender struct{ smtp *SMTPEmailVerificationSender }

func NewSMTPPasswordResetSender(config SMTPEmailVerificationSenderConfig) *SMTPPasswordResetSender {
	return &SMTPPasswordResetSender{smtp: NewSMTPEmailVerificationSender(config)}
}
func (s *SMTPPasswordResetSender) SendPasswordResetCode(ctx context.Context, input domain.SendPasswordResetCodeDTO) error {
	message, err := s.buildMessage(input)
	if err != nil {
		return err
	}
	return s.smtp.sendMessage(ctx, input.Email, message, "password reset")
}
func (s *SMTPPasswordResetSender) buildMessage(input domain.SendPasswordResetCodeDTO) ([]byte, error) {
	recipient, err := mail.ParseAddress(input.Email)
	if err != nil || recipient.Address != input.Email {
		return nil, fmt.Errorf("invalid password reset recipient")
	}
	from := mail.Address{Name: s.smtp.config.FromName, Address: s.smtp.config.FromAddress}
	var body, message bytes.Buffer
	mw := multipart.NewWriter(&body)
	headers := []string{"From: " + from.String(), "To: " + recipient.String(), "Subject: " + mime.QEncoding.Encode("utf-8", "Recupere sua senha do FoundryStack"), "MIME-Version: 1.0", fmt.Sprintf(`Content-Type: multipart/alternative; boundary=%q`, mw.Boundary()), "", ""}
	for _, h := range headers {
		message.WriteString(h)
		message.WriteString("\r\n")
	}
	expires := input.ExpiresAt.UTC().Format(time.RFC3339)
	text := fmt.Sprintf("Recupere sua senha do FoundryStack\r\n\r\nSeu código de recuperação é: %s\r\n\r\nO código expira em: %s.\r\n\r\nSe você não solicitou este código, ignore esta mensagem.\r\n", input.Code, expires)
	if err := writeEmailPart(mw, `text/plain; charset="UTF-8"`, text); err != nil {
		return nil, fmt.Errorf("compose text password reset message: %w", err)
	}
	htmlBody := fmt.Sprintf("<!doctype html><html lang=\"pt-BR\"><body><h1>Recupere sua senha do FoundryStack</h1><p>Seu código de recuperação é:</p><p><strong>%s</strong></p><p>O código expira em: %s.</p><p>Se você não solicitou este código, ignore esta mensagem.</p></body></html>", html.EscapeString(input.Code), html.EscapeString(expires))
	if err := writeEmailPart(mw, `text/html; charset="UTF-8"`, htmlBody); err != nil {
		return nil, fmt.Errorf("compose HTML password reset message: %w", err)
	}
	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("finish password reset message: %w", err)
	}
	message.Write(body.Bytes())
	return message.Bytes(), nil
}

var _ domain.PasswordResetSender = (*MemoryPasswordResetSender)(nil)
var _ domain.PasswordResetSender = (*SMTPPasswordResetSender)(nil)
