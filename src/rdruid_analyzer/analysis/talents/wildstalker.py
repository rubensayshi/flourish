from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent, CastEvent, ApplyBuffEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

# Symbiotic Bloom buff — from Thriving Growth
SYMBIOTIC_BLOOM = 439530

STRATEGIC_INFUSION_CRIT_BONUS = 0.04  # +4% crit chance on periodic heals
CRIT_RATING_PER_PERCENT = 700.0  # same conversion as abundance.py

SWIFTMEND = 18562
WILD_GROWTH = 48438
IMPLANT_TAG = "implant"
IMPLANT_WINDOW_MS = 500


class VigorousCreepersAttributor(TalentAttributor):
    """Symbiotic Blooms increase healing to affected targets by 20%."""

    name = "Vigorous Creepers"
    talent_node_id = 94627

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        # Don't double-count the bloom's own healing ticks
        if event.ability_id == SYMBIOTIC_BLOOM:
            return 0.0
        # Check if target has a Symbiotic Bloom
        if hot_tracker.get(event.target_id, SYMBIOTIC_BLOOM):
            return event.amount - event.amount / 1.2
        return 0.0


class ImplantAttributor(TalentAttributor):
    """Implant: SM/WG spawns a Symbiotic Bloom. Attribute bloom healing from those."""

    name = "Implant"
    talent_node_id = 94628
    talent_id = 117229  # WCL entryId (choice node vs Twin Sprouts)

    def __init__(self):
        super().__init__()
        self._recent_casts: list[tuple[int, int]] = []  # (timestamp, target_id)

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        if isinstance(event, CastEvent) and event.ability_id in (SWIFTMEND, WILD_GROWTH):
            self._recent_casts.append((event.timestamp, event.target_id))
            # Clean old entries
            self._recent_casts = [
                (t, tid) for t, tid in self._recent_casts if event.timestamp - t < IMPLANT_WINDOW_MS * 2
            ]

        if isinstance(event, ApplyBuffEvent) and event.ability_id == SYMBIOTIC_BLOOM:
            for ts, tid in self._recent_casts:
                if event.timestamp - ts < IMPLANT_WINDOW_MS:
                    hot = hot_tracker.get(event.target_id, SYMBIOTIC_BLOOM)
                    if hot:
                        hot.tags.add(IMPLANT_TAG)
                    break

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if event.ability_id != SYMBIOTIC_BLOOM:
            return 0.0
        hot = hot_tracker.get(event.target_id, SYMBIOTIC_BLOOM)
        if hot and IMPLANT_TAG in hot.tags:
            return float(event.amount)
        return 0.0


class StrategicInfusionAttributor(TalentAttributor):
    """Strategic Infusion: periodic heals have +4% crit chance.
    For periodic heal crits, attribute the talent's share of the crit bonus."""

    name = "Strategic Infusion"
    talent_node_id = 94623

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        if not event.tick or event.hit_type != 2:
            return 0.0

        base_crit = 0.0
        if self.combatant_info:
            base_crit = self.combatant_info.crit_spell / CRIT_RATING_PER_PERCENT
        base_crit = max(base_crit, 0.05)

        total_crit = base_crit + STRATEGIC_INFUSION_CRIT_BONUS
        infusion_share = STRATEGIC_INFUSION_CRIT_BONUS / total_crit

        crit_bonus = event.amount / 2.0
        return crit_bonus * infusion_share


class RootNetworkAttributor(TalentAttributor):
    """Root Network: +2% healing per active Symbiotic Bloom."""

    name = "Root Network"
    talent_node_id = 94631
    talent_id = 117233  # WCL entryId (choice node vs Resilient Flourishing)

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        bloom_count = len(hot_tracker.get_all_by_spell(SYMBIOTIC_BLOOM))
        if bloom_count <= 0:
            return 0.0
        multiplier = 0.02 * bloom_count
        return event.amount - event.amount / (1 + multiplier)
