package service

import (
	"context"

	"github.com/snowaa-desigram/backend/services/go/internal/auth/store"
)

// Mailer — отправка писем. Синхронно, в запросе (таймаут — у SMTP-клиента). Реализация — adapter.SMTPMailer.
type Mailer interface {
	Send(ctx context.Context, to, subject, body string) error
}

// Тексты писем: тема получает префикс "<AppName>: ", код — отдельной строкой, e2e достаёт его регуляркой [0-9]{6}.
var mailTemplates = map[store.CodePurpose]struct{ subject, body string }{
	store.PurposeRegister:      {"код подтверждения", "Ваш код подтверждения регистрации: %s\n\nКод действует %d минут."},
	store.PurposePasswordReset: {"сброс пароля", "Ваш код для смены пароля: %s\n\nКод действует %d минут. Если вы не запрашивали сброс — проигнорируйте письмо."},
}
