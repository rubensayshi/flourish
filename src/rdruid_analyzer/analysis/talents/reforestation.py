from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent, CastEvent, ApplyBuffEvent, RemoveBuffEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

SWIFTMEND = 18562
TOL_BUFF = 33891
REFORESTATION_TOL_DURATION_MS = 10000
POTENT_ENCHANTMENTS_TOL_DURATION_MS = 16000
POTENT_ENCHANTMENTS_NODE_ID = 94595
POTENT_ENCHANTMENTS_ENTRY_ID = 117188  # WCL entryId
REJUV_IDS = {774, 155777}


class ReforestationAttributor(TalentAttributor):
    """Reforestation: every 4th Swiftmend triggers a 10 sec Tree of Life.
    During this window, applies ToL healing multipliers.
    Does NOT claim credit if a real (player-cast) ToL is already active."""

    name = "Reforestation"
    talent_node_id = 82069

    def __init__(self):
        super().__init__()
        self._sm_count = 0
        self._reforestation_tol_end = 0
        self._real_tol_active = False

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        # Track real ToL buff
        if isinstance(event, ApplyBuffEvent) and event.ability_id == TOL_BUFF:
            self._real_tol_active = True
        elif isinstance(event, RemoveBuffEvent) and event.ability_id == TOL_BUFF:
            self._real_tol_active = False

        if isinstance(event, CastEvent) and event.ability_id == SWIFTMEND:
            self._sm_count += 1
            if self._sm_count % 4 == 0 and not self._real_tol_active:
                duration = self._get_tol_duration()
                self._reforestation_tol_end = event.timestamp + duration

    def _get_tol_duration(self) -> int:
        """Return ToL duration, extended if Potent Enchantments is selected."""
        if (self.combatant_info
                and POTENT_ENCHANTMENTS_NODE_ID in self.combatant_info.talent_nodes
                and POTENT_ENCHANTMENTS_ENTRY_ID in self.combatant_info.talent_ids):
            return POTENT_ENCHANTMENTS_TOL_DURATION_MS
        return REFORESTATION_TOL_DURATION_MS

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if self._real_tol_active:
            return 0.0  # Real ToL is handling this
        if event.timestamp > self._reforestation_tol_end:
            return 0.0
        if event.ability_id in REJUV_IDS:
            return event.amount - event.amount / 1.5
        return event.amount - event.amount / 1.1
