package audit

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
	userID := c.Query("user")
	if userID == "" {
		userID = c.Query("user_id")
	}

	result, meta, err := h.service.ListWithScope(c.UserContext(), ListQuery{
		Page:      c.QueryInt("page", 1),
		Limit:     c.QueryInt("limit", 20),
		UserID:    userID,
		Module:    c.Query("module"),
		Action:    c.Query("action"),
		StartDate: c.Query("start_date"),
		EndDate:   c.Query("end_date"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}, auditScope(c))
	if err != nil {
		return err
	}

	return response.WithMeta(c, "audit logs retrieved", result, meta)
}

func auditScope(c *fiber.Ctx) ListScope {
	if hasLocalRole(c, "admin") {
		return ListScope{CanViewAll: true}
	}

	if hasLocalRole(c, "project_manager") {
		return ListScope{AllowedModules: projectManagerAuditModules}
	}

	return ListScope{}
}

func hasLocalRole(c *fiber.Ctx, roles ...string) bool {
	allowed := map[string]struct{}{}
	for _, role := range roles {
		normalized := normalize(role)
		if normalized != "" {
			allowed[normalized] = struct{}{}
		}
	}

	for role := range localRoles(c) {
		if _, ok := allowed[role]; ok {
			return true
		}
	}

	return false
}

func localRoles(c *fiber.Ctx) map[string]struct{} {
	roles := map[string]struct{}{}
	switch value := c.Locals(httpx.LocalRoles).(type) {
	case []string:
		for _, role := range value {
			if normalized := normalize(role); normalized != "" {
				roles[normalized] = struct{}{}
			}
		}
	case string:
		if normalized := normalize(value); normalized != "" {
			roles[normalized] = struct{}{}
		}
	}

	if len(roles) == 0 {
		if role, ok := c.Locals(httpx.LocalRole).(string); ok {
			if normalized := normalize(role); normalized != "" {
				roles[normalized] = struct{}{}
			}
		}
	}

	return roles
}
