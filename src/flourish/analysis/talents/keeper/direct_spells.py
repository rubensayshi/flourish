from flourish.analysis.talents.direct_spell import DirectSpellAttributor


class GroveGuardiansAttributor(DirectSpellAttributor):
    """Grove Guardians: Treant pets cast Nourish and direct heals."""
    name = "Grove Guardians"
    talent_node_id = 82043
    spell_ids = {422090, 142421}  # Nourish + direct heal (from treants)
    allow_pet_source = True


class DreamSurgeAttributor(DirectSpellAttributor):
    name = "Dream Surge"
    talent_node_id = 94600
    spell_ids = {434141}  # Dream Bloom


class SpiritOfTheThicketAttributor(DirectSpellAttributor):
    """Spirit of the Thicket: Ironbark summons a Dryad that channels a healing beam.
    The heal comes from a pet source, so we override the pet-source guard."""
    name = "Spirit of the Thicket"
    talent_node_id = 109712
    spell_ids = {1264905}
    allow_pet_source = True
