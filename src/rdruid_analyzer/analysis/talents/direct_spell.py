from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker


class DirectSpellAttributor(TalentAttributor):
    """Attributes all effective healing from specific spell IDs to this talent."""

    name: str = "Unknown"
    spell_ids: set[int] = set()
    allow_pet_source: bool = False

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id in self.spell_ids:
            # Skip pet healing unless explicitly allowed (e.g. Spirit of the Thicket Dryad)
            if (not self.allow_pet_source
                    and self.combatant_info
                    and event.source_id != self.combatant_info.source_id):
                return 0.0
            return float(event.amount)
        return 0.0
