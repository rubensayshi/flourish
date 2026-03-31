from rdruid_analyzer.analysis.attributor import TalentAttributor
from rdruid_analyzer.models.events import HealEvent
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker

REJUV = 774
GERMINATION_REJUV = 155777
REGROWTH_IDS = {8936, 1264664}  # Regrowth + Rampant Growth Regrowth
ABUNDANCE_CRIT_PER_REJUV = 0.08  # 8% crit per active Rejuv


class AbundanceAttributor(TalentAttributor):
    """Abundance: +8% Regrowth crit per active Rejuv, up to 96%.
    For Regrowth crits, attribute abundance_crit / total_crit share of the crit bonus."""

    name = "Abundance"

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        # Only Regrowth crits
        if event.ability_id not in REGROWTH_IDS or event.hit_type != 2:
            return 0.0

        # Count active Rejuvs (both normal and Germination)
        rejuv_count = len(hot_tracker.get_all_by_spell(REJUV)) + len(
            hot_tracker.get_all_by_spell(GERMINATION_REJUV)
        )
        if rejuv_count <= 0:
            return 0.0

        abundance_crit = min(rejuv_count * ABUNDANCE_CRIT_PER_REJUV, 0.96)

        # Get base crit from combatant info
        base_crit = 0.0
        if hasattr(self, "combatant_info") and self.combatant_info:
            # critSpell is in rating, roughly /700 for percentage at this level
            base_crit = self.combatant_info.crit_spell / 700.0
        base_crit = max(base_crit, 0.05)  # minimum 5% base crit

        total_crit = base_crit + abundance_crit
        abundance_share = abundance_crit / total_crit

        # Crit heal = 2x normal, so bonus = amount / 2
        crit_bonus = event.amount / 2.0

        return crit_bonus * abundance_share
