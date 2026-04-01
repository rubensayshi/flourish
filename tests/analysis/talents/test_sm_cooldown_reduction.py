import pytest

from rdruid_analyzer.analysis.pipeline import Pipeline
from rdruid_analyzer.analysis.talents.sm_cooldown_reduction import (
    SmCooldownReductionAttributor,
    WgCooldownReductionAttributor,
    compute_effective_wg_cd,
    DRYADS_DANCE_NODE_ID,
    EARLY_SPRING_NODE_ID,
    EARLY_SPRING_TALENT_ID,
    RENEWING_SURGE_NODE_ID,
)
from rdruid_analyzer.analysis.talents.soul_of_the_forest import SoulOfTheForestAttributor
from rdruid_analyzer.analysis.talents.direct_spells import GroveGuardiansAttributor

SWIFTMEND = 18562
WILD_GROWTH = 48438
DRYAD_TRANQ = 1264659
SOTF_BUFF = 114108
REJUV = 774
GG_NOURISH = 422090


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


# --- compute_effective_cd tests ---

from rdruid_analyzer.analysis.talents.sm_cooldown_reduction import compute_effective_cd


def test_effective_cd_baseline_with_renewing_surge():
    cd = compute_effective_cd(has_renewing_surge=True, has_early_spring=False, dryad_overlap_ms=0)
    assert cd == pytest.approx(12075.0)


def test_effective_cd_with_early_spring():
    cd = compute_effective_cd(has_renewing_surge=True, has_early_spring=True, dryad_overlap_ms=0)
    assert cd == pytest.approx(11075.0)


def test_effective_cd_with_dryad_full_overlap():
    cd = compute_effective_cd(has_renewing_surge=True, has_early_spring=True, dryad_overlap_ms=11075)
    assert cd == pytest.approx(8860.0)


def test_effective_cd_with_dryad_partial_overlap():
    cd = compute_effective_cd(has_renewing_surge=True, has_early_spring=True, dryad_overlap_ms=5000)
    # remaining = 11075 - 5000 = 6075, dryad_wait = 5000/1.25 = 4000, total = 10075
    assert cd == pytest.approx(10075.0)


def test_effective_cd_no_renewing_surge():
    cd = compute_effective_cd(has_renewing_surge=False, has_early_spring=True, dryad_overlap_ms=0)
    assert cd == pytest.approx(14000.0)


# --- finalize tests ---


def make_applybuff(ts, ability, source=1, target=1):
    return {"timestamp": ts, "type": "applybuff", "sourceID": source, "targetID": target, "abilityGameID": ability}


def make_removebuff(ts, ability, source=1, target=1):
    return {"timestamp": ts, "type": "removebuff", "sourceID": source, "targetID": target, "abilityGameID": ability}


