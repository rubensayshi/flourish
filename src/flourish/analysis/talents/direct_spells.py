from flourish.analysis.attributor import TalentAttributor
from flourish.analysis.talents.direct_spell import DirectSpellAttributor


class EverbloomAttributor(DirectSpellAttributor):
    name = "Everbloom"
    talent_node_id = 110424
    spell_ids = {1244341}


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


class RampantGrowthAttributor(TalentAttributor):
    """Rampant Growth: +100% Regrowth HoT healing.
    Credits the bonus portion of periodic ticks only (direct heals unaffected).
    The extra Regrowth on the LB target uses spell 8936 in WCL (indistinguishable)."""
    name = "Rampant Growth"
    talent_node_id = 82058
    REGROWTH = 8936
    MULTIPLIER = 1.0  # +100%

    def process_heal(self, event, hot_tracker, buff_tracker) -> float:
        if event.ability_id == self.REGROWTH and event.tick:
            return event.amount - event.amount / (1 + self.MULTIPLIER)
        return 0.0



class FlourishAttributor(DirectSpellAttributor):
    """Flourish: Tranquility extends HoTs. The extended Tranq ticks
    appear under a different spell ID."""
    name = "Flourish"
    talent_node_id = 82053
    talent_id = 103106  # WCL entryId (choice node vs Inner Peace)
    spell_ids = {1264659}  # Tranquility (Flourish-modified)


class ThrivingVegetationAttributor(DirectSpellAttributor):
    """Thriving Vegetation: Rejuvenation instantly heals for 15/30% of its total periodic effect."""
    name = "Thriving Vegetation"
    talent_node_id = 82068
    spell_ids = {447132}
