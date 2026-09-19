package auth_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/snowaa-desigram/backend/services/go/internal/auth"
)

const (
	testSecret   = "test-secret"
	testEmail    = "user@example.com"
	testPassword = "correct horse"
)

var codeRegexp = regexp.MustCompile(`[0-9]{6}`)

// fixture — сервис на in-memory хранилищах с управляемыми часами.
type fixture struct {
	svc    *auth.Service
	users  *auth.MemoryUserStore
	tokens *auth.MemoryRefreshTokenStore
	codes  *auth.MemoryCodeStore
	mailer *auth.FakeMailer
	now    time.Time
	opts   auth.Options
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	f := &fixture{
		now: time.Date(2026, 9, 19, 12, 0, 0, 0, time.UTC),
		opts: auth.Options{
			AccessTTL: 15 * time.Minute, RefreshTTL: 30 * 24 * time.Hour,
			CodeTTL: 10 * time.Minute, CodeCooldown: time.Minute, CodeMaxAttempts: 3,
			LoginMaxFailures: 3, LoginWindow: 15 * time.Minute,
		},
		mailer: &auth.FakeMailer{},
	}
	clock := func() time.Time { return f.now }
	f.users = auth.NewMemoryUserStore()
	f.tokens = auth.NewMemoryRefreshTokenStore()
	f.codes = auth.NewMemoryCodeStore(clock)
	f.svc = auth.NewService(f.users, f.tokens, f.codes, f.mailer, auth.NewTokenIssuer(testSecret, f.opts.AccessTTL, clock), clock, f.opts)
	return f
}

func (f *fixture) advance(d time.Duration) { f.now = f.now.Add(d) }

// lastCode — код из последнего письма.
func (f *fixture) lastCode(t *testing.T) string {
	t.Helper()
	m := f.mailer.Last()
	if m == nil {
		t.Fatal("no mail sent")
	}
	code := codeRegexp.FindString(m.Body)
	if code == "" {
		t.Fatalf("no code in mail body: %q", m.Body)
	}
	return code
}

// registered — зарегистрированный и подтверждённый пользователь с парой токенов.
func (f *fixture) registered(t *testing.T) *auth.TokenPair {
	t.Helper()
	ctx := context.Background()
	if err := f.svc.Register(ctx, testEmail, testPassword); err != nil {
		t.Fatalf("register: %v", err)
	}
	pair, err := f.svc.ConfirmRegistration(ctx, testEmail, f.lastCode(t))
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	f.advance(f.opts.CodeCooldown + time.Second) // снять cooldown для следующих кодов
	return pair
}

func mustErr(t *testing.T, got error, want *auth.Error) {
	t.Helper()
	if !errors.Is(got, want) {
		t.Fatalf("error = %v, want %v", got, want)
	}
}

func TestRegisterConfirmLogin(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	if err := f.svc.Register(ctx, testEmail, testPassword); err != nil {
		t.Fatalf("register: %v", err)
	}
	if m := f.mailer.Last(); m == nil || m.To != testEmail {
		t.Fatalf("mail = %+v, want to %s", m, testEmail)
	}

	// до подтверждения логин запрещён
	_, err := f.svc.Login(ctx, testEmail, testPassword)
	mustErr(t, err, auth.ErrEmailNotVerified)

	pair, err := f.svc.ConfirmRegistration(ctx, testEmail, f.lastCode(t))
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if pair.AccessToken == "" || pair.RefreshToken == "" || pair.ExpiresIn != 900 {
		t.Fatalf("bad token pair: %+v", pair)
	}

	// код одноразовый
	_, err = f.svc.ConfirmRegistration(ctx, testEmail, f.lastCode(t))
	mustErr(t, err, auth.ErrCodeExpired)

	u, err := f.svc.Me(ctx, claimsOf(t, pair.AccessToken)[auth.ClaimUserID].(string))
	if err != nil || u.Email != testEmail || !u.Verified() {
		t.Fatalf("me = %+v, %v", u, err)
	}

	if _, err := f.svc.Login(ctx, testEmail, testPassword); err != nil {
		t.Fatalf("login: %v", err)
	}
	_, err = f.svc.Login(ctx, testEmail, "wrong password")
	mustErr(t, err, auth.ErrInvalidCredentials)
	_, err = f.svc.Login(ctx, "nobody@example.com", testPassword)
	mustErr(t, err, auth.ErrInvalidCredentials)
}

