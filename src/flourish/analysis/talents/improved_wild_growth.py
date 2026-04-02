from flourish.analysis.attributor import TalentAttributor
from flourish.models.events import HealEvent
from flourish.tracking.hot_tracker import HotTracker
from flourish.tracking.buff_tracker import BuffTracker

WILD_GROWTH = 48438
TOL_BUFF = 33891
BASE_TARGETS = 5
EXTRA_TARGETS = 2


class ImprovedWildGrowthAttributor(TalentAttributor):
    """Improved Wild Growth: adds 2 targets to Wild Growth (5->7).
    Attributes 2/7 of all WG healing as the extra target share.
    Skips during Tree of Life — ToL already accounts for IWG base targets."""

    name = "Improved Wild Growth"
    talent_node_id = 82045

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id != WILD_GROWTH:
            return 0.0
        # ToL attributor handles IWG's extra targets in its own calculation
        if buff_tracker.is_active(TOL_BUFF):
            return 0.0
        total_targets = BASE_TARGETS + EXTRA_TARGETS
        return event.amount * EXTRA_TARGETS / total_targets
