from flourish.analysis.attributor import TalentAttributor
from flourish.models.events import HealEvent
from flourish.tracking.hot_tracker import HotTracker
from flourish.tracking.buff_tracker import BuffTracker


class StaticBuffAttributor(TalentAttributor):
    """For talents that unconditionally buff specific spells by a flat percentage.
    Attributes the bonus portion: amount - amount / (1 + multiplier)."""
    name: str = "Unknown"
    spell_ids: set[int] = set()
    multiplier: float = 0.0

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id in self.spell_ids and self.multiplier > 0:
            return event.amount - event.amount / (1 + self.multiplier)
        return 0.0


class ImprovedSwiftmendAttributor(StaticBuffAttributor):
    """Improved Swiftmend: Swiftmend healing increased by 30%."""
    name = "Improved Swiftmend"
    talent_node_id = 103873
    spell_ids = {18562}  # Swiftmend
    multiplier = 0.3


class LifetreadingAttributor(StaticBuffAttributor):
    """Lifetreading: Efflorescence healing increased by 25%."""
    name = "Lifetreading"
    talent_node_id = 103874
    spell_ids = {81269}  # Efflorescence
    multiplier = 0.25


class UnstoppableGrowthAttributor(StaticBuffAttributor):
    """Unstoppable Growth: WG healing falls off 30% less per rank (2 ranks).
    Net effect: ~27.7% more total WG healing."""
    name = "Unstoppable Growth"
    talent_node_id = 82080
    spell_ids = {48438}  # Wild Growth
    multiplier = 0.277


class IntensityAttributor(TalentAttributor):
    """Intensity: Regrowth crits at 260% instead of 200%.
    On Regrowth crits, attribute the bonus: amount - amount / 1.3"""
    name = "Intensity"
    talent_node_id = 82052

    REGROWTH_IDS = {8936, 1264664}  # Regrowth + Rampant Growth Regrowth

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id in self.REGROWTH_IDS and event.hit_type == 2:  # 2 = crit
            return event.amount - event.amount / 1.3
        return 0.0


class LivelinessAttributor(StaticBuffAttributor):
    """Liveliness: HoTs heal 5% faster = ~5% more total HoT healing."""
    name = "Liveliness"
    talent_node_id = 82074
    talent_id = 103130  # WCL entryId (choice node vs Master Shapeshifter)
    spell_ids = {774, 155777, 8936, 1264664, 48438, 33763, 33778, 1244341}
    # Rejuv, Germination, Regrowth, Rampant Growth RG, WG, LB tick, LB bloom, Everbloom
    multiplier = 0.05


class RegenesisAttributor(StaticBuffAttributor):
    """Regenesis: Rejuv and Tranq healing +up to 30% on low health.
    Approximate as flat 15% (configurable)."""
    name = "Regenesis"
    talent_node_id = 82062
    spell_ids = {774, 155777, 157982, 1264659}  # Rejuv, Germination, Tranquility, Flourish Tranq
    multiplier = 0.15
