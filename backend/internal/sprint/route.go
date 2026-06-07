package sprint

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, requirePermission func(...string) fiber.Handler) {
	group := router.Group("/sprints", authMiddleware, requirePermission("manage_sprints"))

	group.Get("/", handler.List)
	group.Get("/:id", handler.Get)
	group.Post("/", handler.Create)
	group.Patch("/:id/close", handler.Close)
	group.Patch("/:id", handler.Update)
	group.Delete("/:id", handler.Delete)
}
