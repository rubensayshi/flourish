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
		"Bursting Growth": true, "Root Network": true, "Implant": true,
		"Thriving Growth": true, "Symbiotic Bloom Mastery": true, "Strategic Infusion": true,
	},
	"Keeper of the Grove": {
		"Dream Surge": true, "Harmony of the Grove": true, "Power of Nature": true,
		"Bounteous Bloom": true, "Grove's Inspiration": true, "Cenarius' Might": true,
		"Protective Growth": true, "Spirit of the Thicket": true, "Sylvan Beckoning": true,
		"Early Spring + Dryad's Dance": true, "Early Spring (WG)": true,
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
	convokeCfg := config.Talents["convoke_the_spirits"]
	convokeRatio := 0.7
	if convokeCfg.Multiplier != nil {
		convokeRatio = *convokeCfg.Multiplier
	}

	baseStacks := config.Mastery.BaseStacks
	drTable := config.Mastery.DRTable

	sotf := talents.NewSoulOfTheForestAttributor()
	gg := talents.NewGroveGuardiansAttributor()
	smCd := talents.NewSmCooldownReductionAttributor([]talents.TalentAttributor{sotf, gg})
	wgCd := talents.NewWgCooldownReductionAttributor([]talents.TalentAttributor{gg}, false)

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
		talents.NewFlourishAttributor(),
		talents.NewCultivationAttributor(),
		talents.NewYserasGiftAttributor(),
		talents.NewEmbraceOfTheDreamAttributor(),
		talents.NewImprovedSwiftmendAttributor(),
		talents.NewUnstoppableGrowthAttributor(),
		talents.NewLivelinessAttributor(),
		talents.NewRegenesisAttributor(),
		talents.NewThrivingVegetationAttributor(),
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
		talents.NewConvokeAttributor(convokeRatio),
		talents.NewImprovedWildGrowthAttributor(),
		talents.NewReforestationAttributor(),
		talents.NewVigorousCreepersAttributor(),
		talents.NewImplantAttributor(),
		talents.NewRootNetworkAttributor(),
		talents.NewStrategicInfusionAttributor(),
		talents.NewBurstingGrowthAttributor(),
		talents.NewThrivingGrowthAttributor(),
		talents.NewHarmoniousBloomingAttributor(baseStacks, drTable),
		talents.NewSymbioticBloomMasteryAttributor(baseStacks, drTable),
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
