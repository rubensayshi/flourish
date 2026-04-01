import pytest

from rdruid_analyzer.analysis.pipeline import Pipeline
from rdruid_analyzer.analysis.talents.soul_of_the_forest import SoulOfTheForestAttributor

SWIFTMEND = 18562
REJUV = 774
GERMINATION_REJUV = 155777
REGROWTH = 8936
SOTF_BUFF = 114108
PLAYER = 1  # source/self for buff events
TARGET = 10
SPREAD_1 = 20
SPREAD_2 = 30


def make_cast(ts, ability, target=TARGET):
    return {"timestamp": ts, "type": "cast", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_apply(ts, ability, target=TARGET):
    return {"timestamp": ts, "type": "applybuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_refresh(ts, ability, target=TARGET):
    return {"timestamp": ts, "type": "refreshbuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


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
        make_apply(100, SOTF_BUFF, target=PLAYER),
        make_cast(150, REJUV),
        make_apply(150, REJUV),
        make_remove(150, SOTF_BUFF, target=PLAYER),
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


def test_sotf_only_applies_to_consuming_cast():
    """Only the cast that consumes the SotF buff gets tagged, not a later Rejuv outside PotA window."""
    events = [
        make_cast(100, SWIFTMEND),
        make_apply(100, SOTF_BUFF, target=PLAYER),
        # SotF Rejuv on target 10
        make_cast(150, REJUV, target=10),
        make_apply(150, REJUV, target=10),
        make_remove(150, SOTF_BUFF, target=PLAYER),
        # Normal Rejuv on target 20 well outside PotA window (should NOT get SotF)
        make_apply(1000, REJUV, target=20),
        make_heal(1100, REJUV, 10000, target=10),
        make_heal(1110, REJUV, 10000, target=20),
    ]
    pipeline = Pipeline(attributors=[SoulOfTheForestAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["SotF + Power of the Archdruid"] == pytest.approx(3750.0)


def test_sotf_regrowth_direct_heal_attributed():
    """Regrowth direct heal fires before HoT is tagged — should still get SotF bonus."""
    events = [
        make_cast(100, SWIFTMEND),
        make_apply(100, SOTF_BUFF, target=PLAYER),
        make_cast(150, REGROWTH),
        # Direct heal fires before ApplyBuff
        make_heal(150, REGROWTH, 60000),
        make_apply(150, REGROWTH),
        make_remove(150, SOTF_BUFF, target=PLAYER),
        # HoT tick
        make_heal(200, REGROWTH, 5000),
    ]
    pipeline = Pipeline(attributors=[SoulOfTheForestAttributor()])
    results = pipeline.run(events)
    # Direct: 60000 * 0.375 = 22500, HoT tick: 5000 * 0.375 = 1875
    assert results.talent_healing["SotF + Power of the Archdruid"] == pytest.approx(22500.0 + 1875.0)


def test_sotf_germination_rejuv():
    """SotF Rejuv cast applies as Germination (155777) — should still get tagged."""
    events = [
        make_cast(100, SWIFTMEND),
        make_apply(100, SOTF_BUFF, target=PLAYER),
        make_cast(150, REJUV),
        make_apply(150, GERMINATION_REJUV),  # Applied as Germination
        make_remove(150, SOTF_BUFF, target=PLAYER),
        make_heal(200, GERMINATION_REJUV, 10000),
    ]
    pipeline = Pipeline(attributors=[SoulOfTheForestAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["SotF + Power of the Archdruid"] == pytest.approx(3750.0)


# --- PotA spread (SotF consumed → spreads to 2 extra targets) ---


def test_pota_spreads_get_full_attribution():
    """PotA spread HoTs are 100% attributed since they wouldn't exist without PotA."""
    events = [
        make_cast(100, SWIFTMEND),
        make_apply(100, SOTF_BUFF, target=PLAYER),
        make_cast(200, REJUV, target=TARGET),
        make_apply(200, REJUV, target=TARGET),
        make_remove(200, SOTF_BUFF, target=PLAYER),
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
        make_apply(100, SOTF_BUFF, target=PLAYER),
        make_cast(200, REJUV, target=TARGET),
        make_apply(200, REJUV, target=TARGET),
        make_remove(200, SOTF_BUFF, target=PLAYER),
        # Spread outside window
        make_apply(800, REJUV, target=SPREAD_1),
        make_heal(1000, REJUV, 10000, target=TARGET),
        make_heal(1010, REJUV, 10000, target=SPREAD_1),
    ]
    pipeline = Pipeline(attributors=[SoulOfTheForestAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["SotF + Power of the Archdruid"] == pytest.approx(3750.0)


def test_pota_spread_wrong_spell_not_attributed():
    """PotA only spreads the same spell (Rejuv→Rejuv, not Rejuv→Regrowth)."""
    events = [
        make_cast(100, SWIFTMEND),
        make_apply(100, SOTF_BUFF, target=PLAYER),
        make_cast(200, REJUV, target=TARGET),
        make_apply(200, REJUV, target=TARGET),
        make_remove(200, SOTF_BUFF, target=PLAYER),
        # Different spell within window
        make_apply(210, REGROWTH, target=SPREAD_1),
        make_heal(500, REJUV, 10000, target=TARGET),
        make_heal(510, REGROWTH, 10000, target=SPREAD_1),
    ]
    pipeline = Pipeline(attributors=[SoulOfTheForestAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["SotF + Power of the Archdruid"] == pytest.approx(3750.0)


def test_sotf_tag_cleared_on_refresh():
    """Normal Rejuv refresh clears SotF tag — subsequent ticks should not be attributed."""
    events = [
        make_cast(100, SWIFTMEND),
        make_apply(100, SOTF_BUFF, target=PLAYER),
        make_cast(150, REJUV),
        make_apply(150, REJUV),
        make_remove(150, SOTF_BUFF, target=PLAYER),
        make_heal(200, REJUV, 10000),  # SotF tagged → attributed
        # Normal Rejuv refresh → clears SotF tag
        make_refresh(300, REJUV),
        make_heal(400, REJUV, 10000),  # NOT attributed
    ]
    pipeline = Pipeline(attributors=[SoulOfTheForestAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["SotF + Power of the Archdruid"] == pytest.approx(3750.0)


def test_interfering_applybuff_does_not_steal_sotf():
    """ApplyBuff events between Swiftmend and consuming cast should not steal the SotF tag."""
    events = [
        make_cast(100, SWIFTMEND),
        make_apply(100, SOTF_BUFF, target=PLAYER),
        # Interfering event: Dream Surge proc or similar applies a Rejuv on another target
        make_apply(120, REJUV, target=99),
        # Actual SotF-consuming cast
        make_cast(150, REJUV, target=TARGET),
        make_apply(150, REJUV, target=TARGET),
        make_remove(150, SOTF_BUFF, target=PLAYER),
        make_heal(200, REJUV, 10000, target=TARGET),
        make_heal(210, REJUV, 10000, target=99),
    ]
    pipeline = Pipeline(attributors=[SoulOfTheForestAttributor()])
    results = pipeline.run(events)
    # Only TARGET's Rejuv should have SotF, not target 99
    assert results.talent_healing["SotF + Power of the Archdruid"] == pytest.approx(3750.0)
