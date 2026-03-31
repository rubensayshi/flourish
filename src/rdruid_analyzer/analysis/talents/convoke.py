from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent, CastEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

CONVOKE = 391528
CONVOKE_DURATION_MS = 4000
DEFAULT_HEALING_RATIO = 0.7


class ConvokeAttributor(TalentAttributor):
    """Convoke the Spirits: 4 sec channel firing 16 spells.
    Attributes a configurable ratio of all healing during the channel window."""

    name = "Convoke the Spirits"

    def __init__(self, healing_ratio: float = DEFAULT_HEALING_RATIO):
        self._channel_end = 0
        self._healing_ratio = healing_ratio

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        if isinstance(event, CastEvent) and event.ability_id == CONVOKE:
            self._channel_end = event.timestamp + CONVOKE_DURATION_MS

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.timestamp <= self._channel_end:
            return event.amount * self._healing_ratio
        return 0.0
