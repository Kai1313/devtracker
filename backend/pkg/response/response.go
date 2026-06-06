package response

import (
	"net/http"

	apperrors "devtracker/backend/pkg/errors"

	"github.com/gofiber/fiber/v2"
)

type Body struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
	Error   any    `json:"error,omitempty"`
}

type ErrorBody struct {
	Code    string `json:"code"`
	Details any    `json:"details,omitempty"`
}

func JSON(c *fiber.Ctx, status int, message string, data any, meta any) error {
	return c.Status(status).JSON(Body{
		Success: status < http.StatusBadRequest,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func OK(c *fiber.Ctx, message string, data any) error {
	return JSON(c, http.StatusOK, message, data, nil)
}

func Created(c *fiber.Ctx, message string, data any) error {
	return JSON(c, http.StatusCreated, message, data, nil)
}

func WithMeta(c *fiber.Ctx, message string, data any, meta any) error {
	return JSON(c, http.StatusOK, message, data, meta)
}

func Error(c *fiber.Ctx, err error) error {
	appErr := apperrors.From(err)

	return c.Status(appErr.Status).JSON(Body{
		Success: false,
		Message: appErr.Message,
		Error: ErrorBody{
			Code:    appErr.Code,
			Details: appErr.Details,
		},
	})
}

func ErrorHandler(c *fiber.Ctx, err error) error {
	if fiberErr, ok := err.(*fiber.Error); ok {
		return Error(c, apperrors.New(fiberErr.Code, apperrors.CodeBadRequest, fiberErr.Message, nil))
	}

	return Error(c, err)
}
