import pytest

from rdruid_analyzer.analysis.pipeline import Pipeline
from rdruid_analyzer.analysis.talents.tree_of_life import TreeOfLifeAttributor

TOL_BUFF = 33891


def make_applybuff(ts, ability, target=1):
    return {"timestamp": ts, "type": "applybuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_removebuff(ts, ability, target=1):
    return {"timestamp": ts, "type": "removebuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_heal(ts, ability, amount, target=2, overheal=0):
    return {"timestamp": ts, "type": "heal", "sourceID": 1, "targetID": target,
            "abilityGameID": ability, "amount": amount, "overheal": overheal, "hitType": 1}


def test_tol_rejuv_buff():
    events = [
        make_applybuff(100, TOL_BUFF),
        make_heal(200, 774, 15000),
        make_removebuff(500, TOL_BUFF),
    ]
    pipeline = Pipeline(attributors=[TreeOfLifeAttributor()])
    results = pipeline.run(events)
    # +50% = 15000 - 15000/1.5 = 5000
    assert results.talent_healing["Incarnation: Tree of Life"] == pytest.approx(5000.0)


def test_tol_germination_rejuv_buff():
    events = [
        make_applybuff(100, TOL_BUFF),
        make_heal(200, 155777, 15000),
        make_removebuff(500, TOL_BUFF),
    ]
    pipeline = Pipeline(attributors=[TreeOfLifeAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Incarnation: Tree of Life"] == pytest.approx(5000.0)


def test_tol_other_spell_buff():
    events = [
        make_applybuff(100, TOL_BUFF),
        make_heal(200, 8936, 11000),  # Regrowth
        make_removebuff(500, TOL_BUFF),
    ]
    pipeline = Pipeline(attributors=[TreeOfLifeAttributor()])
    results = pipeline.run(events)
    # +10% = 11000 - 11000/1.1 = 1000
    assert results.talent_healing["Incarnation: Tree of Life"] == pytest.approx(1000.0)


def test_tol_no_attribution_outside():
    events = [
        make_heal(50, 774, 10000),
        make_applybuff(100, TOL_BUFF),
        make_removebuff(200, TOL_BUFF),
        make_heal(300, 774, 10000),
    ]
    pipeline = Pipeline(attributors=[TreeOfLifeAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Incarnation: Tree of Life"] == 0.0


def test_tol_wg_base_buff():
    """WG during ToL gets at least the 10% base buff."""
    events = [
        make_applybuff(100, TOL_BUFF),
        # Single WG tick on one target
        make_heal(200, 48438, 10000, target=2),
        # Force flush by ending ToL
        make_removebuff(500, TOL_BUFF),
        # Need a non-WG heal after removebuff to flush the buffer
        make_heal(600, 8936, 0, target=2),
    ]
    pipeline = Pipeline(attributors=[TreeOfLifeAttributor()])
    results = pipeline.run(events)
    # 10% buff = 10000 - 10000/1.1 ~= 909.09
    assert results.talent_healing["Incarnation: Tree of Life"] == pytest.approx(909.09, rel=0.01)
