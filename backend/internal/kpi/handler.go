package kpi

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

func (h *Handler) ListSnapshots(c *fiber.Ctx) error {
	result, err := h.service.ListSnapshots(c.UserContext(), SnapshotQuery{
		SprintID: c.Query("sprint_id"),
	}, snapshotScope(c))
	if err != nil {
		return err
	}

	return response.OK(c, "KPI snapshots retrieved", result)
}

func (h *Handler) DeveloperSnapshots(c *fiber.Ctx) error {
	developerID, err := httpx.ParseUUIDParam(c, "developer_id")
	if err != nil {
		return err
	}

	result, err := h.service.DeveloperSnapshots(c.UserContext(), developerID, snapshotScope(c))
	if err != nil {
		return err
	}

	return response.OK(c, "developer KPI snapshots retrieved", result)
}

func (h *Handler) GenerateSnapshots(c *fiber.Ctx) error {
	actorID, err := httpx.CurrentUserID(c)
	if err != nil {
		return err
	}

	sprintID, err := httpx.ParseUUIDParam(c, "sprint_id")
	if err != nil {
		return err
	}

	result, err := h.service.GenerateSnapshots(c.UserContext(), sprintID, snapshotScopeForUser(c, actorID))
	if err != nil {
		return err
	}

	if err := audit.RecordHTTPRequest(c, h.audit, audit.RecordInput{
		UserID:   &actorID,
		Module:   "kpi_snapshots",
		Action:   "generate",
		EntityID: &sprintID,
		NewValue: map[string]any{
			"sprint_id":      sprintID,
			"snapshot_count": len(result),
			"generated_by":   actorID,
		},
	}); err != nil {
		return err
	}

	return response.OK(c, "KPI snapshots generated", result)
}

func snapshotScope(c *fiber.Ctx) SnapshotScope {
	userID, _ := httpx.CurrentUserID(c)
	return snapshotScopeForUser(c, userID)
}

func snapshotScopeForUser(c *fiber.Ctx, userID uuid.UUID) SnapshotScope {
	return SnapshotScope{
		UserID:       userID,
		IsAdmin:      appmiddleware.HasRole(c, "admin"),
		IsManager:    appmiddleware.HasRole(c, "project_manager"),
		IsManagement: appmiddleware.HasRole(c, "management"),
		IsDeveloper:  appmiddleware.HasRole(c, "developer"),
	}
}
