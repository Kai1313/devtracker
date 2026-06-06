package auth

import (
	apperrors "devtracker/backend/pkg/errors"
	"devtracker/backend/pkg/response"
	appvalidator "devtracker/backend/pkg/validator"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}

	if err := appvalidator.Struct(req); err != nil {
		return err
	}

	result, err := h.service.Login(c.UserContext(), req)
	if err != nil {
		return err
	}

	return response.OK(c, "login successful", result)
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	return response.OK(c, "logout successful", nil)
}

func (h *Handler) BootstrapAdmin(c *fiber.Ctx) error {
	var req BootstrapAdminRequest
	if err := c.BodyParser(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}

	if err := appvalidator.Struct(req); err != nil {
		return err
	}

	result, err := h.service.BootstrapAdmin(c.UserContext(), req)
	if err != nil {
		return err
	}

	return response.Created(c, "bootstrap admin created", result)
}

func (h *Handler) Me(c *fiber.Ctx) error {
	userID, err := currentUserID(c)
	if err != nil {
		return err
	}

	result, err := h.service.Me(c.UserContext(), userID)
	if err != nil {
		return err
	}

	return response.OK(c, "current user retrieved", result)
}

func currentUserID(c *fiber.Ctx) (uuid.UUID, error) {
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
