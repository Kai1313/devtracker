package docs

import (
	"strings"

	"github.com/gofiber/fiber/v2"
)

func RegisterRoutes(app *fiber.App, basePath string) {
	spec := Spec(basePath)

	app.Get("/swagger", swaggerUI)
	app.Get("/swagger/", swaggerUI)
	app.Get("/swagger/doc.json", func(c *fiber.Ctx) error {
		return c.JSON(spec)
	})
}

func swaggerUI(c *fiber.Ctx) error {
	c.Type("html", "utf-8")
	return c.SendString(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>DevTracker API Swagger</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.ui = SwaggerUIBundle({
      url: "/swagger/doc.json",
      dom_id: "#swagger-ui",
      deepLinking: true,
      persistAuthorization: true,
      displayRequestDuration: true
    });
  </script>
</body>
</html>`)
}

func Spec(basePath string) map[string]any {
	basePath = strings.TrimRight(strings.TrimSpace(basePath), "/")
	if basePath == "" {
		basePath = "/"
	}

	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "DevTracker API",
			"description": "Backend API for project, sprint, task, KPI, workload, notification, audit, and RBAC workflows. Protected endpoints use `Authorization: Bearer <access_token>`.",
			"version":     "1.0.0",
		},
		"servers": []any{
			map[string]any{"url": basePath, "description": "Configured API base path"},
		},
		"tags": []any{
			tag("Health", "Service health checks"),
			tag("Auth", "Authentication endpoints and JWT examples"),
			tag("Users", "User administration"),
			tag("Projects", "Project CRUD"),
			tag("Sprints", "Sprint CRUD and close workflow"),
			tag("Task Statuses", "Task status configuration"),
			tag("Tasks", "Task assignment, status changes, and history"),
			tag("Dashboard", "Dashboard summary APIs"),
			tag("KPI", "Developer and project KPI APIs"),
			tag("Audit Logs", "Admin audit trail"),
			tag("Notifications", "Database notifications"),
			tag("Workload", "Developer workload API"),
		},
		"paths":      paths(),
		"components": components(),
	}
}

func paths() map[string]any {
	return map[string]any{
		"/health": map[string]any{
			"get": operation([]string{"Health"}, "Health check", "Checks service and database connectivity.", false, nil, nil, map[string]any{
				"200": response("service is healthy", ref("HealthResponse"), example("service is healthy", map[string]any{"app": "Load Developer Sheets API", "env": "development"})),
			}),
		},
		"/auth/login": map[string]any{
			"post": operation([]string{"Auth"}, "Login", "Authenticates a user and returns a bearer token. Use the returned token in Swagger UI with the `Authorize` button as `Bearer <token>`.", false, nil, request("LoginRequest", map[string]any{
				"email":    "admin@example.com",
				"password": "secret123",
			}), map[string]any{
				"200": response("login successful", ref("LoginResponse"), example("login successful", loginExample())),
				"401": errorResponse("invalid email or password"),
			}),
		},
		"/auth/logout": map[string]any{
			"post": operation([]string{"Auth"}, "Logout", "Logs out the authenticated user. JWT tokens are stateless; clients should discard the token.", true, nil, nil, map[string]any{
				"200": response("logout successful", nil, example("logout successful", nil)),
				"401": errorResponse("missing bearer token"),
			}),
		},
		"/auth/bootstrap": map[string]any{
			"post": operation([]string{"Auth"}, "Bootstrap admin", "Creates the first admin account before any users exist.", false, nil, request("BootstrapAdminRequest", map[string]any{
				"name":     "Admin User",
				"email":    "admin@example.com",
				"password": "secret123",
			}), map[string]any{
				"201": response("bootstrap admin created", ref("User"), example("bootstrap admin created", userExample())),
				"409": errorResponse("bootstrap admin can only be created before any users exist"),
			}),
		},
		"/auth/me": map[string]any{
			"get": operation([]string{"Auth"}, "Current user", "Returns the authenticated user profile. Authentication example: `Authorization: Bearer eyJ...`.", true, nil, nil, map[string]any{
				"200": response("current user retrieved", ref("User"), example("current user retrieved", userExample())),
				"401": errorResponse("authenticated user is missing"),
			}),
		},
		"/users": map[string]any{
			"get": operation([]string{"Users"}, "List users", "Lists users. Requires `manage_users` permission.", true, append(pageParams(), queryParam("search", "string", "Search by name or email", "admin"), queryParam("role_id", "string", "Filter by role UUID", idExample()), queryParam("is_active", "boolean", "Filter by active state", true)), nil, map[string]any{
				"200": listResponse("users retrieved", arrayOf(ref("User")), example("users retrieved", []any{userExample()})),
				"403": errorResponse("insufficient permissions"),
			}),
			"post": operation([]string{"Users"}, "Create user", "Creates a user. Requires `manage_users` permission.", true, nil, request("CreateUserRequest", map[string]any{
				"role_id":   idExample(),
				"name":      "Dev User",
				"email":     "dev@example.com",
				"password":  "secret123",
				"team":      "Backend",
				"position":  "Developer",
				"is_active": true,
			}), map[string]any{
				"201": response("user created", ref("User"), example("user created", userExample())),
				"422": validationErrorResponse(),
			}),
		},
		"/users/{id}": crudPath("Users", "user", "User", "UpdateUserRequest", map[string]any{
			"name":      "Senior Dev User",
			"position":  "Senior Developer",
			"is_active": true,
		}),
		"/projects": map[string]any{
			"get": operation([]string{"Projects"}, "List projects", "Lists projects with pagination and search. Requires `manage_projects` permission.", true, append(pageParams(), queryParam("search", "string", "Search by project_code, project_name, or client_name", "DEV")), nil, map[string]any{
				"200": listResponse("projects retrieved", arrayOf(ref("Project")), example("projects retrieved", []any{projectExample()})),
			}),
			"post": operation([]string{"Projects"}, "Create project", "Creates a project. Requires `manage_projects` permission.", true, nil, request("CreateProjectRequest", map[string]any{
				"project_code": "DEV",
				"project_name": "Dev Tracker",
				"client_name":  "Internal",
				"status":       "active",
				"start_date":   "2026-01-01",
				"end_date":     "2026-03-31",
			}), map[string]any{
				"201": response("project created", ref("Project"), example("project created", projectExample())),
				"422": validationErrorResponse(),
			}),
		},
		"/projects/{id}": crudPath("Projects", "project", "Project", "UpdateProjectRequest", map[string]any{"project_name": "Dev Tracker API", "status": "active"}),
		"/sprints": map[string]any{
			"get": operation([]string{"Sprints"}, "List sprints", "Lists sprints. Requires `manage_sprints` permission.", true, append(pageParams(), queryParam("project_id", "string", "Filter by project UUID", idExample()), queryParam("status", "string", "Filter by planning, active, or closed", "active")), nil, map[string]any{
				"200": listResponse("sprints retrieved", arrayOf(ref("Sprint")), example("sprints retrieved", []any{sprintExample()})),
			}),
			"post": operation([]string{"Sprints"}, "Create sprint", "Creates a sprint. Requires `manage_sprints` permission.", true, nil, request("CreateSprintRequest", map[string]any{
				"project_id":  idExample(),
				"sprint_name": "Sprint 1",
				"start_date":  "2026-01-01",
				"end_date":    "2026-01-14",
				"status":      "active",
			}), map[string]any{
				"201": response("sprint created", ref("Sprint"), example("sprint created", sprintExample())),
				"422": validationErrorResponse(),
			}),
		},
		"/sprints/{id}": crudPath("Sprints", "sprint", "Sprint", "UpdateSprintRequest", map[string]any{"sprint_name": "Sprint 1A", "status": "active"}),
		"/sprints/{id}/close": map[string]any{
			"patch": operation([]string{"Sprints"}, "Close sprint", "Closes a sprint and generates KPI snapshots. Requires `manage_sprints` permission.", true, []any{pathIDParam()}, nil, map[string]any{
				"200": response("sprint closed", ref("Sprint"), example("sprint closed", sprintExampleWithStatus("closed"))),
				"404": errorResponse("sprint not found"),
			}),
		},
		"/statuses": map[string]any{
			"get": operation([]string{"Task Statuses"}, "List task statuses", "Lists task statuses. Requires `manage_task_statuses` permission.", true, append(pageParams(), queryParam("is_active", "boolean", "Filter by active state", true)), nil, map[string]any{
				"200": listResponse("task statuses retrieved", arrayOf(ref("TaskStatus")), example("task statuses retrieved", []any{statusExample()})),
			}),
			"post": operation([]string{"Task Statuses"}, "Create task status", "Creates a task status. Requires `manage_task_statuses` permission.", true, nil, request("CreateTaskStatusRequest", map[string]any{
				"status_name":  "Code Review",
				"color_name":   "purple",
				"color_hex":    "#8B5CF6",
				"status_order": 7,
				"is_done":      false,
				"is_qa_status": true,
				"is_active":    true,
			}), map[string]any{
				"201": response("task status created", ref("TaskStatus"), example("task status created", statusExample())),
				"422": validationErrorResponse(),
			}),
		},
		"/statuses/{id}": crudPath("Task Statuses", "task status", "TaskStatus", "UpdateTaskStatusRequest", map[string]any{"color_hex": "#3B82F6", "is_active": true}),
		"/tasks": map[string]any{
			"get": operation([]string{"Tasks"}, "List tasks", "Lists tasks. Requires `manage_tasks`, `view_assigned_tasks`, or `view_ready_to_check_tasks` permission. Developer results are scoped to assigned tasks; QA results are scoped to Ready to Check tasks.", true, append(pageParams(), queryParam("developer_id", "string", "Filter by developer UUID", idExample()), queryParam("project_id", "string", "Filter by project UUID", idExample()), queryParam("sprint_id", "string", "Filter by sprint UUID", idExample()), queryParam("status_id", "string", "Filter by status UUID", idExample()), queryParam("search", "string", "Search by ticket_number or task_title", "DEV-1")), nil, map[string]any{
				"200": listResponse("tasks retrieved", arrayOf(ref("Task")), example("tasks retrieved", []any{taskExample()})),
			}),
			"post": operation([]string{"Tasks"}, "Create task", "Creates and assigns a task. Requires `manage_tasks` permission.", true, nil, request("CreateTaskRequest", taskRequestExample()), map[string]any{
				"201": response("task created", ref("Task"), example("task created", taskExample())),
				"422": validationErrorResponse(),
			}),
		},
		"/tasks/{id}": map[string]any{
			"get": operation([]string{"Tasks"}, "Get task", "Returns one task by UUID. Requires `manage_tasks`, `view_assigned_tasks`, or `view_ready_to_check_tasks` permission. Developer and QA access is scoped.", true, []any{pathIDParam()}, nil, map[string]any{
				"200": response("task retrieved", ref("Task"), example("task retrieved", taskExample())),
				"404": errorResponse("task not found"),
			}),
			"patch": operation([]string{"Tasks"}, "Update task", "Updates one task by UUID. Requires `manage_tasks` for full updates, `update_own_task_status` for developer status-only updates on assigned tasks, or `update_qa_status` for QA status-only updates to QA statuses.", true, []any{pathIDParam()}, request("UpdateTaskRequest", map[string]any{"status_id": idExample(), "actual_point": 5, "note": "Moved to QA"}), map[string]any{
				"200": response("task updated", ref("Task"), example("task updated", taskExample())),
				"404": errorResponse("task not found"),
				"422": validationErrorResponse(),
			}),
			"delete": operation([]string{"Tasks"}, "Delete task", "Deletes one task by UUID. Requires `manage_tasks` permission.", true, []any{pathIDParam()}, nil, map[string]any{
				"200": response("task deleted", nil, example("task deleted", nil)),
				"404": errorResponse("task not found"),
			}),
		},
		"/tasks/{id}/histories": map[string]any{
			"get": operation([]string{"Tasks"}, "List task histories", "Lists status-change history for a task. Uses the same scoped task-view rule as task detail.", true, []any{pathIDParam()}, nil, map[string]any{
				"200": response("task histories retrieved", arrayOf(ref("TaskHistory")), example("task histories retrieved", []any{taskHistoryExample()})),
				"404": errorResponse("task not found"),
			}),
		},
		"/dashboard/summary": map[string]any{
			"get": operation([]string{"Dashboard"}, "Dashboard summary", "Returns dashboard totals. Requires `view_dashboard` permission.", true, []any{queryParam("sprint_id", "string", "Optional sprint UUID", idExample())}, nil, map[string]any{
				"200": response("dashboard summary retrieved", ref("DashboardSummary"), example("dashboard summary retrieved", dashboardExample())),
			}),
		},
		"/kpi/developers": map[string]any{
			"get": operation([]string{"KPI"}, "Developer KPI", "Returns developer KPI. Closed sprints use KPI snapshots.", true, []any{queryParam("sprint_id", "string", "Optional sprint UUID", idExample())}, nil, map[string]any{
				"200": response("developer KPI retrieved", arrayOf(ref("DeveloperKPI")), example("developer KPI retrieved", []any{developerKPIExample()})),
			}),
		},
		"/kpi/projects": map[string]any{
			"get": operation([]string{"KPI"}, "Project KPI", "Returns project KPI. Closed sprints use KPI snapshots.", true, []any{queryParam("sprint_id", "string", "Optional sprint UUID", idExample())}, nil, map[string]any{
				"200": response("project KPI retrieved", arrayOf(ref("ProjectKPI")), example("project KPI retrieved", []any{projectKPIExample()})),
			}),
		},
		"/audit-logs": map[string]any{
			"get": operation([]string{"Audit Logs"}, "List audit logs", "Lists audit trail entries. Admin only.", true, append(pageParams(), queryParam("user", "string", "Filter by user UUID. Alias: user_id", idExample()), queryParam("module", "string", "Filter by module", "tasks"), queryParam("start_date", "string", "Start date YYYY-MM-DD", "2026-01-01"), queryParam("end_date", "string", "End date YYYY-MM-DD", "2026-01-31")), nil, map[string]any{
				"200": listResponse("audit logs retrieved", arrayOf(ref("AuditLog")), example("audit logs retrieved", []any{auditLogExample()})),
				"403": errorResponse("insufficient permissions"),
			}),
		},
		"/notifications": map[string]any{
			"get": operation([]string{"Notifications"}, "List notifications", "Lists notifications for the authenticated user and returns unread count.", true, pageParams(), nil, map[string]any{
				"200": listResponse("notifications retrieved", ref("NotificationList"), example("notifications retrieved", notificationListExample())),
			}),
		},
		"/notifications/{id}/read": map[string]any{
			"patch": operation([]string{"Notifications"}, "Mark notification read", "Marks one notification as read for the authenticated user.", true, []any{pathIDParam()}, nil, map[string]any{
				"200": response("notification marked as read", ref("NotificationRead"), example("notification marked as read", notificationReadExample())),
				"404": errorResponse("notification not found"),
			}),
		},
		"/workload": map[string]any{
			"get": operation([]string{"Workload"}, "Developer workload", "Returns active workload per developer. Requires `view_kpi` or `manage_tasks` permission.", true, []any{
				queryParam("sprint", "string", "Optional sprint UUID. Alias: sprint_id", idExample()),
				queryParam("project", "string", "Optional project UUID. Alias: project_id", idExample()),
			}, nil, map[string]any{
				"200": response("developer workload retrieved", arrayOf(ref("DeveloperWorkload")), example("developer workload retrieved", []any{workloadExample()})),
			}),
		},
	}
}

func crudPath(tagName, noun, schemaName, updateSchema string, updateExample map[string]any) map[string]any {
	return map[string]any{
		"get": operation([]string{tagName}, "Get "+noun, "Returns one "+noun+" by UUID.", true, []any{pathIDParam()}, nil, map[string]any{
			"200": response(noun+" retrieved", ref(schemaName), example(noun+" retrieved", sampleFor(schemaName))),
			"404": errorResponse(noun + " not found"),
		}),
		"patch": operation([]string{tagName}, "Update "+noun, "Updates one "+noun+" by UUID.", true, []any{pathIDParam()}, request(updateSchema, updateExample), map[string]any{
			"200": response(noun+" updated", ref(schemaName), example(noun+" updated", sampleFor(schemaName))),
			"404": errorResponse(noun + " not found"),
			"422": validationErrorResponse(),
		}),
		"delete": operation([]string{tagName}, "Delete "+noun, "Deletes one "+noun+" by UUID.", true, []any{pathIDParam()}, nil, map[string]any{
			"200": response(noun+" deleted", nil, example(noun+" deleted", nil)),
			"404": errorResponse(noun + " not found"),
		}),
	}
}

func operation(tags []string, summary, description string, protected bool, params []any, requestBody map[string]any, responses map[string]any) map[string]any {
	result := map[string]any{
		"tags":        tags,
		"summary":     summary,
		"description": description,
		"responses":   responses,
	}

	if protected {
		result["security"] = []any{map[string]any{"BearerAuth": []any{}}}
	}
	if len(params) > 0 {
		result["parameters"] = params
	}
	if requestBody != nil {
		result["requestBody"] = requestBody
	}

	return result
}

func components() map[string]any {
	return map[string]any{
		"securitySchemes": map[string]any{
			"BearerAuth": map[string]any{
				"type":         "http",
				"scheme":       "bearer",
				"bearerFormat": "JWT",
				"description":  "JWT authentication example: `Authorization: Bearer <access_token>`.",
			},
		},
		"schemas": schemas(),
	}
}

func schemas() map[string]any {
	return map[string]any{
		"HealthResponse":        object(nil, props("app", str(), "env", str())),
		"LoginRequest":          object([]string{"email", "password"}, props("email", str(), "password", str())),
		"PaginationMeta":        object(nil, props("page", integer(), "limit", integer(), "total", integer())),
		"BootstrapAdminRequest": object([]string{"name", "email", "password"}, props("name", str(), "email", str(), "password", str())),
		"LoginResponse": object(nil, props(
			"access_token", str(),
			"token_type", str(),
			"expires_at", str(),
			"expires_in", integer(),
			"user", ref("User"),
		)),
		"Role":                    object(nil, props("id", uuidSchema(), "name", str(), "description", str())),
		"User":                    object(nil, props("id", uuidSchema(), "role_id", uuidSchema(), "role", ref("Role"), "name", str(), "email", str(), "team", str(), "position", str(), "is_active", boolean(), "created_at", dateTime(), "updated_at", dateTime())),
		"CreateUserRequest":       object([]string{"role_id", "name", "email", "password"}, props("role_id", uuidSchema(), "name", str(), "email", str(), "password", str(), "team", str(), "position", str(), "is_active", boolean())),
		"UpdateUserRequest":       object(nil, props("role_id", uuidSchema(), "name", str(), "email", str(), "password", str(), "team", str(), "position", str(), "is_active", boolean())),
		"Project":                 object(nil, props("id", uuidSchema(), "project_code", str(), "project_name", str(), "client_name", str(), "status", str(), "start_date", date(), "end_date", date(), "created_at", dateTime(), "updated_at", dateTime())),
		"CreateProjectRequest":    object([]string{"project_code", "project_name"}, props("project_code", str(), "project_name", str(), "client_name", str(), "status", str(), "start_date", date(), "end_date", date())),
		"UpdateProjectRequest":    object(nil, props("project_code", str(), "project_name", str(), "client_name", str(), "status", str(), "start_date", date(), "end_date", date())),
		"Sprint":                  object(nil, props("id", uuidSchema(), "project_id", uuidSchema(), "project", ref("Project"), "sprint_name", str(), "start_date", date(), "end_date", date(), "status", str(), "created_at", dateTime(), "updated_at", dateTime())),
		"CreateSprintRequest":     object([]string{"project_id", "sprint_name", "start_date", "end_date"}, props("project_id", uuidSchema(), "sprint_name", str(), "start_date", date(), "end_date", date(), "status", str())),
		"UpdateSprintRequest":     object(nil, props("project_id", uuidSchema(), "sprint_name", str(), "start_date", date(), "end_date", date(), "status", str())),
		"TaskStatus":              object(nil, props("id", uuidSchema(), "status_name", str(), "color_name", str(), "color_hex", str(), "status_order", integer(), "is_done", boolean(), "is_qa_status", boolean(), "is_active", boolean(), "created_at", dateTime(), "updated_at", dateTime())),
		"CreateTaskStatusRequest": object([]string{"status_name", "color_name", "color_hex"}, props("status_name", str(), "color_name", str(), "color_hex", str(), "status_order", integer(), "is_done", boolean(), "is_qa_status", boolean(), "is_active", boolean())),
		"UpdateTaskStatusRequest": object(nil, props("status_name", str(), "color_name", str(), "color_hex", str(), "status_order", integer(), "is_done", boolean(), "is_qa_status", boolean(), "is_active", boolean())),
		"Task":                    object(nil, props("id", uuidSchema(), "developer_id", uuidSchema(), "developer", ref("User"), "project_id", uuidSchema(), "project", ref("Project"), "sprint_id", uuidSchema(), "sprint", ref("Sprint"), "status_id", uuidSchema(), "status", ref("TaskStatus"), "ticket_number", str(), "task_title", str(), "task_description", str(), "priority", str(), "estimated_point", number(), "actual_point", number(), "start_date", date(), "due_date", date(), "completed_date", dateTime(), "qa_checked_date", dateTime(), "created_at", dateTime(), "updated_at", dateTime())),
		"CreateTaskRequest":       object([]string{"developer_id", "project_id", "sprint_id", "task_title", "priority", "status_id"}, props("developer_id", uuidSchema(), "project_id", uuidSchema(), "sprint_id", uuidSchema(), "ticket_number", str(), "task_title", str(), "task_description", str(), "priority", str(), "status_id", uuidSchema(), "estimated_point", number(), "actual_point", number(), "start_date", date(), "due_date", date(), "completed_date", dateTime(), "qa_checked_date", dateTime(), "note", str())),
		"UpdateTaskRequest":       object(nil, props("developer_id", uuidSchema(), "project_id", uuidSchema(), "sprint_id", uuidSchema(), "ticket_number", str(), "task_title", str(), "task_description", str(), "priority", str(), "status_id", uuidSchema(), "estimated_point", number(), "actual_point", number(), "start_date", date(), "due_date", date(), "completed_date", dateTime(), "qa_checked_date", dateTime(), "note", str())),
		"TaskHistory":             object(nil, props("id", uuidSchema(), "task_id", uuidSchema(), "old_status_id", uuidSchema(), "new_status_id", uuidSchema(), "changed_by", uuidSchema(), "changed_at", dateTime(), "note", str())),
		"DashboardSummary":        object(nil, props("total_tasks", integer(), "todo_tasks", integer(), "in_progress_tasks", integer(), "ready_to_check_tasks", integer(), "checked_by_qa_tasks", integer(), "done_tasks", integer(), "blocked_tasks", integer(), "completion_rate", number(), "total_developers", integer(), "total_projects", integer())),
		"DeveloperKPI":            object(nil, props("developer_id", uuidSchema(), "developer_name", str(), "total_assigned", integer(), "total_done", integer(), "total_ready_to_check", integer(), "total_checked_by_qa", integer(), "delayed_tasks", integer(), "completion_rate", number(), "total_estimated_point", number(), "total_actual_point", number())),
		"ProjectKPI":              object(nil, props("project_id", uuidSchema(), "project_name", str(), "total_assigned", integer(), "total_done", integer(), "total_ready_to_check", integer(), "total_checked_by_qa", integer(), "delayed_tasks", integer(), "completion_rate", number(), "total_estimated_point", number(), "total_actual_point", number())),
		"AuditLog":                object(nil, props("id", uuidSchema(), "user_id", uuidSchema(), "module", str(), "action", str(), "old_value", map[string]any{"type": "object", "additionalProperties": true}, "new_value", map[string]any{"type": "object", "additionalProperties": true}, "ip_address", str(), "created_at", dateTime())),
		"Notification":            object(nil, props("id", uuidSchema(), "user_id", uuidSchema(), "task_id", uuidSchema(), "type", str(), "title", str(), "message", str(), "is_read", boolean(), "read_at", dateTime(), "created_at", dateTime())),
		"NotificationList":        object(nil, props("notifications", arrayOf(ref("Notification")), "unread_count", integer())),
		"NotificationRead":        object(nil, props("notification", ref("Notification"), "unread_count", integer())),
		"DeveloperWorkload":       object(nil, props("developer_id", uuidSchema(), "developer_name", str(), "active_tasks", integer(), "total_points", number(), "overdue_tasks", integer(), "current_sprint_tasks", integer(), "workload_classification", map[string]any{"type": "string", "enum": []any{"LOW", "NORMAL", "HIGH", "OVERLOADED"}})),
	}
}

func tag(name, description string) map[string]any {
	return map[string]any{"name": name, "description": description}
}

func request(schemaName string, exampleValue any) map[string]any {
	return map[string]any{
		"required": true,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": ref(schemaName),
				"examples": map[string]any{
					"default": map[string]any{"value": exampleValue},
				},
			},
		},
	}
}

func response(description string, dataSchema any, exampleValue any) map[string]any {
	schema := object(nil, props("success", boolean(), "message", str()))
	if dataSchema != nil {
		schema["properties"].(map[string]any)["data"] = dataSchema
	}

	return map[string]any{
		"description": description,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema":   schema,
				"examples": map[string]any{"default": map[string]any{"value": exampleValue}},
			},
		},
	}
}

func listResponse(description string, dataSchema any, exampleValue any) map[string]any {
	schema := object(nil, props("success", boolean(), "message", str(), "data", dataSchema, "meta", ref("PaginationMeta")))
	return map[string]any{
		"description": description,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": schema,
				"examples": map[string]any{
					"default": map[string]any{
						"value": withMeta(description, exampleValue),
					},
				},
			},
		},
	}
}

func errorResponse(message string) map[string]any {
	return map[string]any{
		"description": message,
		"content": map[string]any{
			"application/json": map[string]any{
				"schema": object(nil, props("success", boolean(), "message", str(), "error", object(nil, props("code", str(), "details", map[string]any{"type": "object", "additionalProperties": true})))),
				"examples": map[string]any{
					"default": map[string]any{"value": map[string]any{
						"success": false,
						"message": message,
						"error":   map[string]any{"code": "bad_request"},
					}},
				},
			},
		},
	}
}

func validationErrorResponse() map[string]any {
	return errorResponse("validation failed")
}

func example(message string, data any) map[string]any {
	result := map[string]any{"success": true, "message": message}
	if data != nil {
		result["data"] = data
	}
	return result
}

func withMeta(message string, data any) map[string]any {
	return map[string]any{
		"success": true,
		"message": message,
		"data":    data,
		"meta":    map[string]any{"page": 1, "limit": 20, "total": 1},
	}
}

func pageParams() []any {
	return []any{
		queryParam("page", "integer", "Page number", 1),
		queryParam("limit", "integer", "Page size, max 100", 20),
	}
}

func pathIDParam() map[string]any {
	return map[string]any{
		"name":        "id",
		"in":          "path",
		"required":    true,
		"description": "Resource UUID",
		"schema":      uuidSchema(),
		"example":     idExample(),
	}
}

func queryParam(name, typ, description string, exampleValue any) map[string]any {
	return map[string]any{
		"name":        name,
		"in":          "query",
		"required":    false,
		"description": description,
		"schema":      map[string]any{"type": typ},
		"example":     exampleValue,
	}
}

func ref(name string) map[string]any {
	return map[string]any{"$ref": "#/components/schemas/" + name}
}

func arrayOf(item any) map[string]any {
	return map[string]any{"type": "array", "items": item}
}

func object(required []string, properties map[string]any) map[string]any {
	result := map[string]any{"type": "object", "properties": properties}
	if len(required) > 0 {
		req := make([]any, 0, len(required))
		for _, item := range required {
			req = append(req, item)
		}
		result["required"] = req
	}
	return result
}

func props(values ...any) map[string]any {
	result := map[string]any{}
	for i := 0; i+1 < len(values); i += 2 {
		result[values[i].(string)] = values[i+1]
	}
	return result
}

func str() map[string]any     { return map[string]any{"type": "string"} }
func boolean() map[string]any { return map[string]any{"type": "boolean"} }
func integer() map[string]any { return map[string]any{"type": "integer", "format": "int64"} }
func number() map[string]any  { return map[string]any{"type": "number", "format": "double"} }
func date() map[string]any {
	return map[string]any{"type": "string", "format": "date", "nullable": true}
}
func dateTime() map[string]any {
	return map[string]any{"type": "string", "format": "date-time", "nullable": true}
}
func uuidSchema() map[string]any {
	return map[string]any{"type": "string", "format": "uuid"}
}

func idExample() string { return "11111111-1111-4111-8111-111111111111" }

func sampleFor(schemaName string) any {
	switch schemaName {
	case "User":
		return userExample()
	case "Project":
		return projectExample()
	case "Sprint":
		return sprintExample()
	case "TaskStatus":
		return statusExample()
	case "Task":
		return taskExample()
	default:
		return map[string]any{"id": idExample()}
	}
}

func roleExample() map[string]any {
	return map[string]any{"id": idExample(), "name": "admin", "description": "System administrator"}
}

func userExample() map[string]any {
	return map[string]any{
		"id":         idExample(),
		"role_id":    idExample(),
		"role":       roleExample(),
		"name":       "Admin User",
		"email":      "admin@example.com",
		"team":       "Backend",
		"position":   "Developer",
		"is_active":  true,
		"created_at": "2026-01-01T00:00:00Z",
		"updated_at": "2026-01-01T00:00:00Z",
	}
}

func loginExample() map[string]any {
	return map[string]any{"access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example", "token_type": "Bearer", "expires_at": "2026-01-02T00:00:00Z", "expires_in": 86400, "user": userExample()}
}

func projectExample() map[string]any {
	return map[string]any{"id": idExample(), "project_code": "DEV", "project_name": "Dev Tracker", "client_name": "Internal", "status": "active", "start_date": "2026-01-01", "end_date": "2026-03-31", "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}
}

func sprintExample() map[string]any {
	return sprintExampleWithStatus("active")
}

func sprintExampleWithStatus(status string) map[string]any {
	return map[string]any{"id": idExample(), "project_id": idExample(), "project": projectExample(), "sprint_name": "Sprint 1", "start_date": "2026-01-01", "end_date": "2026-01-14", "status": status, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}
}

func statusExample() map[string]any {
	return map[string]any{"id": idExample(), "status_name": "Ready to Check", "color_name": "blue", "color_hex": "#3B82F6", "status_order": 3, "is_done": false, "is_qa_status": true, "is_active": true, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}
}

func taskRequestExample() map[string]any {
	return map[string]any{"developer_id": idExample(), "project_id": idExample(), "sprint_id": idExample(), "ticket_number": "DEV-1", "task_title": "Build workload API", "task_description": "Implement endpoint", "priority": "high", "status_id": idExample(), "estimated_point": 5, "actual_point": 3, "start_date": "2026-01-01", "due_date": "2026-01-05", "note": "Assigned"}
}

func taskExample() map[string]any {
	return map[string]any{"id": idExample(), "developer_id": idExample(), "developer": userExample(), "project_id": idExample(), "project": projectExample(), "sprint_id": idExample(), "sprint": sprintExample(), "status_id": idExample(), "status": statusExample(), "ticket_number": "DEV-1", "task_title": "Build workload API", "task_description": "Implement endpoint", "priority": "high", "estimated_point": 5, "actual_point": 3, "start_date": "2026-01-01", "due_date": "2026-01-05", "completed_date": nil, "qa_checked_date": nil, "created_at": "2026-01-01T00:00:00Z", "updated_at": "2026-01-01T00:00:00Z"}
}

func taskHistoryExample() map[string]any {
	return map[string]any{"id": idExample(), "task_id": idExample(), "old_status_id": idExample(), "new_status_id": idExample(), "changed_by": idExample(), "changed_at": "2026-01-01T00:00:00Z", "note": "Moved to QA"}
}

func dashboardExample() map[string]any {
	return map[string]any{"total_tasks": 10, "todo_tasks": 2, "in_progress_tasks": 3, "ready_to_check_tasks": 1, "checked_by_qa_tasks": 1, "done_tasks": 2, "blocked_tasks": 1, "completion_rate": 20, "total_developers": 3, "total_projects": 1}
}

func developerKPIExample() map[string]any {
	return map[string]any{"developer_id": idExample(), "developer_name": "Dev User", "total_assigned": 8, "total_done": 4, "total_ready_to_check": 1, "total_checked_by_qa": 1, "delayed_tasks": 2, "completion_rate": 50, "total_estimated_point": 21, "total_actual_point": 18}
}

func projectKPIExample() map[string]any {
	return map[string]any{"project_id": idExample(), "project_name": "Dev Tracker", "total_assigned": 8, "total_done": 4, "total_ready_to_check": 1, "total_checked_by_qa": 1, "delayed_tasks": 2, "completion_rate": 50, "total_estimated_point": 21, "total_actual_point": 18}
}

func auditLogExample() map[string]any {
	return map[string]any{"id": idExample(), "user_id": idExample(), "module": "tasks", "action": "status_change", "old_value": map[string]any{"status_name": "In Progress"}, "new_value": map[string]any{"status_name": "Ready to Check"}, "ip_address": "127.0.0.1", "created_at": "2026-01-01T00:00:00Z"}
}

func notificationExample() map[string]any {
	return map[string]any{"id": idExample(), "user_id": idExample(), "task_id": idExample(), "type": "task_ready_to_check", "title": "Task ready to check", "message": "Task moved to Ready to Check: Build workload API", "is_read": false, "created_at": "2026-01-01T00:00:00Z"}
}

func notificationListExample() map[string]any {
	return map[string]any{"notifications": []any{notificationExample()}, "unread_count": 1}
}

func notificationReadExample() map[string]any {
	notification := notificationExample()
	notification["is_read"] = true
	notification["read_at"] = "2026-01-01T01:00:00Z"
	return map[string]any{"notification": notification, "unread_count": 0}
}

func workloadExample() map[string]any {
	return map[string]any{"developer_id": idExample(), "developer_name": "Dev User", "active_tasks": 5, "total_points": 18, "overdue_tasks": 1, "current_sprint_tasks": 4, "workload_classification": "HIGH"}
}
