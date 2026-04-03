# Go Backend Rebuild Design

**Date:** 2026-04-03
**Goal:** Rewrite the Python backend in Go to reduce memory footprint. TDD approach — transpile all Python tests to Go first, then implement.

## Project Structure

```
go-backend/
  cmd/flourish/main.go
  internal/
    models/          # event types, config, parse_event
    tracking/        # HotTracker, BuffTracker
    analysis/        # Pipeline, AnalysisResults
    talents/         # TalentAttributor interface + 27 implementations
    wcl/             # GraphQL client, OAuth2, response caching
    web/             # chi router, handlers, middleware
    output/          # table rendering (Rich → terminal table)
  config/talents.yaml
  go.mod
```

Talents are flat — no subdirectories. File naming: `keeper_*.go`, `wildstalker_*.go` for hero talents.

## Data Models

### Events

```go
type Event interface {
    GetBase() *BaseEvent
}

type BaseEvent struct {
    Timestamp int
    SourceID  int
    Type      string
}

type HealEvent struct {
    BaseEvent
    TargetID  int
    AbilityID int
    Amount    int
    Overheal  int
    Absorb    int
    HitType   int  // 1=normal, 2=crit
    Tick      bool
}
// Methods: RawHeal() int, OverhealPct() float64, IsWasted() bool

type CastEvent struct {
    BaseEvent
    TargetID  int
    AbilityID int
}

type ApplyBuffEvent struct {
    BaseEvent
    TargetID  int
    AbilityID int
}

type RefreshBuffEvent struct {
    BaseEvent
    TargetID  int
    AbilityID int
}

type RemoveBuffEvent struct {
    BaseEvent
    TargetID  int
    AbilityID int
}

type SummonEvent struct {
    BaseEvent
    TargetID  int
    AbilityID int
}

type CombatantInfoEvent struct {
    BaseEvent
    TalentNodes  map[int]bool    // set of nodeIDs
    TalentIDs    map[int]bool    // set of WCL entry IDs
    TalentRanks  map[int]int     // entryId -> rank
    CritSpell    float64
    HasteSpell   float64
    Mastery      float64
    SpecID       int
}
```

`ParseEvent(raw map[string]any) Event` — returns concrete type behind `Event` interface. Consumers type-switch.

`OVERHEAL_WASTE_THRESHOLD = 0.5` — constant, same as Python.

### Config

```go
type MasteryConfig struct {
    BaseStacks int       `yaml:"base_stacks"`
    DRTable    []float64 `yaml:"dr_table"`
}

type TalentConfig struct {
    Skip       bool    `yaml:"skip"`
    SkipReason string  `yaml:"skip_reason"`
    Multiplier float64 `yaml:"multiplier"`
}

type Config struct {
    Mastery MasteryConfig
    Talents map[string]TalentConfig
}
```

Loaded via `yaml.v3` with custom `UnmarshalYAML` — the existing `talents.yaml` has `mastery:` as a top-level key mixed with talent keys at the same level, so we parse `mastery` first, then treat remaining keys as talents.

## State Tracking

### HotTracker

```go
type HotInstance struct {
    SpellID     int
    TargetID    int
    AppliedAt   int
    LastRefresh int
    Tags        map[string]bool
}

type HotTracker struct {
    hots map[[2]int]*HotInstance  // key: {targetID, spellID}
}
```

Methods: `Process(Event)`, `Get(targetID, spellID) *HotInstance`, `GetAll(targetID) []*HotInstance`, `GetAllBySpell(spellID) []*HotInstance`.

### BuffTracker

```go
type BuffTracker struct {
    active map[int]int  // buffID -> timestamp
}
```

Methods: `Process(Event)`, `IsActive(buffID) bool`, `GetAppliedAt(buffID) *int`.

## Attribution

### Interface

```go
type TalentAttributor interface {
    Name() string
    TalentNodeID() *int
    TalentID() *int
    SetCombatantInfo(info *CombatantInfoEvent)
    SetPlayerPetIDs(ids map[int]bool)
    IsSelected() bool
    GetTalentRank() *int
    HasTalent(nodeID int) bool
    ProcessEvent(event Event, hot *HotTracker, buff *BuffTracker)
    ProcessHeal(event *HealEvent, hot *HotTracker, buff *BuffTracker) float64
    Finalize() float64
    GetTotalAttributed() float64
    AddTotalAttributed(amount float64)
}
```

### BaseAttributor

Embedded by all concrete talents. Provides default implementations for `IsSelected()`, `GetTalentRank()`, `HasTalent()`, `IsPlayerPet()`, and no-op defaults for `ProcessEvent`/`ProcessHeal`/`Finalize`.

### Implementations

27 attributors, one file per talent (or small logical group):

