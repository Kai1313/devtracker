package notification

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler) {
	group := router.Group("/notifications", authMiddleware)

	// @Summary List notifications
	// @Tags Notifications
	// @Security BearerAuth
	// @Router /notifications [get]
	group.Get("/", handler.List)

	// @Summary Mark notification as read
	// @Tags Notifications
	// @Security BearerAuth
	// @Param id path string true "Notification UUID"
	// @Router /notifications/{id}/read [patch]
	group.Patch("/:id/read", handler.MarkRead)
}
