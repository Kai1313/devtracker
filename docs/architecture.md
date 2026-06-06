# System Architecture
# Load Developer Sheets System

## 1. Architecture Style

Use a modular monolith architecture.

This is recommended because the system domain is still evolving, especially the KPI rules. A modular monolith keeps development simple while still separating business domains cleanly.

## 2. High-Level Architecture

```text
Users
  ↓
Next.js Frontend
  ↓ HTTPS REST API
Golang Backend API
  ↓
PostgreSQL Database
```

Optional components:

```text
Redis for session cache, token blacklist, and frequently accessed dashboard data.
File storage for future import/export files.
```

## 3. Frontend

Technology:

- Next.js
- TypeScript
- Tailwind CSS
- shadcn/ui
- TanStack Query
- TanStack Table
- React Hook Form
- Zod validation

Frontend modules:

- Login
- Dashboard
- Users / Developers
- Projects
- Sprints
- Task Board / Task Table
- Status Management
- KPI Dashboard
- Reports

## 4. Backend

Technology:

- Golang
- Fiber or Gin
- GORM
- PostgreSQL
- JWT authentication
- go-playground/validator
- Zerolog or Zap

Recommended structure:

```text
backend/
  cmd/
    api/
      main.go
  internal/
    config/
    database/
    middleware/
    auth/
    user/
    project/
    sprint/
    task/
    status/
    kpi/
    report/
  pkg/
    response/
    errors/
    validator/
```

Each business module should follow:

```text
handler.go
service.go
repository.go
model.go
dto.go
route.go
```

## 5. Backend Layers

### Handler Layer
Responsible for:

- HTTP request parsing
- Input validation
- Calling service layer
- Returning standardized response

No business logic should be placed here.

### Service Layer
Responsible for:

- Business rules
- Transaction handling
- KPI calculation logic
- Status transition rules

### Repository Layer
Responsible for:

- Database query
- GORM implementation
- Data persistence

### Model Layer
Responsible for:

- Database entity definition
- GORM tags
- Table relationships

## 6. Data Storage

Primary database:

- PostgreSQL

Core tables:

- users
- roles
- projects
- sprints
- task_statuses
- tasks
- task_histories
- comments
- attachments
- kpi_snapshots

## 7. Important Design Decision

Task status changes must always create a row in `task_histories`.

This enables future KPI metrics such as:

- Time from In Progress to Ready to Check
- Time from Ready to Check to QA Checked
- Time from QA Checked to Done
- Delayed task tracking
- Developer completion trend

## 8. Deployment Architecture

```text
DNS
 ↓
Nginx Reverse Proxy
 ↓
Frontend Server: Next.js
 ↓
Backend Server: Golang API
 ↓
PostgreSQL
```

Optional:

```text
Redis
File Storage
Monitoring
CI/CD Pipeline
```

## 9. Security

- Use HTTPS.
- Store JWT in HTTP-only cookie where possible.
- Use role-based access control.
- Validate all API inputs.
- Use parameterized queries through GORM.
- Add audit trail for task status changes.
- Add rate limiting for authentication endpoints.

## 10. KPI-Ready Architecture

KPI should be calculated from raw task and history data.

Do not manually input KPI unless required for adjustment.

Future KPI sources:

- tasks
- task_histories
- sprints
- users
- projects
