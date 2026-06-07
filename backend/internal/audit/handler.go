package audit

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

func (h *Handler) List(c *fiber.Ctx) error {
	userID := c.Query("user")
	if userID == "" {
		userID = c.Query("user_id")
	}

	result, meta, err := h.service.List(c.UserContext(), ListQuery{
		Page:      c.QueryInt("page", 1),
		Limit:     c.QueryInt("limit", 20),
		UserID:    userID,
		Module:    c.Query("module"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
	})
	if err != nil {
		return err
	}

	return response.WithMeta(c, "audit logs retrieved", result, meta)
}
