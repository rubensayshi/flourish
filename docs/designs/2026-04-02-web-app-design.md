# Web App Design

Expose the Resto Druid Talent Analyzer as a web app so users can paste a WCL report URL, select fight + player, and view talent attribution results in the browser.

## Decisions

- **Shared WCL API key** — zero friction for users; rate-limit per IP to protect quota
- **Monorepo, single Fly service** — FastAPI serves both API and built Vue static files
- **Permalink via natural key** — `/results/<code>/<fight>/<player>`, no database needed
- **Fly.io hosting** — free tier eligible, single machine + volume for caching

## API

```
GET  /api/health
  -> 200

GET  /api/report/{code}
  -> { title, fights: [{id, name, kill, duration}], druids: [{id, name, server}] }

GET  /api/analyze/{code}/{fight_id}/{player_name}
  -> { fight_name, player_name, total_healing,
       talents: [{name, attributed, pct, hps}],
       wasted, unattributed }
```

`/api/report` feeds fight/player dropdowns (single WCL query). `/api/analyze` runs the full pipeline and returns results as JSON. Also serves as the permalink data source.

**Caching:** Analysis results cached on disk keyed by `{code}/{fight}/{player}`. Existing `CachedWCLClient` handles WCL API response caching.

**Rate limiting:** IP-based, ~10 analyses/min/IP via FastAPI middleware.

## Frontend

**Stack:** Vue 3 + Tailwind CSS, no component library.

```
frontend/
  src/
    App.vue              # Router wrapper
    views/
      Home.vue           # Landing page + report URL input
      Analyze.vue        # Fight/player selectors -> results table
    components/
      ReportInput.vue    # URL input + validation
      FightSelector.vue  # Dropdown (kill/wipe indicator, duration)
      PlayerSelector.vue # Dropdown for druids
      ResultsTable.vue   # Talent attribution table
      LoadingSpinner.vue
```

**Flow:**

1. User lands on `/` -> URL input
2. Pastes WCL URL -> extracts report code, calls `/api/report/{code}`
3. Routes to `/analyze/{code}` -> fight + player dropdowns
4. Selects both -> calls `/api/analyze/{code}/{fight}/{player}`
5. Results render, URL updates to `/results/{code}/{fight}/{player}` (shareable)

**Permalink:** Direct navigation to `/results/{code}/{fight}/{player}` fetches report metadata and analysis in parallel, skipping selection steps.

**Styling:**

- Dark theme by default (dark slate/gray)
- Resto Druid green accents (#00FF98 / emerald)
- Compact table with alternating stripes, color-coded percentages
- Centered content, max-width ~800px
- Responsive but desktop-optimized

## Backend Structure

```
src/flourish/
  web/
    app.py           # FastAPI app, mounts static files, catch-all for Vue routing
    routes.py        # /api/report, /api/analyze, /api/health
    cache.py         # Disk-based result cache (JSON files)
  # Existing code unchanged:
  wcl/
  analysis/
  models/
  output/
```

Thin wrapper layer. `routes.py` reuses `get_wcl_client()`, `build_attributors()`, and `Pipeline` directly. No changes to the analysis engine.

## Deployment

**Fly.io config:**

- `shared-cpu-1x`, 256MB RAM (free tier)
- 1GB Fly volume at `/data` for caches
- Secrets: `WCL_CLIENT_ID`, `WCL_CLIENT_SECRET` via `fly secrets set`
- Health check: `GET /api/health`
- Internal port 8080, HTTPS on 443

**Dockerfile (multi-stage):**

```dockerfile
# Stage 1: Build Vue
FROM node:22-alpine AS frontend
WORKDIR /app/frontend
COPY frontend/ .
RUN npm ci && npm run build

# Stage 2: Python app
FROM python:3.12-slim
COPY --from=frontend /app/frontend/dist /app/static
COPY . /app
RUN pip install uv && cd /app && uv sync
CMD ["uv", "run", "uvicorn", "flourish.web.app:app", "--host", "0.0.0.0", "--port", "8080"]
```

## WCL API Budget

| Tier     | Points/hour | ~Analyses/hour |
| -------- | ----------- | -------------- |
| Free     | 3,600       | 100-300        |
| Gold     | 9,000       | 300-900        |
| Platinum | 18,000      | 600-1800       |

Free tier is sufficient to start. Upgrade if traffic grows.
