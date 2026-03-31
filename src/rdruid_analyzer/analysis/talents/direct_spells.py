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
