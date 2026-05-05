# health-connect

Go microservice backend for personal health data. Collects from Oura Ring, Strava, Hevy, InBody, and Apple Health. Provides REST API for all health metrics, cross-source aggregation, and AI-powered analysis.

## Quick Start

```bash
cp .env.example .env   # Fill in API credentials
just run               # go run ./cmd/server
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

## AI tool calling

Every API endpoint is also exposed as an LLM-callable tool. The discovery URLs:

| URL | Format |
|-----|--------|
| `GET /openapi.json` | OpenAPI / Swagger 2.0 spec (source of truth) |
| `GET /tools/openai.json` | OpenAI function-calling tool array |
| `GET /tools/anthropic.json` | Anthropic tool-use tool array |

The schemas are pre-converted at server startup. The LLM picks a tool, then **invokes the existing REST route directly** (each tool description ends with `Invoke: <METHOD> <path>`). No proxy/dispatch endpoint is needed.

Auth: set `HEALTH_CONNECT_API_TOKEN=<your-token>` on the server. Hosted agents (Claude, ChatGPT) pass `Authorization: Bearer <token>`. The CLI reads the same env var and forwards it. If the env var is unset, auth is disabled and a startup warning is logged.

After changing handler annotations, regenerate the spec:

```bash
just openapi
```

## Docker

```bash
docker build -t health-connect .
docker run -p 8080:8080 -v ./data:/data --env-file .env health-connect
```

## Architecture

Source-oriented modules — each data source (oura, strava, hevy, inbody, applehealth) is self-contained with its own model, repository, service, handler, and sync client. No cross-imports between source modules.

Summary layer reads across all sources (read-only aggregation). AI analysis calls Claude API.

Background sync runs on configurable cron schedules and can be triggered on-demand via `POST /api/v1/sync/:source`.

OpenAPI annotations (swaggo) live on each handler — `just openapi` regenerates `internal/spec/swagger.json`, which is embedded into the binary and served at `/openapi.json`. `internal/toolspec/` converts the spec to OpenAI/Anthropic tool-call shapes at startup.
