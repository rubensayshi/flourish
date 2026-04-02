from flourish.analysis.attributor import TalentAttributor
from flourish.models.events import (
    HealEvent,
    CastEvent,
    ApplyBuffEvent,
    RefreshBuffEvent,
    RemoveBuffEvent,
)
from flourish.tracking.hot_tracker import HotTracker
from flourish.tracking.buff_tracker import BuffTracker

SWIFTMEND = 18562
REJUV = 774
GERMINATION_REJUV = 155777
REJUV_IDS = {REJUV, GERMINATION_REJUV}
REGROWTH = 8936
SOTF_SPELL_IDS = REJUV_IDS | {REGROWTH}
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
    - Regrowth direct heal: SotF bonus (fires before HoT is tagged)

    We use RemoveBuffEvent(SOTF_BUFF) as the definitive signal that SotF was
    consumed, then retroactively tag the primary HoT. This avoids false
    consumption by interfering events (Dream Surge procs, etc.).
    """

    name = "SotF + PotA"
    talent_node_id = 82055  # SotF node

    def __init__(self):
        super().__init__()
        # PotA spread tracking
        self._primary_target: int | None = None
        self._primary_spell: int | None = None
        self._consume_timestamp: int | None = None
        self._pending_cast: tuple[int, int, int] | None = None  # (ts, target, spell)

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        # Track Rejuv/Regrowth casts as potential SotF consumers
        if isinstance(event, CastEvent) and event.ability_id in (REJUV, REGROWTH):
            self._pending_cast = (event.timestamp, event.target_id, event.ability_id)

        # SotF buff removal = consumed → retroactively tag the primary HoT + start PotA window
        if isinstance(event, RemoveBuffEvent) and event.ability_id == SOTF_BUFF:
            if self._pending_cast is not None:
                _, target_id, spell_id = self._pending_cast
                self._primary_target = target_id
                self._primary_spell = spell_id
                self._consume_timestamp = event.timestamp
                self._pending_cast = None

                # Tag the primary HoT (cast may be Rejuv 774 but applied as Germination 155777)
                if spell_id in REJUV_IDS:
                    for sid in REJUV_IDS:
                        hot = hot_tracker.get(target_id, sid)
                        if hot and SOTF_TAG not in hot.tags:
                            hot.tags.add(SOTF_TAG)
                            break
                else:
                    hot = hot_tracker.get(target_id, spell_id)
                    if hot:
                        hot.tags.add(SOTF_TAG)

        # Tag PotA spread HoTs (within window, different target)
        if isinstance(event, (ApplyBuffEvent, RefreshBuffEvent)) and event.ability_id in SOTF_SPELL_IDS:
            if (
                self._consume_timestamp is not None
                and event.target_id != self._primary_target
                and (event.ability_id in REJUV_IDS if self._primary_spell in REJUV_IDS else event.ability_id == self._primary_spell)
                and event.timestamp - self._consume_timestamp <= POTA_WINDOW_MS
            ):
                hot = hot_tracker.get(event.target_id, event.ability_id)
                if hot:
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
        if event.ability_id not in SOTF_SPELL_IDS:
            return 0.0

        # Regrowth direct heal: fires before HoT is tagged (before RemoveBuffEvent).
        # Use buff_tracker as source of truth — SotF buff is still active during the direct heal.
        if (
            event.ability_id == REGROWTH
            and buff_tracker.is_active(SOTF_BUFF)
            and self._pending_cast is not None
            and event.target_id == self._pending_cast[1]
            and self._pending_cast[2] == REGROWTH
        ):
            return event.amount - event.amount / (1 + SOTF_MULTIPLIER)

        hot = hot_tracker.get(event.target_id, event.ability_id)
        if not hot or SOTF_TAG not in hot.tags:
            return 0.0

        # PotA spread: 100% attributed (HoT wouldn't exist without PotA)
        if POTA_TAG in hot.tags:
            return float(event.amount)

        # Primary SotF: only the bonus portion
        return event.amount - event.amount / (1 + SOTF_MULTIPLIER)
