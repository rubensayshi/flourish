from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

WILD_GROWTH = 48438
BASE_TARGETS = 5
EXTRA_TARGETS = 2


class ImprovedWildGrowthAttributor(TalentAttributor):
    """Improved Wild Growth: adds 2 targets to Wild Growth (5->7).
    Attributes 2/7 of all WG healing as the extra target share."""

    name = "Improved Wild Growth"

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id != WILD_GROWTH:
            return 0.0
        total_targets = BASE_TARGETS + EXTRA_TARGETS
        return event.amount * EXTRA_TARGETS / total_targets
