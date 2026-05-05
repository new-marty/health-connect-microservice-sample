package apperror

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response is the wire-format error envelope. Stable shape so LLM tool callers
// can parse {error.code, error.message, error.details} reliably.
//
// swagger:model ErrorResponse
type Response struct {
	Error Body `json:"error"`
}

type Body struct {
	Code    ErrorCode         `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

func RespondGin(c *gin.Context, err error) {
	var appErr *AppError
	if !errors.As(err, &appErr) {
		c.JSON(http.StatusInternalServerError, Response{Body{
			Code:    CodeInternal,
			Message: err.Error(),
		}})
		return
	}

	status := statusFor(appErr.Code)
	c.JSON(status, Response{Body{
		Code:    appErr.Code,
		Message: appErr.Message,
		Details: appErr.Fields,
	}})
}

func statusFor(code ErrorCode) int {
	switch code {
	case CodeNotFound:
		return http.StatusNotFound
	case CodeInvalidInput:
		return http.StatusBadRequest
	case CodeTimeout:
		return http.StatusGatewayTimeout
	default:
		return http.StatusInternalServerError
	}
}
