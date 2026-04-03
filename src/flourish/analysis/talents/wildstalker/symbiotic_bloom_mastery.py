from flourish.analysis.attributor import TalentAttributor
from flourish.models.events import CombatantInfoEvent, HealEvent
from flourish.tracking.hot_tracker import HotTracker
from flourish.tracking.buff_tracker import BuffTracker

SYMBIOTIC_BLOOM = 439530

# Default DR table: cumulative mastery multiplier per stack count
DEFAULT_DR_TABLE = [1.0, 1.7, 2.3, 2.8, 3.2]


class SymbioticBloomMasteryAttributor(TalentAttributor):
    """Attributes the marginal mastery bonus from Symbiotic Bloom adding an extra HoT stack."""

    name = "Symbiotic Bloom Mastery"
    # Thriving Growth — the talent that creates Symbiotic Blooms
    talent_node_id = 94626

    def __init__(self, base_stacks: int = 2,
                 dr_table: list[float] | None = None):
        super().__init__()
        self._mastery = 0.25  # default, overridden by combatantinfo
        self._base_stacks = base_stacks
        self._dr_table = dr_table or DEFAULT_DR_TABLE
        self._fraction = self._compute_fraction()

    def set_combatant_info(self, info: CombatantInfoEvent):
        super().set_combatant_info(info)
        if info.mastery > 0:
            self._mastery = info.mastery / 100.0
            self._fraction = self._compute_fraction()

    def _compute_fraction(self) -> float:
        """Fraction of each heal attributable to the extra mastery stack."""
        n = self._base_stacks
        table = self._dr_table
        if n < 1 or n >= len(table):
            return 0.0
        mult_base = 1.0 + self._mastery * table[n - 1]
        mult_with_bloom = 1.0 + self._mastery * table[n]
        return 1.0 - mult_base / mult_with_bloom

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id == SYMBIOTIC_BLOOM:
            return 0.0
        if hot_tracker.get(event.target_id, SYMBIOTIC_BLOOM):
            return event.amount * self._fraction
        return 0.0
