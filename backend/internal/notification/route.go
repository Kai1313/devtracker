package notification

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler) {
	group := router.Group("/notifications", authMiddleware)

	group.Get("/", handler.List)
	group.Patch("/:id/read", handler.MarkRead)
}
