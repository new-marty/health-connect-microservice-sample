# CLAUDE.md

## What This Is

A Go microservice backend for personal health data. Collects from Oura Ring, Strava, Hevy, InBody, and Apple Health. Part of a larger microservice ecosystem.

## Key Commands

```bash
make run               # Run the server
make build             # Build binary
make test              # Run tests
go run ./cmd/server    # Run server directly
go run ./cmd/cli       # Run CLI
```

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

- **1:1 CLI-API parity**: Every API endpoint MUST have a corresponding CLI command, and vice versa. When adding a new endpoint, always add both the handler and the CLI command.
- No cross-imports between source modules (oura, strava, hevy, inbody, applehealth, meals). Only `summary/` and `analysis/` may read from multiple sources.
- Follow finance-connect patterns: Go 1.23, Gin, slog, graceful shutdown.
- SQLite with `modernc.org/sqlite` (pure Go, no CGO).
