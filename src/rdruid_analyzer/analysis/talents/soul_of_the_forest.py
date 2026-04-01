from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import (
    HealEvent,
    CastEvent,
    ApplyBuffEvent,
    RemoveBuffEvent,
)
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

SWIFTMEND = 18562
REJUV = 774
REGROWTH = 8936
SOTF_BUFF = 114108
SOTF_MULTIPLIER = 0.6  # +60%
SOTF_TAG = "sotf"
POTA_TAG = "pota"
POTA_WINDOW_MS = 500
POTA_NODE_ID = 82065


class SoulOfTheForestAttributor(TalentAttributor):
    """Combined SotF + Power of the Archdruid attributor.

    SotF: after Swiftmend, next Rejuv/Regrowth gets +60% healing.
    PotA: that SotF-empowered HoT spreads to 2 extra allies.

    Attribution:
    - Primary target: SotF bonus only (the 60% portion)
    - Spread targets: 100% of healing (wouldn't exist without PotA)
    """

    name = "SotF + Power of the Archdruid"
    talent_node_id = 82055  # SotF node

    def __init__(self):
        super().__init__()
        self._sotf_ready = False
        # PotA spread tracking
        self._primary_target: int | None = None
        self._primary_spell: int | None = None
        self._consume_timestamp: int | None = None
        self._pending_cast: tuple[int, int, int] | None = None  # (ts, target, spell)

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        # Track Swiftmend → SotF ready
        if isinstance(event, CastEvent) and event.ability_id == SWIFTMEND:
            self._sotf_ready = True

        # Track Rejuv/Regrowth casts as potential SotF consumers
        if isinstance(event, CastEvent) and event.ability_id in (REJUV, REGROWTH):
            self._pending_cast = (event.timestamp, event.target_id, event.ability_id)

        # SotF buff removal = consumed by the preceding cast → start PotA window
        if isinstance(event, RemoveBuffEvent) and event.ability_id == SOTF_BUFF:
            if self._pending_cast is not None:
                _, target_id, spell_id = self._pending_cast
                self._primary_target = target_id
                self._primary_spell = spell_id
                self._consume_timestamp = event.timestamp
                self._pending_cast = None

        # Tag HoTs on ApplyBuff
        if isinstance(event, ApplyBuffEvent) and event.ability_id in (REJUV, REGROWTH):
            hot = hot_tracker.get(event.target_id, event.ability_id)
            if not hot:
                return

            # Primary SotF target (via _sotf_ready flag)
            if self._sotf_ready:
                hot.tags.add(SOTF_TAG)
                self._sotf_ready = False

            # PotA spread targets (within window, different target, same spell)
            elif (
                self._consume_timestamp is not None
                and event.ability_id == self._primary_spell
                and event.target_id != self._primary_target
                and event.timestamp - self._consume_timestamp <= POTA_WINDOW_MS
            ):
                hot.tags.add(SOTF_TAG)
                hot.tags.add(POTA_TAG)

        # Expire PotA window
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
        if not hot or SOTF_TAG not in hot.tags:
            return 0.0

        # PotA spread: 100% attributed (HoT wouldn't exist without PotA)
        if POTA_TAG in hot.tags:
            return float(event.amount)

        # Primary SotF: only the bonus portion
        return event.amount - event.amount / (1 + SOTF_MULTIPLIER)
