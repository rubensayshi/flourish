import pytest

from rdruid_analyzer.analysis.pipeline import Pipeline
from rdruid_analyzer.analysis.talents.improved_wild_growth import ImprovedWildGrowthAttributor

WILD_GROWTH = 48438


def make_heal(ts, ability, amount, target=2, overheal=0):
    return {"timestamp": ts, "type": "heal", "sourceID": 1, "targetID": target,
            "abilityGameID": ability, "amount": amount, "overheal": overheal, "hitType": 1}


def test_iwg_attributes_extra_target_share():
    events = [
        make_heal(100, WILD_GROWTH, 7000),
    ]
    pipeline = Pipeline(attributors=[ImprovedWildGrowthAttributor()])
    results = pipeline.run(events)
    # 2/7 * 7000 = 2000
    assert results.talent_healing["Improved Wild Growth"] == pytest.approx(2000.0)


def test_iwg_ignores_non_wg():
    events = [
        make_heal(100, 774, 10000),   # Rejuv
        make_heal(200, 8936, 5000),   # Regrowth
    ]
    pipeline = Pipeline(attributors=[ImprovedWildGrowthAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Improved Wild Growth"] == 0.0


def test_iwg_multiple_ticks():
    events = [
        make_heal(100, WILD_GROWTH, 7000, target=2),
        make_heal(100, WILD_GROWTH, 7000, target=3),
        make_heal(100, WILD_GROWTH, 7000, target=4),
    ]
    pipeline = Pipeline(attributors=[ImprovedWildGrowthAttributor()])
    results = pipeline.run(events)
    # 2/7 * 21000 = 6000
    assert results.talent_healing["Improved Wild Growth"] == pytest.approx(6000.0)
