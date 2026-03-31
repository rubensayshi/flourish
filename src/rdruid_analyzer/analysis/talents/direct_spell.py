from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker


class DirectSpellAttributor(TalentAttributor):
    """Attributes all effective healing from specific spell IDs to this talent."""

    name: str = "Unknown"
    spell_ids: set[int] = set()

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id in self.spell_ids:
            return float(event.amount)
        return 0.0
