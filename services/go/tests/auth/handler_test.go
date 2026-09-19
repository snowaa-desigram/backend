package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/snowaa-desigram/backend/services/go/internal/auth"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/handler"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// httpFixture — хендлеры на in-memory сервисе; защищённые маршруты обёрнуты в тот же JWT-middleware
// go-zero (rest/handler.Authorize), что и в проде. Ответы сверяются со схемами из OpenAPI.
type httpFixture struct {
	*fixture
	routes map[string]http.Handler // "METHOD /path"
}

func newHTTPFixture(t *testing.T) *httpFixture {
	t.Helper()
	httpx.SetErrorHandlerCtx(auth.ErrorHandler)
	f := newFixture(t)
	// в HTTP-тестах часы реальные: JWT-middleware go-zero проверяет exp/iat по time.Now
	f.now = time.Now()
	h := &httpFixture{fixture: f, routes: map[string]http.Handler{}}
	public, protected := auth.Routes(auth.NewHandler(f.svc))
	for _, r := range public {
		h.routes[r.Method+" "+r.Path] = r.Handler
	}
	authorize := handler.Authorize(testSecret, handler.WithUnauthorizedCallback(auth.UnauthorizedCallback))
	for _, r := range protected {
		h.routes[r.Method+" "+r.Path] = authorize(r.Handler)
	}
	return h
}

// call делает запрос, проверяет статус и тело по спеке, возвращает декодированное тело.
func (h *httpFixture) call(t *testing.T, method, path string, body any, bearer string, wantStatus int) map[string]any {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if s, ok := body.(string); ok {
			buf.WriteString(s)
		} else if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	rec := httptest.NewRecorder()
	route, ok := h.routes[method+" "+path]
	if !ok {
		t.Fatalf("no route %s %s", method, path)
	}
	route.ServeHTTP(rec, req)

	if rec.Code != wantStatus {
		t.Fatalf("%s %s: status = %d, want %d; body: %s", method, path, rec.Code, wantStatus, rec.Body.String())
	}
	return assertBodyMatchesSpec(t, method, path, rec.Code, rec.Body.Bytes())
}

// assertBodyMatchesSpec: тело ответа валидно по схеме операции из backend/openapi/auth.yaml.
func assertBodyMatchesSpec(t *testing.T, method, path string, status int, body []byte) map[string]any {
	t.Helper()
	op := loadSpec(t).Paths.Find(path).GetOperation(strings.ToUpper(method))
	if op == nil {
		t.Fatalf("spec has no %s %s", method, path)
	}
	resp := op.Responses.Status(status)
	if resp == nil {
		t.Fatalf("spec: %s %s has no response %d", method, path, status)
	}
	media := resp.Value.Content.Get("application/json")
	if media == nil {
		if len(body) != 0 {
			t.Fatalf("spec: %s %s %d has no body, got %s", method, path, status, body)
		}
		return nil
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatalf("body is not a JSON object: %s", body)
	}
	if err := media.Schema.Value.VisitJSON(decoded, openapi3.MultiErrors()); err != nil {
		t.Fatalf("%s %s %d: body does not match schema: %v\n%s", method, path, status, err, body)
	}
	return decoded
}

func TestHTTPFullFlow(t *testing.T) {
	h := newHTTPFixture(t)
	post := http.MethodPost

	h.call(t, post, "/api/auth/register", map[string]string{"email": "  User@Example.com ", "password": testPassword}, "", http.StatusAccepted)
	h.call(t, post, "/api/auth/register", map[string]string{"email": testEmail, "password": testPassword}, "", http.StatusTooManyRequests)

	code := h.lastCode(t)
	body := h.call(t, post, "/api/auth/register/confirm", map[string]string{"email": testEmail, "code": "000000"}, "", http.StatusBadRequest)
	if body["code"] != "invalid_code" {
		t.Fatalf("code = %v", body["code"])
	}
	tokens := h.call(t, post, "/api/auth/register/confirm", map[string]string{"email": testEmail, "code": code}, "", http.StatusOK)
	access, refresh := tokens["accessToken"].(string), tokens["refreshToken"].(string)

	me := h.call(t, http.MethodGet, "/api/auth/me", nil, access, http.StatusOK)
	if me["email"] != testEmail {
		t.Fatalf("me = %v", me)
	}
	h.call(t, http.MethodGet, "/api/auth/me", nil, "", http.StatusUnauthorized)
	h.call(t, http.MethodGet, "/api/auth/me", nil, "not-a-jwt", http.StatusUnauthorized)

	h.call(t, post, "/api/auth/login", map[string]string{"email": testEmail, "password": "wrong password"}, "", http.StatusUnauthorized)
	h.call(t, post, "/api/auth/login", map[string]string{"email": testEmail, "password": testPassword}, "", http.StatusOK)

	next := h.call(t, post, "/api/auth/refresh", map[string]string{"refreshToken": refresh}, "", http.StatusOK)
	h.call(t, post, "/api/auth/refresh", map[string]string{"refreshToken": refresh}, "", http.StatusUnauthorized)

	// logout: без токена — 401, с токеном — 204 без тела
	h.call(t, post, "/api/auth/logout", map[string]string{"refreshToken": next["refreshToken"].(string)}, "", http.StatusUnauthorized)
	h.call(t, post, "/api/auth/logout", map[string]string{"refreshToken": next["refreshToken"].(string)}, access, http.StatusNoContent)
}

