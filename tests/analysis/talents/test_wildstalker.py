import pytest

from flourish.analysis.pipeline import Pipeline
from flourish.analysis.talents.wildstalker.wildstalker import (
    VigorousCreepersAttributor,
    ImplantAttributor,
    RootNetworkAttributor,
    StrategicInfusionAttributor,
    SYMBIOTIC_BLOOM,
    CRIT_RATING_PER_PERCENT,
)

TARGET = 10
TARGET2 = 20
REJUV = 774
SWIFTMEND = 18562
WILD_GROWTH = 48438


def make_cast(ts, ability, target=TARGET):
    return {"timestamp": ts, "type": "cast", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_apply(ts, ability, target=TARGET):
    return {"timestamp": ts, "type": "applybuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_remove(ts, ability, target=TARGET):
    return {"timestamp": ts, "type": "removebuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_heal(ts, ability, amount, target=TARGET, overheal=0, hit_type=1, tick=False):
    evt = {
        "timestamp": ts, "type": "heal", "sourceID": 1, "targetID": target,
        "abilityGameID": ability, "amount": amount, "overheal": overheal, "hitType": hit_type,
    }
    if tick:
        evt["tick"] = True
    return evt


# --- Vigorous Creepers ---

class TestVigorousCreepers:
    def test_buff_on_target_boosts_heal(self):
        events = [
            make_apply(100, SYMBIOTIC_BLOOM),
            make_heal(200, REJUV, 12000),
        ]
        pipeline = Pipeline(attributors=[VigorousCreepersAttributor()])
        results = pipeline.run(events)
        # 12000 - 12000/1.2 = 2000
        assert results.talent_healing["Vigorous Creepers"] == pytest.approx(2000.0)

    def test_no_bloom_no_bonus(self):
        events = [
            make_heal(200, REJUV, 12000),
        ]
        pipeline = Pipeline(attributors=[VigorousCreepersAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Vigorous Creepers"] == 0.0

    def test_bloom_own_healing_not_counted(self):
        events = [
            make_apply(100, SYMBIOTIC_BLOOM),
            make_heal(200, SYMBIOTIC_BLOOM, 5000),
        ]
        pipeline = Pipeline(attributors=[VigorousCreepersAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Vigorous Creepers"] == 0.0

    def test_bloom_removed_no_bonus(self):
        events = [
            make_apply(100, SYMBIOTIC_BLOOM),
            make_remove(150, SYMBIOTIC_BLOOM),
            make_heal(200, REJUV, 12000),
        ]
        pipeline = Pipeline(attributors=[VigorousCreepersAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Vigorous Creepers"] == 0.0


# --- Implant ---

class TestImplant:
    def test_sm_triggers_implant_bloom(self):
        events = [
            make_cast(100, SWIFTMEND),
            make_apply(200, SYMBIOTIC_BLOOM),  # within 500ms window
            make_heal(300, SYMBIOTIC_BLOOM, 8000),
        ]
        pipeline = Pipeline(attributors=[ImplantAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Implant"] == pytest.approx(8000.0)

    def test_wg_triggers_implant_bloom(self):
        """WG is AoE — bloom can appear on any target, not just the cast target."""
        events = [
            make_cast(100, WILD_GROWTH, target=TARGET),
            make_apply(200, SYMBIOTIC_BLOOM, target=TARGET2),  # bloom on different target
            make_heal(300, SYMBIOTIC_BLOOM, 5000, target=TARGET2),
        ]
        pipeline = Pipeline(attributors=[ImplantAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Implant"] == pytest.approx(5000.0)

    def test_natural_bloom_not_attributed(self):
        # Bloom appears without recent SM/WG
        events = [
            make_apply(100, SYMBIOTIC_BLOOM),
            make_heal(200, SYMBIOTIC_BLOOM, 5000),
        ]
        pipeline = Pipeline(attributors=[ImplantAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Implant"] == 0.0

    def test_bloom_outside_window_not_attributed(self):
        events = [
            make_cast(100, SWIFTMEND),
            make_apply(700, SYMBIOTIC_BLOOM),  # 600ms > 500ms window
            make_heal(800, SYMBIOTIC_BLOOM, 5000),
        ]
        pipeline = Pipeline(attributors=[ImplantAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Implant"] == 0.0

    def test_non_bloom_healing_not_attributed(self):
        events = [
            make_cast(100, SWIFTMEND),
            make_apply(200, SYMBIOTIC_BLOOM),
            make_heal(300, REJUV, 10000),  # Rejuv, not bloom
        ]
        pipeline = Pipeline(attributors=[ImplantAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Implant"] == 0.0


# --- Root Network ---

class TestRootNetwork:
    def test_single_bloom_gives_2pct(self):
        events = [
            make_apply(100, SYMBIOTIC_BLOOM, target=TARGET),
            make_heal(200, REJUV, 10000, target=TARGET2),
        ]
        pipeline = Pipeline(attributors=[RootNetworkAttributor()])
        results = pipeline.run(events)
        # 10000 - 10000/1.02 = ~196.08
        expected = 10000 - 10000 / 1.02
        assert results.talent_healing["Root Network"] == pytest.approx(expected)

    def test_multiple_blooms_stack(self):
        events = [
            make_apply(100, SYMBIOTIC_BLOOM, target=TARGET),
            make_apply(110, SYMBIOTIC_BLOOM, target=TARGET2),
            make_heal(200, REJUV, 10000, target=30),
        ]
        pipeline = Pipeline(attributors=[RootNetworkAttributor()])
        results = pipeline.run(events)
        # 2 blooms = 4%: 10000 - 10000/1.04
        expected = 10000 - 10000 / 1.04
        assert results.talent_healing["Root Network"] == pytest.approx(expected)

    def test_no_blooms_no_bonus(self):
        events = [
            make_heal(200, REJUV, 10000),
        ]
        pipeline = Pipeline(attributors=[RootNetworkAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Root Network"] == 0.0


# --- Strategic Infusion ---

def _make_combatant_info(crit_rating=0):
    return {
        "timestamp": 0, "type": "combatantinfo", "sourceID": 1,
        "talentTree": [{"nodeID": 94623, "id": 117223}],
        "critSpell": crit_rating, "hasteSpell": 0, "mastery": 0, "specID": 105,
    }


class TestStrategicInfusion:
    def test_periodic_crit_attributed(self):
        # 21% base crit (0.21 * 700 = 147 rating)
        crit_rating = 0.21 * CRIT_RATING_PER_PERCENT
        events = [
            _make_combatant_info(crit_rating),
            make_heal(100, REJUV, 10000, hit_type=2, tick=True),
        ]
        pipeline = Pipeline(attributors=[StrategicInfusionAttributor()])
        results = pipeline.run(events)
        # crit_bonus = 10000/2 = 5000; share = 0.04 / (0.21 + 0.04) = 0.16
        expected = 5000 * (0.04 / 0.25)
        assert results.talent_healing["Strategic Infusion"] == pytest.approx(expected)

    def test_periodic_non_crit_not_attributed(self):
        events = [
            _make_combatant_info(0.21 * CRIT_RATING_PER_PERCENT),
            make_heal(100, REJUV, 10000, hit_type=1, tick=True),
        ]
        pipeline = Pipeline(attributors=[StrategicInfusionAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Strategic Infusion"] == 0.0

    def test_direct_crit_not_attributed(self):
        events = [
            _make_combatant_info(0.21 * CRIT_RATING_PER_PERCENT),
            make_heal(100, SWIFTMEND, 20000, hit_type=2, tick=False),
        ]
        pipeline = Pipeline(attributors=[StrategicInfusionAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Strategic Infusion"] == 0.0

    def test_no_combatant_info_uses_minimum_crit(self):
        events = [
            make_heal(100, REJUV, 10000, hit_type=2, tick=True),
        ]
        pipeline = Pipeline(attributors=[StrategicInfusionAttributor()])
        results = pipeline.run(events)
        # base_crit floors to 0.05; share = 0.04 / 0.09
        expected = 5000 * (0.04 / 0.09)
        assert results.talent_healing["Strategic Infusion"] == pytest.approx(expected)
