from flourish.analysis.attributor import TalentAttributor
from flourish.models.events import HealEvent, CastEvent, ApplyBuffEvent
from flourish.tracking.hot_tracker import HotTracker
from flourish.tracking.buff_tracker import BuffTracker

CONVOKE_SPELL_IDS = {391528, 323764}  # talent-tree ID + legacy (SL covenant) ID
CONVOKE_DURATION_MS = 4000
CONVOKE_DURATION_CG_MS = 3000  # with Cenarius' Guidance (-25%)
CENARIUS_GUIDANCE_NODE = 82063
DEFAULT_HEALING_RATIO = 0.7
CONVOKE_TAG = "convoke"


class ConvokeAttributor(TalentAttributor):
    """Convoke the Spirits: 4 sec channel firing 16 spells.
    Attributes healing from spells actually cast during the channel:
    - HoTs applied during channel are tagged; their ticks are attributed.
    - Direct heals during channel (no tracked HoT) are attributed.
    Pre-existing HoT ticks during the channel are NOT attributed.
    Duration is reduced to 3s with Cenarius' Guidance."""

    name = "Convoke the Spirits"
    talent_node_id = 82064
    talent_id = 103119  # WCL entryId (choice node vs Tree of Life)

    def __init__(self, healing_ratio: float = DEFAULT_HEALING_RATIO):
        super().__init__()
        self._channel_end = 0
        self._healing_ratio = healing_ratio

    def _channel_duration(self) -> int:
        if self.has_talent(CENARIUS_GUIDANCE_NODE):
            return CONVOKE_DURATION_CG_MS
        return CONVOKE_DURATION_MS

    def _is_channeling(self, timestamp: int) -> bool:
        return self._channel_end > 0 and timestamp <= self._channel_end

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        if isinstance(event, CastEvent) and event.ability_id in CONVOKE_SPELL_IDS:
            self._channel_end = event.timestamp + self._channel_duration()

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
