from __future__ import annotations

from flourish.models.events import CombatantInfoEvent, HealEvent
from flourish.tracking.hot_tracker import HotTracker
from flourish.tracking.buff_tracker import BuffTracker


class TalentAttributor:
    name: str = "Unknown"
    talent_node_id: int | None = None  # Blizzard talent tree nodeID
    talent_id: int | None = None  # WCL entryId (for choice-node disambiguation)

    def __init__(self):
        self.combatant_info: CombatantInfoEvent | None = None
        self.total_attributed: float = 0.0

    def set_combatant_info(self, info: CombatantInfoEvent):
        """Called once at fight start with player's talent/stat info."""
        self.combatant_info = info

    def is_selected(self) -> bool:
        """Check if this talent is in the player's loadout."""
        if self.combatant_info is None:
            return True  # no info yet, assume active
        if self.talent_node_id is None:
            return True  # no node configured, always active
        if self.talent_node_id not in self.combatant_info.talent_nodes:
            return False
        # For choice nodes, also check talent_id (WCL entryId) if set
        if self.talent_id is not None:
            return self.talent_id in self.combatant_info.talent_ids
        return True

    def has_talent(self, node_id: int) -> bool:
        """Check if a specific talent node is active."""
        if self.combatant_info:
            return node_id in self.combatant_info.talent_nodes
        return False

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        """Called for every event — use to update internal state."""
        pass

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        """Called for non-wasted heal events. Return attributed healing amount."""
        return 0.0

    def finalize(self) -> float:
        """Called after all events processed. Return any additional attributed healing."""
        return 0.0
