package models

const OverhealWasteThreshold = 0.5

// Event type strings from WarcraftLogs.
const (
	EventCombatantInfo = "combatantinfo"
	EventCast          = "cast"
	EventBeginCast     = "begincast"
	EventHeal          = "heal"
	EventApplyBuff     = "applybuff"
	EventRefreshBuff   = "refreshbuff"
	EventRemoveBuff    = "removebuff"
	EventSummon        = "summon"
	EventDamage        = "damage"
)

// Event is the interface all parsed events implement.
type Event interface {
	GetBase() *BaseEvent
}

type BaseEvent struct {
	Timestamp int
	SourceID  int
	Type      string
}

func (e *BaseEvent) GetBase() *BaseEvent { return e }

type HealEvent struct {
	BaseEvent
	TargetID  int
	AbilityID int
	Amount    int
	Overheal  int
	Absorb    int
	HitType   int // 1=normal, 2=crit
	Tick      bool
}

func (e *HealEvent) RawHeal() int {
	return e.Amount + e.Overheal + e.Absorb
}

func (e *HealEvent) OverhealPct() float64 {
	raw := e.RawHeal()
	if raw == 0 {
		return 0.0
	}
	return float64(e.Overheal) / float64(raw)
}

func (e *HealEvent) IsWasted() bool {
	return e.OverhealPct() > OverhealWasteThreshold
}

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

type DamageEvent struct {
	BaseEvent
	TargetID  int
	AbilityID int
	Amount    int
	Absorbed  int
	Overkill  int
}

type CombatantInfoEvent struct {
	BaseEvent
	TalentNodes   map[int]bool
	TalentIDs     map[int]bool
	TalentRanks   map[int]int
	Intellect     int
	CritSpell     float64
	HasteSpell    float64
	Mastery       float64
	Versatility   float64 // versatilityHealingDone rating
	SpecID        int
}

// GetInt extracts an int from map[string]any, handling JSON float64.
func GetInt(raw map[string]any, key string, def int) int {
	v, ok := raw[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	default:
		return def
	}
}

func getFloat(raw map[string]any, key string, def float64) float64 {
	v, ok := raw[key]
	if !ok {
		return def
	}
	switch n := v.(type) {
	case float64:
		return n
	case int:
		return float64(n)
	default:
		return def
	}
}

func getBool(raw map[string]any, key string) bool {
	v, ok := raw[key]
	if !ok {
		return false
	}
	b, ok := v.(bool)
	return ok && b
}

func GetString(raw map[string]any, key string) string {
	v, ok := raw[key]
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func ParseEvent(raw map[string]any) Event {
	eventType := GetString(raw, "type")

	if eventType == EventCombatantInfo {
		talentTree, _ := raw["talentTree"].([]any)
		talentNodes := make(map[int]bool)
		talentIDs := make(map[int]bool)
		talentRanks := make(map[int]int)
		for _, entry := range talentTree {
			t, ok := entry.(map[string]any)
			if !ok {
				continue
			}
			nodeID := GetInt(t, "nodeID", 0)
			id := GetInt(t, "id", 0)
			rank := GetInt(t, "rank", 1)
			talentNodes[nodeID] = true
			talentIDs[id] = true
			talentRanks[id] = rank
		}
		return &CombatantInfoEvent{
			BaseEvent:   BaseEvent{Timestamp: GetInt(raw, "timestamp", 0), SourceID: GetInt(raw, "sourceID", 0), Type: eventType},
			TalentNodes: talentNodes,
			TalentIDs:   talentIDs,
			TalentRanks: talentRanks,
			Intellect:   GetInt(raw, "intellect", 0),
			CritSpell:   getFloat(raw, "critSpell", 0),
			HasteSpell:  getFloat(raw, "hasteSpell", 0),
			Mastery:     getFloat(raw, "mastery", 0),
			Versatility: getFloat(raw, "versatilityHealingDone", 0),
			SpecID:      GetInt(raw, "specID", 0),
		}
	}

	base := BaseEvent{
		Timestamp: GetInt(raw, "timestamp", 0),
		SourceID:  GetInt(raw, "sourceID", 0),
		Type:      eventType,
	}
	targetID := GetInt(raw, "targetID", 0)
	abilityID := GetInt(raw, "abilityGameID", 0)

	switch eventType {
	case EventCast, EventBeginCast:
		return &CastEvent{BaseEvent: base, TargetID: targetID, AbilityID: abilityID}
	case EventHeal:
		return &HealEvent{
			BaseEvent: base,
			TargetID:  targetID,
			AbilityID: abilityID,
			Amount:    GetInt(raw, "amount", 0),
			Overheal:  GetInt(raw, "overheal", 0),
			Absorb:    GetInt(raw, "absorb", 0),
			HitType:   GetInt(raw, "hitType", 1),
			Tick:      getBool(raw, "tick"),
		}
	case EventApplyBuff:
		return &ApplyBuffEvent{BaseEvent: base, TargetID: targetID, AbilityID: abilityID}
	case EventRefreshBuff:
		return &RefreshBuffEvent{BaseEvent: base, TargetID: targetID, AbilityID: abilityID}
	case EventRemoveBuff:
		return &RemoveBuffEvent{BaseEvent: base, TargetID: targetID, AbilityID: abilityID}
	case EventSummon:
		return &SummonEvent{BaseEvent: base, TargetID: targetID, AbilityID: abilityID}
	case EventDamage:
		return &DamageEvent{
			BaseEvent: base,
			TargetID:  targetID,
			AbilityID: abilityID,
			Amount:    GetInt(raw, "amount", 0),
			Absorbed:  GetInt(raw, "absorbed", 0),
			Overkill:  GetInt(raw, "overkill", 0),
		}
	default:
		return nil
	}
}
