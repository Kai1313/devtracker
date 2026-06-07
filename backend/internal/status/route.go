package status

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, requirePermission func(...string) fiber.Handler) {
	group := router.Group("/statuses", authMiddleware, requirePermission("manage_task_statuses"))

	group.Get("/", handler.List)
	group.Get("/:id", handler.Get)
	group.Post("/", handler.Create)
	group.Patch("/:id", handler.Update)
	group.Delete("/:id", handler.Delete)
}
