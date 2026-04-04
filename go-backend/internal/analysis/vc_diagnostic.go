package analysis

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/talents"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

// VCDiagnostic collects heal observations to determine whether Vigorous Creepers
// is additive or multiplicative with Mastery: Harmony.
//
// Method: group periodic heal ticks by (spell, crit, totalMasteryStacks), then
// compare average RawHeal for ticks where Symbiotic Bloom IS vs IS NOT one of
// those mastery stacks. Since mastery contribution is identical at the same stack
// count, the only difference is VC's +20%. If multiplicative, the ratio is 1.20.
// If additive, the ratio is less than 1.20.
type VCDiagnostic struct {
	HotTracker  *tracking.HotTracker
	BuffTracker *tracking.BuffTracker
	masteryRaw  float64
	obs         []vcObservation
}

type vcObservation struct {
	SpellID       int
	RawHeal       int
	Crit          bool
	TotalStacks   int  // total mastery HoTs on target (includes SB if present)
	HasSB         bool // Symbiotic Bloom is one of the stacks
}

var masteryHoTs = map[int]bool{
	talents.Rejuvenation:        true,
	talents.GerminationRejuv:    true,
	talents.Regrowth:            true,
	talents.WildGrowth:          true,
	talents.Lifebloom:           true,
	talents.CenarionWard:        true,
	talents.SymbioticBloomSpell: true,
	talents.CultivationSpell:    true,
}

var spellNames = map[int]string{
	talents.Rejuvenation:        "Rejuv",
	talents.GerminationRejuv:    "Germ",
	talents.Regrowth:            "Regrowth",
	talents.WildGrowth:          "WG",
	talents.Lifebloom:           "LB",
	talents.SymbioticBloomSpell: "SB",
	talents.CultivationSpell:    "Cultiv",
	talents.Efflorescence:       "Effl",
}

func NewVCDiagnostic(_ []float64) *VCDiagnostic {
	return &VCDiagnostic{
		HotTracker:  tracking.NewHotTracker(),
		BuffTracker: tracking.NewBuffTracker(),
	}
}

func (d *VCDiagnostic) Run(rawEvents []map[string]any) string {
	var events []models.Event
	for _, raw := range rawEvents {
		if e := models.ParseEvent(raw); e != nil {
			events = append(events, e)
		}
	}

	for _, event := range events {
		if ci, ok := event.(*models.CombatantInfoEvent); ok {
			d.masteryRaw = ci.Mastery
			continue
		}

		switch event.(type) {
		case *models.ApplyBuffEvent, *models.RefreshBuffEvent, *models.RemoveBuffEvent:
			d.HotTracker.Process(event)
			d.BuffTracker.Process(event)
		}

		if he, ok := event.(*models.HealEvent); ok {
			// Only count ticks from mastery-eligible HoTs (excluding SB itself)
			if !he.Tick || !masteryHoTs[he.AbilityID] || he.AbilityID == talents.SymbioticBloomSpell {
				continue
			}

			stacks := d.HotTracker.CountByTarget(he.TargetID, masteryHoTs)
			hasSB := d.HotTracker.Get(he.TargetID, talents.SymbioticBloomSpell) != nil

			d.obs = append(d.obs, vcObservation{
				SpellID:     he.AbilityID,
				RawHeal:     he.RawHeal(),
				Crit:        he.HitType == 2,
				TotalStacks: stacks,
				HasSB:       hasSB,
			})
		}
	}

	return d.report()
}

type groupKey struct {
	SpellID     int
	Crit        bool
	TotalStacks int
}

type groupStats struct {
	withSB    []int
	withoutSB []int
}

