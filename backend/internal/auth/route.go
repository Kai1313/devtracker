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

	group.Post("/login", loginLimiter, handler.Login)
	group.Post("/logout", authMiddleware, handler.Logout)
	group.Post("/bootstrap", loginLimiter, handler.BootstrapAdmin)
	group.Get("/me", authMiddleware, handler.Me)
}
