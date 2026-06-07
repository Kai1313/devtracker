package kpi

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, requirePermission func(...string) fiber.Handler) {
	group := router.Group("/kpi", authMiddleware, requirePermission("view_kpi"))

	group.Get("/developers", handler.Developers)
	group.Get("/projects", handler.Projects)
}
