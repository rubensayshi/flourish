from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import CastEvent, HealEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

SWIFTMEND = 18562
EARLY_SPRING_NODE_ID = 94591
EARLY_SPRING_TALENT_ID = 122907
DRYADS_DANCE_NODE_ID = 109713
RENEWING_SURGE_NODE_ID = 82060

DRYAD_HEAL_SPELLS = {1264659, 1264664}
SPIRIT_THICKET_SPELL = 1264905

BASE_SM_CD_MS = 15000
RENEWING_SURGE_REDUCTION_AVG = 0.195
EARLY_SPRING_REDUCTION_MS = 1000
DRYADS_DANCE_SPEED_FACTOR = 1.25
ON_COOLDOWN_TOLERANCE_MS = 1500
DRYAD_GAP_THRESHOLD_MS = 2000


def compute_effective_cd(
    has_renewing_surge: bool,
    has_early_spring: bool,
    dryad_overlap_ms: float,
) -> float:
    """Compute effective SM cooldown in ms given active talents and Dryad overlap."""
    cd = float(BASE_SM_CD_MS)
    if has_renewing_surge:
        cd *= (1 - RENEWING_SURGE_REDUCTION_AVG)
    if has_early_spring:
        cd -= EARLY_SPRING_REDUCTION_MS
    if dryad_overlap_ms > 0:
        overlap = min(dryad_overlap_ms, cd)
        remaining = cd - overlap
        cd = remaining + overlap / DRYADS_DANCE_SPEED_FACTOR
    return cd


class SmCooldownReductionAttributor(TalentAttributor):
    """Attributes healing value of Early Spring + Dryad's Dance SM CD reduction.

    Computes effective CD per SM cast, checks if pressed on cooldown,
    and attributes a fraction of downstream healing (SotF, PotA, GG)."""

    name = "SM Cooldown Reduction"

    def __init__(self, downstream_attributors=None):
        super().__init__()
        self._sm_cast_timestamps: list[int] = []
        self._dryad_windows: list[tuple[int, int]] = []
        self._dryad_start: int | None = None
        self._dryad_last_heal: int | None = None
        self._downstream = downstream_attributors or []

    def is_selected(self) -> bool:
        if self.combatant_info is None:
            return True
        has_early_spring = (
            EARLY_SPRING_NODE_ID in self.combatant_info.talent_nodes
            and EARLY_SPRING_TALENT_ID in self.combatant_info.talent_ids
        )
        has_dryads_dance = DRYADS_DANCE_NODE_ID in self.combatant_info.talent_nodes
        return has_early_spring or has_dryads_dance

    def _close_dryad_window(self):
        if self._dryad_start is not None and self._dryad_last_heal is not None:
            self._dryad_windows.append((self._dryad_start, self._dryad_last_heal))
            self._dryad_start = None
            self._dryad_last_heal = None

    def process_event(self, event, hot_tracker: HotTracker, buff_tracker: BuffTracker):
        all_dryad_spells = DRYAD_HEAL_SPELLS | {SPIRIT_THICKET_SPELL}

        # Track Dryad windows from pet heal events
        if (
            isinstance(event, HealEvent)
            and event.ability_id in all_dryad_spells
            and self.combatant_info is not None
            and event.source_id != self.combatant_info.source_id
        ):
            # Close previous window if gap exceeded
            if (
                self._dryad_last_heal is not None
                and event.timestamp - self._dryad_last_heal > DRYAD_GAP_THRESHOLD_MS
            ):
                self._close_dryad_window()

            if self._dryad_start is None:
                self._dryad_start = event.timestamp
            self._dryad_last_heal = event.timestamp
        else:
            # Non-dryad event: check if open window should close due to gap
            if (
                self._dryad_last_heal is not None
                and event.timestamp - self._dryad_last_heal > DRYAD_GAP_THRESHOLD_MS
            ):
                self._close_dryad_window()

        if isinstance(event, CastEvent) and event.ability_id == SWIFTMEND:
            self._sm_cast_timestamps.append(event.timestamp)

    def process_heal(self, event, hot_tracker, buff_tracker) -> float:
        return 0.0

    def finalize(self) -> float:
        self._close_dryad_window()

        if len(self._sm_cast_timestamps) < 2:
            return 0.0

        has_renewing_surge = self.has_talent(RENEWING_SURGE_NODE_ID)
        has_early_spring = (
            self.has_talent(EARLY_SPRING_NODE_ID)
            and self.combatant_info is not None
            and EARLY_SPRING_TALENT_ID in self.combatant_info.talent_ids
        )
        has_dryads_dance = self.has_talent(DRYADS_DANCE_NODE_ID)

        # Unreduced CD = baseline with Renewing Surge only (no Early Spring, no Dryad)
        unreduced_cd = compute_effective_cd(
            has_renewing_surge=has_renewing_surge,
            has_early_spring=False,
            dryad_overlap_ms=0,
        )

        total_ratio = 0.0
        total_casts = len(self._sm_cast_timestamps)

        for i in range(1, total_casts):
            prev_ts = self._sm_cast_timestamps[i - 1]
            curr_ts = self._sm_cast_timestamps[i]
            gap = curr_ts - prev_ts

            # Compute Dryad overlap during CD window
            dryad_overlap = 0.0
            if has_dryads_dance:
                cd_start = prev_ts
                cd_end = prev_ts + unreduced_cd
                for d_start, d_end in self._dryad_windows:
                    overlap_start = max(cd_start, d_start)
                    overlap_end = min(cd_end, d_end)
                    if overlap_end > overlap_start:
                        dryad_overlap += overlap_end - overlap_start

            reduced_cd = compute_effective_cd(
                has_renewing_surge=has_renewing_surge,
                has_early_spring=has_early_spring,
                dryad_overlap_ms=dryad_overlap,
            )

            # On-cooldown check
            if gap <= reduced_cd + ON_COOLDOWN_TOLERANCE_MS:
                ratio = 1 - (reduced_cd / unreduced_cd)
                total_ratio += max(ratio, 0.0)

        if total_ratio == 0.0:
            return 0.0

        extra_cast_fraction = total_ratio / total_casts

        # Sum downstream attributor totals
        downstream_total = sum(attr.total_attributed for attr in self._downstream)

        return extra_cast_fraction * downstream_total
