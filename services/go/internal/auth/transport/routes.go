package transport

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

// Routes — таблица маршрутов 1:1 с backend/openapi/auth.yaml (проверяется тестом TestRoutesMatchSpec).
func Routes(h *Handler) (public, protected []rest.Route) {
	public = []rest.Route{
		{Method: http.MethodPost, Path: "/api/auth/register", Handler: h.Register},
		{Method: http.MethodPost, Path: "/api/auth/register/confirm", Handler: h.ConfirmRegistration},
		{Method: http.MethodPost, Path: "/api/auth/register/resend", Handler: h.ResendCode},
		{Method: http.MethodPost, Path: "/api/auth/login", Handler: h.Login},
		{Method: http.MethodPost, Path: "/api/auth/refresh", Handler: h.Refresh},
		{Method: http.MethodPost, Path: "/api/auth/password/forgot", Handler: h.ForgotPassword},
		{Method: http.MethodPost, Path: "/api/auth/password/reset", Handler: h.ResetPassword},
	}
	protected = []rest.Route{
		{Method: http.MethodPost, Path: "/api/auth/logout", Handler: h.Logout},
		{Method: http.MethodGet, Path: "/api/auth/me", Handler: h.Me},
	}
	return public, protected
}

// RegisterRoutes вешает маршруты на сервер: защищённые — под JWT-middleware go-zero.
func RegisterRoutes(server *rest.Server, h *Handler, jwtSecret string) {
	httpx.SetErrorHandlerCtx(ErrorHandler)

	public, protected := Routes(h)
	server.AddRoutes(public)
	server.AddRoutes(protected, rest.WithJwt(jwtSecret))
	server.AddRoute(rest.Route{Method: http.MethodGet, Path: "/health", Handler: Health})
}