func (d *VCDiagnostic) report() string {
	var sb strings.Builder

	sb.WriteString("=== Vigorous Creepers Diagnostic ===\n")
	sb.WriteString(fmt.Sprintf("Mastery rating from WCL: %.0f\n", d.masteryRaw))
	sb.WriteString(fmt.Sprintf("Total periodic observations: %d\n\n", len(d.obs)))

	sb.WriteString("Method: group ticks by (spell, crit, totalMasteryStacks).\n")
	sb.WriteString("Compare avg RawHeal WITH vs WITHOUT Symbiotic Bloom at same stack count.\n")
	sb.WriteString("Since mastery is identical, the ratio isolates VC's effect.\n")
	sb.WriteString("  Multiplicative → ratio = 1.20\n")
	sb.WriteString("  Additive       → ratio < 1.20\n\n")

	// Group observations
	groups := make(map[groupKey]*groupStats)
	for _, o := range d.obs {
		k := groupKey{SpellID: o.SpellID, Crit: o.Crit, TotalStacks: o.TotalStacks}
		g, ok := groups[k]
		if !ok {
			g = &groupStats{}
			groups[k] = g
		}
		if o.HasSB {
			g.withSB = append(g.withSB, o.RawHeal)
		} else {
			g.withoutSB = append(g.withoutSB, o.RawHeal)
		}
	}

	// Sort keys
	var keys []groupKey
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].SpellID != keys[j].SpellID {
			return keys[i].SpellID < keys[j].SpellID
		}
		if keys[i].Crit != keys[j].Crit {
			return !keys[i].Crit
		}
		return keys[i].TotalStacks < keys[j].TotalStacks
	})

	// Header
	sb.WriteString(fmt.Sprintf("%-10s %-5s %-7s %7s %7s %10s %10s %7s\n",
		"Spell", "Crit", "Stacks", "#noSB", "#wSB", "Avg noSB", "Avg wSB", "Ratio"))
	sb.WriteString(strings.Repeat("-", 75) + "\n")

	type ratioObs struct {
		ratio  float64
		weight int // min(#noSB, #wSB) as confidence weight
	}
	var allRatios []ratioObs

	for _, k := range keys {
		g := groups[k]
		// Need reasonable sample sizes in both groups
		if len(g.withSB) < 5 || len(g.withoutSB) < 5 {
			continue
		}

		avgWith := mean(g.withSB)
		avgWithout := mean(g.withoutSB)
		if avgWithout == 0 {
			continue
		}
		ratio := avgWith / avgWithout

		name := spellNames[k.SpellID]
		if name == "" {
			name = fmt.Sprintf("%d", k.SpellID)
		}
		critStr := " "
		if k.Crit {
			critStr = "crit"
		}

		sb.WriteString(fmt.Sprintf("%-10s %-5s %-7d %7d %7d %10.0f %10.0f %7.4f\n",
			name, critStr, k.TotalStacks,
			len(g.withoutSB), len(g.withSB),
			avgWithout, avgWith, ratio))

		weight := len(g.withSB)
		if len(g.withoutSB) < weight {
			weight = len(g.withoutSB)
		}
		allRatios = append(allRatios, ratioObs{ratio, weight})
	}

	// Weighted average ratio
	if len(allRatios) > 0 {
		var sumW, sumWR float64
		for _, r := range allRatios {
			w := float64(r.weight)
			sumW += w
			sumWR += w * r.ratio
		}
		weightedAvg := sumWR / sumW

		// Weighted RMSE from 1.20
		var sumSqMult, sumSqAdd float64
		for _, r := range allRatios {
			w := float64(r.weight) / sumW
			sumSqMult += w * (r.ratio - 1.20) * (r.ratio - 1.20)
			// For additive we don't know the exact expected value per group,
			// but the overall average being significantly below 1.20 is the signal
			_ = sumSqAdd
		}
		rmseMult := math.Sqrt(sumSqMult)

		sb.WriteString(fmt.Sprintf("\n=== Summary (%d groups, %d weighted observations) ===\n",
			len(allRatios), int(sumW)))
		sb.WriteString(fmt.Sprintf("Weighted average ratio (with SB / without SB): %.4f\n", weightedAvg))
		sb.WriteString(fmt.Sprintf("Expected if multiplicative: 1.2000\n"))
		sb.WriteString(fmt.Sprintf("Weighted RMSE vs 1.20:      %.4f\n", rmseMult))
		sb.WriteString(fmt.Sprintf("Deviation from 1.20:        %+.4f (%.1f%%)\n",
			weightedAvg-1.20, (weightedAvg-1.20)/1.20*100))

		if math.Abs(weightedAvg-1.20) < 0.02 {
			sb.WriteString("\n>>> Result: MULTIPLICATIVE — ratio is ~1.20 as expected <<<\n")
		} else if weightedAvg < 1.20 {
			sb.WriteString("\n>>> Result: ADDITIVE — ratio is below 1.20 <<<\n")
		} else {
			sb.WriteString("\n>>> Result: UNCLEAR — ratio is above 1.20, investigate further <<<\n")
		}
	}

	return sb.String()
}

func mean(vals []int) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0
	for _, v := range vals {
		sum += v
	}
	return float64(sum) / float64(len(vals))
}
