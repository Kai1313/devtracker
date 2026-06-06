package middleware

import (
	"errors"
	"time"

	apperrors "devtracker/backend/pkg/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func RequestLogger(log zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		startedAt := time.Now()
		err := c.Next()
		status := responseStatus(c, err)

		event := log.Info()
		if err != nil || status >= fiber.StatusInternalServerError {
			event = log.Error()
		}

		event.
			Err(err).
			Str("request_id", c.GetRespHeader(fiber.HeaderXRequestID)).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", status).
			Dur("latency", time.Since(startedAt)).
			Msg("request completed")

		return err
	}
}

func responseStatus(c *fiber.Ctx, err error) int {
	if err == nil {
		return c.Response().StatusCode()
	}

	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		return appErr.Status
	}

	var fiberErr *fiber.Error
	if errors.As(err, &fiberErr) {
		return fiberErr.Code
	}

	return fiber.StatusInternalServerError
}
