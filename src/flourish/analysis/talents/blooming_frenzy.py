from flourish.analysis.attributor import TalentAttributor
from flourish.models.events import HealEvent, RemoveBuffEvent
from flourish.tracking.hot_tracker import HotTracker
from flourish.tracking.buff_tracker import BuffTracker

SOTF_BUFF = 114108
LB_BLOOM = 33778
FRENZY_WINDOW_MS = 1500
FRENZY_MAX_BLOOMS = 5


class BloomingFrenzyAttributor(TalentAttributor):
    """Everbloom rank 4: Lifebloom blooms 5 times rapidly when SotF is consumed."""

    name = "Blooming Frenzy"
    talent_node_id = 110424

    def __init__(self):
        super().__init__()
        self._frenzy_start: int | None = None
        self._frenzy_count: int = 0

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        if isinstance(event, RemoveBuffEvent) and event.ability_id == SOTF_BUFF:
            self._frenzy_start = event.timestamp
            self._frenzy_count = 0

        # Expire window
        if self._frenzy_start is not None:
            if event.timestamp - self._frenzy_start > FRENZY_WINDOW_MS:
                self._frenzy_start = None
                self._frenzy_count = 0

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id != LB_BLOOM:
            return 0.0
        if self._frenzy_start is None:
            return 0.0
        if event.timestamp - self._frenzy_start > FRENZY_WINDOW_MS:
            return 0.0
        if self._frenzy_count >= FRENZY_MAX_BLOOMS:
            return 0.0

        self._frenzy_count += 1
        return float(event.amount)
