package dashboard

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, requirePermission func(...string) fiber.Handler) {
	group := router.Group("/dashboard", authMiddleware, requirePermission("view_dashboard"))

	// @Summary Dashboard summary
	// @Tags Dashboard
	// @Security BearerAuth
	// @Router /dashboard/summary [get]
	group.Get("/summary", handler.Summary)
}
