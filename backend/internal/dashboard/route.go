package dashboard

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler) {
	group := router.Group("/dashboard", authMiddleware)

	group.Get("/summary", handler.Summary)
}
