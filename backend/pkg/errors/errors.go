package errors

import (
	"errors"
	"net/http"
)

const (
	CodeBadRequest       = "bad_request"
	CodeUnauthorized     = "unauthorized"
	CodeForbidden        = "forbidden"
	CodeNotFound         = "not_found"
	CodeMethodNotAllowed = "method_not_allowed"
	CodeConflict         = "conflict"
	CodeValidation       = "validation_error"
	CodeTooManyRequests  = "too_many_requests"
	CodeInternal         = "internal_error"
)

type AppError struct {
	Status  int
	Code    string
	Message string
	Details any
	Err     error
}

func (e *AppError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	if e.Err != nil {
		return e.Err.Error()
	}

	return e.Code
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(status int, code string, message string, details any) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
		Details: details,
	}
}

func Wrap(err error, status int, code string, message string) *AppError {
	return &AppError{
		Status:  status,
		Code:    code,
		Message: message,
		Err:     err,
	}
}

func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, CodeBadRequest, message, nil)
}

func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, CodeUnauthorized, message, nil)
}

func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, CodeForbidden, message, nil)
}

func NotFound(message string) *AppError {
	return New(http.StatusNotFound, CodeNotFound, message, nil)
}

func Conflict(message string) *AppError {
	return New(http.StatusConflict, CodeConflict, message, nil)
}

func Validation(details any) *AppError {
	return New(http.StatusUnprocessableEntity, CodeValidation, "validation failed", details)
}

func Internal(err error) *AppError {
	return Wrap(err, http.StatusInternalServerError, CodeInternal, "internal server error")
}

func From(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}

	return Internal(err)
}
