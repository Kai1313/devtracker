package workload

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, requirePermission func(...string) fiber.Handler) {
	group := router.Group("/workload", authMiddleware, requirePermission("view_kpi", "manage_tasks"))

	group.Get("/", handler.DeveloperWorkload)
}
