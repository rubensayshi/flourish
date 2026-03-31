from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent
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


class WildSynthesisAttributor(StaticBuffAttributor):
    """Wild Synthesis: Grove Guardians, Efflorescence, and summons heal 30% more."""
    name = "Wild Synthesis"
    spell_ids = {422090, 81269, 434141}  # Nourish, Efflorescence, Dream Bloom
    multiplier = 0.3


class WildstalkersPowerAttributor(StaticBuffAttributor):
    """Wildstalker's Power: Rejuvenation healing increased by 10%."""
    name = "Wildstalker's Power"
    spell_ids = {774, 155777}  # Rejuvenation + Germination Rejuv
    multiplier = 0.1


class PatientCustodianAttributor(StaticBuffAttributor):
    """Patient Custodian: HoT effects are 6% more effective.
    Applies to all HoTs: Rejuv, Regrowth, Wild Growth, Lifebloom, etc."""
    name = "Patient Custodian"
    spell_ids = {774, 155777, 8936, 48438, 33763, 33778, 1244341, 1264664}
    # Rejuv, Germination, Regrowth, Wild Growth, Lifebloom tick, Lifebloom bloom, Everbloom, Rampant Growth Regrowth
    multiplier = 0.06


class LifetreadingAttributor(StaticBuffAttributor):
    """Lifetreading: Efflorescence healing increased by 25%."""
    name = "Lifetreading"
    spell_ids = {81269}  # Efflorescence
    multiplier = 0.25


class HarmonyOfTheGroveAttributor(TalentAttributor):
    """Harmony of the Grove: Each Grove Guardian increases healing done by 5%.
    This is a dynamic buff — we need to track active guardian count.
    For simplicity, we track summon/despawn events. Grove Guardians last 8 sec.
    The buff spell ID in WCL should be 428731."""
    name = "Harmony of the Grove"

    def __init__(self):
        self._guardian_count = 0
        self._guardian_despawn_times: list[int] = []

    def process_event(self, event, hot_tracker, buff_tracker):
        # Clean up expired guardians (8 sec = 8000ms duration)
        if hasattr(event, 'timestamp'):
            self._guardian_despawn_times = [t for t in self._guardian_despawn_times if t > event.timestamp]
            self._guardian_count = len(self._guardian_despawn_times)

        # Track guardian summons via the summon event type or cast events
        # WCL summon events have type "summon" with abilityGameID for the guardian
        from rdruid_analyzer.models.events import CastEvent
        if isinstance(event, CastEvent) and event.ability_id in (18562, 48438):
            # Swiftmend or Wild Growth summons a guardian (Grove Guardians talent)
            self._guardian_despawn_times.append(event.timestamp + 8000)
            self._guardian_count = len([t for t in self._guardian_despawn_times if t > event.timestamp])

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if self._guardian_count <= 0:
            return 0.0
        # 5% per guardian
        multiplier = 0.05 * self._guardian_count
        return event.amount - event.amount / (1 + multiplier)


class GrovesInspirationAttributor(StaticBuffAttributor):
    """Grove's Inspiration: Regrowth, Wild Growth, and Swiftmend healing increased by 9%."""
    name = "Grove's Inspiration"
    spell_ids = {8936, 1264664, 48438, 18562, 142421}
    # Regrowth, Rampant Growth Regrowth, Wild Growth, Swiftmend, Improved Swiftmend
    multiplier = 0.09


class CenariusMightAttributor(StaticBuffAttributor):
    """Cenarius' Might: Swiftmend healing increased by 20%."""
    name = "Cenarius' Might"
    spell_ids = {18562, 142421}  # Swiftmend + Improved Swiftmend
    multiplier = 0.2


class BountifulBloomAttributor(StaticBuffAttributor):
    """Bounteous Bloom: Grove Guardians healing increased by 30%."""
    name = "Bounteous Bloom"
    spell_ids = {422090}  # Nourish (from treants)
    multiplier = 0.3