func TestRegisterExistingEmail(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()

	// неподтверждённый: пароль обновляется, код уходит заново (после cooldown)
	if err := f.svc.Register(ctx, testEmail, "first password"); err != nil {
		t.Fatal(err)
	}
	mustErr(t, f.svc.Register(ctx, testEmail, "second password"), auth.ErrTooManyRequests)
	f.advance(f.opts.CodeCooldown + time.Second)
	if err := f.svc.Register(ctx, testEmail, "second password"); err != nil {
		t.Fatal(err)
	}
	if len(f.mailer.Messages) != 2 {
		t.Fatalf("mails = %d, want 2", len(f.mailer.Messages))
	}
	if _, err := f.svc.ConfirmRegistration(ctx, testEmail, f.lastCode(t)); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.Login(ctx, testEmail, "second password"); err != nil {
		t.Fatalf("login with updated password: %v", err)
	}

	// подтверждённый — занят
	mustErr(t, f.svc.Register(ctx, testEmail, "third password"), auth.ErrEmailTaken)
}

func TestConfirmWrongCode(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	if err := f.svc.Register(ctx, testEmail, testPassword); err != nil {
		t.Fatal(err)
	}

	_, err := f.svc.ConfirmRegistration(ctx, testEmail, "000000")
	mustErr(t, err, auth.ErrInvalidCode)
	_, err = f.svc.ConfirmRegistration(ctx, testEmail, "000000")
	mustErr(t, err, auth.ErrInvalidCode)
	// третья неудача (MaxAttempts=3) — блок, даже с правильным кодом
	_, err = f.svc.ConfirmRegistration(ctx, testEmail, "000000")
	mustErr(t, err, auth.ErrTooManyAttempts)
	_, err = f.svc.ConfirmRegistration(ctx, testEmail, f.lastCode(t))
	mustErr(t, err, auth.ErrTooManyAttempts)

	// новый код снимает блок
	f.advance(f.opts.CodeCooldown + time.Second)
	if err := f.svc.ResendCode(ctx, testEmail); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.ConfirmRegistration(ctx, testEmail, f.lastCode(t)); err != nil {
		t.Fatalf("confirm with fresh code: %v", err)
	}
}

func TestConfirmExpiredCode(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	if err := f.svc.Register(ctx, testEmail, testPassword); err != nil {
		t.Fatal(err)
	}
	code := f.lastCode(t)
	f.advance(f.opts.CodeTTL + time.Second)
	_, err := f.svc.ConfirmRegistration(ctx, testEmail, code)
	mustErr(t, err, auth.ErrCodeExpired)
}

func TestResendCodeSilentForUnknownOrVerified(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	if err := f.svc.ResendCode(ctx, "nobody@example.com"); err != nil {
		t.Fatal(err)
	}
	f.registered(t)
	sent := len(f.mailer.Messages)
	if err := f.svc.ResendCode(ctx, testEmail); err != nil {
		t.Fatal(err)
	}
	if len(f.mailer.Messages) != sent {
		t.Fatal("resend for verified user must not send mail")
	}
}

func TestLoginLockout(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	f.registered(t)

	for i := 0; i < f.opts.LoginMaxFailures; i++ {
		_, err := f.svc.Login(ctx, testEmail, "wrong")
		mustErr(t, err, auth.ErrInvalidCredentials)
	}
	_, err := f.svc.Login(ctx, testEmail, testPassword)
	mustErr(t, err, auth.ErrTooManyAttempts)

	f.advance(f.opts.LoginWindow + time.Second)
	if _, err := f.svc.Login(ctx, testEmail, testPassword); err != nil {
		t.Fatalf("login after window: %v", err)
	}
	// успешный вход сбрасывает счётчик
	if n, _ := f.codes.Failures(ctx, testEmail); n != 0 {
		t.Fatalf("failures = %d, want 0", n)
	}
}

