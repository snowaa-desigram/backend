package auth_test

import (
	"context"
	"sync"
)

// FakeMailer реализует service.Mailer: копит письма в памяти.
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
