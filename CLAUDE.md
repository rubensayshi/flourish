# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Domain Context

Restoration Druid talent analyzer for World of Warcraft. Fetches combat log data from the WarcraftLogs (WCL) v2 GraphQL API and attributes healing done to individual talent choices, helping players evaluate talent effectiveness.

**Before working on talent logic, always read `docs/resto_druid_talents.md` first** to understand all talents and their mechanics.

## Commands

```bash
# Install (uses uv)
uv sync --all-extras

# Run CLI
uv run rdruid-analyzer analyze <report_code> [--fight ID] [--player NAME] [--config-path PATH]

# Tests
uv run pytest                        # all tests
uv run pytest tests/analysis/        # specific module
uv run pytest -k test_name           # single test by name

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
| Attribution        | `analysis/talents/`              | One subclass per talent                                    |
| Pipeline           | `analysis/pipeline.py`           | Orchestrates: parse → track → attribute → `AnalysisResults` |
| Output             | `output/table.py`                | Rich table rendering with HPS and % total                  |
| CLI                | `cli.py`                         | Typer app; interactive fight/player selection               |

### Key concepts

- **Overheal filtering:** heals >50% overheal (`OVERHEAL_WASTE_THRESHOLD`) are excluded from attribution (`is_wasted` on `HealEvent`).
- **HoT tagging:** attributors tag `HotInstance.tags` during `process_event` (e.g., `"sotf"` tag), then check tags during `process_heal` to calculate bonus healing.
- **Adding a new talent:** create a `TalentAttributor` subclass in `analysis/talents/`, implement `process_event`/`process_heal`, and add it to the registry in `cli.py:build_attributors()`. The talent must also have an entry in `config/talents.yaml`.

## Config

`config/talents.yaml` — declares all talents with `skip` (bool), optional `skip_reason`, and optional `multiplier`. Talents marked `skip: true` are excluded from analysis (baseline spells, non-healing utility).
