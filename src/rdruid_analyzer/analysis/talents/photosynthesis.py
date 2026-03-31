from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent, RemoveBuffEvent, RefreshBuffEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

LIFEBLOOM = 33763
LIFEBLOOM_BLOOM = 33778


class PhotosynthesisAttributor(TalentAttributor):
    """Photosynthesis: attribute LB blooms that aren't from expiry/refresh.
    Total blooms - explained blooms (expiry/refresh) = Photosynthesis procs.
    Attribution is deferred to finalize() since we need to look ahead."""

    name = "Photosynthesis"
    talent_node_id = 82073

    def __init__(self):
        super().__init__()
        self._bloom_events: list[tuple[int, int, float]] = []  # (timestamp, target_id, amount)
        self._explained_blooms: set[tuple[int, int]] = set()  # (timestamp, target_id)

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        # Mark blooms as explained if followed by LB remove or refresh
        if isinstance(event, (RemoveBuffEvent, RefreshBuffEvent)) and event.ability_id == LIFEBLOOM:
            for ts, target_id, _ in self._bloom_events:
                if target_id == event.target_id and event.timestamp - ts < 200:
                    self._explained_blooms.add((ts, target_id))

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id == LIFEBLOOM_BLOOM:
            self._bloom_events.append((event.timestamp, event.target_id, event.amount))
        return 0.0

    def finalize(self) -> float:
        total = 0.0
        for ts, target_id, amount in self._bloom_events:
            if (ts, target_id) not in self._explained_blooms:
                total += amount
        return total