def test_full_attribution_on_cooldown():
    """SM pressed on cooldown -> attributes fraction of downstream healing."""
    sotf = SoulOfTheForestAttributor()
    gg = GroveGuardiansAttributor()
    sm_cd = SmCooldownReductionAttributor(downstream_attributors=[sotf, gg])

    # With Early Spring + Renewing Surge (no Dryad's Dance):
    # reduced_cd = 11075ms, unreduced_cd = 12075ms
    # ratio = 1 - 11075/12075 ≈ 0.0828
    events = [
        make_combatant_info(0, talent_nodes=[
            EARLY_SPRING_NODE_ID, RENEWING_SURGE_NODE_ID,
            82055,  # SotF node
            82043,  # GG node
        ], talent_ids=[EARLY_SPRING_TALENT_ID]),
        # SM 1
        make_cast(1000, SWIFTMEND),
        make_applybuff(1001, SOTF_BUFF),
        make_cast(1002, REJUV, target=3),
        make_applybuff(1003, REJUV, target=3),
        make_removebuff(1004, SOTF_BUFF),
        make_heal(1100, REJUV, 10000, target=3),
        make_heal(1200, GG_NOURISH, 5000, source=99),
        # SM 2 — gap=11000 < 11075+1500 → on cooldown
        make_cast(12000, SWIFTMEND),
        make_applybuff(12001, SOTF_BUFF),
        make_cast(12002, REJUV, target=4),
        make_applybuff(12003, REJUV, target=4),
        make_removebuff(12004, SOTF_BUFF),
        make_heal(12100, REJUV, 10000, target=4),
        make_heal(12200, GG_NOURISH, 5000, source=99),
        # SM 3 — gap=11000 < 11075+1500 → on cooldown
        make_cast(23000, SWIFTMEND),
        make_applybuff(23001, SOTF_BUFF),
        make_cast(23002, REJUV, target=5),
        make_applybuff(23003, REJUV, target=5),
        make_removebuff(23004, SOTF_BUFF),
        make_heal(23100, REJUV, 10000, target=5),
        make_heal(23200, GG_NOURISH, 5000, source=99),
    ]
    # sm_cd must be LAST for finalize ordering
    pipeline = Pipeline(attributors=[sotf, gg, sm_cd])
    results = pipeline.run(events)

    # SotF: 3 * (10000 - 10000/1.6) = 3 * 3750 = 11250
    # GG: 3 * 5000 = 15000
    # downstream = 26250
    # 2 on-CD casts (gaps from cast 1→2 and 2→3), ratio ≈ 0.0828 each
    # extra_cast_fraction = (0.0828 * 2) / 3
    # attribution = fraction * 26250
    ratio = 1 - 11075.0 / 12075.0
    fraction = (ratio * 2) / 3
    expected = fraction * 26250
    assert results.talent_healing["SM Cooldown Reduction"] == pytest.approx(expected, rel=0.01)


def test_no_attribution_when_not_on_cooldown():
    """SM not pressed on cooldown -> no attribution."""
    sotf = SoulOfTheForestAttributor()
    sm_cd = SmCooldownReductionAttributor(downstream_attributors=[sotf])

    events = [
        make_combatant_info(0, talent_nodes=[
            EARLY_SPRING_NODE_ID, RENEWING_SURGE_NODE_ID,
            82055,
        ], talent_ids=[EARLY_SPRING_TALENT_ID]),
        make_cast(1000, SWIFTMEND),
        make_applybuff(1001, SOTF_BUFF),
        make_cast(1002, REJUV, target=3),
        make_applybuff(1003, REJUV, target=3),
        make_removebuff(1004, SOTF_BUFF),
        make_heal(1100, REJUV, 10000, target=3),
        # SM 2 — gap=30000 >> effective CD + tolerance
        make_cast(31000, SWIFTMEND),
        make_applybuff(31001, SOTF_BUFF),
        make_cast(31002, REJUV, target=4),
        make_applybuff(31003, REJUV, target=4),
        make_removebuff(31004, SOTF_BUFF),
        make_heal(31100, REJUV, 10000, target=4),
    ]
    pipeline = Pipeline(attributors=[sotf, sm_cd])
    results = pipeline.run(events)
    assert results.talent_healing["SM Cooldown Reduction"] == 0.0


def test_no_attribution_single_sm_cast():
    """Single SM cast can't determine on-cooldown, no attribution."""
    sotf = SoulOfTheForestAttributor()
    sm_cd = SmCooldownReductionAttributor(downstream_attributors=[sotf])
    events = [
        make_combatant_info(0, talent_nodes=[
            EARLY_SPRING_NODE_ID, RENEWING_SURGE_NODE_ID, 82055,
        ], talent_ids=[EARLY_SPRING_TALENT_ID]),
        make_cast(1000, SWIFTMEND),
        make_applybuff(1001, SOTF_BUFF),
        make_cast(1002, REJUV, target=3),
        make_applybuff(1003, REJUV, target=3),
        make_removebuff(1004, SOTF_BUFF),
        make_heal(1100, REJUV, 10000, target=3),
    ]
    pipeline = Pipeline(attributors=[sotf, sm_cd])
    results = pipeline.run(events)
    assert results.talent_healing["SM Cooldown Reduction"] == 0.0


