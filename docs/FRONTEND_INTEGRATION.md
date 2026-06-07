# Frontend Integration Guide

This guide is the frontend-facing API contract for the DevTracker backend.

## Base API URL

Default local backend:

```text
http://localhost:8080/api
```

Swagger UI:

```text
http://localhost:8080/swagger
```

OpenAPI JSON:

```text
http://localhost:8080/swagger/doc.json
```

## Response Envelope

All successful responses use the same wrapper:

```json
{
  "success": true,
  "message": "projects retrieved",
  "data": [],
  "meta": {
    "page": 1,
    "limit": 20,
    "total": 0,
    "sort_by": "created_at",
    "sort_order": "desc"
  }
}
```

`meta` is only present on paginated list endpoints.

Single-resource and command responses omit `meta`:

```json
{
  "success": true,
  "message": "project retrieved",
  "data": {
    "id": "11111111-1111-4111-8111-111111111111",
    "project_code": "DEV",
    "project_name": "Dev Tracker",
    "client_name": "Internal",
    "status": "active",
    "start_date": "2026-01-01",
    "end_date": "2026-03-31",
    "created_at": "2026-01-01T00:00:00Z",
    "updated_at": "2026-01-01T00:00:00Z"
  }
}
```

Delete commands return `data` as `null` or omit it depending on JSON encoding:

```json
{
  "success": true,
  "message": "project deleted"
}
```

## Error Contract

All errors use the same wrapper:

```json
{
  "success": false,
  "message": "validation failed",
  "error": {
    "code": "validation_error",
    "details": [
      {
        "field": "email",
        "message": "must be a valid email address",
        "tag": "email"
      }
    ]
  }
}
```

Common statuses:

| Status | Code | Frontend action |
| --- | --- | --- |
| 400 | `bad_request` | Show request/query error. |
| 401 | `unauthorized` | Clear token and redirect to login. |
| 403 | `forbidden` | Show permission denied state. |
| 404 | `not_found` | Show not-found state or stale-resource message. |
| 409 | `conflict` | Show duplicate/conflict message. |
| 422 | `validation_error` | Bind `details[].field` to form fields. |
| 429 | `too_many_requests` | Show retry/rate-limit message. |
| 500 | `internal_error` | Show generic retry/support message. |

Protected endpoints consistently use the same error envelope for authentication, authorization, not found, validation, and internal server errors.

## Auth Flow

1. Bootstrap the first admin account if the database has no users:

```http
POST /api/auth/bootstrap
Content-Type: application/json

{
  "name": "Admin User",
  "email": "admin@example.com",
  "password": "secret123"
}
```

2. Login:

```http
POST /api/auth/login
Content-Type: application/json

{
  "email": "admin@example.com",
  "password": "secret123"
}
```

Response:

```json
{
  "success": true,
  "message": "login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.example",
    "token_type": "Bearer",
    "expires_at": "2026-01-02T00:00:00Z",
    "expires_in": 86400,
    "user": {
      "id": "11111111-1111-4111-8111-111111111111",
      "role_id": "11111111-1111-4111-8111-111111111111",
      "role": {
        "id": "11111111-1111-4111-8111-111111111111",
        "name": "admin",
        "description": "System administrator"
      },
      "name": "Admin User",
      "email": "admin@example.com",
      "is_active": true,
      "created_at": "2026-01-01T00:00:00Z",
      "updated_at": "2026-01-01T00:00:00Z"
    }
  }
}
```

3. Send the JWT on protected requests:

```http
Authorization: Bearer <access_token>
```

4. Use `GET /api/auth/me` on app load to hydrate the authenticated user.

5. Use `POST /api/auth/logout` for logout. Tokens are stateless, so the frontend should discard the token locally.

## JWT Storage Recommendation

Prefer storing the access token in memory and re-authenticating on page reload. If persistence is required, use secure, same-site, HTTP-only cookies through a backend-for-frontend layer. Avoid `localStorage` for high-risk deployments because injected scripts can read it.

The current backend issues access tokens only; no refresh-token endpoint exists yet.

## Pagination Contract

Paginated list endpoints:

- `GET /api/users`
- `GET /api/projects`
- `GET /api/sprints`
- `GET /api/statuses`
- `GET /api/tasks`
- `GET /api/audit-logs`
- `GET /api/notifications`

Supported query parameters:

| Parameter | Default | Rule |
| --- | --- | --- |
| `page` | `1` | Values below 1 normalize to 1. |
| `limit` | `20` | Values below 1 normalize to 20; values above 100 normalize to 100. |
| `sort_by` | Endpoint default | Must be one of the endpoint allowlisted fields. |
| `sort_order` | Endpoint default | `asc` or `desc`. |

Paginated response meta:

```json
{
  "page": 1,
  "limit": 20,
  "total": 125,
  "sort_by": "created_at",
  "sort_order": "desc"
}
```

The frontend should calculate pages as `Math.ceil(total / limit)`.

## Search Contract

Search is case-insensitive and partial-match.

