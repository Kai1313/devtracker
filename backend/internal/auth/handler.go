package auth

import (
	"devtracker/backend/internal/audit"
	"devtracker/backend/internal/httpx"
	apperrors "devtracker/backend/pkg/errors"
	"devtracker/backend/pkg/response"
	appvalidator "devtracker/backend/pkg/validator"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
	audit   *audit.Service
}

func NewHandler(service *Service, auditService *audit.Service) *Handler {
	return &Handler{service: service, audit: auditService}
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

	userID := result.User.ID
	if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
		UserID:   &userID,
		Module:   "auth",
		Action:   "login",
		EntityID: &userID,
		NewValue: result.User,
	}); err != nil {
		return err
	}

	return response.OK(c, "login successful", result)
}

func (h *Handler) Logout(c *fiber.Ctx) error {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
		UserID:   &userID,
		Module:   "auth",
		Action:   "logout",
		EntityID: &userID,
	}); err != nil {
		return err
	}

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

	userID := result.ID
	if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
		UserID:   &userID,
		Module:   "users",
		Action:   "create",
		EntityID: &userID,
		NewValue: result,
	}); err != nil {
		return err
	}

	return response.Created(c, "bootstrap admin created", result)
}

func (h *Handler) Me(c *fiber.Ctx) error {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	result, err := h.service.Me(c.UserContext(), userID)
	if err != nil {
		return err
	}

	return response.OK(c, "current user retrieved", result)
}
