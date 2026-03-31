from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent, CastEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

SWIFTMEND = 18562
REFORESTATION_TOL_DURATION_MS = 10000
REJUV_IDS = {774, 155777}


class ReforestationAttributor(TalentAttributor):
    """Reforestation: every 4th Swiftmend triggers a 10 sec Tree of Life.
    During this window, applies ToL healing multipliers."""

    name = "Reforestation"

    def __init__(self):
        self._sm_count = 0
        self._tol_end = 0

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        if isinstance(event, CastEvent) and event.ability_id == SWIFTMEND:
            self._sm_count += 1
            if self._sm_count % 4 == 0:
                self._tol_end = event.timestamp + REFORESTATION_TOL_DURATION_MS

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.timestamp > self._tol_end:
            return 0.0
        if event.ability_id in REJUV_IDS:
            return event.amount - event.amount / 1.5
        return event.amount - event.amount / 1.1
