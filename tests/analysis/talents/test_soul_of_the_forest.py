import pytest

from rdruid_analyzer.analysis.pipeline import Pipeline
from rdruid_analyzer.analysis.talents.soul_of_the_forest import SoulOfTheForestAttributor

SWIFTMEND = 18562
REJUV = 774
REGROWTH = 8936
SOTF_BUFF = 114108
TARGET = 10
SPREAD_1 = 20
SPREAD_2 = 30


def make_cast(ts, ability, target=TARGET):
    return {"timestamp": ts, "type": "cast", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_apply(ts, ability, target=TARGET):
    return {"timestamp": ts, "type": "applybuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_remove(ts, ability, target=TARGET):
    return {"timestamp": ts, "type": "removebuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


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


# --- SotF basics ---


def test_sotf_attributes_bonus_from_rejuv():
    events = [
        make_cast(100, SWIFTMEND),
        make_apply(150, REJUV),
        make_heal(200, REJUV, 10000),
    ]
    pipeline = Pipeline(attributors=[SoulOfTheForestAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["SotF + Power of the Archdruid"] == pytest.approx(3750.0)


def test_sotf_does_not_attribute_unbuffed_rejuv():
    events = [
        make_apply(100, REJUV),
        make_heal(200, REJUV, 10000),
    ]
    pipeline = Pipeline(attributors=[SoulOfTheForestAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["SotF + Power of the Archdruid"] == 0.0


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
    # Only the first rejuv gets SotF bonus
    assert results.talent_healing["SotF + Power of the Archdruid"] == pytest.approx(3750.0)


# --- PotA spread (SotF consumed → spreads to 2 extra targets) ---


def test_pota_spreads_get_full_attribution():
    """PotA spread HoTs are 100% attributed since they wouldn't exist without PotA."""
    events = [
        make_cast(100, SWIFTMEND),
        # SotF-empowered Rejuv on primary target
        make_cast(200, REJUV, target=TARGET),
        make_apply(200, REJUV, target=TARGET),
        make_remove(200, SOTF_BUFF, target=1),  # SotF consumed
        # PotA spreads within window
        make_apply(210, REJUV, target=SPREAD_1),
        make_apply(220, REJUV, target=SPREAD_2),
        # Heals
        make_heal(500, REJUV, 10000, target=TARGET),     # primary: SotF bonus only
        make_heal(510, REJUV, 10000, target=SPREAD_1),   # spread: 100%
        make_heal(520, REJUV, 10000, target=SPREAD_2),   # spread: 100%
    ]
    pipeline = Pipeline(attributors=[SoulOfTheForestAttributor()])
    results = pipeline.run(events)
    # Primary: 3750 (60% bonus), Spread1: 10000, Spread2: 10000
    assert results.talent_healing["SotF + Power of the Archdruid"] == pytest.approx(23750.0)


def test_pota_spread_outside_window_not_attributed():
    """Spread ApplyBuff after the 500ms window should not be tagged."""
    events = [
        make_cast(100, SWIFTMEND),
        make_cast(200, REJUV, target=TARGET),
        make_apply(200, REJUV, target=TARGET),
        make_remove(200, SOTF_BUFF, target=1),
        # Spread outside window
        make_apply(800, REJUV, target=SPREAD_1),
        make_heal(1000, REJUV, 10000, target=TARGET),
        make_heal(1010, REJUV, 10000, target=SPREAD_1),
    ]
    pipeline = Pipeline(attributors=[SoulOfTheForestAttributor()])
    results = pipeline.run(events)
    # Only primary SotF bonus
    assert results.talent_healing["SotF + Power of the Archdruid"] == pytest.approx(3750.0)


def test_pota_spread_wrong_spell_not_attributed():
    """PotA only spreads the same spell (Rejuv→Rejuv, not Rejuv→Regrowth)."""
    events = [
        make_cast(100, SWIFTMEND),
        make_cast(200, REJUV, target=TARGET),
        make_apply(200, REJUV, target=TARGET),
        make_remove(200, SOTF_BUFF, target=1),
        # Different spell within window
        make_apply(210, REGROWTH, target=SPREAD_1),
        make_heal(500, REJUV, 10000, target=TARGET),
        make_heal(510, REGROWTH, 10000, target=SPREAD_1),
    ]
    pipeline = Pipeline(attributors=[SoulOfTheForestAttributor()])
    results = pipeline.run(events)
    # Only primary SotF bonus
    assert results.talent_healing["SotF + Power of the Archdruid"] == pytest.approx(3750.0)
