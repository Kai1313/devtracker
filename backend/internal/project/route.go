package project

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, requirePermission func(...string) fiber.Handler) {
	group := router.Group("/projects", authMiddleware, requirePermission("manage_projects"))

	// @Summary List projects
	// @Tags Projects
	// @Security BearerAuth
	// @Router /projects [get]
	group.Get("/", handler.List)

	// @Summary Get project
	// @Tags Projects
	// @Security BearerAuth
	// @Param id path string true "Project UUID"
	// @Router /projects/{id} [get]
	group.Get("/:id", handler.Get)

	// @Summary Create project
	// @Tags Projects
	// @Security BearerAuth
	// @Param payload body CreateProjectRequest true "Project payload"
	// @Router /projects [post]
	group.Post("/", handler.Create)

	// @Summary Update project
	// @Tags Projects
	// @Security BearerAuth
	// @Param id path string true "Project UUID"
	// @Param payload body UpdateProjectRequest true "Project payload"
	// @Router /projects/{id} [patch]
	group.Patch("/:id", handler.Update)

	// @Summary Delete project
	// @Tags Projects
	// @Security BearerAuth
	// @Param id path string true "Project UUID"
	// @Router /projects/{id} [delete]
	group.Delete("/:id", handler.Delete)
}
