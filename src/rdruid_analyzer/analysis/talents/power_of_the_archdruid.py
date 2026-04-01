from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import (
    HealEvent,
    CastEvent,
    ApplyBuffEvent,
    RemoveBuffEvent,
)
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

SOTF_BUFF = 114108
REJUV = 774
REGROWTH = 8936
POTA_TAG = "pota"
POTA_WINDOW_MS = 500


class PowerOfTheArchdruidAttributor(TalentAttributor):
    """Power of the Archdruid: SotF-empowered Rejuv/Regrowth spreads to 2 extra allies."""

    name = "Power of the Archdruid"
    talent_node_id = 82065

    def __init__(self):
        super().__init__()
        # Track the primary cast target from the Rejuv/Regrowth that consumed SotF
        self._primary_target: int | None = None
        self._primary_spell: int | None = None
        self._consume_timestamp: int | None = None
        # Track pending cast to identify the primary target
        self._pending_cast: tuple[int, int, int] | None = None  # (timestamp, target_id, spell_id)

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        # Track Rejuv/Regrowth casts as potential SotF consumers
        if isinstance(event, CastEvent) and event.ability_id in (REJUV, REGROWTH):
            self._pending_cast = (event.timestamp, event.target_id, event.ability_id)

        # SotF buff removal = SotF was consumed by the preceding cast
        if isinstance(event, RemoveBuffEvent) and event.ability_id == SOTF_BUFF:
            if self._pending_cast is not None:
                _, target_id, spell_id = self._pending_cast
                self._primary_target = target_id
                self._primary_spell = spell_id
                self._consume_timestamp = event.timestamp
                self._pending_cast = None

        # Watch for spread ApplyBuff events within the window
        if isinstance(event, ApplyBuffEvent) and event.ability_id in (REJUV, REGROWTH):
            if (
                self._consume_timestamp is not None
                and event.ability_id == self._primary_spell
                and event.target_id != self._primary_target
                and event.timestamp - self._consume_timestamp <= POTA_WINDOW_MS
            ):
                hot = hot_tracker.get(event.target_id, event.ability_id)
                if hot:
                    hot.tags.add(POTA_TAG)

            # Expire the window after enough time has passed
            if (
                self._consume_timestamp is not None
                and event.timestamp - self._consume_timestamp > POTA_WINDOW_MS
            ):
                self._primary_target = None
                self._primary_spell = None
                self._consume_timestamp = None

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id not in (REJUV, REGROWTH):
            return 0.0

        hot = hot_tracker.get(event.target_id, event.ability_id)
        if not hot or POTA_TAG not in hot.tags:
            return 0.0

        # 100% of healing from PotA-spread HoTs is attributed
        return float(event.amount)
