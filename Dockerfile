# syntax=docker/dockerfile:1

FROM golang:1.25-alpine AS builder

WORKDIR /src/backend

RUN apk add --no-cache ca-certificates git

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/devtracker-api ./cmd/api

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata wget \
	&& addgroup -S app \
	&& adduser -S app -G app

WORKDIR /app

COPY --from=builder --chown=app:app /out/devtracker-api /app/devtracker-api

ENV APP_ENV=production \
	APP_PORT=8080 \
	APP_BASE_PATH=/api \
	DB_RUN_MIGRATIONS=true

EXPOSE 8080

USER app

HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
	CMD wget -qO- "http://127.0.0.1:${APP_PORT}${APP_BASE_PATH}/health" >/dev/null || exit 1

ENTRYPOINT ["/app/devtracker-api"]
