package middleware

import (
	"strings"

	"devtracker/backend/internal/auth"
	"devtracker/backend/internal/config"
	apperrors "devtracker/backend/pkg/errors"

	"github.com/gofiber/fiber/v2"
)

func JWTAuth(cfg config.JWTConfig) fiber.Handler {
	tokenManager := auth.NewTokenManager(cfg)

	return func(c *fiber.Ctx) error {
		tokenString := bearerToken(c)
		if tokenString == "" {
			return apperrors.Unauthorized("missing bearer token")
		}

		claims, err := tokenManager.Parse(tokenString)
		if err != nil {
			return auth.TokenFromError(err)
		}

		c.Locals(auth.LocalUserID, claims.UserID)
		c.Locals(auth.LocalEmail, claims.Email)
		c.Locals(auth.LocalName, claims.Name)
		c.Locals(auth.LocalRole, claims.Role)

		return c.Next()
	}
}

func RequireRoles(roles ...string) fiber.Handler {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[strings.ToLower(role)] = struct{}{}
	}

	return func(c *fiber.Ctx) error {
		currentRole, ok := c.Locals(auth.LocalRole).(string)
		if !ok || currentRole == "" {
			return apperrors.Forbidden("role is required")
		}

		if _, ok := allowed[strings.ToLower(currentRole)]; !ok {
			return apperrors.Forbidden("insufficient permissions")
		}

		return c.Next()
	}
}

func bearerToken(c *fiber.Ctx) string {
	header := c.Get(fiber.HeaderAuthorization)
	if strings.HasPrefix(strings.ToLower(header), "bearer ") {
		return strings.TrimSpace(header[7:])
	}

	if cookie := strings.TrimSpace(c.Cookies("access_token")); cookie != "" {
		return cookie
	}

	return ""
}
