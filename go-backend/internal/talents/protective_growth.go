package talents

const pgDRFraction = 0.08

type ProtectiveGrowthAttributor struct {
	BaseAttributor
	damageTaken int
	hasData     bool
}

func NewProtectiveGrowthAttributor(damageTaken int) *ProtectiveGrowthAttributor {
	return &ProtectiveGrowthAttributor{
		BaseAttributor: NewBaseAttributor("Protective Growth", intPtr(94593), nil),
		damageTaken:    damageTaken,
		hasData:        true,
	}
}

func NewProtectiveGrowthAttributorNil() *ProtectiveGrowthAttributor {
	return &ProtectiveGrowthAttributor{
		BaseAttributor: NewBaseAttributor("Protective Growth", intPtr(94593), nil),
		hasData:        false,
	}
}

func (a *ProtectiveGrowthAttributor) Finalize() float64 {
	if !a.hasData || a.damageTaken <= 0 {
		return 0.0
	}
	return float64(a.damageTaken) * pgDRFraction / (1 - pgDRFraction)
}
