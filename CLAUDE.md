# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Domain Context

Restoration Druid talent analyzer for World of Warcraft. Fetches combat log data from the WarcraftLogs (WCL) v2 GraphQL API and attributes healing done to individual talent choices, helping players evaluate talent effectiveness.

**Before working on talent logic, always read `docs/resto_druid_talents.md` first** to understand all talents and their mechanics.

## Commands

```bash
# Run CLI (single command, "analyze" subcommand can be omitted)
cd go-backend && go run ./cmd/flourish/ <report_code> [--fight ID] [--player NAME]

# Run web server
cd go-backend && go run ./cmd/flourish/ serve [--port PORT]

# Tests
cd go-backend && go test ./...                    # all tests
cd go-backend && go test ./internal/talents/      # specific package
cd go-backend && go test -run TestName ./...       # single test by name

# Frontend (Vue 3 + Vite + TailwindCSS)
cd frontend && npm install && npm run dev          # dev server (proxies /api to :8000)
cd frontend && npm run build                       # build to dist/ (served by Go server)

# Environment
cp .env.example .env          # then fill in WCL_CLIENT_ID / WCL_CLIENT_SECRET
```

## Architecture

**Data flow:** WCL API → event parsing → tracking → attribution → output

| Layer           | Package                              | Purpose                                                    |
| --------------- | ------------------------------------ | ---------------------------------------------------------- |
| API client      | `go-backend/internal/wcl/`           | OAuth2 + paginated GraphQL event fetching, disk cache      |
| Event models    | `go-backend/internal/models/`        | Typed structs per WCL event type; `ParseEvent()` maps raw dicts |
| Talent config   | `go-backend/internal/models/`        | Loads `config/talents.yaml` → `Config` struct              |
| State tracking  | `go-backend/internal/tracking/`      | HotTracker (active HoTs with tags) + BuffTracker           |
| Attribution     | `go-backend/internal/talents/`       | `TalentAttributor` interface + `BaseAttributor` embedded struct |
| Pipeline        | `go-backend/internal/analysis/`      | Orchestrates: parse → track → attribute → `AnalysisResults` |
| Output          | `go-backend/internal/output/`        | Text table rendering with HPS and % total                  |
| Web API         | `go-backend/internal/web/`           | chi router, rate limiting, OAuth, result caching           |
| CLI             | `go-backend/cmd/flourish/`           | cobra CLI: `analyze` + `serve` subcommands                 |
| Frontend        | `frontend/`                          | Vue 3 + Vite + TailwindCSS SPA; served from `frontend/dist/` |

### Key concepts

- **Overheal filtering:** heals >50% overheal are excluded from attribution (`IsWasted()` on `HealEvent`).
- **HoT tagging:** attributors tag `HotInstance.Tags` during `ProcessEvent()` (e.g., `"sotf"` tag), then check tags during `ProcessHeal()` to calculate bonus healing.
- **Mastery attribution:** Two mastery-aware attributors (Harmonious Blooming, Symbiotic Bloom) use configurable `BaseStacks` and `DRTable` from `config/talents.yaml`.
- **Adding a new talent:** create a struct implementing `TalentAttributor` in `go-backend/internal/talents/`, implement `ProcessEvent`/`ProcessHeal`, and add it to `BuildAttributors()` in `go-backend/internal/web/attributors.go`. The talent must also have an entry in `config/talents.yaml`.
- **Talent ID types — do not confuse:**
  - `nodeID` — talent tree position. Used for `TalentNodeID()` on attributors. Same across all ID systems.
  - WCL entry ID (`talentTree[].id`) — what WCL returns in combatantinfo. Used for `TalentID()` on attributors to disambiguate choice nodes. Look these up in `docs/raidbots_druid_talents.json` under `entries[].id`.
  - Blizzard definition ID (`talent.id` in Blizzard API) — **not** the same as WCL entry ID. Shown in `docs/resto_druid_talents.md` as "Definition ID". Do **not** use these in attributor code.

## Config

`config/talents.yaml` — declares all talents with `skip` (bool), optional `skip_reason`, and optional `multiplier`. Talents marked `skip: true` are excluded from analysis (baseline spells, non-healing utility).
