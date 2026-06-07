package kpi

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, requirePermission func(...string) fiber.Handler) {
	group := router.Group("/kpi", authMiddleware, requirePermission("view_kpi"))

	// @Summary Developer KPI
	// @Tags KPI
	// @Security BearerAuth
	// @Router /kpi/developers [get]
	group.Get("/developers", handler.Developers)

	// @Summary Project KPI
	// @Tags KPI
	// @Security BearerAuth
	// @Router /kpi/projects [get]
	group.Get("/projects", handler.Projects)
}
