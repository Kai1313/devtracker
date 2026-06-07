package task

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, requirePermission func(...string) fiber.Handler) {
	group := router.Group("/tasks", authMiddleware)
	canView := requirePermission("manage_tasks", "view_assigned_tasks")
	canUpdate := requirePermission("manage_tasks", "update_own_task_status", "update_qa_status")
	canManage := requirePermission("manage_tasks")

	group.Get("/", canView, handler.List)
	group.Get("/:id/histories", canView, handler.ListHistories)
	group.Get("/:id", canView, handler.Get)
	group.Post("/", canManage, handler.Create)
	group.Patch("/:id", canUpdate, handler.Update)
	group.Delete("/:id", canManage, handler.Delete)
}
