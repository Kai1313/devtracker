package dashboard

import (
	"devtracker/backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Summary(c *fiber.Ctx) error {
	result, err := h.service.Summary(c.UserContext(), SummaryQuery{
		SprintID: c.Query("sprint_id"),
	})
	if err != nil {
		return err
	}

	return response.OK(c, "dashboard summary retrieved", result)
}
