# Load Developer Sheets System Starter

This repository starter contains planning and technical documents for building a web-based sprint developer task tracking system.

## Purpose

Replace spreadsheet-based developer assignment tracking with a structured web app using:

- Golang backend
- Next.js frontend
- PostgreSQL database

## Included Files

```text
docs/prd.md              Product requirements document
docs/architecture.md     System architecture document
database/schema.sql      Initial PostgreSQL schema
api/openapi.yaml         Initial REST API contract
CODEX.md                 Instructions for Codex
```

## Recommended Development Order

1. Review PRD.
2. Review architecture.
3. Create database migration.
4. Build backend API.
5. Build frontend screens.
6. Add KPI and reports.

## MVP Modules

- Authentication
- User / Developer Management
- Project Management
- Sprint Management
- Task Tracking
- Status Management
- Task History
- Basic Dashboard

## Running With Docker

Create an environment file from the example, then start the full backend stack:

```bash
cp .env.example .env
docker compose up --build
```

The Docker stack includes the backend API, PostgreSQL, and Redis. PostgreSQL data is persisted in the `postgres_data` volume, Redis data is persisted in `redis_data`, and the backend waits for both services to pass health checks before starting.

The backend runs migrations automatically on startup when `DB_RUN_MIGRATIONS=true`, which is the default in `.env.example`.

Useful URLs:

- API health check: `http://localhost:8080/api/health`
- Swagger UI: `http://localhost:8080/swagger`

## Local Development

Start only the local infrastructure:

```bash
docker compose up -d postgres redis
```

Run the backend from source:

```bash
cd backend
go run ./cmd/api
```

For local source runs, set database variables for your shell if they differ from the defaults. The app reads environment variables directly; Docker Compose loads `.env` for container runs.

## Production Build

Build the production backend image:

```bash
docker build -t devtracker-api:prod .
```

Or build through Compose:

```bash
docker compose build app
```

## CI Pipeline

GitHub Actions runs the backend pipeline from `.github/workflows/backend-ci.yml` on pushes and pull requests that touch backend or deployment files.

The pipeline checks:

- Repository checkout
- Go 1.25 setup
- Dependency download with `go mod download`
- Go formatting with `gofmt -l`
- Static checks with `go vet ./...`
- Tests with `go test ./...`
- Backend binary build
- Docker image build check

CI environment variables:

```text
GO_VERSION=1.25.0
CGO_ENABLED=0
APP_ENV=test
APP_PORT=8080
APP_BASE_PATH=/api
DB_RUN_MIGRATIONS=false
JWT_SECRET=ci-test-secret-change-me
JWT_ISSUER=devtracker-ci
JWT_ACCESS_TOKEN_TTL=24h
```

The current unit test suite does not require external database or Redis services. Add CI secrets or service containers later if integration tests need live infrastructure.

Recommended branch protection:

- Protect `main`.
- Require pull requests before merging.
- Require the `Backend checks` status check to pass.
- Require branches to be up to date before merging.
- Restrict direct pushes to release maintainers.
