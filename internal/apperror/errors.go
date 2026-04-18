package apperror

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	CodeNotFound     ErrorCode = "NOT_FOUND"
	CodeInvalidInput ErrorCode = "INVALID_INPUT"
	CodeInternal     ErrorCode = "INTERNAL_ERROR"
	CodeTimeout      ErrorCode = "TIMEOUT"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Err     error
	Fields  map[string]string
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NotFound(resource string) *AppError {
	return &AppError{Code: CodeNotFound, Message: fmt.Sprintf("%s not found", resource)}
}

func NotFoundWithErr(resource string, err error) *AppError {
	return &AppError{Code: CodeNotFound, Message: fmt.Sprintf("%s not found", resource), Err: err}
}

func InvalidInput(message string) *AppError {
	return &AppError{Code: CodeInvalidInput, Message: message}
}

func InvalidInputWithErr(message string, err error) *AppError {
	return &AppError{Code: CodeInvalidInput, Message: message, Err: err}
}

func InvalidInputFields(fields map[string]string) *AppError {
	return &AppError{Code: CodeInvalidInput, Message: "validation failed", Fields: fields}
}

func Internal(message string) *AppError {
	return &AppError{Code: CodeInternal, Message: message}
}

func InternalWithErr(message string, err error) *AppError {
	return &AppError{Code: CodeInternal, Message: message, Err: err}
}

func Timeout(message string) *AppError {
	return &AppError{Code: CodeTimeout, Message: message}
}

func IsNotFound(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == CodeNotFound
	}
	return false
}

func IsInvalidInput(err error) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == CodeInvalidInput
	}
	return false
}
