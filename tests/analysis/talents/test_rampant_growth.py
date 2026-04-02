import pytest
from flourish.analysis.pipeline import Pipeline
from flourish.analysis.talents.direct_spells import RampantGrowthAttributor


REGROWTH = 8936


def make_heal(ts, ability, amount, overheal=0, tick=False):
    evt = {"timestamp": ts, "type": "heal", "sourceID": 1, "targetID": 2,
           "abilityGameID": ability, "amount": amount, "overheal": overheal, "hitType": 1}
    if tick:
        evt["tick"] = True
    return evt


def test_rampant_growth_attributes_bonus_on_regrowth_ticks():
    """100% HoT increase → credit half of each tick."""
    events = [make_heal(100, REGROWTH, 10000, tick=True)]
    pipeline = Pipeline(attributors=[RampantGrowthAttributor()])
    results = pipeline.run(events)
    # bonus = 10000 - 10000/2.0 = 5000
    assert results.talent_healing["Rampant Growth"] == pytest.approx(5000.0, rel=0.01)


def test_rampant_growth_ignores_regrowth_direct_heals():
    """Direct (non-tick) Regrowth heals are unaffected."""
    events = [make_heal(100, REGROWTH, 10000, tick=False)]
    pipeline = Pipeline(attributors=[RampantGrowthAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Rampant Growth"] == 0.0


def test_rampant_growth_ignores_other_spells():
    events = [make_heal(100, 774, 10000, tick=True)]  # Rejuv tick
    pipeline = Pipeline(attributors=[RampantGrowthAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Rampant Growth"] == 0.0


def test_rampant_growth_skips_wasted_heals():
    events = [make_heal(100, REGROWTH, 2000, overheal=3000, tick=True)]
    pipeline = Pipeline(attributors=[RampantGrowthAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Rampant Growth"] == 0.0
