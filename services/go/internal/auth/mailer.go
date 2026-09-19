package auth

import (
	"context"
	"fmt"
	"sync"

	"github.com/wneessen/go-mail"
)

// Mailer — отправка писем. Синхронно, в запросе (таймаут — у SMTP-клиента).
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

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

// FakeMailer копит письма в памяти (тесты).
type FakeMailer struct {
	mu       sync.Mutex
	Messages []FakeMessage
}

type FakeMessage struct{ To, Subject, Body string }

func (m *FakeMailer) Send(_ context.Context, to, subject, body string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Messages = append(m.Messages, FakeMessage{To: to, Subject: subject, Body: body})
	return nil
}

func (m *FakeMailer) Last() *FakeMessage {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.Messages) == 0 {
		return nil
	}
	return &m.Messages[len(m.Messages)-1]
}

// Тексты писем: тема получает префикс "<AppName>: ", код — отдельной строкой, e2e достаёт его регуляркой [0-9]{6}.
var mailTemplates = map[CodePurpose]struct{ subject, body string }{
	PurposeRegister:      {"код подтверждения", "Ваш код подтверждения регистрации: %s\n\nКод действует %d минут."},
	PurposePasswordReset: {"сброс пароля", "Ваш код для смены пароля: %s\n\nКод действует %d минут. Если вы не запрашивали сброс — проигнорируйте письмо."},
}
