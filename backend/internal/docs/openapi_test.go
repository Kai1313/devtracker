package docs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestSpecIsSerializableAndDocumentsRoutes(t *testing.T) {
	spec := Spec("/api")
	if _, err := json.Marshal(spec); err != nil {
		t.Fatalf("marshal OpenAPI spec: %v", err)
	}

	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		t.Fatal("expected paths map")
	}

	expectedPaths := []string{
		"/health",
		"/auth/login",
		"/auth/logout",
		"/auth/bootstrap",
		"/auth/me",
		"/users",
		"/users/{id}",
		"/projects",
		"/projects/{id}",
		"/sprints",
		"/sprints/{id}",
		"/sprints/{id}/close",
		"/statuses",
		"/statuses/{id}",
		"/tasks",
		"/tasks/{id}",
		"/tasks/{id}/histories",
		"/dashboard/summary",
		"/kpi/developers",
		"/kpi/projects",
		"/kpi/snapshots",
		"/kpi/snapshots/developer/{developer_id}",
		"/kpi/snapshots/generate/{sprint_id}",
		"/audit-logs",
		"/notifications",
		"/notifications/unread-count",
		"/notifications/read-all",
		"/notifications/{id}/read",
		"/workload",
	}

	for _, path := range expectedPaths {
		if _, ok := paths[path]; !ok {
			t.Fatalf("expected path %s to be documented", path)
		}
	}
}

func TestSpecIncludesBearerAuth(t *testing.T) {
	spec := Spec("/api")
	components := spec["components"].(map[string]any)
	securitySchemes := components["securitySchemes"].(map[string]any)

	if _, ok := securitySchemes["BearerAuth"]; !ok {
		t.Fatal("expected BearerAuth security scheme")
	}
}

func TestSpecIncludesFeatureTagsAndRBACPolicy(t *testing.T) {
	spec := Spec("/api")
	rawTags := spec["tags"].([]any)
	expectedTags := map[string]bool{
		"Task Histories": false,
		"KPI Snapshots":  false,
		"RBAC":           false,
	}

	for _, rawTag := range rawTags {
		tag, ok := rawTag.(map[string]any)
		if !ok {
			t.Fatalf("expected tag map, got %T", rawTag)
		}
		name, _ := tag["name"].(string)
		if _, ok := expectedTags[name]; ok {
			expectedTags[name] = true
		}
	}

	for name, found := range expectedTags {
		if !found {
			t.Fatalf("expected tag %s to be documented", name)
		}
	}

	rbac, ok := spec["x-rbac"].(map[string]any)
	if !ok {
		t.Fatal("expected x-rbac policy")
	}
	roles := rbac["roles"].(map[string]any)
	if _, ok := roles["admin"]; !ok {
		t.Fatal("expected admin role in x-rbac policy")
	}
}

func TestProtectedOperationsIncludeAuthErrorExamples(t *testing.T) {
	spec := Spec("/api")
	paths := spec["paths"].(map[string]any)
	projects := paths["/projects"].(map[string]any)
	listProjects := projects["get"].(map[string]any)
	responses := listProjects["responses"].(map[string]any)

	if _, ok := responses["401"]; !ok {
		t.Fatal("expected protected operation to document 401")
	}
	if _, ok := responses["403"]; !ok {
		t.Fatal("expected protected operation to document 403")
	}
}

func TestRegisterRoutesServesSwaggerWildcard(t *testing.T) {
	app := fiber.New()
	RegisterRoutes(app, "/api")

	uiResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil))
	if err != nil {
		t.Fatalf("request swagger UI: %v", err)
	}
	if uiResp.StatusCode != http.StatusOK {
		t.Fatalf("expected swagger UI status 200, got %d", uiResp.StatusCode)
	}
	if contentType := uiResp.Header.Get("Content-Type"); !strings.Contains(contentType, "text/html") {
		t.Fatalf("expected swagger UI content type text/html, got %q", contentType)
	}

	specResp, err := app.Test(httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil))
	if err != nil {
		t.Fatalf("request swagger spec: %v", err)
	}
	if specResp.StatusCode != http.StatusOK {
		t.Fatalf("expected swagger spec status 200, got %d", specResp.StatusCode)
	}
	if contentType := specResp.Header.Get("Content-Type"); !strings.Contains(contentType, "application/json") {
		t.Fatalf("expected swagger spec content type application/json, got %q", contentType)
	}
}
