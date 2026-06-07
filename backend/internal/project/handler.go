package project

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

func (h *Handler) List(c *fiber.Ctx) error {
	query := ListProjectsQuery{
		Page:      c.QueryInt("page", 1),
		Limit:     c.QueryInt("limit", 20),
		Search:    c.Query("search"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	result, meta, err := h.service.List(c.UserContext(), query)
	if err != nil {
		return err
	}

	return response.WithMeta(c, "projects retrieved", result, meta)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	id, err := httpx.ParseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	result, err := h.service.Get(c.UserContext(), id)
	if err != nil {
		return err
	}

	return response.OK(c, "project retrieved", result)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	actorID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	var req CreateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}

	if err := appvalidator.Struct(req); err != nil {
		return err
	}

	result, err := h.service.Create(c.UserContext(), req)
	if err != nil {
		return err
	}

	if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
		UserID:   &actorID,
		Module:   "projects",
		Action:   "create",
		EntityID: &result.ID,
		NewValue: result,
	}); err != nil {
		return err
	}

	return response.Created(c, "project created", result)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := httpx.ParseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	actorID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	var req UpdateProjectRequest
	if err := c.BodyParser(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}

	if err := appvalidator.Struct(req); err != nil {
		return err
	}

	oldValue, err := h.service.Get(c.UserContext(), id)
	if err != nil {
		return err
	}

	result, err := h.service.Update(c.UserContext(), id, req)
	if err != nil {
		return err
	}

	if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
		UserID:   &actorID,
		Module:   "projects",
		Action:   "update",
		EntityID: &result.ID,
		OldValue: oldValue,
		NewValue: result,
	}); err != nil {
		return err
	}

	return response.OK(c, "project updated", result)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := httpx.ParseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	actorID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	oldValue, err := h.service.Get(c.UserContext(), id)
	if err != nil {
		return err
	}

	if err := h.service.Delete(c.UserContext(), id); err != nil {
		return err
	}

	if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
		UserID:   &actorID,
		Module:   "projects",
		Action:   "delete",
		EntityID: &id,
		OldValue: oldValue,
	}); err != nil {
		return err
	}

	return response.OK(c, "project deleted", nil)
}