| File                          | Attributor(s)                           |
| ----------------------------- | --------------------------------------- |
| `soul_of_the_forest.go`       | SoulOfTheForest                         |
| `abundance.go`                | Abundance                               |
| `convoke.go`                  | Convoke                                 |
| `tree_of_life.go`             | TreeOfLife                              |
| `photosynthesis.go`           | Photosynthesis                          |
| `reforestation.go`            | Reforestation                           |
| `harmonious_blooming.go`      | HarmoniousBlooming                      |
| `improved_wild_growth.go`     | ImprovedWildGrowth                      |
| `nurturing_dormancy.go`       | NurturingDormancy                       |
| `protective_growth.go`        | ProtectiveGrowth                        |
| `blooming_frenzy.go`          | BloomingFrenzy                          |
| `direct_spells.go`            | GroveGuardians, Flourish, Verdancy, etc |
| `buff_multipliers.go`         | CenarionWard, NaturesSwiftness, etc     |
| `keeper_direct_spells.go`     | Keeper hero talent direct spells        |
| `keeper_buff_multipliers.go`  | Keeper hero talent buff multipliers     |
| `keeper_sm_cooldown.go`       | SM cooldown reduction                   |
| `keeper_sylvan_beckoning.go`  | Sylvan Beckoning                        |
| `wildstalker.go`              | Wildstalker base                        |
| `wildstalker_direct_spells.go`| Wildstalker direct spells               |
| `wildstalker_buff_mults.go`   | Wildstalker buff multipliers            |
| `wildstalker_symbiotic.go`    | Symbiotic Bloom Mastery                 |

`BuildAttributors(config Config, ...) []TalentAttributor` — registry function.

## Pipeline

```go
type AnalysisResults struct {
    TotalHealing    int
    Wasted          int
    TalentHealing   map[string]float64
    TalentRanks     map[string]int
    FightDurationMs int
    CombatantInfo   *CombatantInfoEvent
}

type Pipeline struct {
    Attributors  []TalentAttributor
    HotTracker   *tracking.HotTracker
    BuffTracker  *tracking.BuffTracker
    PetIDs       map[int]bool
    PlayerPetIDs map[int]bool
}

func (p *Pipeline) Run(rawEvents []map[string]any) *AnalysisResults
```

Same flow as Python: parse → track → attribute → finalize.

## Web Layer

### Router (chi)

```
GET  /api/health
GET  /api/report/{code}
GET  /api/analyze/{code}/{fightID}/{playerName}   ?base_stacks=N
GET  /api/auth/login
GET  /api/auth/callback
GET  /                                             (static files from frontend/dist/)
```

### Middleware

- CORS
- Rate limiting — in-memory per-IP token bucket (`golang.org/x/time/rate`)
- Anonymous analysis limit (2 free, same logic)
- Request logging via `zerolog`

### WCL Client

- `net/http` + `golang.org/x/oauth2/clientcredentials`
- Raw GraphQL POSTs, `encoding/json` for request/response
- Paginated event fetching (same cursor logic as Python)
- In-memory result cache (map + sync.RWMutex)

## CLI (cobra)

```
flourish analyze <report_code> [--fight ID] [--player NAME] [--config-path PATH]
flourish serve [--port 8080]
```

Interactive fight/player selection via stdin when flags omitted.

## Dependencies

| Module                             | Purpose              |
| ---------------------------------- | -------------------- |
| `github.com/go-chi/chi/v5`        | HTTP router          |
| `github.com/rs/zerolog`           | Structured logging   |
| `github.com/spf13/cobra`          | CLI framework        |
| `github.com/stretchr/testify`     | Test assertions      |
| `gopkg.in/yaml.v3`               | Config parsing       |
| `golang.org/x/oauth2`            | WCL OAuth2           |
| `golang.org/x/time/rate`         | Rate limiting        |

## TDD Strategy

### Phase 1: Test Transpilation

Port all 24 Python test files to Go `_test.go` files using `testify/require`. Same helper patterns (`makeHeal`, `makeApply`, etc.). Tests won't compile initially.

### Phase 2: Implementation (red → green)

Build bottom-up:
1. `models/` — events, config, parse_event
2. `tracking/` — HotTracker, BuffTracker
3. `talents/` — attributor interface + BaseAttributor + all 27 implementations
4. `analysis/` — Pipeline
5. `output/` — table rendering
6. `wcl/` — GraphQL client
7. `web/` — chi router, handlers, middleware
8. `cmd/` — CLI

Each layer: tests already exist from Phase 1, write code until `go test ./...` passes.

### Test File Mapping

| Python                                  | Go                                    |
| --------------------------------------- | ------------------------------------- |
| `tests/models/test_events.py`           | `internal/models/events_test.go`      |
| `tests/models/test_config.py`           | `internal/models/config_test.go`      |
| `tests/models/test_combatantinfo.py`    | `internal/models/combatantinfo_test.go` |
| `tests/tracking/test_hot_tracker.py`    | `internal/tracking/hot_tracker_test.go` |
| `tests/tracking/test_buff_tracker.py`   | `internal/tracking/buff_tracker_test.go` |
| `tests/analysis/test_pipeline.py`       | `internal/analysis/pipeline_test.go`  |
| `tests/analysis/talents/test_*.py`      | `internal/talents/*_test.go`          |
| `tests/wcl/test_client.py`             | `internal/wcl/client_test.go`         |
| `tests/wcl/test_cache.py`              | `internal/wcl/cache_test.go`          |
| `tests/web/test_routes.py`             | `internal/web/routes_test.go`         |
| `tests/web/test_cache.py`              | `internal/web/cache_test.go`          |
| `tests/output/test_table.py`           | `internal/output/table_test.go`       |
