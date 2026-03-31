from rdruid_analyzer.analysis.talents.direct_spell import DirectSpellAttributor


class EverbloomAttributor(DirectSpellAttributor):
    name = "Everbloom"
    talent_node_id = 110424
    spell_ids = {1244341}


class GroveGuardiansAttributor(DirectSpellAttributor):
    name = "Grove Guardians"
    talent_node_id = 82043
    spell_ids = {422090, 142421}  # Nourish + direct heal (from treants)


class DreamSurgeAttributor(DirectSpellAttributor):
    name = "Dream Surge"
    talent_node_id = 94600
    spell_ids = {434141}  # Dream Bloom


class EfflorescenceAttributor(DirectSpellAttributor):
    name = "Efflorescence"
    talent_node_id = 82057
    spell_ids = {81269}


class VerdancyAttributor(DirectSpellAttributor):
    name = "Verdancy"
    talent_node_id = 82059
    spell_ids = {392329}


class NaturesBountyAttributor(DirectSpellAttributor):
    name = "Nature's Bounty"
    talent_node_id = 82072
    spell_ids = {1264376}


class RegenerativeHeartwoodAttributor(DirectSpellAttributor):
    name = "Regenerative Heartwood"
    talent_node_id = 82075
    spell_ids = {392117}


class CultivationAttributor(DirectSpellAttributor):
    name = "Cultivation"
    talent_node_id = 82056
    spell_ids = {200390}


class YserasGiftAttributor(DirectSpellAttributor):
    name = "Ysera's Gift"
    talent_node_id = 82048
    spell_ids = {145108, 145109, 145110}


class EmbraceOfTheDreamAttributor(DirectSpellAttributor):
    name = "Embrace of the Dream"
    talent_node_id = 82070
    spell_ids = {392124}


class RampantGrowthAttributor(DirectSpellAttributor):
    """Rampant Growth causes Regrowth to also apply to Lifebloom target.
    The extra Regrowth has its own spell ID in WCL."""
    name = "Rampant Growth"
    talent_node_id = 82058
    spell_ids = {1264664}  # Regrowth (from Rampant Growth)


class ImprovedSwiftmendAttributor(DirectSpellAttributor):
    """Improved Swiftmend: +30% Swiftmend healing.
    WCL logs the bonus portion under a separate spell ID."""
    name = "Improved Swiftmend"
    talent_node_id = 103873
    spell_ids = {142421}  # Swiftmend (bonus from Improved Swiftmend)


class FlourishAttributor(DirectSpellAttributor):
    """Flourish: Tranquility extends HoTs. The extended Tranq ticks
    appear under a different spell ID."""
    name = "Flourish"
    talent_node_id = 82053
    talent_id = 108111
    spell_ids = {1264659}  # Tranquility (Flourish-modified)


class BurstingGrowthAttributor(DirectSpellAttributor):
    """Bursting Growth: AoE heal when Symbiotic Blooms expire or Rejuv on bloom target."""
    name = "Bursting Growth"
    talent_node_id = 109716
    spell_ids = {440120}
