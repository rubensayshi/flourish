package talents

import (
	"math"

	"github.com/rdruid-talent-analyzer/go-backend/internal/models"
	"github.com/rdruid-talent-analyzer/go-backend/internal/tracking"
)

const (
	EarlySpringNodeID   = 94591
	EarlySpringTalentID = 117895
	DryadsDanceNodeID   = 109713
	RenewingSurgeNodeID = 82060
	ProsperityNodeID    = 82079
	ProsperityTalentID  = 103136

	baseSMCDMS               = 15000
	baseWGCDMS               = 10000
	wg4pcReductionMS         = 2000
	renewingSurgeReductionAvg = 0.195
	earlySpringReductionMS   = 1000
	dryadsDanceSpeedFactor   = 1.25
	onCooldownToleranceMS    = 1500
	dryadGapThresholdMS      = 2000
)

func ComputeEffectiveCd(hasRenewingSurge, hasEarlySpring bool, dryadOverlapMS float64) float64 {
	cd := float64(baseSMCDMS)
	if hasRenewingSurge {
		cd *= (1 - renewingSurgeReductionAvg)
	}
	if hasEarlySpring {
		cd -= earlySpringReductionMS
	}
	if dryadOverlapMS > 0 {
		overlap := math.Min(dryadOverlapMS, cd)
		remaining := cd - overlap
		cd = remaining + overlap/dryadsDanceSpeedFactor
	}
	return cd
}

func ComputeEffectiveWgCd(hasEarlySpring, has4pc bool) float64 {
	cd := float64(baseWGCDMS)
	if has4pc {
		cd -= wg4pcReductionMS
	}
	if hasEarlySpring {
		cd -= earlySpringReductionMS
	}
	return cd
}

// SmCooldownReductionAttributor tracks SM cast timestamps and Dryad windows.
type SmCooldownReductionAttributor struct {
	BaseAttributor
	smCasts        []int
	dryadWindowsVal [][2]int // {start, end}
	dryadStart     *int
	dryadLastHeal  *int
	downstream     []TalentAttributor
}

func NewSmCooldownReductionAttributor(downstream []TalentAttributor) *SmCooldownReductionAttributor {
	return &SmCooldownReductionAttributor{
		BaseAttributor: NewBaseAttributor("Early Spring + Dryad's Dance", nil, nil),
		downstream:     downstream,
	}
}

func (a *SmCooldownReductionAttributor) IsSelected() bool {
	if a.CombatantInfo == nil {
		return true
	}
	hasES := a.CombatantInfo.TalentNodes[EarlySpringNodeID] && a.CombatantInfo.TalentIDs[EarlySpringTalentID]
	hasDD := a.CombatantInfo.TalentNodes[DryadsDanceNodeID]
	return hasES || hasDD
}

func (a *SmCooldownReductionAttributor) SmCastTimestamps() []int { return a.smCasts }
func (a *SmCooldownReductionAttributor) DryadWindows() [][2]int  { return a.dryadWindowsVal }

func (a *SmCooldownReductionAttributor) closeDryadWindow() {
	if a.dryadStart != nil && a.dryadLastHeal != nil {
		a.dryadWindowsVal = append(a.dryadWindowsVal, [2]int{*a.dryadStart, *a.dryadLastHeal})
		a.dryadStart = nil
		a.dryadLastHeal = nil
	}
}

func isDryadHealSpell(id int) bool {
	return id == DryadTranquility || id == DryadRegrowthSpell || id == SpiritOfTheThicket
}

func (a *SmCooldownReductionAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
	// Track Dryad windows from pet heal events
	if he, ok := event.(*models.HealEvent); ok && isDryadHealSpell(he.AbilityID) && a.IsPlayerPet(he.SourceID) {
		if a.dryadLastHeal != nil && he.Timestamp-*a.dryadLastHeal > dryadGapThresholdMS {
			a.closeDryadWindow()
		}
		if a.dryadStart == nil {
			ts := he.Timestamp
			a.dryadStart = &ts
		}
		ts := he.Timestamp
		a.dryadLastHeal = &ts
	} else {
		// Non-dryad event: close window if gap exceeded
		if a.dryadLastHeal != nil && event.GetBase().Timestamp-*a.dryadLastHeal > dryadGapThresholdMS {
			a.closeDryadWindow()
		}
	}

	if ce, ok := event.(*models.CastEvent); ok && ce.AbilityID == Swiftmend {
		a.smCasts = append(a.smCasts, ce.Timestamp)
	}
}

func (a *SmCooldownReductionAttributor) dryadOverlapInWindow(windowStart, windowDuration float64) float64 {
	cdEnd := windowStart + windowDuration
	overlap := 0.0
	for _, w := range a.dryadWindowsVal {
		oStart := math.Max(windowStart, float64(w[0]))
		oEnd := math.Min(cdEnd, float64(w[1]))
		if oEnd > oStart {
			overlap += oEnd - oStart
		}
	}
	return overlap
}

type pendingCharge struct {
	completion float64
	reducedCD  float64
}

