package kpi

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

func (h *Handler) Developers(c *fiber.Ctx) error {
	result, err := h.service.Developers(c.UserContext(), Query{
		SprintID: c.Query("sprint_id"),
	})
	if err != nil {
		return err
	}

	return response.OK(c, "developer KPI retrieved", result)
}

func (h *Handler) Projects(c *fiber.Ctx) error {
	result, err := h.service.Projects(c.UserContext(), Query{
		SprintID: c.Query("sprint_id"),
	})
	if err != nil {
		return err
	}

	return response.OK(c, "project KPI retrieved", result)
}
