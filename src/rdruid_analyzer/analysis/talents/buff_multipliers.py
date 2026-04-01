from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent, CastEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker


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


class HarmonyOfTheGroveAttributor(TalentAttributor):
    """Harmony of the Grove: Each Grove Guardian increases healing done by 5%.
    This is a dynamic buff — we need to track active guardian count.
    For simplicity, we track summon/despawn events. Grove Guardians last 8 sec.
    The buff spell ID in WCL should be 428731."""
    name = "Harmony of the Grove"
    talent_node_id = 94606

    def __init__(self):
        super().__init__()
        self._guardian_count = 0
        self._guardian_despawn_times: list[int] = []

    def process_event(self, event, hot_tracker, buff_tracker):
        # Clean up expired guardians (8 sec = 8000ms duration)
        if hasattr(event, 'timestamp'):
            self._guardian_despawn_times = [t for t in self._guardian_despawn_times if t > event.timestamp]
            self._guardian_count = len(self._guardian_despawn_times)

        # Track guardian summons via cast events
        if isinstance(event, CastEvent) and event.ability_id in (18562, 48438):
            # Only count if Grove Guardians talent is taken
            if self.has_talent(82043):
                self._guardian_despawn_times.append(event.timestamp + 8000)
                self._guardian_count = len([t for t in self._guardian_despawn_times if t > event.timestamp])

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if self._guardian_count <= 0:
            return 0.0
        # 5% per guardian
        multiplier = 0.05 * self._guardian_count
        return event.amount - event.amount / (1 + multiplier)


class PowerOfNatureAttributor(TalentAttributor):
    """Power of Nature: Grove Guardians increase Rejuv, Efflorescence, and Lifebloom
    healing by 10% while active. Same guardian tracking as HarmonyOfTheGrove."""
    name = "Power of Nature"
    talent_node_id = 94605
    talent_id = 122213  # Choice node with Durability of Nature

    SPELL_IDS = {774, 155777, 81269, 33763, 33778}  # Rejuv, Germ Rejuv, Efflor, LB tick, LB bloom

    def __init__(self):
        super().__init__()
        self._guardian_count = 0
        self._guardian_despawn_times: list[int] = []

    def process_event(self, event, hot_tracker, buff_tracker):
        if hasattr(event, 'timestamp'):
            self._guardian_despawn_times = [t for t in self._guardian_despawn_times if t > event.timestamp]
            self._guardian_count = len(self._guardian_despawn_times)

        if isinstance(event, CastEvent) and event.ability_id in (18562, 48438):
            if self.has_talent(82043):
                self._guardian_despawn_times.append(event.timestamp + 8000)
                self._guardian_count = len([t for t in self._guardian_despawn_times if t > event.timestamp])

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if self._guardian_count <= 0 or event.ability_id not in self.SPELL_IDS:
            return 0.0
        multiplier = 0.10 * self._guardian_count
        return event.amount - event.amount / (1 + multiplier)


class GrovesInspirationAttributor(StaticBuffAttributor):
    """Grove's Inspiration: Regrowth, Wild Growth, and Swiftmend healing increased by 9%."""
    name = "Grove's Inspiration"
    talent_node_id = 94595
    talent_id = 122201
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
    talent_id = 122196
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
    talent_id = 108135
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
