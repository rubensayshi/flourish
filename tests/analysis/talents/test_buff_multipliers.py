import pytest
from flourish.analysis.pipeline import Pipeline
from flourish.analysis.talents.buff_multipliers import (
    StaticBuffAttributor, LifetreadingAttributor,
)
from flourish.analysis.talents.keeper.buff_multipliers import WildSynthesisAttributor
from flourish.analysis.talents.wildstalker.buff_multipliers import WildstalkersPowerAttributor


def make_heal(ts, ability, amount, overheal=0):
    return {"timestamp": ts, "type": "heal", "sourceID": 1, "targetID": 2,
            "abilityGameID": ability, "amount": amount, "overheal": overheal, "hitType": 1}


def test_static_buff_attributes_bonus_portion():
    """Wild Synthesis: +30% to Nourish (422090)"""
    events = [make_heal(100, 422090, 13000)]  # 13000 effective
    pipeline = Pipeline(attributors=[WildSynthesisAttributor()])
    results = pipeline.run(events)
    # bonus = 13000 - 13000/1.3 = 3000
    assert results.talent_healing["Wild Synthesis"] == pytest.approx(3000.0, rel=0.01)


def test_wildstalkers_power_on_rejuv():
    events = [make_heal(100, 774, 11000)]
    pipeline = Pipeline(attributors=[WildstalkersPowerAttributor()])
    results = pipeline.run(events)
    # bonus = 11000 - 11000/1.1 = 1000
    assert results.talent_healing["Wildstalker's Power"] == pytest.approx(1000.0, rel=0.01)


def test_static_buff_ignores_unrelated_spells():
    events = [make_heal(100, 999, 10000)]
    pipeline = Pipeline(attributors=[WildSynthesisAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Wild Synthesis"] == 0.0


def test_lifetreading_on_efflorescence():
    events = [make_heal(100, 81269, 12500)]
    pipeline = Pipeline(attributors=[LifetreadingAttributor()])
    results = pipeline.run(events)
    # bonus = 12500 - 12500/1.25 = 2500
    assert results.talent_healing["Lifetreading"] == pytest.approx(2500.0, rel=0.01)


def test_static_buff_skips_wasted():
    events = [make_heal(100, 422090, 2000, overheal=3000)]
    pipeline = Pipeline(attributors=[WildSynthesisAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Wild Synthesis"] == 0.0
