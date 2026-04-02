import pytest

from flourish.analysis.pipeline import Pipeline
from flourish.analysis.talents.reforestation import ReforestationAttributor

SWIFTMEND = 18562


def make_cast(ts, ability, target=1):
    return {"timestamp": ts, "type": "cast", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_heal(ts, ability, amount, target=2, overheal=0):
    return {"timestamp": ts, "type": "heal", "sourceID": 1, "targetID": target,
            "abilityGameID": ability, "amount": amount, "overheal": overheal, "hitType": 1}


def test_reforestation_triggers_after_4th_swiftmend():
    events = [
        make_cast(100, SWIFTMEND),
        make_cast(200, SWIFTMEND),
        make_cast(300, SWIFTMEND),
        make_cast(400, SWIFTMEND),  # 4th SM triggers ToL
        make_heal(500, 8936, 11000),  # Regrowth during reforestation ToL
    ]
    pipeline = Pipeline(attributors=[ReforestationAttributor()])
    results = pipeline.run(events)
    # +10% = 11000 - 11000/1.1 = 1000
    assert results.talent_healing["Reforestation"] == pytest.approx(1000.0)


def test_reforestation_rejuv_gets_50pct():
    events = [
        make_cast(100, SWIFTMEND),
        make_cast(200, SWIFTMEND),
        make_cast(300, SWIFTMEND),
        make_cast(400, SWIFTMEND),
        make_heal(500, 774, 15000),  # Rejuv during reforestation ToL
    ]
    pipeline = Pipeline(attributors=[ReforestationAttributor()])
    results = pipeline.run(events)
    # +50% = 15000 - 15000/1.5 = 5000
    assert results.talent_healing["Reforestation"] == pytest.approx(5000.0)


def test_reforestation_no_trigger_before_4th():
    events = [
        make_cast(100, SWIFTMEND),
        make_cast(200, SWIFTMEND),
        make_cast(300, SWIFTMEND),
        # Only 3 SMs
        make_heal(400, 774, 10000),
    ]
    pipeline = Pipeline(attributors=[ReforestationAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Reforestation"] == 0.0


def test_reforestation_expires_after_10sec():
    events = [
        make_cast(100, SWIFTMEND),
        make_cast(200, SWIFTMEND),
        make_cast(300, SWIFTMEND),
        make_cast(400, SWIFTMEND),  # Triggers at t=400, expires at t=10400
        make_heal(10500, 774, 10000),  # After expiry
    ]
    pipeline = Pipeline(attributors=[ReforestationAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Reforestation"] == 0.0


def test_reforestation_triggers_again_at_8th():
    events = [
        # First 4 SMs
        make_cast(100, SWIFTMEND),
        make_cast(200, SWIFTMEND),
        make_cast(300, SWIFTMEND),
        make_cast(400, SWIFTMEND),
        # Window expires
        # Next 4 SMs
        make_cast(20000, SWIFTMEND),
        make_cast(20100, SWIFTMEND),
        make_cast(20200, SWIFTMEND),
        make_cast(20300, SWIFTMEND),  # 8th SM triggers again
        make_heal(20400, 8936, 11000),
    ]
    pipeline = Pipeline(attributors=[ReforestationAttributor()])
    results = pipeline.run(events)
    # +10% = 1000
    assert results.talent_healing["Reforestation"] == pytest.approx(1000.0)


TOL_BUFF = 33891


def make_applybuff(ts, ability, target=1):
    return {"timestamp": ts, "type": "applybuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_removebuff(ts, ability, target=1):
    return {"timestamp": ts, "type": "removebuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def test_reforestation_no_trigger_during_real_tol():
    """Reforestation should NOT trigger if real ToL is already active."""
    events = [
        make_applybuff(50, TOL_BUFF),  # Real ToL active
        make_cast(100, SWIFTMEND),
        make_cast(200, SWIFTMEND),
        make_cast(300, SWIFTMEND),
        make_cast(400, SWIFTMEND),  # 4th SM — but real ToL is active, so no Reforestation trigger
        make_removebuff(450, TOL_BUFF),
        make_heal(500, 8936, 11000),  # After real ToL ends — should not be attributed
    ]
    pipeline = Pipeline(attributors=[ReforestationAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Reforestation"] == 0.0


def test_reforestation_no_attribution_during_real_tol():
    """Even if Reforestation window is active, real ToL takes priority."""
    events = [
        make_cast(100, SWIFTMEND),
        make_cast(200, SWIFTMEND),
        make_cast(300, SWIFTMEND),
        make_cast(400, SWIFTMEND),  # Triggers Reforestation ToL (no real ToL active)
        make_applybuff(500, TOL_BUFF),  # Real ToL activates during Reforestation window
        make_heal(600, 8936, 11000),  # During real ToL — Reforestation should NOT claim this
        make_removebuff(700, TOL_BUFF),
    ]
    pipeline = Pipeline(attributors=[ReforestationAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Reforestation"] == 0.0
