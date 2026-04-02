from flourish.analysis.talents.direct_spell import DirectSpellAttributor


class BurstingGrowthAttributor(DirectSpellAttributor):
    """Bursting Growth: AoE heal when Symbiotic Blooms expire or Rejuv on bloom target."""
    name = "Bursting Growth"
    talent_node_id = 109716
    spell_ids = {440121}


class ThrivingGrowthAttributor(DirectSpellAttributor):
    """Thriving Growth: Wild Growth/Regrowth/Efflorescence can proc Symbiotic Blooms."""
    name = "Thriving Growth"
    talent_node_id = 94626
    spell_ids = {474760}  # Symbiotic Bloom healing (logged as "Symbiotic Relationship")
