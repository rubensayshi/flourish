package talents

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
