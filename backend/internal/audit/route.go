package audit

import "github.com/gofiber/fiber/v2"

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler, auditViewer fiber.Handler) {
	group := router.Group("/audit-logs", authMiddleware, auditViewer)

	// @Summary List audit logs
	// @Tags Audit Logs
	// @Security BearerAuth
	// @Router /audit-logs [get]
	group.Get("/", handler.List)
}