func (a *SmCooldownReductionAttributor) Finalize() float64 {
	a.closeDryadWindow()

	if len(a.smCasts) < 2 {
		return 0.0
	}

	hasRS := a.HasTalent(RenewingSurgeNodeID)
	hasES := a.HasTalent(EarlySpringNodeID) && a.CombatantInfo != nil && a.CombatantInfo.TalentIDs[EarlySpringTalentID]
	hasDD := a.HasTalent(DryadsDanceNodeID)
	hasProsperity := a.HasTalent(ProsperityNodeID) && a.CombatantInfo != nil && a.CombatantInfo.TalentIDs[ProsperityTalentID]

	maxCharges := 1
	if hasProsperity {
		maxCharges = 2
	}

	unreducedCD := ComputeEffectiveCd(hasRS, false, 0)

	charges := maxCharges
	var pending []pendingCharge
	totalRatio := 0.0
	totalCasts := len(a.smCasts)

	for _, castTS := range a.smCasts {
		wasDepleted := charges == 0
		var lastRestore *pendingCharge

		for len(pending) > 0 && pending[0].completion <= float64(castTS) {
			entry := pending[0]
			pending = pending[1:]
			if charges < maxCharges {
				charges++
			}
			if wasDepleted {
				e := entry
				lastRestore = &e
				wasDepleted = charges == 0
			}
		}

		onCooldown := false
		reducedCDUsed := 0.0

		if charges == 0 && len(pending) > 0 {
			entry := pending[0]
			if entry.completion <= float64(castTS)+onCooldownToleranceMS {
				onCooldown = true
				reducedCDUsed = entry.reducedCD
				pending = pending[1:]
				charges++
			} else {
				pending = pending[1:]
				charges++
			}
		} else if lastRestore != nil {
			if float64(castTS)-lastRestore.completion <= onCooldownToleranceMS {
				onCooldown = true
				reducedCDUsed = lastRestore.reducedCD
			}
		}

		charges--

		rechargeStart := float64(castTS)
		if len(pending) > 0 {
			rechargeStart = pending[len(pending)-1].completion
		}
		dryadOverlap := 0.0
		if hasDD {
			dryadOverlap = a.dryadOverlapInWindow(rechargeStart, unreducedCD)
		}
		reducedCD := ComputeEffectiveCd(hasRS, hasES, dryadOverlap)
		pending = append(pending, pendingCharge{rechargeStart + reducedCD, reducedCD})

		if onCooldown {
			ratio := 1 - (reducedCDUsed / unreducedCD)
			totalRatio += math.Max(ratio, 0.0)
		}
	}

	if totalRatio == 0.0 {
		return 0.0
	}

	extraCastFraction := totalRatio / float64(totalCasts)
	downstreamTotal := 0.0
	for _, attr := range a.downstream {
		downstreamTotal += attr.GetTotalAttributed()
	}

	return extraCastFraction * downstreamTotal
}

// WgCooldownReductionAttributor tracks WG cast timestamps.
type WgCooldownReductionAttributor struct {
	BaseAttributor
	wgCasts    []int
	downstream []TalentAttributor
	has4pc     bool
}

func NewWgCooldownReductionAttributor(downstream []TalentAttributor, has4pc bool) *WgCooldownReductionAttributor {
	return &WgCooldownReductionAttributor{
		BaseAttributor: NewBaseAttributor("Early Spring (WG)", nil, nil),
		downstream:     downstream,
		has4pc:         has4pc,
	}
}

func (a *WgCooldownReductionAttributor) IsSelected() bool {
	if a.CombatantInfo == nil {
		return true
	}
	return a.CombatantInfo.TalentNodes[EarlySpringNodeID] && a.CombatantInfo.TalentIDs[EarlySpringTalentID]
}

func (a *WgCooldownReductionAttributor) WgCastTimestamps() []int { return a.wgCasts }

func (a *WgCooldownReductionAttributor) ProcessEvent(event models.Event, hot *tracking.HotTracker, buff *tracking.BuffTracker) {
	if ce, ok := event.(*models.CastEvent); ok && ce.Type == models.EventBeginCast && ce.AbilityID == WildGrowth {
		a.wgCasts = append(a.wgCasts, ce.Timestamp)
	}
}

func (a *WgCooldownReductionAttributor) Finalize() float64 {
	if len(a.wgCasts) < 2 {
		return 0.0
	}

	hasES := a.HasTalent(EarlySpringNodeID) && a.CombatantInfo != nil && a.CombatantInfo.TalentIDs[EarlySpringTalentID]

	unreducedCD := ComputeEffectiveWgCd(false, a.has4pc)
	reducedCD := ComputeEffectiveWgCd(hasES, a.has4pc)

	totalRatio := 0.0
	totalCasts := len(a.wgCasts)

	for i := 1; i < totalCasts; i++ {
		gap := a.wgCasts[i] - a.wgCasts[i-1]
		if float64(gap) <= reducedCD+onCooldownToleranceMS {
			ratio := 1 - (reducedCD / unreducedCD)
			totalRatio += math.Max(ratio, 0.0)
		}
	}

	if totalRatio == 0.0 {
		return 0.0
	}

	extraCastFraction := totalRatio / float64(totalCasts)
	downstreamTotal := 0.0
	for _, attr := range a.downstream {
		downstreamTotal += attr.GetTotalAttributed()
	}

	return extraCastFraction * downstreamTotal
}
