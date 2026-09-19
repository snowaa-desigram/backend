package auth

import (
	"encoding/json"
	"net/http"
	"net/mail"
	"regexp"
	"strings"

	"github.com/zeromicro/go-zero/rest/httpx"

	authapi "github.com/snowaa-desigram/backend/services/go/gen/openapi/auth"
)

// Handler — HTTP-слой: разобрать тело → провалидировать → Service → ответ по OpenAPI.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req authapi.RegisterRequest
	if !parse(w, r, &req, func(d details) {
		d.email(req.Email)
		d.password("password", req.Password)
	}) {
		return
	}
	if err := h.svc.Register(r.Context(), normalizeEmail(req.Email), req.Password); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	accepted(w, r)
}

func (h *Handler) ConfirmRegistration(w http.ResponseWriter, r *http.Request) {
	var req authapi.CodeRequest
	if !parse(w, r, &req, func(d details) {
		d.email(req.Email)
		d.code(req.Code)
	}) {
		return
	}
	pair, err := h.svc.ConfirmRegistration(r.Context(), normalizeEmail(req.Email), req.Code)
	respondTokens(w, r, pair, err)
}

func (h *Handler) ResendCode(w http.ResponseWriter, r *http.Request) {
	var req authapi.EmailRequest
	if !parse(w, r, &req, func(d details) { d.email(req.Email) }) {
		return
	}
	if err := h.svc.ResendCode(r.Context(), normalizeEmail(req.Email)); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	accepted(w, r)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req authapi.LoginRequest
	if !parse(w, r, &req, func(d details) {
		d.email(req.Email)
		d.password("password", req.Password)
	}) {
		return
	}
	pair, err := h.svc.Login(r.Context(), normalizeEmail(req.Email), req.Password)
	respondTokens(w, r, pair, err)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req authapi.RefreshRequest
	if !parse(w, r, &req, func(d details) { d.required("refreshToken", req.RefreshToken) }) {
		return
	}
	pair, err := h.svc.Refresh(r.Context(), req.RefreshToken)
	respondTokens(w, r, pair, err)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req authapi.RefreshRequest
	if !parse(w, r, &req, func(d details) { d.required("refreshToken", req.RefreshToken) }) {
		return
	}
	if err := h.svc.Logout(r.Context(), UserIDFromContext(r), req.RefreshToken); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	var req authapi.EmailRequest
	if !parse(w, r, &req, func(d details) { d.email(req.Email) }) {
		return
	}
	if err := h.svc.ForgotPassword(r.Context(), normalizeEmail(req.Email)); err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	accepted(w, r)
}

func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var req authapi.ResetPasswordRequest
	if !parse(w, r, &req, func(d details) {
		d.email(req.Email)
		d.code(req.Code)
		d.password("newPassword", req.NewPassword)
	}) {
		return
	}
	pair, err := h.svc.ResetPassword(r.Context(), normalizeEmail(req.Email), req.Code, req.NewPassword)
	respondTokens(w, r, pair, err)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	u, err := h.svc.Me(r.Context(), UserIDFromContext(r))
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	httpx.OkJsonCtx(r.Context(), w, authapi.User{Id: u.ID, Email: u.Email, CreatedAt: u.CreatedAt})
}

func Health(w http.ResponseWriter, r *http.Request) {
	httpx.OkJsonCtx(r.Context(), w, map[string]string{"status": "ok"})
}

// UserIDFromContext — id пользователя, который положил в контекст JWT-middleware go-zero (claim uid).
func UserIDFromContext(r *http.Request) string {
	id, _ := r.Context().Value(ClaimUserID).(string)
	return id
}

// ---- разбор и валидация ----

// parse читает JSON-тело и прогоняет validate; при ошибке пишет 400 и возвращает false.
// encoding/json, а не httpx.ParseJsonBody: go-zero считает поля обязательными по умолчанию,
// а сообщения об ошибках валидации нужны в нашем формате (details: поле → текст).
func parse(w http.ResponseWriter, r *http.Request, req any, validate func(details)) bool {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		httpx.ErrorCtx(r.Context(), w, ValidationError(map[string]string{"body": "malformed json"}))
		return false
	}
	d := details{}
	validate(d)
	if len(d) > 0 {
		httpx.ErrorCtx(r.Context(), w, ValidationError(d))
		return false
	}
	return true
}

type details map[string]string

var codeRe = regexp.MustCompile(`^[0-9]{6}$`)

func (d details) email(v string) {
	v = normalizeEmail(v)
	if v == "" || len(v) > 255 {
		d["email"] = "must be a valid email"
		return
	}
	addr, err := mail.ParseAddress(v)
	if err != nil || addr.Address != v {
		d["email"] = "must be a valid email"
	}
}

// password: 8–72 байта — верхняя граница bcrypt.
func (d details) password(field, v string) {
	if len(v) < 8 || len(v) > 72 {
		d[field] = "must be 8-72 characters"
	}
}

func (d details) code(v string) {
	if !codeRe.MatchString(v) {
		d["code"] = "must be 6 digits"
	}
}

func (d details) required(field, v string) {
	if v == "" {
		d[field] = "is required"
	}
}

func normalizeEmail(v string) string { return strings.ToLower(strings.TrimSpace(v)) }

// ---- ответы ----

func accepted(w http.ResponseWriter, r *http.Request) {
	httpx.WriteJsonCtx(r.Context(), w, http.StatusAccepted, authapi.Empty{})
}

func respondTokens(w http.ResponseWriter, r *http.Request, pair *TokenPair, err error) {
	if err != nil {
		httpx.ErrorCtx(r.Context(), w, err)
		return
	}
	httpx.OkJsonCtx(r.Context(), w, authapi.TokenPair{
		AccessToken:  pair.AccessToken,
		RefreshToken: pair.RefreshToken,
		ExpiresIn:    pair.ExpiresIn,
	})
}
