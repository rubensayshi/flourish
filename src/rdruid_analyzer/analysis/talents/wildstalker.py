from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent, CastEvent, ApplyBuffEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

# Symbiotic Bloom buff — from Thriving Growth
SYMBIOTIC_BLOOM = 439528

SWIFTMEND = 18562
WILD_GROWTH = 48438
IMPLANT_TAG = "implant"
IMPLANT_WINDOW_MS = 500


class VigorousCreepersAttributor(TalentAttributor):
    """Symbiotic Blooms increase healing to affected targets by 20%."""

    name = "Vigorous Creepers"

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

    def __init__(self):
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


class RootNetworkAttributor(TalentAttributor):
    """Root Network: +2% healing per active Symbiotic Bloom."""

    name = "Root Network"

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        bloom_count = len(hot_tracker.get_all_by_spell(SYMBIOTIC_BLOOM))
        if bloom_count <= 0:
            return 0.0
        multiplier = 0.02 * bloom_count
        return event.amount - event.amount / (1 + multiplier)
