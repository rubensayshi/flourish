from flourish.analysis.attributor import TalentAttributor
from flourish.models.events import CastEvent, HealEvent
from flourish.tracking.hot_tracker import HotTracker
from flourish.tracking.buff_tracker import BuffTracker

SWIFTMEND = 18562
WILD_GROWTH = 48438
EARLY_SPRING_NODE_ID = 94591
EARLY_SPRING_TALENT_ID = 117895  # WCL entryId (choice node vs Bounteous Bloom)
DRYADS_DANCE_NODE_ID = 109713
RENEWING_SURGE_NODE_ID = 82060

PROSPERITY_NODE_ID = 82079
PROSPERITY_TALENT_ID = 103136  # WCL entryId (choice node vs Verdant Infusion)

DRYAD_HEAL_SPELLS = {1264659, 1264664}
SPIRIT_THICKET_SPELL = 1264905

BASE_SM_CD_MS = 15000
BASE_WG_CD_MS = 10000
WG_4PC_REDUCTION_MS = 2000
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


def compute_effective_wg_cd(
    has_early_spring: bool,
    has_4pc: bool,
) -> float:
    """Compute effective WG cooldown in ms."""
    cd = float(BASE_WG_CD_MS)
    if has_4pc:
        cd -= WG_4PC_REDUCTION_MS
    if has_early_spring:
        cd -= EARLY_SPRING_REDUCTION_MS
    return cd


class WgCooldownReductionAttributor(TalentAttributor):
    """Attributes healing value of Early Spring WG CD reduction.

    Tracks WG cast gaps, checks on-cooldown usage, and attributes
    a fraction of downstream GG healing."""

    name = "WG Cooldown Reduction"

    def __init__(self, downstream_attributors=None, has_4pc=True):
        super().__init__()
        self._wg_cast_timestamps: list[int] = []
        self._downstream = downstream_attributors or []
        self._has_4pc = has_4pc

    def is_selected(self) -> bool:
        if self.combatant_info is None:
            return True
        return (
            EARLY_SPRING_NODE_ID in self.combatant_info.talent_nodes
            and EARLY_SPRING_TALENT_ID in self.combatant_info.talent_ids
        )

    def process_event(self, event, hot_tracker, buff_tracker):
        # Track begincast (not cast) for WG since it has a cast time.
        # The CD starts when you press the button, not when the cast finishes.
        if isinstance(event, CastEvent) and event.type == "begincast" and event.ability_id == WILD_GROWTH:
            self._wg_cast_timestamps.append(event.timestamp)

    def process_heal(self, event, hot_tracker, buff_tracker) -> float:
        return 0.0

    def finalize(self) -> float:
        if len(self._wg_cast_timestamps) < 2:
            return 0.0

        has_early_spring = (
            self.has_talent(EARLY_SPRING_NODE_ID)
            and self.combatant_info is not None
            and EARLY_SPRING_TALENT_ID in self.combatant_info.talent_ids
        )

        unreduced_cd = compute_effective_wg_cd(
            has_early_spring=False,
            has_4pc=self._has_4pc,
        )

        reduced_cd = compute_effective_wg_cd(
            has_early_spring=has_early_spring,
            has_4pc=self._has_4pc,
        )

        total_ratio = 0.0
        total_casts = len(self._wg_cast_timestamps)

        for i in range(1, total_casts):
            gap = self._wg_cast_timestamps[i] - self._wg_cast_timestamps[i - 1]
            if gap <= reduced_cd + ON_COOLDOWN_TOLERANCE_MS:
                ratio = 1 - (reduced_cd / unreduced_cd)
                total_ratio += max(ratio, 0.0)

        if total_ratio == 0.0:
            return 0.0

        extra_cast_fraction = total_ratio / total_casts
        downstream_total = sum(attr.total_attributed for attr in self._downstream)

        return extra_cast_fraction * downstream_total


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

    def _dryad_overlap_in_window(self, window_start: float, window_duration: float) -> float:
        """Compute total Dryad overlap during a recharge window."""
        cd_end = window_start + window_duration
        overlap = 0.0
        for d_start, d_end in self._dryad_windows:
            o_start = max(window_start, d_start)
            o_end = min(cd_end, d_end)
            if o_end > o_start:
                overlap += o_end - o_start
        return overlap

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
        has_prosperity = (
            self.has_talent(PROSPERITY_NODE_ID)
            and self.combatant_info is not None
            and PROSPERITY_TALENT_ID in self.combatant_info.talent_ids
        )

        max_charges = 2 if has_prosperity else 1

        # Unreduced CD = baseline with Renewing Surge only (no Early Spring, no Dryad)
        unreduced_cd = compute_effective_cd(
            has_renewing_surge=has_renewing_surge,
            has_early_spring=False,
            dryad_overlap_ms=0,
        )

        # Charge simulation: track pending recharges as (completion_time, reduced_cd)
        charges = max_charges
        pending: list[tuple[float, float]] = []

        total_ratio = 0.0
        total_casts = len(self._sm_cast_timestamps)

        for cast_ts in self._sm_cast_timestamps:
            # Restore charges that have finished recharging
            was_depleted = charges == 0
            last_restore: tuple[float, float] | None = None

            while pending and pending[0][0] <= cast_ts:
                entry = pending.pop(0)
                charges = min(charges + 1, max_charges)
                if was_depleted:
                    last_restore = entry
                    was_depleted = charges == 0

            on_cooldown = False
            reduced_cd_used = 0.0

            if charges == 0 and pending:
                # Still depleted — check if next charge arrives within tolerance
                completion, rcd = pending[0]
                if completion <= cast_ts + ON_COOLDOWN_TOLERANCE_MS:
                    on_cooldown = True
                    reduced_cd_used = rcd
                    pending.pop(0)
                    charges += 1
                else:
                    # Model imprecision: cast happened but no charge predicted.
                    # Force-restore without counting as on-cooldown.
                    pending.pop(0)
                    charges += 1
            elif last_restore is not None:
                # Was depleted, charge recently restored — check tolerance
                completion, rcd = last_restore
                if cast_ts - completion <= ON_COOLDOWN_TOLERANCE_MS:
                    on_cooldown = True
                    reduced_cd_used = rcd

            charges -= 1

            # Schedule recharge for consumed charge
            recharge_start = pending[-1][0] if pending else cast_ts
            dryad_overlap = (
                self._dryad_overlap_in_window(recharge_start, unreduced_cd)
                if has_dryads_dance
                else 0.0
            )
            reduced_cd = compute_effective_cd(
                has_renewing_surge=has_renewing_surge,
                has_early_spring=has_early_spring,
                dryad_overlap_ms=dryad_overlap,
            )
            pending.append((recharge_start + reduced_cd, reduced_cd))

            if on_cooldown:
                ratio = 1 - (reduced_cd_used / unreduced_cd)
                total_ratio += max(ratio, 0.0)

        if total_ratio == 0.0:
            return 0.0

        extra_cast_fraction = total_ratio / total_casts

        # Sum downstream attributor totals
        downstream_total = sum(attr.total_attributed for attr in self._downstream)

        return extra_cast_fraction * downstream_total
