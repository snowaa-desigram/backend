package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Options — параметры логики (см. Config.Options).
type Options struct {
	AppName          string
	AccessTTL        time.Duration
	RefreshTTL       time.Duration
	CodeTTL          time.Duration
	CodeCooldown     time.Duration
	CodeMaxAttempts  int
	LoginMaxFailures int
	LoginWindow      time.Duration
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// Service — вся логика auth: один метод на эндпоинт. Ошибки клиенту — *Error (см. errors.go).
type Service struct {
	users  UserStore
	tokens RefreshTokenStore
	codes  CodeStore
	mailer Mailer
	issuer *TokenIssuer
	now    func() time.Time
	opts   Options
}

func NewService(users UserStore, tokens RefreshTokenStore, codes CodeStore, mailer Mailer,
	issuer *TokenIssuer, now func() time.Time, opts Options) *Service {
	return &Service{users: users, tokens: tokens, codes: codes, mailer: mailer, issuer: issuer, now: now, opts: opts}
}

// Register создаёт неподтверждённого пользователя и шлёт код. Повторная регистрация неподтверждённого
// email обновляет пароль и шлёт код заново; подтверждённого — ErrEmailTaken.
func (s *Service) Register(ctx context.Context, email, password string) (err error) {
	defer track("register", &err)()
	hash, err := HashPassword(password)
	if err != nil {
		return err
	}

	u, err := s.users.FindByEmail(ctx, email)
	switch {
	case err == nil && u.Verified():
		return ErrEmailTaken
	case err == nil:
		u.PasswordHash, u.UpdatedAt = hash, s.now()
		if err := s.users.Update(ctx, u); err != nil {
			return err
		}
	case errors.Is(err, ErrNotFound):
		now := s.now()
		u = &User{ID: uuid.NewString(), Email: email, PasswordHash: hash, CreatedAt: now, UpdatedAt: now}
		if err := s.users.Create(ctx, u); err != nil {
			if errors.Is(err, ErrDuplicate) {
				return ErrEmailTaken
			}
			return err
		}
	default:
		return err
	}

	return s.sendCode(ctx, PurposeRegister, email)
}

// ResendCode шлёт код регистрации заново. Для неизвестного или уже подтверждённого email молча ничего не делает.
func (s *Service) ResendCode(ctx context.Context, email string) (err error) {
	defer track("resend_code", &err)()
	u, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, ErrNotFound) || (err == nil && u.Verified()) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.sendCode(ctx, PurposeRegister, email)
}

// ConfirmRegistration проверяет код, помечает email подтверждённым и выдаёт токены.
func (s *Service) ConfirmRegistration(ctx context.Context, email, code string) (pair *TokenPair, err error) {
	defer track("confirm_registration", &err)()
	if err := s.verifyCode(ctx, PurposeRegister, email, code); err != nil {
		return nil, err
	}
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if !u.Verified() {
		now := s.now()
		u.EmailVerifiedAt, u.UpdatedAt = &now, now
		if err := s.users.Update(ctx, u); err != nil {
			return nil, err
		}
	}
	return s.issueTokens(ctx, u)
}

// Login — email+password → токены. После LoginMaxFailures неудач подряд в окне — ErrTooManyAttempts.
func (s *Service) Login(ctx context.Context, email, password string) (pair *TokenPair, err error) {
	defer track("login", &err)()
	failures, err := s.codes.Failures(ctx, email)
	if err != nil {
		return nil, err
	}
	if failures >= s.opts.LoginMaxFailures {
		return nil, ErrTooManyAttempts
	}

	u, err := s.users.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, err
	}
	if u == nil || !CheckPassword(u.PasswordHash, password) {
		if _, err := s.codes.IncrFailures(ctx, email, s.opts.LoginWindow); err != nil {
			return nil, err
		}
		return nil, ErrInvalidCredentials
	}
	if !u.Verified() {
		return nil, ErrEmailNotVerified
	}
	if err := s.codes.ResetFailures(ctx, email); err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, u)
}

// Refresh ротирует refresh-токен. Повторное использование уже ротированного — признак кражи: сносим все сессии.
func (s *Service) Refresh(ctx context.Context, raw string) (pair *TokenPair, err error) {
	defer track("refresh", &err)()
	t, err := s.tokens.FindByHash(ctx, HashToken(raw))
	if errors.Is(err, ErrNotFound) {
		return nil, ErrInvalidToken
	}
	if err != nil {
		return nil, err
	}
	now := s.now()
	if t.RevokedAt != nil {
		if err := s.tokens.DeleteAllForUser(ctx, t.UserID); err != nil {
			return nil, err
		}
		return nil, ErrInvalidToken
	}
	if !t.ExpiresAt.After(now) {
		return nil, ErrInvalidToken
	}
	u, err := s.users.FindByID(ctx, t.UserID)
	if err != nil {
		return nil, err
	}
	if err := s.tokens.Revoke(ctx, t.ID, now); err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, u)
}

