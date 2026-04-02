from flourish.analysis.talents.protective_growth import ProtectiveGrowthAttributor


class TestProtectiveGrowth:
    def test_attributes_dr_as_healing(self):
        attr = ProtectiveGrowthAttributor(damage_taken_with_regrowth=100_000)
        result = attr.finalize()
        # 100k damage seen after 8% DR => prevented = 100000 * 0.08 / 0.92
        expected = 100_000 * 0.08 / 0.92
        assert abs(result - expected) < 1

    def test_zero_damage(self):
        attr = ProtectiveGrowthAttributor(damage_taken_with_regrowth=0)
        assert attr.finalize() == 0.0

    def test_none_damage(self):
        """No API data available — attribute nothing."""
        attr = ProtectiveGrowthAttributor(damage_taken_with_regrowth=None)
        assert attr.finalize() == 0.0
