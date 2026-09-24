package infrastructure

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"html"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/mail"
	"net/smtp"
	"net/textproto"
	"time"

	"github.com/rudsonalves/foundry-stack/api/internal/users/domain"
)

const (
	smtpTLSModeImplicit = "tls"
	smtpTLSModeSTARTTLS = "starttls"
	smtpSendMaxAttempts = 2
)

type smtpDialContextFunc func(
	ctx context.Context,
	network string,
	address string,
) (net.Conn, error)

type SMTPEmailVerificationSenderConfig struct {
	Host        string
	Port        int
	Username    string
	Password    string
	FromAddress string
	FromName    string
	TLSMode     string
	Timeout     time.Duration
}

type SMTPEmailVerificationSender struct {
	config      SMTPEmailVerificationSenderConfig
	dialContext smtpDialContextFunc
}

var _ domain.EmailVerificationSender = (*SMTPEmailVerificationSender)(nil)

func NewSMTPEmailVerificationSender(
	config SMTPEmailVerificationSenderConfig,
) *SMTPEmailVerificationSender {
	dialer := &net.Dialer{}

	return &SMTPEmailVerificationSender{
		config:      config,
		dialContext: dialer.DialContext,
	}
}

func (s *SMTPEmailVerificationSender) SendVerificationCode(
	ctx context.Context,
	input domain.SendEmailVerificationCodeDTO,
) error {
	message, err := s.buildMessage(input)
	if err != nil {
		return err
	}
	return s.sendMessage(ctx, input.Email, message, "email verification")
}

func (s *SMTPEmailVerificationSender) sendMessage(ctx context.Context, recipient string, message []byte, operation string) error {
	var lastErr error

	for attempt := 1; attempt <= smtpSendMaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := s.sendOnce(ctx, recipient, message); err != nil {
			lastErr = err
			continue
		}

		return nil
	}

	return fmt.Errorf(
		"send %s message after %d attempts: %w",
		operation,
		smtpSendMaxAttempts,
		lastErr,
	)
}

func (s *SMTPEmailVerificationSender) sendOnce(
	ctx context.Context,
	recipient string,
	message []byte,
) error {
	attemptCtx, cancel := context.WithTimeout(
		ctx,
		s.config.Timeout,
	)
	defer cancel()

	address := net.JoinHostPort(
		s.config.Host,
		fmt.Sprintf("%d", s.config.Port),
	)

	connection, err := s.dialContext(
		attemptCtx,
		"tcp",
		address,
	)
	if err != nil {
		return fmt.Errorf("connect to SMTP server: %w", err)
	}

	if deadline, ok := attemptCtx.Deadline(); ok {
		if err := connection.SetDeadline(deadline); err != nil {
			_ = connection.Close()
			return fmt.Errorf("set SMTP connection deadline: %w", err)
		}
	}

	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: s.config.Host,
	}

	if s.config.TLSMode == smtpTLSModeImplicit {
		tlsConnection := tls.Client(connection, tlsConfig)

		if err := tlsConnection.HandshakeContext(attemptCtx); err != nil {
			_ = connection.Close()
			return fmt.Errorf("establish implicit SMTP TLS: %w", err)
		}

		connection = tlsConnection
	}

	client, err := smtp.NewClient(connection, s.config.Host)
	if err != nil {
		_ = connection.Close()
		return fmt.Errorf("initialize SMTP client: %w", err)
	}
	defer client.Close()

	switch s.config.TLSMode {
	case smtpTLSModeSTARTTLS:
		if err := client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("establish SMTP STARTTLS: %w", err)
		}

	case smtpTLSModeImplicit:
		// The connection completed its TLS handshake before SMTP started.

	default:
		return fmt.Errorf(
			"unsupported SMTP TLS mode %q",
			s.config.TLSMode,
		)
	}

	auth := smtp.PlainAuth(
		"",
		s.config.Username,
		s.config.Password,
		s.config.Host,
	)

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authenticate with SMTP server: %w", err)
	}

	if err := client.Mail(s.config.FromAddress); err != nil {
		return fmt.Errorf("set SMTP sender: %w", err)
	}

	if err := client.Rcpt(recipient); err != nil {
		return fmt.Errorf("set SMTP recipient: %w", err)
	}

	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("start SMTP message body: %w", err)
	}

	if _, err := io.Copy(writer, bytes.NewReader(message)); err != nil {
		_ = writer.Close()
		return fmt.Errorf("write SMTP message body: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("finish SMTP message body: %w", err)
	}

	// Depois que DATA foi aceito, a mensagem já pertence ao servidor.
	// Uma falha em QUIT não deve provocar reenvio e possível duplicação.
	_ = client.Quit()

	return nil
}

