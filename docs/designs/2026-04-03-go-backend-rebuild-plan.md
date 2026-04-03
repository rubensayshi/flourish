# Go Backend Rebuild — Implementation Plan

**Goal:** Rewrite the Python backend in Go using TDD — transpile all Python tests to Go first, then implement bottom-up until all tests pass.

**Architecture:** Go module at `go-backend/` with `internal/` packages for models, tracking, talents, analysis, wcl, web, output. Chi router, zerolog, cobra CLI. Same REST API contract as Python.

**Tech Stack:** Go 1.22+, chi/v5, zerolog, cobra, testify, yaml.v3, golang.org/x/oauth2, golang.org/x/time/rate

---

## Phase 1: Project Scaffolding

### Task 1: Initialize Go module

**Files:**
- Create: `go-backend/go.mod`
- Create: `go-backend/cmd/flourish/main.go`

- [ ] **Step 1: Create go module**

```bash
mkdir -p go-backend/cmd/flourish
cd go-backend && go mod init github.com/rdruid-talent-analyzer/go-backend
```

- [ ] **Step 2: Create minimal main.go**

```go
// go-backend/cmd/flourish/main.go
package main

import "fmt"

func main() {
	fmt.Println("flourish")
}
```

- [ ] **Step 3: Verify it compiles**

```bash
cd go-backend && go build ./cmd/flourish
```

Expected: builds successfully.

- [ ] **Step 4: Add dependencies**

```bash
cd go-backend
go get github.com/go-chi/chi/v5
go get github.com/rs/zerolog
go get github.com/spf13/cobra
go get github.com/stretchr/testify
go get gopkg.in/yaml.v3
go get golang.org/x/oauth2
go get golang.org/x/time/rate
```

- [ ] **Step 5: Commit**

```bash
git add go-backend/
git commit -m "chore: initialize Go module with dependencies"
```

---

## Phase 2: Test Transpilation — Models

### Task 2: Transpile event model tests

**Files:**
- Create: `go-backend/internal/models/events_test.go`

- [ ] **Step 1: Write events_test.go**

```go
package models_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/stretchr/testify/require"
)

func TestParseHealEvent(t *testing.T) {
	raw := map[string]any{
		"timestamp":    1000,
		"type":         "heal",
		"sourceID":     1,
		"targetID":     2,
		"abilityGameID": 774,
		"amount":       5000,
		"overheal":     1000,
		"hitType":      1,
	}
	event := models.ParseEvent(raw)
	require.NotNil(t, event)
	he, ok := event.(*models.HealEvent)
	require.True(t, ok)
	require.Equal(t, 5000, he.Amount)
	require.Equal(t, 1000, he.Overheal)
	require.Equal(t, 6000, he.RawHeal())
	require.InDelta(t, 1000.0/6000.0, he.OverhealPct(), 0.0001)
}

func TestParseCastEvent(t *testing.T) {
	raw := map[string]any{
		"timestamp":    1000,
		"type":         "cast",
		"sourceID":     1,
		"targetID":     2,
		"abilityGameID": 18562,
	}
	event := models.ParseEvent(raw)
	require.NotNil(t, event)
	ce, ok := event.(*models.CastEvent)
	require.True(t, ok)
	require.Equal(t, 18562, ce.AbilityID)
}

func TestParseApplyBuffEvent(t *testing.T) {
	raw := map[string]any{
		"timestamp":    1000,
		"type":         "applybuff",
		"sourceID":     1,
		"targetID":     2,
		"abilityGameID": 774,
	}
	event := models.ParseEvent(raw)
	require.NotNil(t, event)
	_, ok := event.(*models.ApplyBuffEvent)
	require.True(t, ok)
}

func TestParseUnknownEventReturnsNil(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000,
		"type":      "totally_unknown",
		"sourceID":  1,
	}
	event := models.ParseEvent(raw)
	require.Nil(t, event)
}

func TestHealEventIsWasted(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000, "type": "heal", "sourceID": 1,
		"targetID": 2, "abilityGameID": 774,
		"amount": 2000, "overheal": 3000, "hitType": 1,
	}
	event := models.ParseEvent(raw)
	he := event.(*models.HealEvent)
	require.True(t, he.IsWasted()) // 3000/5000 = 60% > 50%
}

func TestHealEventNotWasted(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000, "type": "heal", "sourceID": 1,
		"targetID": 2, "abilityGameID": 774,
		"amount": 4000, "overheal": 1000, "hitType": 1,
	}
	event := models.ParseEvent(raw)
	he := event.(*models.HealEvent)
	require.False(t, he.IsWasted()) // 1000/5000 = 20% < 50%
}

func TestHealEventAbsorbIncludedInRawHeal(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000, "type": "heal", "sourceID": 1,
		"targetID": 2, "abilityGameID": 774,
		"amount": 5000, "overheal": 1000, "absorb": 500, "hitType": 1,
	}
	event := models.ParseEvent(raw)
	he := event.(*models.HealEvent)
	require.Equal(t, 500, he.Absorb)
	require.Equal(t, 6500, he.RawHeal()) // 5000 + 1000 + 500
}

func TestHealEventAbsorbDefaultsToZero(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000, "type": "heal", "sourceID": 1,
		"targetID": 2, "abilityGameID": 774,
		"amount": 5000, "overheal": 1000, "hitType": 1,
	}
	event := models.ParseEvent(raw)
	he := event.(*models.HealEvent)
	require.Equal(t, 0, he.Absorb)
	require.Equal(t, 6000, he.RawHeal())
}

func TestHealEventTickParsed(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000, "type": "heal", "sourceID": 1,
		"targetID": 2, "abilityGameID": 774,
		"amount": 5000, "overheal": 0, "hitType": 1,
		"tick": true,
	}
	event := models.ParseEvent(raw)
	he := event.(*models.HealEvent)
	require.True(t, he.Tick)
}

func TestHealEventTickDefaultsToFalse(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000, "type": "heal", "sourceID": 1,
		"targetID": 2, "abilityGameID": 774,
		"amount": 5000, "overheal": 0, "hitType": 1,
	}
	event := models.ParseEvent(raw)
	he := event.(*models.HealEvent)
	require.False(t, he.Tick)
}
```

- [ ] **Step 2: Verify tests fail to compile**

```bash
cd go-backend && go test ./internal/models/ 2>&1 | head -5
```

