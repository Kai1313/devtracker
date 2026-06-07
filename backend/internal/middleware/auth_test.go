package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"devtracker/backend/internal/httpx"
	"devtracker/backend/pkg/response"

	"github.com/gofiber/fiber/v2"
)

func TestRequirePermissionAllowsMatchingPermission(t *testing.T) {
	app := testRBACApp(func(c *fiber.Ctx) error {
		c.Locals(httpx.LocalRoles, []string{"developer"})
		c.Locals(httpx.LocalPermissions, []string{"view_assigned_tasks"})
		return c.Next()
	})
	app.Get("/", RequirePermission("view_assigned_tasks"), noContent)

	resp := performRequest(t, app)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}

func TestRequirePermissionAllowsAdminWithoutExplicitPermission(t *testing.T) {
	app := testRBACApp(func(c *fiber.Ctx) error {
		c.Locals(httpx.LocalRoles, []string{"admin"})
		return c.Next()
	})
	app.Get("/", RequirePermission("manage_projects"), noContent)

	resp := performRequest(t, app)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}

func TestRequirePermissionRejectsMissingPermission(t *testing.T) {
	app := testRBACApp(func(c *fiber.Ctx) error {
		c.Locals(httpx.LocalRoles, []string{"developer"})
		c.Locals(httpx.LocalPermissions, []string{"view_assigned_tasks"})
		return c.Next()
	})
	app.Get("/", RequirePermission("manage_projects"), noContent)

	resp := performRequest(t, app)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, resp.StatusCode)
	}
}

func TestRequireRoleAllowsMatchingRole(t *testing.T) {
	app := testRBACApp(func(c *fiber.Ctx) error {
		c.Locals(httpx.LocalRoles, []string{"project_manager"})
		return c.Next()
	})
	app.Get("/", RequireRole("project_manager"), noContent)

	resp := performRequest(t, app)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}

func TestRequireRoleAllowsAdminBypass(t *testing.T) {
	app := testRBACApp(func(c *fiber.Ctx) error {
		c.Locals(httpx.LocalRole, "admin")
		return c.Next()
	})
	app.Get("/", RequireRole("project_manager"), noContent)

	resp := performRequest(t, app)
	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, resp.StatusCode)
	}
}

func testRBACApp(seed fiber.Handler) *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: response.ErrorHandler})
	app.Use(seed)
	return app
}

func performRequest(t *testing.T, app *fiber.App) *http.Response {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("perform request: %v", err)
	}

	return resp
}

func noContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}
