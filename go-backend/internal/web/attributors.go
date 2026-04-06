package web

import (
	"strings"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
)

// HeroTrees maps hero tree names to their talent names.
var HeroTrees = map[string]map[string]bool{
	"Wildstalker": {
		"Wildstalker's Power": true, "Patient Custodian": true, "Vigorous Creepers": true,
		"Bursting Growth": true, "Root Network": true, "Implant": true, "Twin Sprouts": true,
		"Thriving Growth": true, "Symbiotic Bloom Mastery": true, "Strategic Infusion": true,
	},
	"Keeper of the Grove": {
		"Dream Surge": true, "Harmony of the Grove": true, "Power of Nature": true,
		"Bounteous Bloom": true, "Grove's Inspiration": true, "Cenarius' Might": true,
		"Protective Growth": true, "Spirit of the Thicket": true, "Sylvan Beckoning": true,
		"Early Spring + Dryad's Dance": true, "Early Spring (WG)": true, "Potent Enchantments": true,
	},
}

func heroTreeFor(name string) string {
	for tree, talents := range HeroTrees {
		if talents[name] {
			return tree
		}
	}
	return ""
}

// BuildAttributors creates the full set of talent attributors from config.
func BuildAttributors(config *models.Config, damageTaken int) []talents.TalentAttributor {
	drTable := config.Mastery.DRTable

	sotf := talents.NewSoulOfTheForestAttributor()
	gg := talents.NewGroveGuardiansAttributor()
	smCd := talents.NewSmCooldownReductionAttributor([]talents.TalentAttributor{sotf, gg})
	wgCd := talents.NewWgCooldownReductionAttributor([]talents.TalentAttributor{gg}, false)

	potentEnch := talents.NewPotentEnchantmentsAttributor()

	hb := talents.NewHarmoniousBloomingAttributor(drTable)
	sb := talents.NewSymbioticBloomMasteryAttributor(drTable)
	// SB needs to know if HB is active to account for LB's +2 virtual mastery stacks.
	// Check if HB is not skipped in config.
	if cfg, ok := config.Talents["harmonious_blooming"]; !ok || !cfg.Skip {
		sb.SetHBActive(true)
	}

	all := []talents.TalentAttributor{
		sotf,
		talents.NewEverbloomSplashAttributor(),
		talents.NewBloomingFrenzyAttributor(),
		gg,
		talents.NewDreamSurgeAttributor(),
		talents.NewEfflorescenceAttributor(),
		talents.NewVerdancyAttributor(),
		talents.NewNaturesBountyAttributor(),
		talents.NewRegenerativeHeartwoodAttributor(),
		talents.NewRampantGrowthAttributor(),
		talents.NewCultivationAttributor(),
		talents.NewYserasGiftAttributor(),
		talents.NewEmbraceOfTheDreamAttributor(),
		talents.NewImprovedSwiftmendAttributor(),
		talents.NewUnstoppableGrowthAttributor(),
		talents.NewLivelinessAttributor(),
		talents.NewRegenesisAttributor(),
		talents.NewThrivingVegetationRejuvAttributor(),
		talents.NewThrivingVegetationRegrowthAttributor(),
		talents.NewWildSynthesisAttributor(),
		talents.NewGrovesInspirationAttributor(),
		talents.NewCenariusMightAttributor(),
		talents.NewBountifulBloomAttributor(),
		talents.NewHarmonyOfTheGroveAttributor(),
		talents.NewPowerOfNatureAttributor(),
		talents.NewSpiritOfTheThicketAttributor(),
		talents.NewSylvanBeckoningAttributor(),
		talents.NewWildstalkersPowerAttributor(),
		talents.NewPatientCustodianAttributor(),
		talents.NewLifetreadingAttributor(),
		talents.NewTreeOfLifeAttributor(),
		talents.NewConvokeAttributor(),
		talents.NewImprovedWildGrowthAttributor(),
		potentEnch,
		talents.NewReforestationAttributor(potentEnch),
		talents.NewVigorousCreepersAttributor(),
		talents.NewImplantAttributor(),
		talents.NewTwinSproutsAttributor(),
		talents.NewRootNetworkAttributor(),
		talents.NewStrategicInfusionAttributor(),
		talents.NewBurstingGrowthAttributor(),
		talents.NewThrivingGrowthAttributor(),
		hb,
		sb,
		talents.NewIntensityAttributor(),
		talents.NewAbundanceAttributor(),
		talents.NewPhotosynthesisAttributor(),
		talents.NewNurturingDormancyAttributor(),
		talents.NewProtectiveGrowthAttributor(damageTaken),
		smCd,
		wgCd,
	}

	var active []talents.TalentAttributor
	for _, a := range all {
		key := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(a.Name(), " ", "_"), "'", ""))
		if cfg, ok := config.Talents[key]; ok && cfg.Skip {
			continue
		}
		active = append(active, a)
	}
	return active
}