func TestHTTPPasswordReset(t *testing.T) {
	h := newHTTPFixture(t)
	post := http.MethodPost
	h.registered(t)

	h.call(t, post, "/api/auth/password/forgot", map[string]string{"email": "nobody@example.com"}, "", http.StatusAccepted)
	h.call(t, post, "/api/auth/password/forgot", map[string]string{"email": testEmail}, "", http.StatusAccepted)
	code := h.lastCode(t)

	h.call(t, post, "/api/auth/password/reset", map[string]string{"email": testEmail, "code": code, "newPassword": "short"}, "", http.StatusBadRequest)
	h.call(t, post, "/api/auth/password/reset", map[string]string{"email": testEmail, "code": code, "newPassword": "brand new password"}, "", http.StatusOK)
	h.call(t, post, "/api/auth/password/reset", map[string]string{"email": testEmail, "code": code, "newPassword": "brand new password"}, "", http.StatusGone)
	h.call(t, post, "/api/auth/login", map[string]string{"email": testEmail, "password": "brand new password"}, "", http.StatusOK)
}

func TestHTTPValidation(t *testing.T) {
	h := newHTTPFixture(t)
	post := http.MethodPost

	cases := []struct {
		name string
		path string
		body any
		want string // ключ в details
	}{
		{"malformed json", "/api/auth/register", "{not json", "body"},
		{"bad email", "/api/auth/register", map[string]string{"email": "not-an-email", "password": testPassword}, "email"},
		{"email with name", "/api/auth/register", map[string]string{"email": "Bob <bob@example.com>", "password": testPassword}, "email"},
		{"short password", "/api/auth/register", map[string]string{"email": testEmail, "password": "1234567"}, "password"},
		{"long password", "/api/auth/register", map[string]string{"email": testEmail, "password": strings.Repeat("x", 73)}, "password"},
		{"bad code", "/api/auth/register/confirm", map[string]string{"email": testEmail, "code": "12345"}, "code"},
		{"missing refresh", "/api/auth/refresh", map[string]string{}, "refreshToken"},
		{"missing email", "/api/auth/password/forgot", map[string]string{}, "email"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			body := h.call(t, post, c.path, c.body, "", http.StatusBadRequest)
			if body["code"] != "validation" {
				t.Fatalf("code = %v", body["code"])
			}
			details, _ := body["details"].(map[string]any)
			if _, ok := details[c.want]; !ok {
				t.Fatalf("details = %v, want key %q", details, c.want)
			}
		})
	}
	if len(h.mailer.Messages) != 0 {
		t.Fatal("invalid requests must not send mail")
	}
}

func TestHTTPExpiredAccessToken(t *testing.T) {
	h := newHTTPFixture(t)
	past := time.Now().Add(-time.Hour)
	tok, err := auth.NewTokenIssuer(testSecret, time.Minute, func() time.Time { return past }).IssueAccess(&auth.User{ID: "u1", Email: testEmail})
	if err != nil {
		t.Fatal(err)
	}
	h.call(t, http.MethodGet, "/api/auth/me", nil, tok, http.StatusUnauthorized)
}

func TestErrorHandlerHidesInternal(t *testing.T) {
	status, body := auth.ErrorHandler(context.Background(), fmt.Errorf("db exploded"))
	if status != http.StatusInternalServerError || strings.Contains(fmt.Sprint(body), "exploded") {
		t.Fatalf("status=%d body=%v", status, body)
	}
}

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	auth.Health(rec, httptest.NewRequest(http.MethodGet, "/health", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"ok"`) {
		t.Fatalf("health: %d %s", rec.Code, rec.Body.String())
	}
}

var _ rest.Route
