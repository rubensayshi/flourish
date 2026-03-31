from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent, CastEvent, ApplyBuffEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

CONVOKE = 391528
CONVOKE_DURATION_MS = 4000
DEFAULT_HEALING_RATIO = 0.7
CONVOKE_TAG = "convoke"


class ConvokeAttributor(TalentAttributor):
    """Convoke the Spirits: 4 sec channel firing 16 spells.
    Attributes healing from spells actually cast during the channel:
    - HoTs applied during channel are tagged; their ticks are attributed.
    - Direct heals during channel (no tracked HoT) are attributed.
    Pre-existing HoT ticks during the channel are NOT attributed."""

    name = "Convoke the Spirits"
    talent_node_id = 82064
    talent_id = 108124

    def __init__(self, healing_ratio: float = DEFAULT_HEALING_RATIO):
        super().__init__()
        self._channel_end = 0
        self._healing_ratio = healing_ratio

    def _is_channeling(self, timestamp: int) -> bool:
        return 0 < self._channel_end >= timestamp

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        if isinstance(event, CastEvent) and event.ability_id == CONVOKE:
            self._channel_end = event.timestamp + CONVOKE_DURATION_MS

        # Tag HoTs applied during Convoke channel
        if self._is_channeling(event.timestamp) and isinstance(event, ApplyBuffEvent):
            hot = hot_tracker.get(event.target_id, event.ability_id)
            if hot:
                hot.tags.add(CONVOKE_TAG)

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        hot = hot_tracker.get(event.target_id, event.ability_id)

        # Convoke-tagged HoT tick (during or after channel)
        if hot and CONVOKE_TAG in hot.tags:
            return event.amount * self._healing_ratio

        # Direct heal during channel (no tracked HoT, e.g. Swiftmend)
        if self._is_channeling(event.timestamp) and hot is None:
            return event.amount * self._healing_ratio

        return 0.0
