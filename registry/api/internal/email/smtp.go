package email

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
)

// SMTP sends email via SMTP (e.g. Mailhog in dev).
type SMTP struct {
	addr string
	from string
	auth smtp.Auth
}

// NewSMTP creates an SMTP sender. host may include port (e.g. "mailhog:1025").
// If no port, 25 is used. For Mailhog use host "mailhog" or "localhost", port 1025.
func NewSMTP(host, port, from, user, pass string) (*SMTP, error) {
	addr := host
	if port != "" {
		addr = host + ":" + port
	} else {
		addr = host + ":25"
	}
	var auth smtp.Auth
	if user != "" && pass != "" {
		auth = smtp.PlainAuth("", user, pass, host)
	}
	return &SMTP{addr: addr, from: from, auth: auth}, nil
}

// Send sends a plain text (and optional HTML) email.
func (s *SMTP) Send(ctx context.Context, to, subject, bodyText, bodyHTML string) error {
	// Build message (simple format; Mailhog accepts it)
	headers := []string{
		"From: " + s.from,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
	}
	body := bodyText
	if bodyHTML != "" {
		headers = append(headers, "Content-Type: text/html; charset=UTF-8")
		body = bodyHTML
	} else {
		headers = append(headers, "Content-Type: text/plain; charset=UTF-8")
	}
	msg := strings.Join(headers, "\r\n") + "\r\n\r\n" + body

	// Send (no TLS for Mailhog)
	err := smtp.SendMail(s.addr, s.auth, s.from, []string{to}, []byte(msg))
	if err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}
	return nil
}