func TestRefreshRotation(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	pair := f.registered(t)

	next, err := f.svc.Refresh(ctx, pair.RefreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if next.RefreshToken == pair.RefreshToken {
		t.Fatal("refresh token must rotate")
	}

	// повторное использование ротированного — кража: сносим всё, включая новый
	_, err = f.svc.Refresh(ctx, pair.RefreshToken)
	mustErr(t, err, auth.ErrInvalidToken)
	_, err = f.svc.Refresh(ctx, next.RefreshToken)
	mustErr(t, err, auth.ErrInvalidToken)

	_, err = f.svc.Refresh(ctx, "garbage")
	mustErr(t, err, auth.ErrInvalidToken)
}

func TestRefreshExpired(t *testing.T) {
	f := newFixture(t)
	pair := f.registered(t)
	f.advance(f.opts.RefreshTTL + time.Second)
	_, err := f.svc.Refresh(context.Background(), pair.RefreshToken)
	mustErr(t, err, auth.ErrInvalidToken)
}

func TestLogout(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	pair := f.registered(t)
	uid := claimsOf(t, pair.AccessToken)[auth.ClaimUserID].(string)

	if err := f.svc.Logout(ctx, uid, pair.RefreshToken); err != nil {
		t.Fatal(err)
	}
	_, err := f.svc.Refresh(ctx, pair.RefreshToken)
	mustErr(t, err, auth.ErrInvalidToken)

	// идемпотентно; чужой токен — no-op
	if err := f.svc.Logout(ctx, uid, pair.RefreshToken); err != nil {
		t.Fatal(err)
	}
	other, _ := f.svc.Login(ctx, testEmail, testPassword)
	if err := f.svc.Logout(ctx, "someone-else", other.RefreshToken); err != nil {
		t.Fatal(err)
	}
	if _, err := f.svc.Refresh(ctx, other.RefreshToken); err != nil {
		t.Fatalf("token of another user must stay valid: %v", err)
	}
}

func TestPasswordReset(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	pair := f.registered(t)

	// неизвестный email — молча
	if err := f.svc.ForgotPassword(ctx, "nobody@example.com"); err != nil {
		t.Fatal(err)
	}
	sent := len(f.mailer.Messages)
	if err := f.svc.ForgotPassword(ctx, testEmail); err != nil {
		t.Fatal(err)
	}
	if len(f.mailer.Messages) != sent+1 {
		t.Fatal("reset mail not sent")
	}
	code := f.lastCode(t)

	// код сброса не годится для подтверждения регистрации и наоборот
	_, err := f.svc.ConfirmRegistration(ctx, testEmail, code)
	mustErr(t, err, auth.ErrCodeExpired)

	next, err := f.svc.ResetPassword(ctx, testEmail, code, "new password!")
	if err != nil {
		t.Fatalf("reset: %v", err)
	}
	_, err = f.svc.Login(ctx, testEmail, testPassword)
	mustErr(t, err, auth.ErrInvalidCredentials)
	if _, err := f.svc.Login(ctx, testEmail, "new password!"); err != nil {
		t.Fatalf("login with new password: %v", err)
	}
	// старые сессии удалены (без каскада на новую), новая — жива
	_, err = f.svc.Refresh(ctx, pair.RefreshToken)
	mustErr(t, err, auth.ErrInvalidToken)
	if _, err := f.svc.Refresh(ctx, next.RefreshToken); err != nil {
		t.Fatalf("new session: %v", err)
	}
}

func TestForgotPasswordUnverifiedIsSilent(t *testing.T) {
	f := newFixture(t)
	ctx := context.Background()
	if err := f.svc.Register(ctx, testEmail, testPassword); err != nil {
		t.Fatal(err)
	}
	sent := len(f.mailer.Messages)
	if err := f.svc.ForgotPassword(ctx, testEmail); err != nil {
		t.Fatal(err)
	}
	if len(f.mailer.Messages) != sent {
		t.Fatal("must not send reset code to unverified email")
	}
}

func TestMeUnknownUser(t *testing.T) {
	f := newFixture(t)
	_, err := f.svc.Me(context.Background(), "missing")
	mustErr(t, err, auth.ErrUnauthorized)
}
