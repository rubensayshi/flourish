from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent, ApplyBuffEvent, RemoveBuffEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

TOL_BUFF = 33891
REJUV_IDS = {774, 155777}
WILD_GROWTH = 48438
IWG_NODE = 82045  # Improved Wild Growth talent node
TICK_WINDOW_MS = 200


class TreeOfLifeAttributor(TalentAttributor):
    """Incarnation: Tree of Life increases healing by 10%, Rejuv by 50%,
    and adds 2 extra Wild Growth targets."""

    name = "Incarnation: Tree of Life"

    def __init__(self):
        super().__init__()
        self._tol_active = False
        self._wg_buffer: list[HealEvent] = []
        self._buffer_start = 0
        self._deferred_wg_healing = 0.0

    def _base_wg_targets(self) -> int:
        return 7 if self.has_talent(IWG_NODE) else 5

    def _flush_wg_buffer(self) -> float:
        if not self._wg_buffer:
            return 0.0
        targets = len({e.target_id for e in self._wg_buffer})
        total_healing = sum(e.amount for e in self._wg_buffer)
        base_targets = self._base_wg_targets()
        # 10% base buff on all ticks
        base_buff = total_healing - total_healing / 1.1
        # Extra target attribution: 2 extra targets on top of base
        if targets > base_targets:
            extra_share = total_healing * (targets - base_targets) / targets
        else:
            extra_share = 0.0
        self._wg_buffer.clear()
        return base_buff + extra_share

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        if isinstance(event, ApplyBuffEvent) and event.ability_id == TOL_BUFF:
            self._tol_active = True
        elif isinstance(event, RemoveBuffEvent) and event.ability_id == TOL_BUFF:
            self._tol_active = False
            # Flush pending WG buffer when ToL ends
            self._deferred_wg_healing += self._flush_wg_buffer()

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if not self._tol_active:
            return 0.0

        if event.ability_id in REJUV_IDS:
            return event.amount - event.amount / 1.5

        if event.ability_id == WILD_GROWTH:
            flushed = 0.0
            if self._wg_buffer and event.timestamp - self._buffer_start > TICK_WINDOW_MS:
                flushed = self._flush_wg_buffer()
            if not self._wg_buffer:
                self._buffer_start = event.timestamp
            self._wg_buffer.append(event)
            return flushed

        # All other heals: +10%
        return event.amount - event.amount / 1.1

    def finalize(self) -> float:
        return self._deferred_wg_healing + self._flush_wg_buffer()
