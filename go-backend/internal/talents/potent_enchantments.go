package talents

// PotentEnchantmentsAttributor attributes the extra healing from the extended
// Reforestation duration (10→16s). ReforestationAttributor feeds it healing
// that occurs in the 10-16s window via AddHealing.
type PotentEnchantmentsAttributor struct {
	BaseAttributor
	deferredHealing float64
}

func NewPotentEnchantmentsAttributor() *PotentEnchantmentsAttributor {
	return &PotentEnchantmentsAttributor{
		BaseAttributor: NewBaseAttributor("Potent Enchantments", intPtr(PotentEnchantmentsNode), intPtr(PotentEnchantmentsTalent)),
	}
}

// AddHealing is called by ReforestationAttributor to feed healing from the extended window.
func (a *PotentEnchantmentsAttributor) AddHealing(amount float64) {
	a.deferredHealing += amount
}

func (a *PotentEnchantmentsAttributor) Finalize() float64 {
	return a.deferredHealing
}
