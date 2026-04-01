from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

# Spells cast by the Sylvan Beckoning Dryad
DRYAD_TRANQ = 1264659
DRYAD_REGROWTH = 1264664
DRYAD_SPELLS = {DRYAD_TRANQ, DRYAD_REGROWTH}


class SylvanBeckoningAttributor(TalentAttributor):
    """Sylvan Beckoning: periodic heals empower Swiftmend to summon a Dryad.
    The Dryad casts Tranquility and Regrowth, logged as pet healing events
    with the same spell IDs as Flourish Tranq and Rampant Growth Regrowth.
    We distinguish by checking source_id != player source_id."""

    name = "Sylvan Beckoning"
    talent_node_id = 109714

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id not in DRYAD_SPELLS:
            return 0.0
        # If source is not the player, it's from the Dryad pet
        if self.combatant_info and event.source_id != self.combatant_info.source_id:
            return float(event.amount)
        return 0.0
