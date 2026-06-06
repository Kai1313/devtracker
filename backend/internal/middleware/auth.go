package middleware

import (
	"errors"
	"strings"

	"devtracker/backend/internal/auth"
	"devtracker/backend/internal/config"
	"devtracker/backend/internal/user"
	apperrors "devtracker/backend/pkg/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func JWTAuth(cfg config.JWTConfig, users user.Repository) fiber.Handler {
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

		userID, err := uuid.Parse(claims.UserID)
		if err != nil {
			return apperrors.Unauthorized("invalid token claims")
		}

		account, err := users.FindByID(c.UserContext(), userID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.Unauthorized("user no longer exists")
			}

			return err
		}

		if !account.IsActive {
			return apperrors.Forbidden("user account is inactive")
		}

		c.Locals(auth.LocalUserID, account.ID.String())
		c.Locals(auth.LocalEmail, account.Email)
		c.Locals(auth.LocalName, account.Name)
		c.Locals(auth.LocalRole, account.Role.Name)

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
