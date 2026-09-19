package service

import (
	"net/http"

	"github.com/snowaa-desigram/backend/services/go/gen/openapi/common"
)

// Error — ошибка, которую сервис отдаёт клиенту в формате authapi.Error из OpenAPI (в HTTP превращает transport.ErrorHandler).
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
