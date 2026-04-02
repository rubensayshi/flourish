from flourish.analysis.attributor import TalentAttributor
from flourish.models.events import HealEvent
from flourish.tracking.hot_tracker import HotTracker
from flourish.tracking.buff_tracker import BuffTracker

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
        if self.is_player_pet(event.source_id):
            return float(event.amount)
        return 0.0
