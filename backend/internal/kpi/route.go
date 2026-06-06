package kpi

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler) {
	group := router.Group("/kpi", authMiddleware)

	group.Get("/developers", handler.Developers)
	group.Get("/projects", handler.Projects)
}
