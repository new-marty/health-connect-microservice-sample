# CLAUDE.md

## What This Is

A Go microservice backend for personal health data. Collects from Oura Ring, Strava, Hevy, InBody, and Apple Health. Part of a larger microservice ecosystem.

Three callable surfaces, all backed by the same handlers: REST API, Cobra CLI, and LLM tool-calling (auto-generated from swaggo annotations and served at `/openapi.json`, `/tools/openai.json`, `/tools/anthropic.json`).

## Key Commands

```bash
just run               # Run the server
just build             # Build binaries (server + cli)
just test              # Run tests
just openapi           # Regenerate OpenAPI spec from swaggo annotations
just cli oura sleep list   # Run CLI subcommand
go run ./cmd/server    # Run server directly
```

`HEALTH_CONNECT_API_TOKEN=<token>` enables Bearer auth on the API. The CLI reads the same env var and injects the header automatically. If the env var is unset, the API is open (local-only convenience).

## Architecture

Source-oriented modules — each data source is self-contained under `internal/{source}/` with its own model, repository, service, handler, and sync client. No cross-imports between source modules.

- `internal/oura/` — Oura Ring (sleep, readiness, activity, HRV)
- `internal/strava/` — Strava (running, cycling)
- `internal/hevy/` — Hevy (gym workouts + sets)
- `internal/inbody/` — InBody (body composition scans)
- `internal/applehealth/` — Apple Health (weight, vitals, webhook ingest)
- `internal/meals/` — Manual food logging
- `internal/summary/` — Cross-source aggregation (read-only from other repos)
- `internal/analysis/` — AI-powered health insights (Claude API)
- `internal/sync/` — Background cron scheduler + on-demand sync

## Rules

- **1:1 CLI ↔ API ↔ tool parity**: Every API endpoint MUST have (a) a matching CLI command and (b) full swaggo annotations on the handler so it shows up as an LLM tool. After adding/changing a handler annotation, run `just openapi` to regenerate the spec at `internal/spec/swagger.json`.
- All handlers use `apperror.RespondGin(c, err)` for error responses so the wire envelope (`{error: {code, message, details}}`) stays uniform — LLM callers parse this shape.
- No cross-imports between source modules (oura, strava, hevy, inbody, applehealth, meals). Only `summary/` and `analysis/` may read from multiple sources.
- Follow finance-connect patterns: Go 1.23, Gin, slog, graceful shutdown.
- SQLite with `modernc.org/sqlite` (pure Go, no CGO).
