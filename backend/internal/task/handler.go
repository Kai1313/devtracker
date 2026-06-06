package task

import (
	"devtracker/backend/internal/auth"
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
	query := ListTasksQuery{
		Page:        c.QueryInt("page", 1),
		Limit:       c.QueryInt("limit", 20),
		DeveloperID: c.Query("developer_id"),
		ProjectID:   c.Query("project_id"),
		SprintID:    c.Query("sprint_id"),
		StatusID:    c.Query("status_id"),
		Search:      c.Query("search"),
	}

	result, meta, err := h.service.List(c.UserContext(), query)
	if err != nil {
		return err
	}

	return response.WithMeta(c, "tasks retrieved", result, meta)
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

	return response.OK(c, "task retrieved", result)
}

func (h *Handler) ListHistories(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	result, err := h.service.ListHistories(c.UserContext(), id)
	if err != nil {
		return err
	}

	return response.OK(c, "task histories retrieved", result)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	actorID, err := currentUserID(c)
	if err != nil {
		return err
	}

	var req CreateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}

	if err := appvalidator.Struct(req); err != nil {
		return err
	}

	result, err := h.service.Create(c.UserContext(), req, actorID)
	if err != nil {
		return err
	}

	return response.Created(c, "task created", result)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	actorID, err := currentUserID(c)
	if err != nil {
		return err
	}

	id, err := parseID(c)
	if err != nil {
		return err
	}

	var req UpdateTaskRequest
	if err := c.BodyParser(&req); err != nil {
		return apperrors.BadRequest("invalid request body")
	}

	if err := appvalidator.Struct(req); err != nil {
		return err
	}

	result, err := h.service.Update(c.UserContext(), id, req, actorID)
	if err != nil {
		return err
	}

	return response.OK(c, "task updated", result)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}

	if err := h.service.Delete(c.UserContext(), id); err != nil {
		return err
	}

	return response.OK(c, "task deleted", nil)
}

func parseID(c *fiber.Ctx) (uuid.UUID, error) {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return uuid.Nil, apperrors.BadRequest("id must be a valid UUID")
	}

	return id, nil
}

func currentUserID(c *fiber.Ctx) (uuid.UUID, error) {
	raw, ok := c.Locals(auth.LocalUserID).(string)
	if !ok || raw == "" {
		return uuid.Nil, apperrors.Unauthorized("authenticated user is missing")
	}

	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, apperrors.Unauthorized("authenticated user is invalid")
	}

	return id, nil
}
