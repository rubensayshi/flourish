import pytest
from rdruid_analyzer.models.events import parse_event, HealEvent, CastEvent, ApplyBuffEvent


def test_parse_heal_event():
    raw = {
        "timestamp": 1000,
        "type": "heal",
        "sourceID": 1,
        "targetID": 2,
        "abilityGameID": 774,  # Rejuvenation
        "amount": 5000,
        "overheal": 1000,
        "hitType": 1,
    }
    event = parse_event(raw)
    assert isinstance(event, HealEvent)
    assert event.amount == 5000
    assert event.overheal == 1000
    assert event.raw_heal == 6000
    assert event.overheal_pct == pytest.approx(1000 / 6000)


def test_parse_cast_event():
    raw = {
        "timestamp": 1000,
        "type": "cast",
        "sourceID": 1,
        "targetID": 2,
        "abilityGameID": 18562,  # Swiftmend
    }
    event = parse_event(raw)
    assert isinstance(event, CastEvent)
    assert event.ability_id == 18562


def test_parse_applybuff_event():
    raw = {
        "timestamp": 1000,
        "type": "applybuff",
        "sourceID": 1,
        "targetID": 2,
        "abilityGameID": 774,
    }
    event = parse_event(raw)
    assert isinstance(event, ApplyBuffEvent)


def test_parse_unknown_event_returns_none():
    raw = {"timestamp": 1000, "type": "totally_unknown", "sourceID": 1}
    event = parse_event(raw)
    assert event is None


def test_heal_event_is_wasted():
    raw = {
        "timestamp": 1000, "type": "heal", "sourceID": 1,
        "targetID": 2, "abilityGameID": 774,
        "amount": 2000, "overheal": 3000, "hitType": 1,
    }
    event = parse_event(raw)
    assert event.is_wasted  # 3000/5000 = 60% > 50%


def test_heal_event_not_wasted():
    raw = {
        "timestamp": 1000, "type": "heal", "sourceID": 1,
        "targetID": 2, "abilityGameID": 774,
        "amount": 4000, "overheal": 1000, "hitType": 1,
    }
    event = parse_event(raw)
    assert not event.is_wasted  # 1000/5000 = 20% < 50%


def test_heal_event_absorb_included_in_raw_heal():
    raw = {
        "timestamp": 1000, "type": "heal", "sourceID": 1,
        "targetID": 2, "abilityGameID": 774,
        "amount": 5000, "overheal": 1000, "absorb": 500, "hitType": 1,
    }
    event = parse_event(raw)
    assert event.absorb == 500
    assert event.raw_heal == 6500  # 5000 + 1000 + 500


def test_heal_event_absorb_defaults_to_zero():
    raw = {
        "timestamp": 1000, "type": "heal", "sourceID": 1,
        "targetID": 2, "abilityGameID": 774,
        "amount": 5000, "overheal": 1000, "hitType": 1,
    }
    event = parse_event(raw)
    assert event.absorb == 0
    assert event.raw_heal == 6000