# --- WG CD reduction tests ---


def test_effective_wg_cd_with_4pc_and_early_spring():
    cd = compute_effective_wg_cd(has_early_spring=True, has_4pc=True)
    # 10000 - 2000 - 1000 = 7000
    assert cd == pytest.approx(7000.0)


def test_effective_wg_cd_with_4pc_no_early_spring():
    cd = compute_effective_wg_cd(has_early_spring=False, has_4pc=True)
    assert cd == pytest.approx(8000.0)


def test_effective_wg_cd_no_4pc_with_early_spring():
    cd = compute_effective_wg_cd(has_early_spring=True, has_4pc=False)
    assert cd == pytest.approx(9000.0)


def test_wg_tracks_casts():
    attr = WgCooldownReductionAttributor()
    pipeline = Pipeline(attributors=[attr])
    events = [
        make_cast(0, WILD_GROWTH),
        make_cast(8000, WILD_GROWTH),
        make_cast(16000, WILD_GROWTH),
    ]
    pipeline.run(events)
    assert attr._wg_cast_timestamps == [0, 8000, 16000]


def test_wg_attribution_on_cooldown():
    """WG pressed on cooldown with 4pc + Early Spring -> attributes fraction of GG."""
    gg = GroveGuardiansAttributor()
    wg_cd = WgCooldownReductionAttributor(downstream_attributors=[gg], has_4pc=True)

    # reduced_cd = 7000, unreduced_cd = 8000
    # ratio = 1 - 7000/8000 = 0.125
    GG_NODE = 82043
    events = [
        make_combatant_info(0, talent_nodes=[
            EARLY_SPRING_NODE_ID, GG_NODE,
        ], talent_ids=[EARLY_SPRING_TALENT_ID]),
        make_cast(1000, WILD_GROWTH),
        make_heal(1100, GG_NOURISH, 5000, source=99),
        # WG 2 — gap=7500 < 7000+1500 → on cooldown
        make_cast(8500, WILD_GROWTH),
        make_heal(8600, GG_NOURISH, 5000, source=99),
        # WG 3 — gap=7500 → on cooldown
        make_cast(16000, WILD_GROWTH),
        make_heal(16100, GG_NOURISH, 5000, source=99),
    ]
    pipeline = Pipeline(attributors=[gg, wg_cd])
    results = pipeline.run(events)

    # GG total: 3 * 5000 = 15000
    # 2 on-CD casts, ratio = 0.125 each
    # fraction = (0.125 * 2) / 3
    # attribution = fraction * 15000
    ratio = 1 - 7000.0 / 8000.0
    fraction = (ratio * 2) / 3
    expected = fraction * 15000
    assert results.talent_healing["WG Cooldown Reduction"] == pytest.approx(expected, rel=0.01)


def test_wg_no_attribution_when_not_on_cooldown():
    gg = GroveGuardiansAttributor()
    wg_cd = WgCooldownReductionAttributor(downstream_attributors=[gg], has_4pc=True)

    events = [
        make_combatant_info(0, talent_nodes=[
            EARLY_SPRING_NODE_ID, 82043,
        ], talent_ids=[EARLY_SPRING_TALENT_ID]),
        make_cast(1000, WILD_GROWTH),
        make_heal(1100, GG_NOURISH, 5000, source=99),
        # WG 2 — gap=30000 >> 7000+1500
        make_cast(31000, WILD_GROWTH),
        make_heal(31100, GG_NOURISH, 5000, source=99),
    ]
    pipeline = Pipeline(attributors=[gg, wg_cd])
    results = pipeline.run(events)
    assert results.talent_healing["WG Cooldown Reduction"] == 0.0
