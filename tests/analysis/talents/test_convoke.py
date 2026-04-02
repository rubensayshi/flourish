import pytest

from flourish.analysis.pipeline import Pipeline
from flourish.analysis.talents.convoke import ConvokeAttributor

CONVOKE = 391528
CONVOKE_LEGACY = 323764


def make_cast(ts, ability, target=1):
    return {"timestamp": ts, "type": "cast", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_heal(ts, ability, amount, target=2, overheal=0):
    return {"timestamp": ts, "type": "heal", "sourceID": 1, "targetID": target,
            "abilityGameID": ability, "amount": amount, "overheal": overheal, "hitType": 1}


def test_convoke_attributes_during_channel():
    events = [
        make_cast(1000, CONVOKE),
        make_heal(1500, 774, 10000),   # During channel
        make_heal(2000, 8936, 5000),   # During channel
    ]
    pipeline = Pipeline(attributors=[ConvokeAttributor()])
    results = pipeline.run(events)
    # 70% of (10000 + 5000) = 10500
    assert results.talent_healing["Convoke the Spirits"] == pytest.approx(10500.0)


def test_convoke_legacy_spell_id():
    """WCL may log Convoke cast with the legacy Shadowlands spell ID 323764."""
    events = [
        make_cast(1000, CONVOKE_LEGACY),
        make_heal(1500, 774, 10000),
    ]
    pipeline = Pipeline(attributors=[ConvokeAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Convoke the Spirits"] == pytest.approx(7000.0)


def test_convoke_no_attribution_outside_channel():
    events = [
        make_heal(500, 774, 10000),    # Before cast
        make_cast(1000, CONVOKE),
        make_heal(6000, 774, 10000),   # After 4s window (1000+4000=5000)
    ]
    pipeline = Pipeline(attributors=[ConvokeAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Convoke the Spirits"] == 0.0


def test_convoke_boundary_at_window_end():
    events = [
        make_cast(1000, CONVOKE),
        make_heal(5000, 774, 10000),   # Exactly at channel end (1000+4000)
    ]
    pipeline = Pipeline(attributors=[ConvokeAttributor()])
    results = pipeline.run(events)
    # At boundary (<=), should be attributed
    assert results.talent_healing["Convoke the Spirits"] == pytest.approx(7000.0)


def test_convoke_custom_ratio():
    events = [
        make_cast(1000, CONVOKE),
        make_heal(1500, 774, 10000),
    ]
    pipeline = Pipeline(attributors=[ConvokeAttributor(healing_ratio=0.5)])
    results = pipeline.run(events)
    assert results.talent_healing["Convoke the Spirits"] == pytest.approx(5000.0)


def make_applybuff(ts, ability, target=2):
    return {"timestamp": ts, "type": "applybuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def test_convoke_tags_hot_applied_during_channel():
    """HoT applied during Convoke channel is tagged; ticks after channel are attributed."""
    events = [
        make_cast(1000, CONVOKE),
        make_applybuff(1200, 774, target=5),      # Rejuv applied during channel
        make_heal(1500, 774, 3000, target=5),      # Tick during channel — tagged HoT
        make_heal(6000, 774, 3000, target=5),      # Tick AFTER channel — still attributed via tag
    ]
    pipeline = Pipeline(attributors=[ConvokeAttributor()])
    results = pipeline.run(events)
    # Both ticks attributed at 70%: 2 * 3000 * 0.7 = 4200
    assert results.talent_healing["Convoke the Spirits"] == pytest.approx(4200.0)


def test_convoke_preexisting_hot_not_attributed():
    """Pre-existing HoT ticks during Convoke channel should NOT be attributed."""
    events = [
        make_applybuff(100, 774, target=5),        # Rejuv applied before Convoke
        make_cast(1000, CONVOKE),
        make_heal(1500, 774, 10000, target=5),     # Tick during channel — but HoT existed before
    ]
    pipeline = Pipeline(attributors=[ConvokeAttributor()])
    results = pipeline.run(events)
    # HoT was tracked (not None) but NOT tagged by Convoke — should be 0
    assert results.talent_healing["Convoke the Spirits"] == 0.0