| Endpoint | Parameter | Fields searched |
| --- | --- | --- |
| `GET /api/users` | `search` | `name`, `email` |
| `GET /api/projects` | `search` | `project_code`, `project_name`, `client_name` |
| `GET /api/statuses` | `search` | `status_name`, `color_name`, `color_hex` |
| `GET /api/tasks` | `search` | `ticket_number`, `task_title` |

Example:

```http
GET /api/tasks?page=1&limit=20&search=DEV&sort_by=due_date&sort_order=asc
Authorization: Bearer <access_token>
```

## Filtering Contract

| Endpoint | Filters |
| --- | --- |
| `GET /api/users` | `role_id`, `is_active` |
| `GET /api/sprints` | `project_id`, `status` |
| `GET /api/statuses` | `is_active` |
| `GET /api/tasks` | `developer_id`, `project_id`, `sprint_id`, `status_id` |
| `GET /api/dashboard/summary` | `sprint_id` |
| `GET /api/kpi/developers` | `sprint_id` |
| `GET /api/kpi/projects` | `sprint_id` |
| `GET /api/kpi/snapshots` | `sprint_id` |
| `GET /api/audit-logs` | `user_id`, `user`, `module`, `action`, `start_date`, `end_date` |
| `GET /api/workload` | `sprint_id`, `sprint`, `project_id`, `project`, `developer_id`, `status_id`, `start_date`, `end_date` |

UUID filters must be valid UUID strings. Date filters use `YYYY-MM-DD`.

## Sorting Contract

| Endpoint | Default | Allowed `sort_by` |
| --- | --- | --- |
| `GET /api/users` | `created_at desc` | `name`, `email`, `team`, `position`, `is_active`, `created_at`, `updated_at` |
| `GET /api/projects` | `created_at desc` | `project_code`, `project_name`, `client_name`, `status`, `start_date`, `end_date`, `created_at`, `updated_at` |
| `GET /api/sprints` | `start_date desc` | `sprint_name`, `start_date`, `end_date`, `status`, `created_at`, `updated_at` |
| `GET /api/statuses` | `status_order asc` | `status_name`, `color_name`, `color_hex`, `status_order`, `is_done`, `is_qa_status`, `is_active`, `created_at`, `updated_at` |
| `GET /api/tasks` | `created_at desc` | `ticket_number`, `task_title`, `priority`, `estimated_point`, `actual_point`, `start_date`, `due_date`, `completed_date`, `qa_checked_date`, `created_at`, `updated_at` |
| `GET /api/audit-logs` | `created_at desc` | `user_id`, `module`, `action`, `created_at` |
| `GET /api/notifications` | `created_at desc` | `title`, `type`, `is_read`, `read_at`, `created_at` |

Invalid `sort_by` returns `400 bad_request`. Invalid `sort_order` returns `400 bad_request`.

## Endpoint List

### Auth

| Method | Path | Notes |
| --- | --- | --- |
| `POST` | `/api/auth/login` | Public login. |
| `POST` | `/api/auth/logout` | Authenticated user. |
| `POST` | `/api/auth/bootstrap` | Public only before users exist. |
| `GET` | `/api/auth/me` | Authenticated user. |

### Users

| Method | Path | Notes |
| --- | --- | --- |
| `GET` | `/api/users` | Paginated. Requires `manage_users`. |
| `POST` | `/api/users` | Requires `manage_users`. |
| `GET` | `/api/users/:id` | Requires `manage_users`. |
| `PATCH` | `/api/users/:id` | Requires `manage_users`. |
| `DELETE` | `/api/users/:id` | Requires `manage_users`. |

### Projects

| Method | Path | Notes |
| --- | --- | --- |
| `GET` | `/api/projects` | Paginated. Requires `manage_projects`. |
| `POST` | `/api/projects` | Requires `manage_projects`. |
| `GET` | `/api/projects/:id` | Requires `manage_projects`. |
| `PATCH` | `/api/projects/:id` | Requires `manage_projects`. |
| `DELETE` | `/api/projects/:id` | Requires `manage_projects`. |

### Sprints

| Method | Path | Notes |
| --- | --- | --- |
| `GET` | `/api/sprints` | Paginated. Requires `manage_sprints`. |
| `POST` | `/api/sprints` | Requires `manage_sprints`. |
| `GET` | `/api/sprints/:id` | Requires `manage_sprints`. |
| `PATCH` | `/api/sprints/:id` | Requires `manage_sprints`. |
| `PATCH` | `/api/sprints/:id/close` | Requires `manage_sprints`; generates KPI snapshots. |
| `DELETE` | `/api/sprints/:id` | Requires `manage_sprints`. |

### Task Statuses

| Method | Path | Notes |
| --- | --- | --- |
| `GET` | `/api/statuses` | Paginated. Requires `manage_task_statuses`. |
| `POST` | `/api/statuses` | Requires `manage_task_statuses`. |
| `GET` | `/api/statuses/:id` | Requires `manage_task_statuses`. |
| `PATCH` | `/api/statuses/:id` | Requires `manage_task_statuses`. |
| `DELETE` | `/api/statuses/:id` | Requires `manage_task_statuses`. |

