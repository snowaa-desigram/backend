package adapter

import (
	"context"
	"fmt"

	"github.com/wneessen/go-mail"
)

// SMTPMailer реализует service.Mailer через SMTP (go-mail).

type SMTPConfig struct {
	Host     string
	Port     int    `json:",default=25"`
	User     string `json:",optional"`
	Password string `json:",optional"`
	From     string
	// TLS — обязательный STARTTLS (prod); false — plain (Mailpit).
	TLS bool `json:",default=false"`
}

type SMTPMailer struct {
	client *mail.Client
	from   string
}

func NewSMTPMailer(c SMTPConfig) (*SMTPMailer, error) {
	opts := []mail.Option{mail.WithPort(c.Port), mail.WithTLSPolicy(mail.NoTLS)}
	if c.TLS {
		opts = append(opts, mail.WithTLSPolicy(mail.TLSMandatory))
	}
	if c.User != "" {
		opts = append(opts, mail.WithSMTPAuth(mail.SMTPAuthPlain), mail.WithUsername(c.User), mail.WithPassword(c.Password))
	}
	client, err := mail.NewClient(c.Host, opts...)
	if err != nil {
		return nil, fmt.Errorf("smtp client: %w", err)
	}
	return &SMTPMailer{client: client, from: c.From}, nil
}

func (m *SMTPMailer) Send(ctx context.Context, to, subject, body string) error {
	msg := mail.NewMsg()
	if err := msg.From(m.from); err != nil {
		return err
	}
	if err := msg.To(to); err != nil {
		return err
	}
	msg.Subject(subject)
	msg.SetBodyString(mail.TypeTextPlain, body)
	return m.client.DialAndSendWithContext(ctx, msg)
}
