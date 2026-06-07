package audit

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, adminOnly fiber.Handler) {
	group := router.Group("/audit-logs", authMiddleware, adminOnly)

	group.Get("/", handler.List)
}