### Tasks And Histories

| Method | Path | Notes |
| --- | --- | --- |
| `GET` | `/api/tasks` | Paginated. Scoped by role. |
| `POST` | `/api/tasks` | Requires `manage_tasks`. |
| `GET` | `/api/tasks/:id` | Scoped by role. |
| `PATCH` | `/api/tasks/:id` | Manager full update; Developer/QA status-only rules apply. |
| `DELETE` | `/api/tasks/:id` | Requires `manage_tasks`. |
| `GET` | `/api/tasks/:id/histories` | Scoped by task view permission. |

### Dashboard, KPI, Workload

| Method | Path | Notes |
| --- | --- | --- |
| `GET` | `/api/dashboard/summary` | Requires `view_dashboard`. |
| `GET` | `/api/kpi/developers` | Requires `view_kpi`. Closed sprints use snapshots. |
| `GET` | `/api/kpi/projects` | Requires `view_kpi`. Closed sprints use snapshots. |
| `GET` | `/api/kpi/snapshots` | Snapshot list scoped by role. |
| `GET` | `/api/kpi/snapshots/developer/:developer_id` | Developer can view own only. |
| `POST` | `/api/kpi/snapshots/generate/:sprint_id` | Admin and Project Manager. |
| `GET` | `/api/workload` | Workload list scoped by role. |

### Notifications And Audit Logs

| Method | Path | Notes |
| --- | --- | --- |
| `GET` | `/api/notifications` | Paginated. User sees own; Admin sees all. |
| `GET` | `/api/notifications/unread-count` | User sees own count; Admin sees global count. |
| `PATCH` | `/api/notifications/:id/read` | Marks one visible notification as read. |
| `PATCH` | `/api/notifications/read-all` | Marks visible notifications as read. |
| `GET` | `/api/audit-logs` | Paginated. Admin all; Project Manager task/project/sprint only. |

## DTO Summary

### Auth DTOs

`LoginRequest`:

```json
{
  "email": "admin@example.com",
  "password": "secret123"
}
```

`LoginResponse.data` includes `access_token`, `token_type`, `expires_at`, `expires_in`, and `user`.

### Project DTOs

Create:

```json
{
  "project_code": "DEV",
  "project_name": "Dev Tracker",
  "client_name": "Internal",
  "status": "active",
  "start_date": "2026-01-01",
  "end_date": "2026-03-31"
}
```

Response fields: `id`, `project_code`, `project_name`, `client_name`, `status`, `start_date`, `end_date`, `created_at`, `updated_at`.

### Sprint DTOs

Create:

```json
{
  "project_id": "11111111-1111-4111-8111-111111111111",
  "sprint_name": "Sprint 1",
  "start_date": "2026-01-01",
  "end_date": "2026-01-14",
  "status": "active"
}
```

Response includes the linked `project` object.

### Task Status DTOs

Create:

```json
{
  "status_name": "Ready to Check",
  "color_name": "blue",
  "color_hex": "#3B82F6",
  "status_order": 3,
  "is_done": false,
  "is_qa_status": true,
  "is_active": true
}
```

### Task DTOs

Create:

```json
{
  "developer_id": "11111111-1111-4111-8111-111111111111",
  "project_id": "11111111-1111-4111-8111-111111111111",
  "sprint_id": "11111111-1111-4111-8111-111111111111",
  "ticket_number": "DEV-1",
  "task_title": "Build task board",
  "task_description": "Create backend contract",
  "priority": "high",
  "status_id": "11111111-1111-4111-8111-111111111111",
  "estimated_point": 5,
  "actual_point": 3,
  "start_date": "2026-01-01",
  "due_date": "2026-01-05",
  "note": "Assigned"
}
```

Response includes linked `developer`, `project`, `sprint`, and `status` objects.

### Notification DTOs

`GET /api/notifications` returns:

```json
{
  "notifications": [
    {
      "id": "11111111-1111-4111-8111-111111111111",
      "user_id": "11111111-1111-4111-8111-111111111111",
      "title": "Task ready to check",
      "message": "Task moved to Ready to Check: Build task board",
      "type": "task_ready_to_check",
      "reference_module": "tasks",
      "reference_id": "11111111-1111-4111-8111-111111111111",
      "is_read": false,
      "created_at": "2026-01-01T00:00:00Z"
    }
  ],
  "unread_count": 1
}
```

## Frontend Error Handling Guide

- On `401`, clear auth state and route to login.
- On `403`, keep auth state and show an access-denied message.
- On `404`, invalidate the missing resource from cache and show not found.
- On `409`, show the conflict message near the relevant field or action.
- On `422`, map `error.details[].field` to form controls.
- On `500`, show a generic retry message and preserve user-entered form data.

Use the `message` field for user-facing text unless the UI has a more specific localized message for the error code.
