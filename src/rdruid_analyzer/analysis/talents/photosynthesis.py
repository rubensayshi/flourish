from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import CastEvent, HealEvent, RemoveBuffEvent, RefreshBuffEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

LIFEBLOOM = 33763
LIFEBLOOM_BLOOM = 33778
SOTF_BUFF = 114108
WINDOW_MS = 200
EVERBLOOM_WINDOW_MS = 1200  # Everbloom produces 5 blooms over ~1000ms after SotF


class PhotosynthesisAttributor(TalentAttributor):
    """Photosynthesis: attribute LB blooms not caused by expiry/refresh/Everbloom.

    A bloom is "explained" (not a photo proc) if:
    1. A RemoveBuffEvent or RefreshBuffEvent for LB on the same target is
       within WINDOW_MS in either direction (natural expiry / pandemic refresh).
    2. A CastEvent for LB on the same target is within WINDOW_MS
       (LB recast — WCL sometimes omits the RefreshBuffEvent).
    3. A SotF consumption (RemoveBuffEvent for 114108) occurred within
       EVERBLOOM_WINDOW_MS before the bloom (Everbloom triggers 5 rapid blooms
       when SotF is consumed).

    Attribution is deferred to finalize() since we need the full event stream."""

    name = "Photosynthesis"
    talent_node_id = 82073

    def __init__(self):
        super().__init__()
        self._bloom_events: list[tuple[int, int, float]] = []  # (timestamp, target_id, amount)
        self._lb_transitions: list[tuple[int, int]] = []  # (timestamp, target_id) for remove/refresh
        self._lb_casts: list[tuple[int, int]] = []  # (timestamp, target_id) for LB casts
        self._sotf_consumptions: list[int] = []  # timestamps of SotF removal

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        if isinstance(event, (RemoveBuffEvent, RefreshBuffEvent)) and event.ability_id == LIFEBLOOM:
            self._lb_transitions.append((event.timestamp, event.target_id))
        elif isinstance(event, RemoveBuffEvent) and event.ability_id == SOTF_BUFF:
            self._sotf_consumptions.append(event.timestamp)
        elif isinstance(event, CastEvent) and event.ability_id == LIFEBLOOM:
            self._lb_casts.append((event.timestamp, event.target_id))

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id == LIFEBLOOM_BLOOM:
            self._bloom_events.append((event.timestamp, event.target_id, event.amount))
        return 0.0

    def finalize(self) -> float:
        total = 0.0
        for ts, target_id, amount in self._bloom_events:
            # 1. Explained by natural expiry or pandemic refresh?
            explained = any(
                tid == target_id and abs(tts - ts) < WINDOW_MS
                for tts, tid in self._lb_transitions
            )
            # 2. Explained by LB recast (CastEvent on same target)?
            if not explained:
                explained = any(
                    tid == target_id and abs(tts - ts) < WINDOW_MS
                    for tts, tid in self._lb_casts
                )
            # 3. Explained by Everbloom (SotF consumed within window before bloom)?
            if not explained:
                explained = any(
                    0 <= ts - sts <= EVERBLOOM_WINDOW_MS
                    for sts in self._sotf_consumptions
                )
            if not explained:
                total += amount
        return total
