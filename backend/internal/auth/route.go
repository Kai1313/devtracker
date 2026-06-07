package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func RegisterRoutes(router fiber.Router, handler *Handler, authMiddleware fiber.Handler) {
	group := router.Group("/auth")
	loginLimiter := limiter.New(limiter.Config{
		Max:        5,
		Expiration: time.Minute,
	})

	// @Summary Login
	// @Tags Auth
	// @Accept json
	// @Produce json
	// @Param payload body LoginRequest true "Login credentials"
	// @Success 200 {object} response.Body
	// @Router /auth/login [post]
	group.Post("/login", loginLimiter, handler.Login)

	// @Summary Logout
	// @Tags Auth
	// @Security BearerAuth
	// @Success 200 {object} response.Body
	// @Router /auth/logout [post]
	group.Post("/logout", authMiddleware, handler.Logout)

	// @Summary Bootstrap admin
	// @Tags Auth
	// @Accept json
	// @Produce json
	// @Param payload body BootstrapAdminRequest true "Initial admin account"
	// @Success 201 {object} response.Body
	// @Router /auth/bootstrap [post]
	group.Post("/bootstrap", loginLimiter, handler.BootstrapAdmin)

	// @Summary Current user
	// @Tags Auth
	// @Security BearerAuth
	// @Success 200 {object} response.Body
	// @Router /auth/me [get]
	group.Get("/me", authMiddleware, handler.Me)
}
