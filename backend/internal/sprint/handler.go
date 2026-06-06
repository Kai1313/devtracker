package sprint

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

func (h *Handler) List(c *fiber.Ctx) error {
	query := ListSprintsQuery{
		Page:      c.QueryInt("page", 1),
		Limit:     c.QueryInt("limit", 20),
		ProjectID: c.Query("project_id"),
		Status:    c.Query("status"),
	}

	result, meta, err := h.service.List(c.UserContext(), query)
	if err != nil {
		return err
	}

	return response.WithMeta(c, "sprints retrieved", result, meta)
}

func (h *Handler) Get(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	result, err := h.service.Get(c.UserContext(), id)
	if err != nil {
		return err
	}

	return response.OK(c, "sprint retrieved", result)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	var req CreateSprintRequest
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

	return response.Created(c, "sprint created", result)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	var req UpdateSprintRequest
	if err := c.BodyParser(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}

	if err := appvalidator.Struct(req); err != nil {
		return err
	}

	result, err := h.service.Update(c.UserContext(), id, req)
	if err != nil {
		return err
	}

	return response.OK(c, "sprint updated", result)
}

func (h *Handler) Close(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	result, err := h.service.Close(c.UserContext(), id)
	if err != nil {
		return err
	}

	return response.OK(c, "sprint closed", result)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	if err := h.service.Delete(c.UserContext(), id); err != nil {
		return err
	}

	return response.OK(c, "sprint deleted", nil)
}

func parseID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, apperrors.BadRequest("id must be a valid UUID")
	}

	return id, nil
}
