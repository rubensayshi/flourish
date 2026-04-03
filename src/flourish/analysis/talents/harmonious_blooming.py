from flourish.analysis.attributor import TalentAttributor
from flourish.models.events import CombatantInfoEvent, HealEvent
from flourish.tracking.hot_tracker import HotTracker
from flourish.tracking.buff_tracker import BuffTracker

LIFEBLOOM = 33763

# Default DR table: cumulative mastery multiplier per stack count
DEFAULT_DR_TABLE = [1.0, 1.7, 2.3, 2.8, 3.2]


class HarmoniousBloomingAttributor(TalentAttributor):
    """Attributes the marginal mastery bonus from Lifebloom counting as 3 mastery stacks instead of 1."""

    name = "Harmonious Blooming"
    talent_node_id = 82077

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
        """Fraction of each heal attributable to the 2 extra mastery stacks from Harmonious Blooming.

        Without talent: Lifebloom = 1 stack → base_stacks includes that 1.
        With talent: Lifebloom = 3 stacks → base_stacks + 2.
        """
        n = self._base_stacks
        table = self._dr_table
        # With talent: 2 extra stacks
        n_with = min(n + 2, len(table) - 1)
        if n < 1 or n >= len(table):
            return 0.0
        mult_base = 1.0 + self._mastery * table[n]
        mult_with = 1.0 + self._mastery * table[n_with]
        return 1.0 - mult_base / mult_with

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        # Don't attribute Lifebloom's own healing — only the mastery bonus on other heals
        if event.ability_id == LIFEBLOOM:
            return 0.0
        if hot_tracker.get(event.target_id, LIFEBLOOM):
            return event.amount * self._fraction
        return 0.0
