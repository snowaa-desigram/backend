package auth

import (
	"context"
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	authapi "github.com/snowaa-desigram/backend/services/go/gen/openapi/auth"
	"github.com/snowaa-desigram/backend/services/go/gen/openapi/common"
)

// Error — ошибка, которую сервис отдаёт клиенту в формате authapi.Error из OpenAPI.
type Error struct {
	Status  int
	Code    common.ErrorCode
	Message string
	Details map[string]string
}

func (e *Error) Error() string { return string(e.Code) + ": " + e.Message }

var (
	ErrEmailTaken         = &Error{Status: http.StatusConflict, Code: common.ErrorCodeEmailTaken, Message: "email is already registered"}
	ErrInvalidCredentials = &Error{Status: http.StatusUnauthorized, Code: common.ErrorCodeInvalidCredentials, Message: "invalid email or password"}
	ErrEmailNotVerified   = &Error{Status: http.StatusForbidden, Code: common.ErrorCodeEmailNotVerified, Message: "email is not verified"}
	ErrInvalidCode        = &Error{Status: http.StatusBadRequest, Code: common.ErrorCodeInvalidCode, Message: "invalid code"}
	ErrCodeExpired        = &Error{Status: http.StatusGone, Code: common.ErrorCodeCodeExpired, Message: "code expired, request a new one"}
	ErrInvalidToken       = &Error{Status: http.StatusUnauthorized, Code: common.ErrorCodeInvalidToken, Message: "invalid refresh token"}
	ErrUnauthorized       = &Error{Status: http.StatusUnauthorized, Code: common.ErrorCodeUnauthorized, Message: "unauthorized"}
	ErrTooManyAttempts    = &Error{Status: http.StatusTooManyRequests, Code: common.ErrorCodeTooManyAttempts, Message: "too many attempts, try later"}
	ErrTooManyRequests    = &Error{Status: http.StatusTooManyRequests, Code: common.ErrorCodeTooManyRequests, Message: "code was sent recently, try later"}
)

// ValidationError — 400 с полем → сообщение.
func ValidationError(details map[string]string) *Error {
	return &Error{Status: http.StatusBadRequest, Code: common.ErrorCodeValidation, Message: "invalid request", Details: details}
}

// ErrorHandler для httpx.SetErrorHandlerCtx: *Error → свой статус, всё остальное — 500 без деталей.
func ErrorHandler(ctx context.Context, err error) (int, any) {
	var e *Error
	if errors.As(err, &e) {
		body := authapi.Error{Code: e.Code, Message: e.Message}
		if len(e.Details) > 0 {
			body.Details = &e.Details
		}
		return e.Status, body
	}

	logx.WithContext(ctx).Errorf("internal error: %v", err)
	return http.StatusInternalServerError, authapi.Error{Code: common.ErrorCodeInternal, Message: "internal error"}
}

// UnauthorizedCallback — ответ JWT-middleware go-zero в формате authapi.Error.
func UnauthorizedCallback(w http.ResponseWriter, r *http.Request, _ error) {
	httpx.WriteJsonCtx(r.Context(), w, ErrUnauthorized.Status, authapi.Error{Code: ErrUnauthorized.Code, Message: ErrUnauthorized.Message})
}
