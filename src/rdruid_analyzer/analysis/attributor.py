from __future__ import annotations

from rdruid_analyzer.models.events import CombatantInfoEvent, HealEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker


class TalentAttributor:
    name: str = "Unknown"

    def set_combatant_info(self, info: CombatantInfoEvent):
        """Called once at fight start with player's talent/stat info."""
        self.combatant_info = info

    def has_talent(self, node_id: int) -> bool:
        """Check if a specific talent node is active."""
        if hasattr(self, "combatant_info") and self.combatant_info:
            return node_id in self.combatant_info.talent_nodes
        return False

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        """Called for every event — use to update internal state."""
        pass

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        """Called for non-wasted heal events. Return attributed healing amount."""
        return 0.0
