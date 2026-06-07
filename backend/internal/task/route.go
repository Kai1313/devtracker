package task

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, requirePermission func(...string) fiber.Handler) {
	group := router.Group("/tasks", authMiddleware)
	canView := requirePermission("manage_tasks", "view_assigned_tasks", "view_ready_to_check_tasks")
	canUpdate := requirePermission("manage_tasks", "update_own_task_status", "update_qa_status")
	canManage := requirePermission("manage_tasks")

	// @Summary List tasks
	// @Tags Tasks
	// @Security BearerAuth
	// @Router /tasks [get]
	group.Get("/", canView, handler.List)

	// @Summary List task histories
	// @Tags Tasks
	// @Security BearerAuth
	// @Param id path string true "Task UUID"
	// @Router /tasks/{id}/histories [get]
	group.Get("/:id/histories", canView, handler.ListHistories)

	// @Summary Get task
	// @Tags Tasks
	// @Security BearerAuth
	// @Param id path string true "Task UUID"
	// @Router /tasks/{id} [get]
	group.Get("/:id", canView, handler.Get)

	// @Summary Create task
	// @Tags Tasks
	// @Security BearerAuth
	// @Param payload body CreateTaskRequest true "Task payload"
	// @Router /tasks [post]
	group.Post("/", canManage, handler.Create)

	// @Summary Update task
	// @Tags Tasks
	// @Security BearerAuth
	// @Param id path string true "Task UUID"
	// @Param payload body UpdateTaskRequest true "Task payload"
	// @Router /tasks/{id} [patch]
	group.Patch("/:id", canUpdate, handler.Update)

	// @Summary Delete task
	// @Tags Tasks
	// @Security BearerAuth
	// @Param id path string true "Task UUID"
	// @Router /tasks/{id} [delete]
	group.Delete("/:id", canManage, handler.Delete)
}
