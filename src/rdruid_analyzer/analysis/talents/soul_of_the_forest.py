from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent, CastEvent, ApplyBuffEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

SWIFTMEND = 18562
REJUV = 774
REGROWTH = 8936
SOTF_MULTIPLIER = 0.6  # +60%
SOTF_TAG = "sotf"


class SoulOfTheForestAttributor(TalentAttributor):
    name = "Soul of the Forest"

    def __init__(self):
        super().__init__()
        self._sotf_ready = False

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        if isinstance(event, CastEvent) and event.ability_id == SWIFTMEND:
            self._sotf_ready = True

        elif isinstance(event, ApplyBuffEvent) and event.ability_id in (REJUV, REGROWTH):
            if self._sotf_ready:
                hot = hot_tracker.get(event.target_id, event.ability_id)
                if hot:
                    hot.tags.add(SOTF_TAG)
                self._sotf_ready = False

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id not in (REJUV, REGROWTH):
            return 0.0

        hot = hot_tracker.get(event.target_id, event.ability_id)
        # NOTE: SotF tag persists through Rejuv refresh. This may or may not be
        # correct — if the game resets the buff on refresh, we're over-attributing.
        # See docs/ingame-testing.md for verification plan.
        if not hot or SOTF_TAG not in hot.tags:
            return 0.0

        return event.amount - event.amount / (1 + SOTF_MULTIPLIER)
