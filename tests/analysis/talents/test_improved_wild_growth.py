import pytest

from flourish.analysis.pipeline import Pipeline
from flourish.analysis.talents.improved_wild_growth import ImprovedWildGrowthAttributor

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


TOL_BUFF = 33891


def make_applybuff(ts, ability, target=2):
    return {"timestamp": ts, "type": "applybuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_removebuff(ts, ability, target=2):
    return {"timestamp": ts, "type": "removebuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def test_iwg_skips_during_tol():
    """IWG should not attribute during Tree of Life — ToL handles IWG targets itself."""
    events = [
        make_applybuff(50, TOL_BUFF, target=1),
        make_heal(100, WILD_GROWTH, 7000),
        make_heal(200, WILD_GROWTH, 7000),
        make_removebuff(300, TOL_BUFF, target=1),
        make_heal(400, WILD_GROWTH, 7000),  # after ToL ends
    ]
    pipeline = Pipeline(attributors=[ImprovedWildGrowthAttributor()])
    results = pipeline.run(events)
    # Only the last tick (after ToL) should be attributed: 2/7 * 7000 = 2000
    assert results.talent_healing["Improved Wild Growth"] == pytest.approx(2000.0)
