package user

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, requirePermission func(...string) fiber.Handler) {
	group := router.Group("/users", authMiddleware, requirePermission("manage_users"))

	// @Summary List users
	// @Tags Users
	// @Security BearerAuth
	// @Router /users [get]
	group.Get("/", handler.List)

	// @Summary Get user
	// @Tags Users
	// @Security BearerAuth
	// @Param id path string true "User UUID"
	// @Router /users/{id} [get]
	group.Get("/:id", handler.Get)

	// @Summary Create user
	// @Tags Users
	// @Security BearerAuth
	// @Param payload body CreateUserRequest true "User payload"
	// @Router /users [post]
	group.Post("/", handler.Create)

	// @Summary Update user
	// @Tags Users
	// @Security BearerAuth
	// @Param id path string true "User UUID"
	// @Param payload body UpdateUserRequest true "User payload"
	// @Router /users/{id} [patch]
	group.Patch("/:id", handler.Update)

	// @Summary Delete user
	// @Tags Users
	// @Security BearerAuth
	// @Param id path string true "User UUID"
	// @Router /users/{id} [delete]
	group.Delete("/:id", handler.Delete)
}
