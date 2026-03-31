import pytest

from rdruid_analyzer.analysis.pipeline import Pipeline
from rdruid_analyzer.analysis.talents.soul_of_the_forest import SoulOfTheForestAttributor

SWIFTMEND = 18562
REJUV = 774
REGROWTH = 8936
TARGET = 10


def make_cast(ts, ability, target=TARGET):
    return {"timestamp": ts, "type": "cast", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_apply(ts, ability, target=TARGET):
    return {"timestamp": ts, "type": "applybuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_heal(ts, ability, amount, target=TARGET, overheal=0):
    return {
        "timestamp": ts,
        "type": "heal",
        "sourceID": 1,
        "targetID": target,
        "abilityGameID": ability,
        "amount": amount,
        "overheal": overheal,
        "hitType": 1,
    }


def test_sotf_attributes_bonus_from_rejuv():
    events = [
        make_cast(100, SWIFTMEND),
        make_apply(150, REJUV),
        make_heal(200, REJUV, 10000),
    ]
    pipeline = Pipeline(attributors=[SoulOfTheForestAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Soul of the Forest"] == pytest.approx(3750.0)


def test_sotf_does_not_attribute_unbuffed_rejuv():
    events = [
        make_apply(100, REJUV),
        make_heal(200, REJUV, 10000),
    ]
    pipeline = Pipeline(attributors=[SoulOfTheForestAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Soul of the Forest"] == 0.0


def test_sotf_only_applies_to_first_hot_after_swiftmend():
    events = [
        make_cast(100, SWIFTMEND),
        make_apply(150, REJUV, target=10),
        make_apply(160, REJUV, target=20),
        make_heal(200, REJUV, 10000, target=10),
        make_heal(210, REJUV, 10000, target=20),
    ]
    pipeline = Pipeline(attributors=[SoulOfTheForestAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Soul of the Forest"] == pytest.approx(3750.0)
