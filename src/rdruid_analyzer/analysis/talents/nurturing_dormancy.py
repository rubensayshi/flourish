from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

REJUV = 774
BASE_REJUV_DURATION_MS = 12000  # 12 sec base


class NurturingDormancyAttributor(TalentAttributor):
    """Nurturing Dormancy: Rejuv on full-health targets gets extended.
    Attribute Rejuv ticks past the base 12s duration."""

    name = "Nurturing Dormancy"

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id != REJUV:
            return 0.0
        hot = hot_tracker.get(event.target_id, REJUV)
        if not hot:
            return 0.0
        # Check if this tick is past the base duration
        elapsed = event.timestamp - hot.applied_at
        if elapsed > BASE_REJUV_DURATION_MS:
            return float(event.amount)
        return 0.0
