package workload

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

func (h *Handler) DeveloperWorkload(c *fiber.Ctx) error {
	sprintID := c.Query("sprint")
	if sprintID == "" {
		sprintID = c.Query("sprint_id")
	}

	projectID := c.Query("project")
	if projectID == "" {
		projectID = c.Query("project_id")
	}

	result, err := h.service.DeveloperWorkload(c.UserContext(), Query{
		SprintID:  sprintID,
		ProjectID: projectID,
	})
	if err != nil {
		return err
	}

	return response.OK(c, "developer workload retrieved", result)
}
