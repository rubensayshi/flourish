from __future__ import annotations

from rdruid_analyzer.analysis.attributor import TalentAttributor

DR_FRACTION = 0.08  # 8% damage reduction


class ProtectiveGrowthAttributor(TalentAttributor):
    name = "Protective Growth"
    talent_node_id = 94593

    def __init__(self, damage_taken_with_regrowth: int | None = None):
        super().__init__()
        self.damage_taken_with_regrowth = damage_taken_with_regrowth or 0

    def finalize(self) -> float:
        if self.damage_taken_with_regrowth <= 0:
            return 0.0
        return self.damage_taken_with_regrowth * DR_FRACTION / (1 - DR_FRACTION)
