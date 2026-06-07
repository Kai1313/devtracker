package middleware

import (
	"errors"
	"strings"

	"devtracker/backend/internal/auth"
	"devtracker/backend/internal/config"
	"devtracker/backend/internal/httpx"
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

		roles, permissions := collectAccess(account)
		c.Locals(httpx.LocalUserID, account.ID.String())
		c.Locals(httpx.LocalEmail, account.Email)
		c.Locals(httpx.LocalName, account.Name)
		c.Locals(httpx.LocalRole, account.Role.Name)
		c.Locals(httpx.LocalRoles, roles)
		c.Locals(httpx.LocalPermissions, permissions)

		return c.Next()
	}
}

func RequireRole(roles ...string) fiber.Handler {
	allowed := normalizeSet(roles...)

	return func(c *fiber.Ctx) error {
		currentRoles := localSet(c, httpx.LocalRoles)
		if len(currentRoles) == 0 {
			currentRoles = normalizeSet(localString(c, httpx.LocalRole))
		}

		if _, ok := currentRoles["admin"]; ok {
			return c.Next()
		}

		for role := range allowed {
			if _, ok := currentRoles[role]; ok {
				return c.Next()
			}
		}

		return apperrors.Forbidden("insufficient permissions")
	}
}

func RequireRoles(roles ...string) fiber.Handler {
	return RequireRole(roles...)
}

func RequirePermission(permissions ...string) fiber.Handler {
	required := normalizeSet(permissions...)

	return func(c *fiber.Ctx) error {
		currentRoles := localSet(c, httpx.LocalRoles)
		if len(currentRoles) == 0 {
			currentRoles = normalizeSet(localString(c, httpx.LocalRole))
		}

		if _, ok := currentRoles["admin"]; ok {
			return c.Next()
		}

		currentPermissions := localSet(c, httpx.LocalPermissions)
		for permission := range required {
			if _, ok := currentPermissions[permission]; ok {
				return c.Next()
			}
		}

		return apperrors.Forbidden("insufficient permissions")
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

func collectAccess(account *user.User) ([]string, []string) {
	roles := map[string]struct{}{}
	permissions := map[string]struct{}{}

	addRoleAccess(roles, permissions, account.Role)
	for _, role := range account.Roles {
		addRoleAccess(roles, permissions, role)
	}

	return setValues(roles), setValues(permissions)
}

func addRoleAccess(roles, permissions map[string]struct{}, role user.Role) {
	roleName := normalize(role.Name)
	if roleName == "" {
		return
	}

	roles[roleName] = struct{}{}
	for _, permission := range role.Permissions {
		permissionName := normalize(permission.Name)
		if permissionName != "" {
			permissions[permissionName] = struct{}{}
		}
	}
}

func localSet(c *fiber.Ctx, key string) map[string]struct{} {
	switch values := c.Locals(key).(type) {
	case []string:
		return normalizeSet(values...)
	case string:
		return normalizeSet(values)
	default:
		return map[string]struct{}{}
	}
}

func localString(c *fiber.Ctx, key string) string {
	value, _ := c.Locals(key).(string)
	return value
}

func normalizeSet(values ...string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized := normalize(value)
		if normalized != "" {
			set[normalized] = struct{}{}
		}
	}

	return set
}

func setValues(set map[string]struct{}) []string {
	values := make([]string, 0, len(set))
	for value := range set {
		values = append(values, value)
	}

	return values
}

func normalize(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}