func (s *SMTPEmailVerificationSender) buildMessage(
	input domain.SendEmailVerificationCodeDTO,
) ([]byte, error) {
	recipient, err := mail.ParseAddress(input.Email)
	if err != nil || recipient.Address != input.Email {
		return nil, fmt.Errorf("invalid email verification recipient")
	}

	from := mail.Address{
		Name:    s.config.FromName,
		Address: s.config.FromAddress,
	}

	var body bytes.Buffer
	multipartWriter := multipart.NewWriter(&body)

	var message bytes.Buffer

	headers := []string{
		"From: " + from.String(),
		"To: " + recipient.String(),
		"Subject: " + mime.QEncoding.Encode(
			"utf-8",
			"Confirme seu e-mail no FoundryStack",
		),
		"MIME-Version: 1.0",
		fmt.Sprintf(
			`Content-Type: multipart/alternative; boundary=%q`,
			multipartWriter.Boundary(),
		),
		"",
		"",
	}

	for _, header := range headers {
		message.WriteString(header)
		message.WriteString("\r\n")
	}

	expiresAt := input.ExpiresAt.UTC().Format(time.RFC3339)

	textBody := fmt.Sprintf(
		"Confirme seu e-mail no FoundryStack\r\n\r\n"+
			"Seu código de verificação é: %s\r\n\r\n"+
			"O código expira em: %s.\r\n\r\n"+
			"Se você não solicitou este código, ignore esta mensagem.\r\n",
		input.Code,
		expiresAt,
	)

	if err := writeEmailPart(
		multipartWriter,
		`text/plain; charset="UTF-8"`,
		textBody,
	); err != nil {
		return nil, fmt.Errorf(
			"compose text email verification message: %w",
			err,
		)
	}

	htmlBody := fmt.Sprintf(
		"<!doctype html>"+
			"<html lang=\"pt-BR\">"+
			"<body>"+
			"<h1>Confirme seu e-mail no FoundryStack</h1>"+
			"<p>Seu código de verificação é:</p>"+
			"<p><strong>%s</strong></p>"+
			"<p>O código expira em: %s.</p>"+
			"<p>Se você não solicitou este código, ignore esta mensagem.</p>"+
			"</body>"+
			"</html>",
		html.EscapeString(input.Code),
		html.EscapeString(expiresAt),
	)

	if err := writeEmailPart(
		multipartWriter,
		`text/html; charset="UTF-8"`,
		htmlBody,
	); err != nil {
		return nil, fmt.Errorf(
			"compose HTML email verification message: %w",
			err,
		)
	}

	if err := multipartWriter.Close(); err != nil {
		return nil, fmt.Errorf(
			"finish email verification message: %w",
			err,
		)
	}

	message.Write(body.Bytes())

	return message.Bytes(), nil
}

func writeEmailPart(
	writer *multipart.Writer,
	contentType string,
	content string,
) error {
	headers := make(textproto.MIMEHeader)
	headers.Set("Content-Type", contentType)
	headers.Set("Content-Transfer-Encoding", "quoted-printable")

	part, err := writer.CreatePart(headers)
	if err != nil {
		return err
	}

	encoded := quotedprintable.NewWriter(part)

	if _, err := io.WriteString(encoded, content); err != nil {
		_ = encoded.Close()
		return err
	}

	return encoded.Close()
}