Expected: compilation errors (package doesn't exist yet).

- [ ] **Step 3: Commit**

```bash
git add go-backend/internal/models/events_test.go
git commit -m "test: transpile event model tests to Go"
```

### Task 3: Transpile combatant info tests

**Files:**
- Create: `go-backend/internal/models/combatantinfo_test.go`

- [ ] **Step 1: Write combatantinfo_test.go**

```go
package models_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/stretchr/testify/require"
)

func TestParseCombatantInfo(t *testing.T) {
	raw := map[string]any{
		"timestamp": 1000,
		"type":      "combatantinfo",
		"sourceID":  3,
		"talentTree": []any{
			map[string]any{"id": 103098, "rank": 1, "nodeID": 82047},
			map[string]any{"id": 103100, "rank": 1, "nodeID": 82049},
		},
		"critSpell":  256,
		"hasteSpell": 564,
		"mastery":    893,
		"specID":     105,
	}
	event := models.ParseEvent(raw)
	require.NotNil(t, event)
	ci, ok := event.(*models.CombatantInfoEvent)
	require.True(t, ok)
	require.True(t, ci.TalentNodes[82047])
	require.True(t, ci.TalentNodes[82049])
	require.Equal(t, 256.0, ci.CritSpell)
	require.Equal(t, 105, ci.SpecID)
}
```

- [ ] **Step 2: Commit**

```bash
git add go-backend/internal/models/combatantinfo_test.go
git commit -m "test: transpile combatant info tests to Go"
```

### Task 4: Transpile config tests

**Files:**
- Create: `go-backend/internal/models/config_test.go`

- [ ] **Step 1: Write config_test.go**

```go
package models_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig(t *testing.T) {
	yamlContent := `
mastery:
  base_stacks: 4

soul_of_the_forest:
  skip: false
  multiplier: 0.6
wild_growth:
  skip: true
  skip_reason: "always take"
`
	dir := t.TempDir()
	p := filepath.Join(dir, "talents.yaml")
	err := os.WriteFile(p, []byte(yamlContent), 0644)
	require.NoError(t, err)

	config, err := models.LoadConfig(p)
	require.NoError(t, err)
	require.Equal(t, 4, config.Mastery.BaseStacks)
	require.False(t, config.Talents["soul_of_the_forest"].Skip)
	require.NotNil(t, config.Talents["soul_of_the_forest"].Multiplier)
	require.Equal(t, 0.6, *config.Talents["soul_of_the_forest"].Multiplier)
	require.True(t, config.Talents["wild_growth"].Skip)
}

func TestMissingTalentUsesDefaults(t *testing.T) {
	tc := models.TalentConfig{}
	require.False(t, tc.Skip)
	require.Nil(t, tc.Multiplier)
}
```

- [ ] **Step 2: Commit**

```bash
git add go-backend/internal/models/config_test.go
git commit -m "test: transpile config tests to Go"
```

---

## Phase 3: Test Transpilation — Tracking

### Task 5: Transpile HotTracker tests

**Files:**
- Create: `go-backend/internal/tracking/hot_tracker_test.go`

- [ ] **Step 1: Write hot_tracker_test.go**

```go
package tracking_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
	"github.com/stretchr/testify/require"
)

const (
	rejuvID = 774
	targetA = 10
)

func makeApplyBuff(ts, target, ability int) *models.ApplyBuffEvent {
	return &models.ApplyBuffEvent{
		BaseEvent: models.BaseEvent{Timestamp: ts, SourceID: 1, Type: "applybuff"},
		TargetID:  target, AbilityID: ability,
	}
}

func makeRemoveBuff(ts, target, ability int) *models.RemoveBuffEvent {
	return &models.RemoveBuffEvent{
		BaseEvent: models.BaseEvent{Timestamp: ts, SourceID: 1, Type: "removebuff"},
		TargetID:  target, AbilityID: ability,
	}
}

func makeRefreshBuff(ts, target, ability int) *models.RefreshBuffEvent {
	return &models.RefreshBuffEvent{
		BaseEvent: models.BaseEvent{Timestamp: ts, SourceID: 1, Type: "refreshbuff"},
		TargetID:  target, AbilityID: ability,
	}
}

func TestApplyCreatesHot(t *testing.T) {
	tracker := tracking.NewHotTracker()
	tracker.Process(makeApplyBuff(100, targetA, rejuvID))
	hot := tracker.Get(targetA, rejuvID)
	require.NotNil(t, hot)
	require.Equal(t, rejuvID, hot.SpellID)
	require.Equal(t, 100, hot.AppliedAt)
}

func TestRemoveClearsHot(t *testing.T) {
	tracker := tracking.NewHotTracker()
	tracker.Process(makeApplyBuff(100, targetA, rejuvID))
	tracker.Process(makeRemoveBuff(200, targetA, rejuvID))
	require.Nil(t, tracker.Get(targetA, rejuvID))
}

func TestRefreshUpdatesTimestamp(t *testing.T) {
	tracker := tracking.NewHotTracker()
	tracker.Process(makeApplyBuff(100, targetA, rejuvID))
	tracker.Process(makeRefreshBuff(200, targetA, rejuvID))
	hot := tracker.Get(targetA, rejuvID)
	require.Equal(t, 100, hot.AppliedAt) // original apply time preserved
	require.Equal(t, 200, hot.LastRefresh)
}

func TestTagsClearedOnRefresh(t *testing.T) {
	tracker := tracking.NewHotTracker()
	tracker.Process(makeApplyBuff(100, targetA, rejuvID))
	hot := tracker.Get(targetA, rejuvID)
	hot.Tags["sotf"] = true
	tracker.Process(makeRefreshBuff(200, targetA, rejuvID))
	require.False(t, tracker.Get(targetA, rejuvID).Tags["sotf"])
}

func TestGetAllHotsOnTarget(t *testing.T) {
	tracker := tracking.NewHotTracker()
	tracker.Process(makeApplyBuff(100, targetA, 774))
	tracker.Process(makeApplyBuff(100, targetA, 33763)) // Lifebloom
	hots := tracker.GetAll(targetA)
	require.Len(t, hots, 2)
}
```

- [ ] **Step 2: Commit**

```bash
git add go-backend/internal/tracking/hot_tracker_test.go
git commit -m "test: transpile HotTracker tests to Go"
```

### Task 6: Transpile BuffTracker tests

**Files:**
- Create: `go-backend/internal/tracking/buff_tracker_test.go`

- [ ] **Step 1: Write buff_tracker_test.go**

```go
package tracking_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
	"github.com/stretchr/testify/require"
)

const sotfBuff = 114108

func TestBuffApplied(t *testing.T) {
	tracker := tracking.NewBuffTracker()
	tracker.Process(&models.ApplyBuffEvent{
		BaseEvent: models.BaseEvent{Timestamp: 100, SourceID: 1, Type: "applybuff"},
		TargetID: 1, AbilityID: sotfBuff,
	})
	require.True(t, tracker.IsActive(sotfBuff))
}

func TestBuffRemoved(t *testing.T) {
	tracker := tracking.NewBuffTracker()
	tracker.Process(&models.ApplyBuffEvent{
		BaseEvent: models.BaseEvent{Timestamp: 100, SourceID: 1, Type: "applybuff"},
		TargetID: 1, AbilityID: sotfBuff,
	})
	tracker.Process(&models.RemoveBuffEvent{
		BaseEvent: models.BaseEvent{Timestamp: 200, SourceID: 1, Type: "removebuff"},
		TargetID: 1, AbilityID: sotfBuff,
	})
	require.False(t, tracker.IsActive(sotfBuff))
}

func TestBuffNotActiveByDefault(t *testing.T) {
	tracker := tracking.NewBuffTracker()
	require.False(t, tracker.IsActive(sotfBuff))
}
```

- [ ] **Step 2: Commit**

```bash
git add go-backend/internal/tracking/buff_tracker_test.go
git commit -m "test: transpile BuffTracker tests to Go"
```

---

## Phase 4: Test Transpilation — Pipeline & Attributor

### Task 7: Transpile pipeline tests

**Files:**
- Create: `go-backend/internal/analysis/pipeline_test.go`

- [ ] **Step 1: Write pipeline_test.go**

```go
package analysis_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)

// FakeAttributor claims 50% of Rejuv (774) healing.
type FakeAttributor struct {
	talents.BaseAttributor
}

func NewFakeAttributor() *FakeAttributor {
	return &FakeAttributor{
		BaseAttributor: talents.NewBaseAttributor("Fake Talent", nil, nil),
	}
}

func (f *FakeAttributor) ProcessHeal(event *models.HealEvent, hot *tracking.HotTracker, buff *tracking.BuffTracker) float64 {
	if event.AbilityID == 774 {
		return float64(event.Amount) * 0.5
	}
	return 0.0
}

func makeHealRaw(ts, ability, amount, overheal int) map[string]any {
	return map[string]any{
		"timestamp":    ts,
		"type":         "heal",
		"sourceID":     1,
		"targetID":     2,
		"abilityGameID": ability,
		"amount":       amount,
		"overheal":     overheal,
		"hitType":      1,
	}
}

func TestPipelineAttributesHealing(t *testing.T) {
	rawEvents := []map[string]any{makeHealRaw(100, 774, 10000, 0)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{NewFakeAttributor()}, nil, nil)
	results := pipeline.Run(rawEvents)
	require.Equal(t, 5000.0, results.TalentHealing["Fake Talent"])
}

func TestPipelineSkipsWastedHeals(t *testing.T) {
	rawEvents := []map[string]any{makeHealRaw(100, 774, 2000, 3000)} // 60% OH
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{NewFakeAttributor()}, nil, nil)
	results := pipeline.Run(rawEvents)
	require.Equal(t, 0.0, results.TalentHealing["Fake Talent"])
	require.Greater(t, results.Wasted, 0)
}

func TestPipelineTracksTotalHealing(t *testing.T) {
	rawEvents := []map[string]any{
		makeHealRaw(100, 774, 10000, 0),
		makeHealRaw(200, 48438, 5000, 0),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{NewFakeAttributor()}, nil, nil)
	results := pipeline.Run(rawEvents)
	require.Equal(t, 15000, results.TotalHealing)
	require.Equal(t, 5000.0, results.TalentHealing["Fake Talent"])
}
```

- [ ] **Step 2: Commit**

```bash
git add go-backend/internal/analysis/pipeline_test.go
git commit -m "test: transpile pipeline tests to Go"
```

### Task 8: Transpile combatant info pipeline tests

**Files:**
- Create: `go-backend/internal/analysis/combatantinfo_pipeline_test.go`

- [ ] **Step 1: Write combatantinfo_pipeline_test.go**

```go
package analysis_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)

// TestAttributor tracks whether it received combatant info.
type TestAttributor struct {
	talents.BaseAttributor
	SawInfo bool
}

func NewTestAttributor() *TestAttributor {
	return &TestAttributor{
		BaseAttributor: talents.NewBaseAttributor("Test", nil, nil),
	}
}

func (a *TestAttributor) SetCombatantInfo(info *models.CombatantInfoEvent) {
	a.BaseAttributor.SetCombatantInfo(info)
	a.SawInfo = true
}

func makeCombatantInfoRaw() map[string]any {
	return map[string]any{
		"timestamp":  1000,
		"type":       "combatantinfo",
		"sourceID":   3,
		"talentTree": []any{map[string]any{"id": 1, "rank": 1, "nodeID": 82047}},
		"critSpell":  256,
		"hasteSpell": 564,
		"mastery":    893,
		"specID":     105,
	}
}

func TestCombatantInfoPassedToAttributors(t *testing.T) {
	attr := NewTestAttributor()
	rawEvents := []map[string]any{
		makeCombatantInfoRaw(),
		makeHealRaw(2000, 774, 1000, 0),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{attr}, nil, nil)
	pipeline.Run(rawEvents)
	require.True(t, attr.SawInfo)
	require.True(t, attr.HasTalent(82047))
	require.False(t, attr.HasTalent(99999))
}

func TestTalentFilteringRemovesUnselected(t *testing.T) {
	nodeID82047 := 82047
	nodeID99999 := 99999

	selected := &talents.BaseAttributor{}
	*selected = talents.NewBaseAttributor("Selected", &nodeID82047, nil)
	unselected := &talents.BaseAttributor{}
	*unselected = talents.NewBaseAttributor("Unselected", &nodeID99999, nil)
	noNode := &talents.BaseAttributor{}
	*noNode = talents.NewBaseAttributor("NoNode", nil, nil)

	rawEvents := []map[string]any{
		{
			"timestamp":  1000,
			"type":       "combatantinfo",
			"sourceID":   3,
			"talentTree": []any{map[string]any{"id": 103098, "rank": 1, "nodeID": 82047}},
			"critSpell":  256, "hasteSpell": 564, "mastery": 893, "specID": 105,
		},
	}
	pipeline := analysis.NewPipeline(
		[]talents.TalentAttributor{selected, unselected, noNode}, nil, nil,
	)
	results := pipeline.Run(rawEvents)
	_, hasSelected := results.TalentHealing["Selected"]
	_, hasUnselected := results.TalentHealing["Unselected"]
	_, hasNoNode := results.TalentHealing["NoNode"]
	require.True(t, hasSelected)
	require.False(t, hasUnselected)
	require.True(t, hasNoNode)
}

func TestChoiceNodeFiltering(t *testing.T) {
	nodeID := 82064
	talentIDa := 108125
	talentIDb := 108124

	choiceA := &talents.BaseAttributor{}
	*choiceA = talents.NewBaseAttributor("Choice A", &nodeID, &talentIDa)
	choiceB := &talents.BaseAttributor{}
	*choiceB = talents.NewBaseAttributor("Choice B", &nodeID, &talentIDb)

	rawEvents := []map[string]any{
		{
			"timestamp":  1000,
			"type":       "combatantinfo",
			"sourceID":   3,
			"talentTree": []any{map[string]any{"id": 108125, "rank": 1, "nodeID": 82064}},
			"critSpell":  256, "hasteSpell": 564, "mastery": 893, "specID": 105,
		},
	}
	pipeline := analysis.NewPipeline(
		[]talents.TalentAttributor{choiceA, choiceB}, nil, nil,
	)
	results := pipeline.Run(rawEvents)
	_, hasA := results.TalentHealing["Choice A"]
	_, hasB := results.TalentHealing["Choice B"]
	require.True(t, hasA)
	require.False(t, hasB)
}
```

- [ ] **Step 2: Commit**

```bash
git add go-backend/internal/analysis/combatantinfo_pipeline_test.go
git commit -m "test: transpile combatant info pipeline tests to Go"
```

---

## Phase 5: Test Transpilation — Talents

### Task 9: Transpile talent test helpers

Create a shared test helpers file for talent tests. Many talent tests use the same helper functions.

**Files:**
- Create: `go-backend/internal/talents/test_helpers_test.go`

- [ ] **Step 1: Write test_helpers_test.go**

```go
package talents_test

// Shared test helpers for talent tests.
// These construct raw event dicts matching WCL format.

func makeHeal(ts, ability, amount int, opts ...func(map[string]any)) map[string]any {
	m := map[string]any{
		"timestamp":    ts,
		"type":         "heal",
		"sourceID":     1,
		"targetID":     2,
		"abilityGameID": ability,
		"amount":       amount,
		"overheal":     0,
		"hitType":      1,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func withTarget(target int) func(map[string]any) {
	return func(m map[string]any) { m["targetID"] = target }
}

func withOverheal(oh int) func(map[string]any) {
	return func(m map[string]any) { m["overheal"] = oh }
}

func withHitType(ht int) func(map[string]any) {
	return func(m map[string]any) { m["hitType"] = ht }
}

func withTick() func(map[string]any) {
	return func(m map[string]any) { m["tick"] = true }
}

func withSource(src int) func(map[string]any) {
	return func(m map[string]any) { m["sourceID"] = src }
}

func makeCast(ts, ability int, opts ...func(map[string]any)) map[string]any {
	m := map[string]any{
		"timestamp":    ts,
		"type":         "cast",
		"sourceID":     1,
		"targetID":     2,
		"abilityGameID": ability,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func makeBegincast(ts, ability int, opts ...func(map[string]any)) map[string]any {
	m := map[string]any{
		"timestamp":    ts,
		"type":         "begincast",
		"sourceID":     1,
		"targetID":     2,
		"abilityGameID": ability,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func makeApply(ts, ability int, opts ...func(map[string]any)) map[string]any {
	m := map[string]any{
		"timestamp":    ts,
		"type":         "applybuff",
		"sourceID":     1,
		"targetID":     2,
		"abilityGameID": ability,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func makeRefresh(ts, ability int, opts ...func(map[string]any)) map[string]any {
	m := map[string]any{
		"timestamp":    ts,
		"type":         "refreshbuff",
		"sourceID":     1,
		"targetID":     2,
		"abilityGameID": ability,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func makeRemove(ts, ability int, opts ...func(map[string]any)) map[string]any {
	m := map[string]any{
		"timestamp":    ts,
		"type":         "removebuff",
		"sourceID":     1,
		"targetID":     2,
		"abilityGameID": ability,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func makeCombatantInfo(ts int, opts ...func(map[string]any)) map[string]any {
	m := map[string]any{
		"timestamp":  ts,
		"type":       "combatantinfo",
		"sourceID":   1,
		"talentTree": []any{},
		"critSpell":  0,
		"hasteSpell": 0,
		"mastery":    0,
		"specID":     105,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func withCritSpell(v float64) func(map[string]any) {
	return func(m map[string]any) { m["critSpell"] = v }
}

func withTalentTree(nodes []map[string]any) func(map[string]any) {
	return func(m map[string]any) {
		tree := make([]any, len(nodes))
		for i, n := range nodes {
			tree[i] = n
		}
		m["talentTree"] = tree
	}
}
```

- [ ] **Step 2: Commit**

```bash
git add go-backend/internal/talents/test_helpers_test.go
git commit -m "test: add shared talent test helpers"
```

### Task 10: Transpile direct spell talent tests

**Files:**
- Create: `go-backend/internal/talents/direct_spells_test.go`

- [ ] **Step 1: Write direct_spells_test.go**

```go
package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)

func TestEverbloomAttributesAllHealing(t *testing.T) {
	events := []map[string]any{makeHeal(100, 1244341, 5000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewEverbloomSplashAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 5000.0, results.TalentHealing["Everbloom: Splash"])
}

func TestGroveGuardiansAttributesNourish(t *testing.T) {
	events := []map[string]any{makeHeal(100, 422090, 3000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewGroveGuardiansAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 3000.0, results.TalentHealing["Grove Guardians"])
}

func TestDreamSurgeAttributesDreamBloom(t *testing.T) {
	events := []map[string]any{makeHeal(100, 434141, 2000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewDreamSurgeAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 2000.0, results.TalentHealing["Dream Surge"])
}

func TestDirectSpellIgnoresUnrelatedSpells(t *testing.T) {
	events := []map[string]any{makeHeal(100, 774, 10000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewEverbloomSplashAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Everbloom: Splash"])
}

func TestDirectSpellSkipsWastedHeals(t *testing.T) {
	events := []map[string]any{makeHeal(100, 1244341, 2000, withOverheal(3000))}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewEverbloomSplashAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Everbloom: Splash"])
}

func TestMultipleDirectAttributors(t *testing.T) {
	events := []map[string]any{
		makeHeal(100, 1244341, 5000),
		makeHeal(200, 422090, 3000),
		makeHeal(300, 434141, 2000),
		makeHeal(400, 81269, 1000),
		makeHeal(500, 774, 8000),
	}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{
		talents.NewEverbloomSplashAttributor(),
		talents.NewGroveGuardiansAttributor(),
		talents.NewDreamSurgeAttributor(),
		talents.NewEfflorescenceAttributor(),
	}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 5000.0, results.TalentHealing["Everbloom: Splash"])
	require.Equal(t, 3000.0, results.TalentHealing["Grove Guardians"])
	require.Equal(t, 2000.0, results.TalentHealing["Dream Surge"])
	require.Equal(t, 1000.0, results.TalentHealing["Efflorescence"])
	require.Equal(t, 19000, results.TotalHealing)
}

func TestRampantGrowthAttributesBonusOnRegrowthTicks(t *testing.T) {
	events := []map[string]any{makeHeal(100, 8936, 10000, withTick())}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewRampantGrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 5000.0, results.TalentHealing["Rampant Growth"], 1.0)
}

func TestRampantGrowthIgnoresDirectHeals(t *testing.T) {
	events := []map[string]any{makeHeal(100, 8936, 10000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewRampantGrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Rampant Growth"])
}

func TestRampantGrowthIgnoresOtherSpells(t *testing.T) {
	events := []map[string]any{makeHeal(100, 774, 10000, withTick())}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewRampantGrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Rampant Growth"])
}

func TestRampantGrowthSkipsWastedHeals(t *testing.T) {
	events := []map[string]any{makeHeal(100, 8936, 2000, withOverheal(3000), withTick())}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewRampantGrowthAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Rampant Growth"])
}
```

- [ ] **Step 2: Commit**

```bash
git add go-backend/internal/talents/direct_spells_test.go
git commit -m "test: transpile direct spell talent tests to Go"
```

### Task 11: Transpile buff multiplier talent tests

**Files:**
- Create: `go-backend/internal/talents/buff_multipliers_test.go`

- [ ] **Step 1: Write buff_multipliers_test.go**

```go
package talents_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/stretchr/testify/require"
)

func TestWildSynthesisAttributesBonusPortion(t *testing.T) {
	events := []map[string]any{makeHeal(100, 422090, 13000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewWildSynthesisAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 3000.0, results.TalentHealing["Wild Synthesis"], 50.0)
}

func TestWildstalkersPowerOnRejuv(t *testing.T) {
	events := []map[string]any{makeHeal(100, 774, 11000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewWildstalkersPowerAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 1000.0, results.TalentHealing["Wildstalker's Power"], 50.0)
}

func TestStaticBuffIgnoresUnrelatedSpells(t *testing.T) {
	events := []map[string]any{makeHeal(100, 999, 10000)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewWildSynthesisAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Wild Synthesis"])
}

func TestLifetreadingOnEfflorescence(t *testing.T) {
	events := []map[string]any{makeHeal(100, 81269, 12500)}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewLifetreadingAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.InDelta(t, 2500.0, results.TalentHealing["Lifetreading"], 50.0)
}

func TestStaticBuffSkipsWasted(t *testing.T) {
	events := []map[string]any{makeHeal(100, 422090, 2000, withOverheal(3000))}
	pipeline := analysis.NewPipeline([]talents.TalentAttributor{talents.NewWildSynthesisAttributor()}, nil, nil)
	results := pipeline.Run(events)
	require.Equal(t, 0.0, results.TalentHealing["Wild Synthesis"])
}
```

- [ ] **Step 2: Commit**

```bash
git add go-backend/internal/talents/buff_multipliers_test.go
git commit -m "test: transpile buff multiplier talent tests to Go"
```

### Task 12: Transpile remaining talent tests

The remaining talent tests follow the same pattern. Each becomes a `*_test.go` file in the `talents` package. Due to the volume (~800 lines across SotF, Convoke, ToL, IWG, Reforestation, Misc, Wildstalker, SM cooldown), transpile each test file.

**Files to create (one per Python test file):**
- `go-backend/internal/talents/soul_of_the_forest_test.go` — from `test_soul_of_the_forest.py`
- `go-backend/internal/talents/convoke_test.go` — from `test_convoke.py`
- `go-backend/internal/talents/tree_of_life_test.go` — from `test_tree_of_life.py`
- `go-backend/internal/talents/improved_wild_growth_test.go` — from `test_improved_wild_growth.py`
- `go-backend/internal/talents/reforestation_test.go` — from `test_reforestation.py`
- `go-backend/internal/talents/misc_test.go` — from `test_misc.py` (Abundance, Photosynthesis, Nurturing Dormancy)
- `go-backend/internal/talents/wildstalker_test.go` — from `test_wildstalker.py`
- `go-backend/internal/talents/sm_cooldown_reduction_test.go` — from `test_sm_cooldown_reduction.py`
- `go-backend/internal/talents/protective_growth_test.go` — from `test_protective_growth.py`

For each file: follow the same pattern as Tasks 10-11 — use the shared test helpers, construct raw event slices, run through `analysis.NewPipeline`, assert on `results.TalentHealing[name]` using `require.InDelta` for float comparisons.

Each test file follows this template:
1. Use `makeHeal`, `makeCast`, `makeApply`, `makeRemove`, `makeRefresh` with option functions
2. Construct `[]map[string]any` event slices matching the Python test
3. Create pipeline with `analysis.NewPipeline([]talents.TalentAttributor{...}, nil, nil)`
4. Assert `results.TalentHealing["TalentName"]`

**The exact test logic for each file is a 1:1 translation of the Python tests already read above.** Each Python `def test_*` becomes a Go `func Test*(t *testing.T)`. `pytest.approx(x)` becomes `require.InDelta(t, x, result, 1.0)`. `assert x == 0.0` becomes `require.Equal(t, 0.0, result)`.

Commit each file individually:

```bash
git add go-backend/internal/talents/<file>_test.go
git commit -m "test: transpile <talent> tests to Go"
```

---

## Phase 6: Test Transpilation — WCL, Web, Output

### Task 13: Transpile WCL client tests

**Files:**
- Create: `go-backend/internal/wcl/client_test.go`

- [ ] **Step 1: Write client_test.go**

The Python tests mock `httpx`. In Go, use `net/http/httptest` to create a fake server.

```go
package wcl_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/wcl"
	"github.com/stretchr/testify/require"
)

func TestAuthenticateAndQuery(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// OAuth token request
			json.NewEncoder(w).Encode(map[string]any{"access_token": "fake_token"})
		} else {
			// GraphQL query
			json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"reportData": map[string]any{
						"report": map[string]any{"title": "Test"},
					},
				},
			})
		}
	}))
	defer server.Close()

	client := wcl.NewClient("test_id", "test_secret", wcl.WithBaseURL(server.URL))
	result, err := client.GetReport("abc123")
	require.NoError(t, err)
	require.Equal(t, "Test", result["title"])
	require.Equal(t, 2, callCount)
}

func TestGetEventsPaginates(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		var resp map[string]any
		if callCount == 1 {
			resp = map[string]any{
				"data": map[string]any{
					"reportData": map[string]any{
						"report": map[string]any{
							"events": map[string]any{
								"data":              []any{map[string]any{"type": "heal", "timestamp": 1}},
								"nextPageTimestamp": 5000,
							},
						},
					},
				},
			}
		} else {
			resp = map[string]any{
				"data": map[string]any{
					"reportData": map[string]any{
						"report": map[string]any{
							"events": map[string]any{
								"data":              []any{map[string]any{"type": "heal", "timestamp": 5001}},
								"nextPageTimestamp": nil,
							},
						},
					},
				},
			}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := wcl.NewClient("test_id", "test_secret",
		wcl.WithBaseURL(server.URL),
		wcl.WithToken("fake_token"),
	)
	events, err := client.GetEvents("abc", 1, 1, 0, 10000)
	require.NoError(t, err)
	require.Len(t, events, 2)
	require.Equal(t, 2, callCount)
}
```

- [ ] **Step 2: Commit**

```bash
git add go-backend/internal/wcl/client_test.go
git commit -m "test: transpile WCL client tests to Go"
```

### Task 14: Transpile WCL cache tests

**Files:**
- Create: `go-backend/internal/wcl/cache_test.go`

- [ ] **Step 1: Write cache_test.go**

```go
package wcl_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/wcl"
	"github.com/stretchr/testify/require"
)

// MockWCLClient implements the WCLQuerier interface for testing.
type MockWCLClient struct {
	GetReportResult map[string]any
	GetEventsResult []map[string]any
	ReportCalled    bool
	EventsCalled    bool
}

func (m *MockWCLClient) GetReport(code string) (map[string]any, error) {
	m.ReportCalled = true
	return m.GetReportResult, nil
}

func (m *MockWCLClient) GetEvents(code string, fightID, sourceID int, startTime, endTime float64) ([]map[string]any, error) {
	m.EventsCalled = true
	return m.GetEventsResult, nil
}

func (m *MockWCLClient) GetDamageTaken(code string, fightID, sourceID int, startTime, endTime float64, filter string) (int, error) {
	return 0, nil
}

func TestGetReportCachesToDisk(t *testing.T) {
	inner := &MockWCLClient{
		GetReportResult: map[string]any{"title": "Test Report", "fights": []any{}},
	}
	client := wcl.NewCachedClient(inner, t.TempDir())

	result, err := client.GetReport("ABC123")
	require.NoError(t, err)
	require.Equal(t, "Test Report", result["title"])
}

func TestGetReportReadsFromCache(t *testing.T) {
	inner := &MockWCLClient{
		GetReportResult: map[string]any{"title": "Test Report"},
	}
	dir := t.TempDir()
	client := wcl.NewCachedClient(inner, dir)

	// First call caches
	_, err := client.GetReport("ABC123")
	require.NoError(t, err)

	// Second call reads from cache
	inner2 := &MockWCLClient{}
	client2 := wcl.NewCachedClient(inner2, dir)
	result, err := client2.GetReport("ABC123")
	require.NoError(t, err)
	require.Equal(t, "Test Report", result["title"])
	require.False(t, inner2.ReportCalled)
}

func TestGetEventsCachesToDisk(t *testing.T) {
	inner := &MockWCLClient{
		GetEventsResult: []map[string]any{{"type": "heal", "timestamp": 1}},
	}
	client := wcl.NewCachedClient(inner, t.TempDir())

	result, err := client.GetEvents("ABC123", 1, 5, 0, 10000)
	require.NoError(t, err)
	require.Len(t, result, 1)
}

func TestGetEventsReadsFromCache(t *testing.T) {
	inner := &MockWCLClient{
		GetEventsResult: []map[string]any{{"type": "heal", "timestamp": 1}},
	}
	dir := t.TempDir()
	client := wcl.NewCachedClient(inner, dir)

	_, err := client.GetEvents("ABC123", 1, 5, 0, 10000)
	require.NoError(t, err)

	inner2 := &MockWCLClient{}
	client2 := wcl.NewCachedClient(inner2, dir)
	result, err := client2.GetEvents("ABC123", 1, 5, 0, 10000)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.False(t, inner2.EventsCalled)
}
```

- [ ] **Step 2: Commit**

```bash
git add go-backend/internal/wcl/cache_test.go
git commit -m "test: transpile WCL cache tests to Go"
```

### Task 15: Transpile web cache tests

**Files:**
- Create: `go-backend/internal/web/cache_test.go`

- [ ] **Step 1: Write cache_test.go**

```go
package web_test

import (
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/web"
	"github.com/stretchr/testify/require"
)

func TestGetReturnsNilWhenMissing(t *testing.T) {
	cache := web.NewResultCache(t.TempDir())
	result := cache.Get("ABC123", 1, "Player")
	require.Nil(t, result)
}

func TestSetThenGetRoundtrips(t *testing.T) {
	cache := web.NewResultCache(t.TempDir())
	data := map[string]any{"total_healing": 1000.0, "talents": []any{map[string]any{"name": "SotF", "attributed": 500.0}}}
	cache.Set("ABC123", 1, "Player", data)
	result := cache.Get("ABC123", 1, "Player")
	require.NotNil(t, result)
	require.Equal(t, 1000.0, result["total_healing"])
}

func TestCacheKeyIsCaseInsensitiveForPlayer(t *testing.T) {
	cache := web.NewResultCache(t.TempDir())
	data := map[string]any{"total_healing": 1000.0}
	cache.Set("ABC123", 1, "Saikó", data)
	result := cache.Get("ABC123", 1, "saikó")
	require.NotNil(t, result)
	require.Equal(t, 1000.0, result["total_healing"])
}
```

- [ ] **Step 2: Commit**

```bash
git add go-backend/internal/web/cache_test.go
git commit -m "test: transpile web cache tests to Go"
```

### Task 16: Transpile web routes tests

**Files:**
- Create: `go-backend/internal/web/routes_test.go`

- [ ] **Step 1: Write routes_test.go**

```go
package web_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/web"
	"github.com/stretchr/testify/require"
)

type MockReportClient struct{}

func (m *MockReportClient) GetReport(code string) (map[string]any, error) {
	if code == "INVALID" {
		return nil, errors.New("not found")
	}
	return map[string]any{
		"title": "Test Raid",
		"fights": []any{
			map[string]any{"id": 1, "name": "Boss", "kill": true, "startTime": 0,
				"endTime": 60000, "encounterID": 123, "difficulty": 4},
			map[string]any{"id": 2, "name": "Trash", "kill": true, "startTime": 0,
				"endTime": 30000, "encounterID": 0, "difficulty": 0},
		},
		"masterData": map[string]any{
			"actors": []any{
				map[string]any{"id": 10, "name": "Saikó", "type": "Player",
					"subType": "Druid", "server": "Draenor"},
				map[string]any{"id": 11, "name": "Warrior", "type": "Player",
					"subType": "Warrior", "server": "Draenor"},
			},
		},
	}, nil
}

func TestReportEndpointReturnsFightsAndDruids(t *testing.T) {
	router := web.NewRouter(&MockReportClient{}, t.TempDir())
	req := httptest.NewRequest("GET", "/api/report/ABC123", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, 200, w.Code)
	var data map[string]any
	json.NewDecoder(w.Body).Decode(&data)
	require.Equal(t, "Test Raid", data["title"])
	fights := data["fights"].([]any)
	require.Len(t, fights, 1)
	druids := data["druids"].([]any)
	require.Len(t, druids, 1)
}

func TestHealthEndpoint(t *testing.T) {
	router := web.NewRouter(&MockReportClient{}, t.TempDir())
	req := httptest.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, 200, w.Code)
}
```

- [ ] **Step 2: Commit**

```bash
git add go-backend/internal/web/routes_test.go
git commit -m "test: transpile web routes tests to Go"
```

### Task 17: Transpile output table tests

**Files:**
- Create: `go-backend/internal/output/table_test.go`

- [ ] **Step 1: Write table_test.go**

```go
package output_test

import (
	"strings"
	"testing"

	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/output"
	"github.com/stretchr/testify/require"
)

func TestRenderResultsReturnsString(t *testing.T) {
	results := &analysis.AnalysisResults{
		TotalHealing:    100000,
		Wasted:          10000,
		TalentHealing:   map[string]float64{"Soul of the Forest": 15000.0, "Cultivation": 8000.0},
		FightDurationMs: 300000,
	}
	out := output.RenderResults(results, "Mythic Boss", "TestDruid")
	require.Contains(t, out, "Soul of the Forest")
	require.Contains(t, out, "Cultivation")
	require.Contains(t, out, "Wasted")
}

func TestRenderOverlapDisclaimerWhenExceedingTotal(t *testing.T) {
	results := &analysis.AnalysisResults{
		TotalHealing:    100000,
		TalentHealing:   map[string]float64{"Talent A": 80000.0, "Talent B": 50000.0},
		FightDurationMs: 300000,
	}
	out := output.RenderResults(results, "", "")
	require.Contains(t, out, "Talents can overlap")
}

func TestRenderNoDisclaimerWhenWithinTotal(t *testing.T) {
	results := &analysis.AnalysisResults{
		TotalHealing:    100000,
		TalentHealing:   map[string]float64{"Talent A": 30000.0, "Talent B": 20000.0},
		FightDurationMs: 300000,
	}
	out := output.RenderResults(results, "", "")
	require.NotContains(t, out, "Talents can overlap")
}

func TestRenderUnattributedDashWhenNegative(t *testing.T) {
	results := &analysis.AnalysisResults{
		TotalHealing:    100000,
		TalentHealing:   map[string]float64{"Talent A": 80000.0, "Talent B": 50000.0},
		FightDurationMs: 300000,
	}
	out := output.RenderResults(results, "", "")
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		if strings.Contains(line, "Unattributed") {
			require.Contains(t, line, "—")
		}
	}
}
```

- [ ] **Step 2: Commit**

```bash
git add go-backend/internal/output/table_test.go
git commit -m "test: transpile output table tests to Go"
```

---

## Phase 7: Implementation — Models (Red → Green)

### Task 18: Implement event models and ParseEvent

**Files:**
- Create: `go-backend/internal/models/events.go`

- [ ] **Step 1: Implement events.go**

Implement all event structs (`BaseEvent`, `HealEvent`, `CastEvent`, `ApplyBuffEvent`, `RefreshBuffEvent`, `RemoveBuffEvent`, `SummonEvent`, `CombatantInfoEvent`), the `Event` interface, `ParseEvent`, and helper methods (`RawHeal`, `OverhealPct`, `IsWasted`).

Key implementation details:
- `ParseEvent(raw map[string]any) Event` — type-switch on `raw["type"]`, extract fields with safe type assertions using helper `getInt`, `getFloat`, `getBool`, `getString` functions
- `OverhealWasteThreshold = 0.5` constant
- `HealEvent.RawHeal() = Amount + Overheal + Absorb`
- `HealEvent.IsWasted() = OverhealPct() > OverhealWasteThreshold`
- `CombatantInfoEvent` uses `map[int]bool` for TalentNodes and TalentIDs (Go set idiom)
- JSON numbers from `map[string]any` come as `float64` — cast to `int` where needed

- [ ] **Step 2: Run tests**

```bash
cd go-backend && go test ./internal/models/ -v
```

Expected: all events_test.go and combatantinfo_test.go tests PASS.

- [ ] **Step 3: Commit**

```bash
git add go-backend/internal/models/events.go
git commit -m "feat: implement event models and ParseEvent"
```

### Task 19: Implement config loading

**Files:**
- Create: `go-backend/internal/models/config.go`

- [ ] **Step 1: Implement config.go**

Key implementation details:
- Custom `UnmarshalYAML` on `Config` — unmarshal to `map[string]any`, extract `mastery` key first, then iterate remaining keys as `TalentConfig`
- `TalentConfig.Multiplier` is `*float64` (nil = not set)
- `LoadConfig(path string) (*Config, error)`

- [ ] **Step 2: Run tests**

```bash
cd go-backend && go test ./internal/models/ -v -run Config
```

Expected: config_test.go tests PASS.

- [ ] **Step 3: Commit**

```bash
git add go-backend/internal/models/config.go
git commit -m "feat: implement config loading"
```

---

## Phase 8: Implementation — Tracking

### Task 20: Implement HotTracker

**Files:**
- Create: `go-backend/internal/tracking/hot_tracker.go`

- [ ] **Step 1: Implement hot_tracker.go**

Key implementation details:
- `HotInstance` struct with `Tags map[string]bool`
- `HotTracker` with `hots map[[2]int]*HotInstance`
- `NewHotTracker() *HotTracker`
- `Process(event models.Event)` — type-switch on ApplyBuff/RefreshBuff/RemoveBuff
- Refresh resets Tags to empty map
- `Get`, `GetAll`, `GetAllBySpell` methods

- [ ] **Step 2: Run tests**

```bash
cd go-backend && go test ./internal/tracking/ -v
```

Expected: hot_tracker_test.go tests PASS.

- [ ] **Step 3: Commit**

```bash
git add go-backend/internal/tracking/hot_tracker.go
git commit -m "feat: implement HotTracker"
```

### Task 21: Implement BuffTracker

**Files:**
- Create: `go-backend/internal/tracking/buff_tracker.go`

- [ ] **Step 1: Implement buff_tracker.go**

Key implementation details:
- `active map[int]int` (buffID → timestamp)
- `Process`, `IsActive`, `GetAppliedAt`

- [ ] **Step 2: Run tests**

```bash
cd go-backend && go test ./internal/tracking/ -v
```

Expected: all tracking tests PASS.

- [ ] **Step 3: Commit**

```bash
git add go-backend/internal/tracking/buff_tracker.go
git commit -m "feat: implement BuffTracker"
```

---

## Phase 9: Implementation — Talents & Pipeline

### Task 22: Implement TalentAttributor interface and BaseAttributor

**Files:**
- Create: `go-backend/internal/talents/attributor.go`

- [ ] **Step 1: Implement attributor.go**

Key implementation details:
- `TalentAttributor` interface with all methods from the design spec
- `BaseAttributor` struct with default implementations
- `NewBaseAttributor(name string, nodeID *int, talentID *int) BaseAttributor`
- Default `ProcessEvent`/`ProcessHeal`/`Finalize` return 0
- `IsSelected()` checks `CombatantInfo.TalentNodes` and `TalentIDs`
- `HasTalent(nodeID int) bool`
- `IsPlayerPet(sourceID int) bool`

- [ ] **Step 2: Verify it compiles**

```bash
cd go-backend && go build ./internal/talents/
```

- [ ] **Step 3: Commit**

```bash
git add go-backend/internal/talents/attributor.go
git commit -m "feat: implement TalentAttributor interface and BaseAttributor"
```

### Task 23: Implement Pipeline

**Files:**
- Create: `go-backend/internal/analysis/pipeline.go`

- [ ] **Step 1: Implement pipeline.go**

Port the Python `Pipeline.run()` logic exactly:
1. Parse all raw events
2. Calculate fight duration
3. For each event: handle CombatantInfo (set info, filter attributors), update trackers, call ProcessEvent, process heals (skip pets, track total, skip wasted, attribute)
4. Finalize attributors

- [ ] **Step 2: Run pipeline tests**

```bash
cd go-backend && go test ./internal/analysis/ -v
```

Expected: pipeline_test.go and combatantinfo_pipeline_test.go tests PASS.

- [ ] **Step 3: Commit**

```bash
git add go-backend/internal/analysis/pipeline.go
git commit -m "feat: implement analysis Pipeline"
```

### Task 24: Implement direct spell attributors

**Files:**
- Create: `go-backend/internal/talents/direct_spells.go`

- [ ] **Step 1: Implement direct_spells.go**

Port `DirectSpellAttributor` pattern: each attributor matches specific spell IDs and claims 100% of the healing. Also `RampantGrowthAttributor` (Regrowth ticks, bonus = amount - amount/2.0).

Constructors: `NewEverbloomSplashAttributor()`, `NewGroveGuardiansAttributor()`, `NewDreamSurgeAttributor()`, `NewEfflorescenceAttributor()`, `NewVerdancyAttributor()`, `NewRampantGrowthAttributor()`.

- [ ] **Step 2: Run tests**

```bash
cd go-backend && go test ./internal/talents/ -v -run "Everbloom|GroveGuardians|DreamSurge|DirectSpell|Rampant"
```

Expected: direct_spells_test.go tests PASS.

- [ ] **Step 3: Commit**

```bash
git add go-backend/internal/talents/direct_spells.go
git commit -m "feat: implement direct spell attributors"
```

### Task 25: Implement buff multiplier attributors

**Files:**
- Create: `go-backend/internal/talents/buff_multipliers.go`

- [ ] **Step 1: Implement buff_multipliers.go**

Port `StaticBuffAttributor` pattern: matches spell IDs, calculates `amount - amount / (1 + multiplier)`. Also `LifetreadingAttributor`.

Constructors: `NewWildSynthesisAttributor()`, `NewWildstalkersPowerAttributor()`, `NewLifetreadingAttributor()`.

- [ ] **Step 2: Run tests**

```bash
cd go-backend && go test ./internal/talents/ -v -run "WildSynthesis|Wildstalker|Lifetreading|StaticBuff"
```

Expected: buff_multipliers_test.go tests PASS.

- [ ] **Step 3: Commit**

```bash
git add go-backend/internal/talents/buff_multipliers.go
git commit -m "feat: implement buff multiplier attributors"
```

### Task 26: Implement remaining talent attributors

Implement each remaining talent attributor one at a time, running its tests after each:

1. **SoulOfTheForest** — tracks SotF buff, tags HoTs on consuming cast, attributes 60% bonus. Also PotA spread tracking.
2. **Convoke** — tracks channel window (4s), attributes 70% of healing during channel, tags HoTs applied during channel.
3. **TreeOfLife** — tracks ToL buff, +50% to Rejuv, +10% to other spells, WG buffer for extra targets.
4. **ImprovedWildGrowth** — attributes 2/7 of WG healing (2 extra targets out of 7). Skips during ToL.
5. **Reforestation** — tracks Swiftmend count, triggers mini-ToL every 4th SM, 10s duration.
6. **Abundance** — counts active Rejuvs, attributes crit bonus share on Regrowth crits.
7. **Photosynthesis** — detects unexplained Lifebloom blooms (no nearby remove/refresh/recast/SotF).
8. **NurturingDormancy** — attributes Rejuv ticks past base duration (17s).
9. **ProtectiveGrowth** — finalize-only: damage taken * 0.08 / 0.92.
10. **Wildstalker talents** — VigorousCreepers, Implant, RootNetwork, StrategicInfusion.
11. **SM/WG cooldown reduction** — tracks cast timestamps, computes effective CD, attributes fraction of downstream healing.

For each:

```bash
cd go-backend && go test ./internal/talents/ -v -run <TestName>
# Fix until PASS
git add go-backend/internal/talents/<file>.go
git commit -m "feat: implement <TalentName> attributor"
```

---

## Phase 10: Implementation — WCL Client

### Task 27: Implement WCL client

**Files:**
- Create: `go-backend/internal/wcl/client.go`
- Create: `go-backend/internal/wcl/queries.go`

- [ ] **Step 1: Implement client.go**

Key implementation details:
- `Client` struct with `clientID`, `clientSecret`, `token`, `httpClient`, `baseURL`, `oauthURL`
- `WithBaseURL` and `WithToken` option functions (for testing)
- `authenticate()` — POST to oauth URL with client credentials
- `query()` — POST GraphQL to API URL with bearer token
- `GetReport`, `GetEvents` (with pagination), `GetDamageTaken`
- `WCLQuerier` interface for mocking

- [ ] **Step 2: Implement queries.go**

Copy the 3 GraphQL query strings from Python.

- [ ] **Step 3: Run tests**

```bash
cd go-backend && go test ./internal/wcl/ -v -run "Authenticate|Paginate"
```

Expected: client_test.go tests PASS.

- [ ] **Step 4: Commit**

```bash
git add go-backend/internal/wcl/
git commit -m "feat: implement WCL GraphQL client"
```

### Task 28: Implement WCL cache

**Files:**
- Create: `go-backend/internal/wcl/cache.go`

- [ ] **Step 1: Implement cache.go**

Disk-based JSON cache wrapping `WCLQuerier`. Same pattern as Python `CachedWCLClient`.

- [ ] **Step 2: Run tests**

```bash
cd go-backend && go test ./internal/wcl/ -v -run Cache
```

Expected: cache_test.go tests PASS.

- [ ] **Step 3: Commit**

```bash
git add go-backend/internal/wcl/cache.go
git commit -m "feat: implement WCL response cache"
```

---

## Phase 11: Implementation — Web & Output

### Task 29: Implement web result cache

**Files:**
- Create: `go-backend/internal/web/cache.go`

- [ ] **Step 1: Implement cache.go**

Disk-based JSON cache for analysis results. Case-insensitive player name in key.

- [ ] **Step 2: Run tests**

```bash
cd go-backend && go test ./internal/web/ -v -run Cache
```

Expected: cache_test.go tests PASS.

- [ ] **Step 3: Commit**

```bash
git add go-backend/internal/web/cache.go
git commit -m "feat: implement web result cache"
```

### Task 30: Implement web router and handlers

**Files:**
- Create: `go-backend/internal/web/router.go`

- [ ] **Step 1: Implement router.go**

Key implementation details:
- `NewRouter(client WCLQuerier, cacheDir string) http.Handler`
- Chi router with `/api/health`, `/api/report/{code}`, `/api/analyze/{code}/{fightID}/{playerName}`
- Report endpoint: fetch report, filter fights (encounterID > 0), filter druids
- Analyze endpoint: full pipeline execution (same logic as Python routes.py)
- Rate limiting middleware, CORS, zerolog request logging

- [ ] **Step 2: Run tests**

```bash
cd go-backend && go test ./internal/web/ -v -run "Report|Health"
```

Expected: routes_test.go tests PASS.

- [ ] **Step 3: Commit**

```bash
git add go-backend/internal/web/router.go
git commit -m "feat: implement web router and handlers"
```

### Task 31: Implement output table rendering

**Files:**
- Create: `go-backend/internal/output/table.go`

- [ ] **Step 1: Implement table.go**

`RenderResults(results *AnalysisResults, fightName, playerName string) string` — format results as a text table. Include overlap disclaimer when attributed > total. Show "—" for negative unattributed.

- [ ] **Step 2: Run tests**

```bash
cd go-backend && go test ./internal/output/ -v
```

Expected: table_test.go tests PASS.

- [ ] **Step 3: Commit**

```bash
git add go-backend/internal/output/table.go
git commit -m "feat: implement output table rendering"
```

---

## Phase 12: CLI & Integration

### Task 32: Implement CLI with cobra

**Files:**
- Create: `go-backend/cmd/flourish/main.go` (replace placeholder)
- Create: `go-backend/cmd/flourish/analyze.go`
- Create: `go-backend/cmd/flourish/serve.go`

- [ ] **Step 1: Implement CLI**

- `analyze` command: load config, create WCL client, fetch report, interactive selection, run pipeline, render output
- `serve` command: start HTTP server with chi router
- Flags: `--fight`, `--player`, `--config-path`, `--port`
- Load `.env` for `WCL_CLIENT_ID` / `WCL_CLIENT_SECRET`

- [ ] **Step 2: Verify build**

```bash
cd go-backend && go build ./cmd/flourish && ./flourish --help
```

Expected: shows help text with `analyze` and `serve` subcommands.

- [ ] **Step 3: Commit**

```bash
git add go-backend/cmd/
git commit -m "feat: implement CLI with analyze and serve commands"
```

### Task 33: Copy talents.yaml config

**Files:**
- Create: `go-backend/config/talents.yaml` (copy from root)

- [ ] **Step 1: Copy config**

```bash
cp config/talents.yaml go-backend/config/talents.yaml
```

- [ ] **Step 2: Commit**

```bash
git add go-backend/config/
git commit -m "chore: copy talents.yaml to go-backend"
```

### Task 34: Final integration test

- [ ] **Step 1: Run all tests**

```bash
cd go-backend && go test ./... -v
```

Expected: ALL tests pass.

- [ ] **Step 2: Build and verify**

```bash
cd go-backend && go build ./cmd/flourish
```

Expected: clean build.

- [ ] **Step 3: Final commit**

```bash
git add -A go-backend/
git commit -m "chore: all Go backend tests passing"
```
