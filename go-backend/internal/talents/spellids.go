package talents

import "fmt"

// Spell and ability IDs from World of Warcraft used across talent attributors.
const (
	// Core restoration spells
	Rejuvenation     = 774
	GerminationRejuv = 155777
	Regrowth         = 8936
	WildGrowth       = 48438
	Swiftmend        = 18562
	Lifebloom        = 33763
	LifebloomBloom   = 33778
	Efflorescence    = 81269
	CenarionWard     = 157982

	// Buffs
	TreeOfLifeBuff      = 33891
	SoulOfTheForestBuff = 114108

	// Talent-granted healing spells
	SymbioticBloomSpell   = 439530
	EverbloomSplash       = 1244341
	Verdancy              = 392329
	DreamSurge            = 434141
	BurstingGrowthSpell   = 440121
	ThrivingGrowthSpell   = 474760
	NaturesBountySpell    = 1264376
	RegenerativeHeartwood = 392117
	EmbraceOfTheDream     = 392124
	ThrivingVegetation    = 447132
	CultivationSpell      = 200390

	// Grove Guardian spells
	GroveGuardianNourish = 422090
	GroveGuardianHeal    = 142421
	GroveGuardianSummon  = 102693

	// Dryad / pet spells
	DryadTranquility   = 1264659
	DryadRegrowthSpell = 1264664
	SpiritOfTheThicket = 1264905

	// Ysera's Gift
	YserasGift1 = 145108
	YserasGift2 = 145109
	YserasGift3 = 145110

	// Convoke
	ConvokeTheSpirits = 391528
	ConvokeLegacy     = 323764

	// Tree of Life multipliers (used by ToL and Reforestation)
	TolRejuvDivisor = 1.5
	TolOtherDivisor = 1.1

	// Crit rating per percent (shared by Abundance and Strategic Infusion)
	CritRatingPerPercent = 700.0
)

// RegrowthDamageTakenFilter is the WCL filter for damage taken while Regrowth HoT is active.
const RegrowthDamageTakenFilter = `IN RANGE FROM (type = "applybuff" OR type = "refreshbuff") AND ability.id = 8936 TO type = "removebuff" AND ability.id = 8936 GROUP BY target ON target END`

// IsRejuv returns true if the spell ID is Rejuvenation or Germination Rejuvenation.
func IsRejuv(id int) bool {
	return id == Rejuvenation || id == GerminationRejuv
}

// SpellName returns a human-readable name for known restoration druid spell IDs.
func SpellName(id int) string {
	if name, ok := spellNames[id]; ok {
		return name
	}
	return fmt.Sprintf("Unknown (%d)", id)
}

var spellNames = map[int]string{
	Rejuvenation:          "Rejuvenation",
	GerminationRejuv:      "Rejuvenation (Germ)",
	Regrowth:              "Regrowth",
	WildGrowth:            "Wild Growth",
	Swiftmend:             "Swiftmend",
	Lifebloom:             "Lifebloom",
	LifebloomBloom:        "Lifebloom (Bloom)",
	Efflorescence:         "Efflorescence",
	CenarionWard:          "Cenarion Ward",
	SymbioticBloomSpell:   "Symbiotic Bloom",
	EverbloomSplash:       "Everbloom",
	Verdancy:              "Verdancy",
	DreamSurge:            "Dream Surge",
	BurstingGrowthSpell:   "Bursting Growth",
	ThrivingGrowthSpell:   "Thriving Growth",
	NaturesBountySpell:    "Nature's Bounty",
	RegenerativeHeartwood: "Regenerative Heartwood",
	EmbraceOfTheDream:     "Embrace of the Dream",
	ThrivingVegetation:    "Thriving Vegetation",
	CultivationSpell:      "Cultivation",
	GroveGuardianNourish:  "Grove Guardian (Nourish)",
	GroveGuardianHeal:     "Grove Guardian (Heal)",
	DryadTranquility:      "Dryad Tranquility",
	DryadRegrowthSpell:    "Dryad Regrowth",
	SpiritOfTheThicket:    "Spirit of the Thicket",
	YserasGift1:           "Ysera's Gift",
	YserasGift2:           "Ysera's Gift",
	YserasGift3:           "Ysera's Gift",
	ConvokeTheSpirits:     "Convoke",
	ConvokeLegacy:         "Convoke",
	// Tranquility
	740:   "Tranquility",
	44203: "Tranquility",
	// Nature's Swiftness Regrowth
	132158: "Nature's Swiftness",

	// Other druid talents / abilities
	22842:  "Frenzied Regeneration",
	474683: "Aessina's Renewal",
	439902: "Flower Walk",
	455474: "Lethal Preservation",
	455470: "Lethal Preservation",

	// External / consumable healing
	143924:  "Leech",
	1265145: "Refreshing Drink",
	1234768: "Health Potion",
	6262:    "Healthstone",
	361509:  "Living Flame",
	374251:  "Cauterizing Flame",
	361195:  "Verdant Embrace",
	130654:  "Chi Burst",
	413786:  "Fate Mirror",
}
