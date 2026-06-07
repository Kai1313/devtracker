package kpi

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(
	router fiber.Router,
	handler *Handler,
	authMiddleware fiber.Handler,
	requirePermission func(...string) fiber.Handler,
	requireRole func(...string) fiber.Handler,
) {
	group := router.Group("/kpi", authMiddleware)
	canViewKPI := requirePermission("view_kpi")
	canViewSnapshots := requireRole("admin", "project_manager", "management", "developer")
	canGenerateSnapshots := requireRole("admin", "project_manager")

	// @Summary Developer KPI
	// @Tags KPI
	// @Security BearerAuth
	// @Router /kpi/developers [get]
	group.Get("/developers", canViewKPI, handler.Developers)

	// @Summary Project KPI
	// @Tags KPI
	// @Security BearerAuth
	// @Router /kpi/projects [get]
	group.Get("/projects", canViewKPI, handler.Projects)

	// @Summary List KPI snapshots
	// @Tags KPI
	// @Security BearerAuth
	// @Router /kpi/snapshots [get]
	group.Get("/snapshots", canViewSnapshots, handler.ListSnapshots)

	// @Summary Developer KPI snapshots
	// @Tags KPI
	// @Security BearerAuth
	// @Param developer_id path string true "Developer UUID"
	// @Router /kpi/snapshots/developer/{developer_id} [get]
	group.Get("/snapshots/developer/:developer_id", canViewSnapshots, handler.DeveloperSnapshots)

	// @Summary Generate KPI snapshots
	// @Tags KPI
	// @Security BearerAuth
	// @Param sprint_id path string true "Sprint UUID"
	// @Router /kpi/snapshots/generate/{sprint_id} [post]
	group.Post("/snapshots/generate/:sprint_id", canGenerateSnapshots, handler.GenerateSnapshots)
}
