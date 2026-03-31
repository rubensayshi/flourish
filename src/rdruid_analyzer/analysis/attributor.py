from rdruid_analyzer.models.events import HealEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker


class TalentAttributor:
    name: str = "Unknown"

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        """Called for every event — use to update internal state."""
        pass

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        """Called for non-wasted heal events. Return attributed healing amount."""
        return 0.0
