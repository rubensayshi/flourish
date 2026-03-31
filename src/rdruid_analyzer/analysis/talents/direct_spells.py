from rdruid_analyzer.analysis.talents.direct_spell import DirectSpellAttributor


class EverbloomAttributor(DirectSpellAttributor):
    name = "Everbloom"
    spell_ids = {1244341}


class GroveGuardiansAttributor(DirectSpellAttributor):
    name = "Grove Guardians"
    spell_ids = {422090}  # Nourish (from treants)


class DreamSurgeAttributor(DirectSpellAttributor):
    name = "Dream Surge"
    spell_ids = {434141}  # Dream Bloom


class EfflorescenceAttributor(DirectSpellAttributor):
    name = "Efflorescence"
    spell_ids = {81269}


class VerdancyAttributor(DirectSpellAttributor):
    name = "Verdancy"
    spell_ids = {392329}


class NaturesBountyAttributor(DirectSpellAttributor):
    name = "Nature's Bounty"
    spell_ids = {1264376}


class RegenerativeHeartwoodAttributor(DirectSpellAttributor):
    name = "Regenerative Heartwood"
    spell_ids = {392117}


class CultivationAttributor(DirectSpellAttributor):
    name = "Cultivation"
    spell_ids = {200390}


class YserasGiftAttributor(DirectSpellAttributor):
    name = "Ysera's Gift"
    spell_ids = {145108, 145109, 145110}


class EmbraceOfTheDreamAttributor(DirectSpellAttributor):
    name = "Embrace of the Dream"
    spell_ids = {392124}


class RampantGrowthAttributor(DirectSpellAttributor):
    """Rampant Growth causes Regrowth to also apply to Lifebloom target.
    The extra Regrowth has its own spell ID in WCL."""
    name = "Rampant Growth"
    spell_ids = {1264664}  # Regrowth (from Rampant Growth)


class ImprovedSwiftmendAttributor(DirectSpellAttributor):
    """Improved Swiftmend: +30% Swiftmend healing.
    WCL logs the bonus portion under a separate spell ID."""
    name = "Improved Swiftmend"
    spell_ids = {142421}  # Swiftmend (bonus from Improved Swiftmend)


class FlourishAttributor(DirectSpellAttributor):
    """Flourish: Tranquility extends HoTs. The extended Tranq ticks
    appear under a different spell ID."""
    name = "Flourish"
    spell_ids = {1264659}  # Tranquility (Flourish-modified)
