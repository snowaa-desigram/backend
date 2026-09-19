package transport

import (
	"context"
	"errors"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"

	authapi "github.com/snowaa-desigram/backend/services/go/gen/openapi/auth"
	"github.com/snowaa-desigram/backend/services/go/gen/openapi/common"
	"github.com/snowaa-desigram/backend/services/go/internal/auth/service"
)

// ValidationError — 400 с полем → сообщение.
func ValidationError(details map[string]string) *service.Error {
	return &service.Error{Status: http.StatusBadRequest, Code: common.ErrorCodeValidation, Message: "invalid request", Details: details}
}

// ErrorHandler для httpx.SetErrorHandlerCtx: *Error → свой статус, всё остальное — 500 без деталей.
func ErrorHandler(ctx context.Context, err error) (int, any) {
	var e *service.Error
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
	httpx.WriteJsonCtx(r.Context(), w, service.ErrUnauthorized.Status, authapi.Error{Code: service.ErrUnauthorized.Code, Message: service.ErrUnauthorized.Message})
}
