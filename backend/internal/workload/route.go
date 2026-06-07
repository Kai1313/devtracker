package workload

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, requireRole func(...string) fiber.Handler) {
	group := router.Group(
		"/workload",
		authMiddleware,
		requireRole("admin", "project_manager", "management", "developer", "qa"),
	)

	// @Summary Developer workload
	// @Tags Workload
	// @Security BearerAuth
	// @Router /workload [get]
	group.Get("/", handler.DeveloperWorkload)
}