// Logout удаляет refresh-токен пользователя. Идемпотентен: чужой или неизвестный токен — no-op.
func (s *Service) Logout(ctx context.Context, userID, raw string) (err error) {
	defer track("logout", &err)()
	t, err := s.tokens.FindByHash(ctx, HashToken(raw))
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	if t.UserID != userID {
		return nil
	}
	return s.tokens.Delete(ctx, t.ID)
}

// ForgotPassword шлёт код сброса. Неизвестный или неподтверждённый email — молча nil (не раскрываем наличие аккаунта).
func (s *Service) ForgotPassword(ctx context.Context, email string) (err error) {
	defer track("forgot_password", &err)()
	u, err := s.users.FindByEmail(ctx, email)
	if errors.Is(err, ErrNotFound) || (err == nil && !u.Verified()) {
		return nil
	}
	if err != nil {
		return err
	}
	return s.sendCode(ctx, PurposePasswordReset, email)
}

// ResetPassword по коду ставит новый пароль, удаляет все refresh-токены и выдаёт новую пару.
func (s *Service) ResetPassword(ctx context.Context, email, code, newPassword string) (pair *TokenPair, err error) {
	defer track("reset_password", &err)()
	if err := s.verifyCode(ctx, PurposePasswordReset, email, code); err != nil {
		return nil, err
	}
	u, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	hash, err := HashPassword(newPassword)
	if err != nil {
		return nil, err
	}
	now := s.now()
	u.PasswordHash, u.UpdatedAt = hash, now
	if err := s.users.Update(ctx, u); err != nil {
		return nil, err
	}
	if err := s.tokens.DeleteAllForUser(ctx, u.ID); err != nil {
		return nil, err
	}
	if err := s.codes.ResetFailures(ctx, email); err != nil {
		return nil, err
	}
	return s.issueTokens(ctx, u)
}

// Me — пользователь по id из access-токена.
func (s *Service) Me(ctx context.Context, userID string) (u *User, err error) {
	defer track("me", &err)()

	u, err = s.users.FindByID(ctx, userID)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrUnauthorized
	}
	return u, err
}

func (s *Service) sendCode(ctx context.Context, purpose CodePurpose, email string) error {
	ok, err := s.codes.SetCooldown(ctx, purpose, email, s.opts.CodeCooldown)
	if err != nil {
		return err
	}
	if !ok {
		return ErrTooManyRequests
	}
	code, err := NewCode()
	if err != nil {
		return err
	}
	if err := s.codes.Put(ctx, purpose, email, HashCode(code), s.opts.CodeTTL); err != nil {
		return err
	}
	t := mailTemplates[purpose]
	start := time.Now()
	err = s.mailer.Send(ctx, email, s.opts.AppName+": "+t.subject, fmt.Sprintf(t.body, code, int(s.opts.CodeTTL.Minutes())))
	trackMail(purpose, start, err)
	return err
}

func (s *Service) verifyCode(ctx context.Context, purpose CodePurpose, email, code string) error {
	hash, attempts, err := s.codes.Get(ctx, purpose, email)
	if errors.Is(err, ErrNotFound) {
		return ErrCodeExpired
	}
	if err != nil {
		return err
	}
	if attempts >= s.opts.CodeMaxAttempts {
		return ErrTooManyAttempts
	}
	if !hashEqual(hash, HashCode(code)) {
		n, err := s.codes.IncrAttempts(ctx, purpose, email)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return err
		}
		if n >= s.opts.CodeMaxAttempts {
			return ErrTooManyAttempts
		}
		return ErrInvalidCode
	}
	return s.codes.Delete(ctx, purpose, email)
}

func (s *Service) issueTokens(ctx context.Context, u *User) (*TokenPair, error) {
	access, err := s.issuer.IssueAccess(u)
	if err != nil {
		return nil, err
	}
	raw, hash, err := NewRefreshToken()
	if err != nil {
		return nil, err
	}
	now := s.now()
	t := &RefreshToken{ID: uuid.NewString(), UserID: u.ID, TokenHash: hash, ExpiresAt: now.Add(s.opts.RefreshTTL), CreatedAt: now}
	if err := s.tokens.Create(ctx, t); err != nil {
		return nil, err
	}
	return &TokenPair{AccessToken: access, RefreshToken: raw, ExpiresIn: int64(s.issuer.TTL().Seconds())}, nil
}
