import pytest

from rdruid_analyzer.analysis.pipeline import Pipeline
from rdruid_analyzer.analysis.talents.sm_cooldown_reduction import (
    SmCooldownReductionAttributor,
    DRYADS_DANCE_NODE_ID,
    EARLY_SPRING_NODE_ID,
    EARLY_SPRING_TALENT_ID,
)

SWIFTMEND = 18562
DRYAD_TRANQ = 1264659


def make_cast(ts, ability, source=1, target=2):
    return {"timestamp": ts, "type": "cast", "sourceID": source, "targetID": target, "abilityGameID": ability}


def make_heal(ts, ability, amount, source=1, target=2, overheal=0):
    return {"timestamp": ts, "type": "heal", "sourceID": source, "targetID": target,
            "abilityGameID": ability, "amount": amount, "overheal": overheal, "hitType": 1}


def make_combatant_info(ts, source=1, talent_nodes=None, talent_ids=None):
    """Helper to create combatantinfo events with proper talentTree structure."""
    tree = []
    for nid in (talent_nodes or []):
        tree.append({"nodeID": nid, "id": nid})
    for tid in (talent_ids or []):
        if tid not in (talent_nodes or []):
            tree.append({"nodeID": 0, "id": tid})
    return {
        "timestamp": ts, "type": "combatantinfo", "sourceID": source,
        "talentTree": tree,
        "critSpell": 0, "hasteSpell": 0, "mastery": 0, "specID": 105,
    }


def test_tracks_sm_casts():
    attr = SmCooldownReductionAttributor()
    pipeline = Pipeline(attributors=[attr])
    events = [
        make_cast(0, SWIFTMEND),
        make_cast(12000, SWIFTMEND),
        make_cast(24000, SWIFTMEND),
    ]
    pipeline.run(events)
    assert attr._sm_cast_timestamps == [0, 12000, 24000]


def test_tracks_dryad_windows_from_pet_heals():
    attr = SmCooldownReductionAttributor()
    pipeline = Pipeline(attributors=[attr])
    events = [
        make_combatant_info(0, talent_nodes=[DRYADS_DANCE_NODE_ID, EARLY_SPRING_NODE_ID],
                           talent_ids=[EARLY_SPRING_TALENT_ID]),
        make_heal(1000, DRYAD_TRANQ, 500, source=99),
        make_heal(1500, DRYAD_TRANQ, 500, source=99),
        make_heal(2000, DRYAD_TRANQ, 500, source=99),
        # Gap > 2s -> window closes on next event
        make_heal(10000, DRYAD_TRANQ, 500, source=99),
        make_heal(10500, DRYAD_TRANQ, 500, source=99),
    ]
    pipeline.run(events)
    assert len(attr._dryad_windows) == 2
    assert attr._dryad_windows[0] == (1000, 2000)
    assert attr._dryad_windows[1] == (10000, 10500)


def test_dryad_window_closes_in_finalize():
    """Open Dryad window at end of fight should be closed in finalize."""
    attr = SmCooldownReductionAttributor()
    pipeline = Pipeline(attributors=[attr])
    events = [
        make_combatant_info(0, talent_nodes=[DRYADS_DANCE_NODE_ID, EARLY_SPRING_NODE_ID],
                           talent_ids=[EARLY_SPRING_TALENT_ID]),
        make_heal(1000, DRYAD_TRANQ, 500, source=99),
        make_heal(1500, DRYAD_TRANQ, 500, source=99),
        # Fight ends without gap -> finalize closes window
    ]
    pipeline.run(events)
    assert len(attr._dryad_windows) == 1
    assert attr._dryad_windows[0] == (1000, 1500)


def test_ignores_player_source_heals():
    """Heals from Dryad spells but from player source should be ignored."""
    attr = SmCooldownReductionAttributor()
    pipeline = Pipeline(attributors=[attr])
    events = [
        make_combatant_info(0, source=1, talent_nodes=[DRYADS_DANCE_NODE_ID, EARLY_SPRING_NODE_ID],
                           talent_ids=[EARLY_SPRING_TALENT_ID]),
        make_heal(1000, DRYAD_TRANQ, 500, source=1),  # Player source, not pet
    ]
    pipeline.run(events)
    assert len(attr._dryad_windows) == 0
