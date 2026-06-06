package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

func RequestLogger(log zerolog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		startedAt := time.Now()
		err := c.Next()

		event := log.Info()
		if err != nil || c.Response().StatusCode() >= fiber.StatusInternalServerError {
			event = log.Error()
		}

		event.
			Err(err).
			Str("request_id", c.GetRespHeader(fiber.HeaderXRequestID)).
			Str("method", c.Method()).
			Str("path", c.Path()).
			Int("status", c.Response().StatusCode()).
			Dur("latency", time.Since(startedAt)).
			Msg("request completed")

		return err
	}
}
