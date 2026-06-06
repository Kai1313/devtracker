# CODEX Instructions
# Load Developer Sheets System

## Project Summary

Build a web-based replacement for a spreadsheet used to track developer sprint assignments.

The system records developer, project, sprint, ticket number, task details, task status, and status colors. The data must support future KPI calculations.

## Tech Stack

### Backend
- Golang
- Fiber preferred
- GORM
- PostgreSQL
- JWT authentication
- Modular monolith architecture

### Frontend
- Next.js
- TypeScript
- Tailwind CSS
- shadcn/ui
- TanStack Query
- TanStack Table
- React Hook Form
- Zod

### Database
- PostgreSQL

### Deployment
- Docker
- Nginx reverse proxy

## Code Rules

### Backend Rules

- Use repository-service-handler pattern.
- Do not put business logic in handlers.
- Use DTOs for request and response payloads.
- Use standardized JSON response format.
- Use standardized error response format.
- Use transactions when updating task status and inserting task history.
- Use soft delete for main business tables.
- Always validate request input.
- Never skip task history when task status changes.

### Frontend Rules

- Use TypeScript strictly.
- Use reusable components.
- Use TanStack Query for API state.
- Use React Hook Form and Zod for forms.
- Keep API functions in `lib/api` or `services`.
- Do not hardcode status colors if they come from API.

## Important Business Rules

1. Every task belongs to one project and one sprint.
2. Every task must have one assigned developer.
3. Every task must have one current status.
4. Every task status change must create a task history record.
5. Status color must be configurable.
6. KPI must be calculated from task and task history data.
7. Done status should set completed_date.
8. Checked by QA status should set qa_checked_date.

## Recommended First Tasks for Codex

1. Generate backend project structure.
2. Generate database migration based on `database/schema.sql`.
3. Generate shared response and error package.
4. Generate auth module.
5. Generate user module.
6. Generate project module.
7. Generate sprint module.
8. Generate task status module.
9. Generate task module with history logging.
10. Generate Next.js frontend layout and task table.

## Do Not Do

- Do not build microservices.
- Do not skip authentication.
- Do not hardcode KPI results.
- Do not remove task history.
- Do not place SQL directly in handlers.
