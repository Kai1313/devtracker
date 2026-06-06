# Product Requirements Document
# Load Developer Sheets System

## 1. Purpose

The Load Developer Sheets System is a web-based application to replace spreadsheet-based sprint developer task tracking.

The system records developer assignments, project names, sprint information, ticket numbers, task details, and task statuses using configurable colors. The data will later become the source for developer KPI, sprint performance, QA tracking, and workload analysis.

## 2. Main Goals

- Replace manual spreadsheet tracking with a centralized web app.
- Track developer tasks per sprint and project.
- Track task status visually using configurable status colors.
- Record status change history for KPI calculation.
- Provide dashboards for sprint progress and developer workload.
- Prepare structured data for future KPI reports.

## 3. User Roles

### Admin
- Manage users and roles.
- Manage projects.
- Manage sprints.
- Manage task statuses and colors.
- View all reports and KPI.

### Project Manager
- Create and manage sprint tasks.
- Assign developers to tasks.
- Update task details.
- View workload and sprint progress.

### Developer
- View assigned tasks.
- Update task progress.
- Add task notes or comments.

### QA
- View tasks ready for checking.
- Update QA-related statuses.
- Mark tasks as checked or returned.

### Management
- View dashboard, reports, and KPI only.

## 4. Core Features

### Authentication and Authorization
- Login and logout.
- Role-based access control.
- JWT authentication.
- Protected API endpoints.

### User / Developer Management
- Create, update, delete, and view users.
- Assign roles.
- Mark users as active or inactive.

### Project Management
- Create, update, delete, and view projects.
- Store project name, code, client name, and status.

### Sprint Management
- Create, update, delete, and view sprints.
- Link sprints to projects.
- Store sprint start date, end date, and status.

### Task Assignment Tracking
- Create, update, delete, and view tasks.
- Assign developer, project, sprint, ticket number, task title, description, priority, and status.
- Filter by sprint, project, developer, status, and ticket number.

### Status Management
- Create configurable task statuses.
- Store status name, color, order, and category.
- Example statuses:
  - Todo: gray
  - In Progress: yellow
  - Ready to Check: blue
  - Checked by QA: orange
  - Done: green
  - Blocked: red

### Task History
- Every status change must be logged.
- Store old status, new status, changed by, changed at, and note.
- This table will be the foundation for KPI calculation.

### Dashboard
- Total active tasks.
- Tasks by status.
- Sprint progress.
- Developer workload.
- Tasks ready for QA.
- Blocked tasks.

### KPI Foundation
- Total assigned tasks per developer.
- Total completed tasks per developer.
- Completion rate.
- QA checked count.
- Delayed task count.
- Average completion duration.

## 5. MVP Scope

The first version should include:

- Authentication
- User management
- Project management
- Sprint management
- Task tracking
- Status color management
- Task history logging
- Basic dashboard

Advanced KPI, exports, and detailed reporting can be improved after the MVP.

## 6. Non-Functional Requirements

- Web-based responsive UI.
- REST API backend.
- PostgreSQL database.
- Audit-friendly task history.
- Docker-based deployment.
- Clean modular code structure.
- Easy to extend into KPI and reporting modules.
