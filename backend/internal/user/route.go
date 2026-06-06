package user

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, adminOnly fiber.Handler) {
	group := router.Group("/users", authMiddleware)

	group.Get("/", handler.List)
	group.Get("/:id", handler.Get)
	group.Post("/", adminOnly, handler.Create)
	group.Patch("/:id", adminOnly, handler.Update)
	group.Delete("/:id", adminOnly, handler.Delete)
}
