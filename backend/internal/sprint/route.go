package sprint

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, requirePermission func(...string) fiber.Handler) {
	group := router.Group("/sprints", authMiddleware, requirePermission("manage_sprints"))

	// @Summary List sprints
	// @Tags Sprints
	// @Security BearerAuth
	// @Router /sprints [get]
	group.Get("/", handler.List)

	// @Summary Get sprint
	// @Tags Sprints
	// @Security BearerAuth
	// @Param id path string true "Sprint UUID"
	// @Router /sprints/{id} [get]
	group.Get("/:id", handler.Get)

	// @Summary Create sprint
	// @Tags Sprints
	// @Security BearerAuth
	// @Param payload body CreateSprintRequest true "Sprint payload"
	// @Router /sprints [post]
	group.Post("/", handler.Create)

	// @Summary Close sprint
	// @Tags Sprints
	// @Security BearerAuth
	// @Param id path string true "Sprint UUID"
	// @Router /sprints/{id}/close [patch]
	group.Patch("/:id/close", handler.Close)

	// @Summary Update sprint
	// @Tags Sprints
	// @Security BearerAuth
	// @Param id path string true "Sprint UUID"
	// @Param payload body UpdateSprintRequest true "Sprint payload"
	// @Router /sprints/{id} [patch]
	group.Patch("/:id", handler.Update)

	// @Summary Delete sprint
	// @Tags Sprints
	// @Security BearerAuth
	// @Param id path string true "Sprint UUID"
	// @Router /sprints/{id} [delete]
	group.Delete("/:id", handler.Delete)
}
