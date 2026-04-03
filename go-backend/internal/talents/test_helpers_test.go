package talents_test

// Shared test helpers and constants for talent tests.

// Spell IDs
const (
	swiftmend        = 18562
	rejuv            = 774
	germinationRejuv = 155777
	regrowth         = 8936
	wildGrowth       = 48438
	lifebloom        = 33763
	lifebloomBloom   = 33778
	convoke          = 391528
	convokeLegacy    = 323764
	nourish          = 422090
	treantHeal       = 142421
	efflorescence    = 81269
	dreamBloom       = 434141
	dryadBeam        = 1264905
	dryadTranq       = 1264659
	dryadRegrowth    = 1264664
	symBloom         = 439530
	sotfBuffID       = 114108
	guardianAbil     = 102693
)

// Actor IDs
const (
	player    = 1
	target    = 10
	spread1   = 20
	spread2   = 30
	petSource = 50
)

func makeHeal(ts, ability, amount int, opts ...func(map[string]any)) map[string]any {
	m := map[string]any{
		"timestamp":     ts,
		"type":          "heal",
		"sourceID":      1,
		"targetID":      2,
		"abilityGameID": ability,
		"amount":        amount,
		"overheal":      0,
		"hitType":       1,
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
		"timestamp":     ts,
		"type":          "cast",
		"sourceID":      1,
		"targetID":      2,
		"abilityGameID": ability,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func makeBegincast(ts, ability int, opts ...func(map[string]any)) map[string]any {
	m := map[string]any{
		"timestamp":     ts,
		"type":          "begincast",
		"sourceID":      1,
		"targetID":      2,
		"abilityGameID": ability,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func makeApply(ts, ability int, opts ...func(map[string]any)) map[string]any {
	m := map[string]any{
		"timestamp":     ts,
		"type":          "applybuff",
		"sourceID":      1,
		"targetID":      2,
		"abilityGameID": ability,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func makeRefresh(ts, ability int, opts ...func(map[string]any)) map[string]any {
	m := map[string]any{
		"timestamp":     ts,
		"type":          "refreshbuff",
		"sourceID":      1,
		"targetID":      2,
		"abilityGameID": ability,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func makeRemove(ts, ability int, opts ...func(map[string]any)) map[string]any {
	m := map[string]any{
		"timestamp":     ts,
		"type":          "removebuff",
		"sourceID":      1,
		"targetID":      2,
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

func makeSummon(ts, ability int, opts ...func(map[string]any)) map[string]any {
	m := map[string]any{
		"timestamp":     ts,
		"type":          "summon",
		"sourceID":      1,
		"targetID":      2,
		"abilityGameID": ability,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func withMastery(v float64) func(map[string]any) {
	return func(m map[string]any) { m["mastery"] = v }
}
