package status

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, requirePermission func(...string) fiber.Handler) {
	group := router.Group("/statuses", authMiddleware, requirePermission("manage_task_statuses"))

	// @Summary List task statuses
	// @Tags Task Statuses
	// @Security BearerAuth
	// @Router /statuses [get]
	group.Get("/", handler.List)

	// @Summary Get task status
	// @Tags Task Statuses
	// @Security BearerAuth
	// @Param id path string true "Task status UUID"
	// @Router /statuses/{id} [get]
	group.Get("/:id", handler.Get)

	// @Summary Create task status
	// @Tags Task Statuses
	// @Security BearerAuth
	// @Param payload body CreateTaskStatusRequest true "Task status payload"
	// @Router /statuses [post]
	group.Post("/", handler.Create)

	// @Summary Update task status
	// @Tags Task Statuses
	// @Security BearerAuth
	// @Param id path string true "Task status UUID"
	// @Param payload body UpdateTaskStatusRequest true "Task status payload"
	// @Router /statuses/{id} [patch]
	group.Patch("/:id", handler.Update)

	// @Summary Delete task status
	// @Tags Task Statuses
	// @Security BearerAuth
	// @Param id path string true "Task status UUID"
	// @Router /statuses/{id} [delete]
	group.Delete("/:id", handler.Delete)
}
