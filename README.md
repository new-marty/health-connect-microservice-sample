# health-connect

Go microservice backend for personal health data. Collects from Oura Ring, Strava, Hevy, InBody, and Apple Health. Provides REST API for all health metrics, cross-source aggregation, and AI-powered analysis.

## Quick Start

```bash
cp .env.example .env   # Fill in API credentials
make run               # go run ./cmd/server
```

## API

All endpoints under `/api/v1/`.

| Source | Endpoints |
|--------|-----------|
| Oura | `GET /oura/sleep`, `/oura/sleep/:day`, `/oura/scores`, `/oura/scores/:day` |
| Strava | `GET /strava/activities`, `/strava/activities/:id` |
| Hevy | `GET /hevy/workouts`, `/hevy/workouts/:id` |
| InBody | `GET /inbody/scans`, `/inbody/scans/latest` |
| Apple Health | `GET /apple-health/weight`, `POST /apple-health/weight`, `GET /apple-health/vitals`, `POST /apple-health/ingest` |
| Meals | `GET /meals`, `POST /meals`, `DELETE /meals/:id` |
| Summary | `GET /summary/daily`, `/summary/weekly`, `/summary/report` |
| Analysis | `POST /analysis` |
| Sync | `GET /sync/status`, `POST /sync/:source` |

Health check: `GET /api/health`

## Docker

```bash
docker build -t health-connect .
docker run -p 8080:8080 -v ./data:/data --env-file .env health-connect
```

## Architecture

Source-oriented modules — each data source (oura, strava, hevy, inbody, applehealth) is self-contained with its own model, repository, service, handler, and sync client. No cross-imports between source modules.

Summary layer reads across all sources (read-only aggregation). AI analysis calls Claude API.

Background sync runs on configurable cron schedules and can be triggered on-demand via `POST /api/v1/sync/:source`.
