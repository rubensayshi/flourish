from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent, CastEvent, SummonEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

GROVE_GUARDIAN_SUMMON_ID = 102693  # WCL summon ability for Grove Guardian treants
GROVE_GUARDIAN_BASE_DURATION_MS = 8000
GROVE_GUARDIAN_DURABILITY_BONUS = 0.2  # Durability of Nature: +20% duration
DURABILITY_OF_NATURE_NODE = 94605
DURABILITY_OF_NATURE_TALENT_ID = 117200  # WCL entryId


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


class WildSynthesisAttributor(StaticBuffAttributor):
    """Wild Synthesis: Grove Guardians, Efflorescence, and summons heal 30% more."""
    name = "Wild Synthesis"
    talent_node_id = 94535
    spell_ids = {422090, 142421, 81269, 434141}  # Nourish, Treant heal, Efflorescence, Dream Bloom
    multiplier = 0.3


class WildstalkersPowerAttributor(StaticBuffAttributor):
    """Wildstalker's Power: Rejuvenation, Efflorescence, and Lifebloom healing increased by 10%."""
    name = "Wildstalker's Power"
    talent_node_id = 94621
    spell_ids = {774, 155777, 81269, 33763, 33778}  # Rejuv, Germination Rejuv, Efflorescence, LB tick, LB bloom
    multiplier = 0.1


class PatientCustodianAttributor(StaticBuffAttributor):
    """Patient Custodian: HoT effects are 6% more effective.
    Applies to all HoTs: Rejuv, Regrowth, Wild Growth, Lifebloom, etc."""
    name = "Patient Custodian"
    talent_node_id = 94630
    spell_ids = {774, 155777, 8936, 48438, 33763, 33778, 1244341, 1264664}
    # Rejuv, Germination, Regrowth, Wild Growth, Lifebloom tick, Lifebloom bloom, Everbloom, Rampant Growth Regrowth
    multiplier = 0.06


class LifetreadingAttributor(StaticBuffAttributor):
    """Lifetreading: Efflorescence healing increased by 25%."""
    name = "Lifetreading"
    talent_node_id = 103874
    spell_ids = {81269}  # Efflorescence
    multiplier = 0.25


class _GuardianTrackingMixin:
    """Shared guardian tracking via WCL summon events (ability 102693).
    Accounts for Durability of Nature extending duration by 20%."""

    def __init__(self):
        super().__init__()
        self._guardian_count = 0
        self._guardian_despawn_times: list[int] = []
        self._guardian_duration_ms: int = GROVE_GUARDIAN_BASE_DURATION_MS

    def set_combatant_info(self, info):
        super().set_combatant_info(info)
        # Check Durability of Nature for extended guardian duration
        if info and DURABILITY_OF_NATURE_NODE in info.talent_nodes:
            if DURABILITY_OF_NATURE_TALENT_ID in info.talent_ids:
                self._guardian_duration_ms = int(
                    GROVE_GUARDIAN_BASE_DURATION_MS * (1 + GROVE_GUARDIAN_DURABILITY_BONUS)
                )

    def _update_guardians(self, event):
        if hasattr(event, 'timestamp'):
            self._guardian_despawn_times = [t for t in self._guardian_despawn_times if t > event.timestamp]
            self._guardian_count = len(self._guardian_despawn_times)

        if isinstance(event, SummonEvent) and event.ability_id == GROVE_GUARDIAN_SUMMON_ID:
            self._guardian_despawn_times.append(event.timestamp + self._guardian_duration_ms)
            self._guardian_count = len([t for t in self._guardian_despawn_times if t > event.timestamp])


class HarmonyOfTheGroveAttributor(_GuardianTrackingMixin, TalentAttributor):
    """Harmony of the Grove: Each Grove Guardian increases healing done by 5%.
    Tracks guardians via WCL summon events, capturing all sources (casts, Convoke, etc.)."""
    name = "Harmony of the Grove"
    talent_node_id = 94606

    def process_event(self, event, hot_tracker, buff_tracker):
        self._update_guardians(event)

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if self._guardian_count <= 0:
            return 0.0
        multiplier = 0.05 * self._guardian_count
        return event.amount - event.amount / (1 + multiplier)


class PowerOfNatureAttributor(_GuardianTrackingMixin, TalentAttributor):
    """Power of Nature: Grove Guardians increase Rejuv, Efflorescence, and Lifebloom
    healing by 10% while active."""
    name = "Power of Nature"
    talent_node_id = 94605
    talent_id = 117201  # WCL entryId (choice node vs Durability of Nature)

    SPELL_IDS = {774, 155777, 81269, 33763, 33778}  # Rejuv, Germ Rejuv, Efflor, LB tick, LB bloom

    def process_event(self, event, hot_tracker, buff_tracker):
        self._update_guardians(event)

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if self._guardian_count <= 0 or event.ability_id not in self.SPELL_IDS:
            return 0.0
        multiplier = 0.10 * self._guardian_count
        return event.amount - event.amount / (1 + multiplier)


class GrovesInspirationAttributor(StaticBuffAttributor):
    """Grove's Inspiration: Regrowth, Wild Growth, and Swiftmend healing increased by 9%."""
    name = "Grove's Inspiration"
    talent_node_id = 94595
    talent_id = 117189  # WCL entryId (choice node vs Potent Enchantments)
    spell_ids = {8936, 1264664, 48438, 18562}
    # Regrowth, Rampant Growth Regrowth, Wild Growth, Swiftmend
    multiplier = 0.09


class CenariusMightAttributor(StaticBuffAttributor):
    """Cenarius' Might: Swiftmend healing increased by 20%."""
    name = "Cenarius' Might"
    talent_node_id = 94604
    spell_ids = {18562}  # Swiftmend
    multiplier = 0.2


class BountifulBloomAttributor(StaticBuffAttributor):
    """Bounteous Bloom: Grove Guardians healing increased by 30%."""
    name = "Bounteous Bloom"
    talent_node_id = 94591
    talent_id = 117184  # WCL entryId (choice node vs Early Spring)
    spell_ids = {422090, 142421}  # Nourish + direct heal (from treants)
    multiplier = 0.3


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
