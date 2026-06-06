package auth

import (
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
		return err
	}

	if err := appvalidator.Struct(req); err != nil {
		return err
	}

	result, err := h.service.Login(c.Context(), req)
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
		return err
	}

	if err := appvalidator.Struct(req); err != nil {
		return err
	}

	result, err := h.service.BootstrapAdmin(c.Context(), req)
	if err != nil {
		return err
	}

	return response.Created(c, "bootstrap admin created", result)
}

func (h *Handler) Me(c *fiber.Ctx) error {
	userID, err := uuid.Parse(c.Locals(LocalUserID).(string))
	if err != nil {
		return err
	}

	result, err := h.service.Me(c.Context(), userID)
	if err != nil {
		return err
	}

	return response.OK(c, "current user retrieved", result)
}
