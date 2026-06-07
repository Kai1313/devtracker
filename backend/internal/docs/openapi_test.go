package docs

import (
	"encoding/json"
	"testing"
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
		"/audit-logs",
		"/notifications",
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
