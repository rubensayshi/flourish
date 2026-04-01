from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

REJUV_IDS = {774, 155777}  # Rejuv + Germination
BASE_REJUV_DURATION_MS = 17000  # 12s base + 3s Lingering Healing + 2s Germination


class NurturingDormancyAttributor(TalentAttributor):
    """Nurturing Dormancy: Rejuv on full-health targets gets extended.
    Attribute Rejuv ticks past the base 12s duration."""

    name = "Nurturing Dormancy"
    talent_node_id = 82076

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id not in REJUV_IDS:
            return 0.0
        hot = hot_tracker.get(event.target_id, event.ability_id)
        if not hot:
            return 0.0
        # Use the most recent application/refresh as the base
        base_time = hot.last_refresh if hot.last_refresh > 0 else hot.applied_at
        elapsed = event.timestamp - base_time
        if elapsed > BASE_REJUV_DURATION_MS:
            return float(event.amount)
        return 0.0
