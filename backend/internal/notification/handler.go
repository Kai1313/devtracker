package notification

import (
	"devtracker/backend/internal/httpx"
	"devtracker/backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *fiber.Ctx) error {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	result, meta, err := h.service.List(c.UserContext(), ListQuery{
		Page:   c.QueryInt("page", 1),
		Limit:  c.QueryInt("limit", 20),
		UserID: userID,
	})
	if err != nil {
		return err
	}

	return response.WithMeta(c, "notifications retrieved", result, meta)
}

func (h *Handler) MarkRead(c *fiber.Ctx) error {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	id, err := httpx.ParseUUIDParam(c, "id")
	if err != nil {
		return err
	}

	result, err := h.service.MarkRead(c.UserContext(), id, userID)
	if err != nil {
		return err
	}

	return response.OK(c, "notification marked as read", result)
}
