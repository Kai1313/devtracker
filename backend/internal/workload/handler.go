package workload

import (
	"devtracker/backend/internal/audit"
	"devtracker/backend/internal/httpx"
	appmiddleware "devtracker/backend/internal/middleware"
	"devtracker/backend/pkg/response"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Handler struct {
	service *Service
	audit   *audit.Service
}

func NewHandler(service *Service, auditService *audit.Service) *Handler {
	return &Handler{service: service, audit: auditService}
}

func (h *Handler) DeveloperWorkload(c *fiber.Ctx) error {
	userID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	sprintID := c.Query("sprint")
	if sprintID == "" {
		sprintID = c.Query("sprint_id")
	}

	projectID := c.Query("project")
	if projectID == "" {
		projectID = c.Query("project_id")
	}

	query := Query{
		SprintID:    sprintID,
		ProjectID:   projectID,
		DeveloperID: c.Query("developer_id"),
		StatusID:    c.Query("status_id"),
		StartDate:   c.Query("start_date"),
		EndDate:     c.Query("end_date"),
	}

	result, err := h.service.DeveloperWorkloadWithScope(c.UserContext(), query, workloadScopeForUser(c, userID))
	if err != nil {
		return err
	}

	if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
		UserID: &userID,
		Module: "workload",
		Action: "view",
		NewValue: map[string]any{
			"filters": query,
			"count":   len(result),
		},
	}); err != nil {
		return err
	}

	return response.OK(c, "developer workload retrieved", result)
}

func workloadScopeForUser(c *fiber.Ctx, userID uuid.UUID) AccessScope {
	return AccessScope{
		UserID:       userID,
		IsAdmin:      appmiddleware.HasRole(c, "admin"),
		IsManager:    appmiddleware.HasRole(c, "project_manager"),
		IsManagement: appmiddleware.HasRole(c, "management"),
		IsDeveloper:  appmiddleware.HasRole(c, "developer"),
		IsQA:         appmiddleware.HasRole(c, "qa"),
	}
}
