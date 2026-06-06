package httpx

import (
	apperrors "devtracker/backend/pkg/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func ParseUUIDParam(c *fiber.Ctx, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params(name))
	if err != nil {
		return uuid.Nil, apperrors.BadRequest(name + " must be a valid UUID")
	}

	return id, nil
}

func CurrentUserID(c *fiber.Ctx) (uuid.UUID, error) {
	raw, ok := c.Locals(LocalUserID).(string)
	if !ok || raw == "" {
		return uuid.Nil, apperrors.Unauthorized("authenticated user is missing")
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperrors.Unauthorized("authenticated user is invalid")
	}

	return id, nil
}
