// Package mail sends emails. The rest of the app only knows the Mailer
// interface, so the delivery method is a configuration choice:
//
//   - MAIL_DRIVER=log  (default) prints the email to the server log. Use it in
//     development: the password-reset link appears in the terminal.
//   - MAIL_DRIVER=smtp sends through any SMTP provider (Gmail, Brevo, Mailgun,
//     Amazon SES, your host's mail server, ...). Use port 587 (STARTTLS).
package mail

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mime"
	"net"
	"net/smtp"
	"strings"
	"time"
)

// Message is a plain-text email.
type Message struct {
	To      string
	Subject string
	Text    string
}

// Mailer delivers messages.
type Mailer interface {
	Send(ctx context.Context, m Message) error
}

// Config selects and configures the mailer.
type Config struct {
	Driver   string // "log" or "smtp"
	Host     string
	Port     string
	Username string
	Password string
	From     string // e.g. "StockWise <no-reply@example.com>"
}

// New builds the mailer described by cfg.
func New(cfg Config) (Mailer, error) {
	switch cfg.Driver {
	case "", "log":
		return LogMailer{}, nil
	case "smtp":
		if cfg.Host == "" || cfg.From == "" {
			return nil, errors.New("MAIL_DRIVER=smtp needs SMTP_HOST and SMTP_FROM")
		}
		if cfg.Port == "" {
			cfg.Port = "587"
		}
		return &SMTPMailer{cfg: cfg}, nil
	default:
		return nil, fmt.Errorf("unknown MAIL_DRIVER %q (use log or smtp)", cfg.Driver)
	}
}

// LogMailer writes emails to the log instead of sending them.
type LogMailer struct{}

func (LogMailer) Send(_ context.Context, m Message) error {
	slog.Info("email (MAIL_DRIVER=log, not sent)", "to", m.To, "subject", m.Subject, "body", m.Text)
	return nil
}

// SMTPMailer sends through an SMTP server, upgrading to TLS when offered.
type SMTPMailer struct {
	cfg Config
}

func (s *SMTPMailer) Send(ctx context.Context, m Message) error {
	if strings.ContainsAny(m.To, "\r\n") || strings.ContainsAny(m.Subject, "\r\n") {
		return errors.New("invalid header value") // blocks header injection
	}
	addr := net.JoinHostPort(s.cfg.Host, s.cfg.Port)
	var auth smtp.Auth
	if s.cfg.Username != "" {
		auth = smtp.PlainAuth("", s.cfg.Username, s.cfg.Password, s.cfg.Host)
	}
	fromAddr := s.cfg.From
	if i := strings.LastIndex(fromAddr, "<"); i >= 0 {
		fromAddr = strings.TrimSuffix(fromAddr[i+1:], ">")
	}

	body := strings.Join([]string{
		"From: " + s.cfg.From,
		"To: " + m.To,
		"Subject: " + mime.QEncoding.Encode("utf-8", m.Subject),
		"Date: " + time.Now().Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=utf-8",
		"Content-Transfer-Encoding: 8bit",
		"",
		strings.ReplaceAll(m.Text, "\n", "\r\n"),
	}, "\r\n")

	done := make(chan error, 1)
	go func() { done <- smtp.SendMail(addr, auth, fromAddr, []string{m.To}, []byte(body)) }()
	select {
	case err := <-done:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}
