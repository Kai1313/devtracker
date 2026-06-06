package sprint

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler) {
	group := router.Group("/sprints", authMiddleware)

	group.Get("/", handler.List)
	group.Get("/:id", handler.Get)
	group.Post("/", handler.Create)
	group.Patch("/:id/close", handler.Close)
	group.Patch("/:id", handler.Update)
	group.Delete("/:id", handler.Delete)
}
