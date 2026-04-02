# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Domain Context

Restoration Druid talent analyzer for World of Warcraft. Fetches combat log data from the WarcraftLogs (WCL) v2 GraphQL API and attributes healing done to individual talent choices, helping players evaluate talent effectiveness.

**Before working on talent logic, always read `docs/resto_druid_talents.md` first** to understand all talents and their mechanics.

## Commands

```bash
# Install (uses uv)
uv sync --all-extras

# Run CLI (single command, "analyze" subcommand can be omitted)
uv run flourish <report_code> [--fight ID] [--player NAME] [--config-path PATH]

# Run web UI (FastAPI + Vue 3 frontend)
uv run uvicorn flourish.web.app:create_app --factory --reload

# Tests
uv run pytest                        # all tests
uv run pytest tests/analysis/        # specific module
uv run pytest -k test_name           # single test by name

# Frontend (Vue 3 + Vite + TailwindCSS)
cd frontend && npm install && npm run dev   # dev server
cd frontend && npm run build                # build to dist/ (served by FastAPI)

# Environment
cp .env.example .env          # then fill in WCL_CLIENT_ID / WCL_CLIENT_SECRET
```

## Architecture

**Data flow:** WCL API → event parsing → tracking → attribution → output

| Layer              | Package                          | Purpose                                                    |
| ------------------ | -------------------------------- | ---------------------------------------------------------- |
| API client         | `wcl/client.py`                  | OAuth2 + paginated GraphQL event fetching                  |
| Event models       | `models/events.py`               | Typed dataclasses per WCL event type; `parse_event()` maps raw dicts |
| Talent config      | `models/config.py`               | Loads `config/talents.yaml` → `dict[str, TalentConfig]`   |
| State tracking     | `tracking/hot_tracker.py`        | Tracks active HoTs per (target, spell) with taggable `HotInstance` |
| State tracking     | `tracking/buff_tracker.py`       | Tracks active self-buffs by buff_id                        |
| Attribution        | `analysis/attributor.py`         | `TalentAttributor` ABC: `process_event()` + `process_heal()` |
| Attribution        | `analysis/talents/`              | One subclass per talent (47 attributors across spec + hero trees) |
| Attribution        | `analysis/talents/keeper/`       | Keeper of the Grove hero talent attributors                 |
| Attribution        | `analysis/talents/wildstalker/`  | Wildstalker hero talent attributors                         |
| Pipeline           | `analysis/pipeline.py`           | Orchestrates: parse → track → attribute → `AnalysisResults` |
| Output             | `output/table.py`                | Rich table rendering with HPS and % total                  |
| CLI                | `cli.py`                         | Typer app; interactive fight/player selection               |
| Web API            | `web/app.py`                     | FastAPI app with rate limiting (SlowAPI) and result caching |
| Web API            | `web/routes.py`                  | `/api/report/{code}`, `/api/analyze/{code}/{fight}/{player}` |
| Frontend           | `frontend/`                      | Vue 3 + Vite + TailwindCSS SPA; served from `frontend/dist/` |

### Key concepts

- **Overheal filtering:** heals >50% overheal (`OVERHEAL_WASTE_THRESHOLD`) are excluded from attribution (`is_wasted` on `HealEvent`).
- **HoT tagging:** attributors tag `HotInstance.tags` during `process_event` (e.g., `"sotf"` tag), then check tags during `process_heal` to calculate bonus healing.
- **Mastery attribution:** Two mastery-aware attributors (Harmonious Blooming, Symbiotic Bloom) use configurable `base_stacks` and `dr_table` from `config/talents.yaml` to calculate diminishing-returns mastery bonus from extra HoT stacks.
- **Adding a new talent:** create a `TalentAttributor` subclass in `analysis/talents/` (or `keeper/`/`wildstalker/` for hero talents), implement `process_event`/`process_heal`, and add it to the registry in `cli.py:build_attributors()`. The talent must also have an entry in `config/talents.yaml`.
- **Talent ID types — do not confuse:**
  - `nodeID` — talent tree position. Used for `talent_node_id` on attributors. Same across all ID systems.
  - WCL entry ID (`talentTree[].id`) — what WCL returns in combatantinfo. Used for `talent_id` on attributors to disambiguate choice nodes. Look these up in `docs/raidbots_druid_talents.json` under `entries[].id`.
  - Blizzard definition ID (`talent.id` in Blizzard API) — **not** the same as WCL entry ID. Shown in `docs/resto_druid_talents.md` as "Definition ID". Do **not** use these in attributor code.

## Config

`config/talents.yaml` — declares all talents with `skip` (bool), optional `skip_reason`, and optional `multiplier`. Talents marked `skip: true` are excluded from analysis (baseline spells, non-healing utility).
