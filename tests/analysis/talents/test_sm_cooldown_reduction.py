import pytest

from flourish.analysis.pipeline import Pipeline
from flourish.analysis.talents.keeper.sm_cooldown_reduction import (
    SmCooldownReductionAttributor,
    WgCooldownReductionAttributor,
    compute_effective_wg_cd,
    DRYADS_DANCE_NODE_ID,
    EARLY_SPRING_NODE_ID,
    EARLY_SPRING_TALENT_ID,
    PROSPERITY_NODE_ID,
    PROSPERITY_TALENT_ID,
    RENEWING_SURGE_NODE_ID,
)
from flourish.analysis.talents.soul_of_the_forest import SoulOfTheForestAttributor
from flourish.analysis.talents.keeper.direct_spells import GroveGuardiansAttributor

SWIFTMEND = 18562
WILD_GROWTH = 48438
DRYAD_TRANQ = 1264659
SOTF_BUFF = 114108
REJUV = 774
GG_NOURISH = 422090


def make_cast(ts, ability, source=1, target=2):
    return {"timestamp": ts, "type": "cast", "sourceID": source, "targetID": target, "abilityGameID": ability}


def make_begincast(ts, ability, source=1, target=2):
    return {"timestamp": ts, "type": "begincast", "sourceID": source, "targetID": target, "abilityGameID": ability}


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
    pipeline = Pipeline(attributors=[attr], player_pet_ids={99})
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
    pipeline = Pipeline(attributors=[attr], player_pet_ids={99})
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

from flourish.analysis.talents.keeper.sm_cooldown_reduction import compute_effective_cd


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
    pipeline = Pipeline(attributors=[sotf, gg, sm_cd], player_pet_ids={99})
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
    assert results.talent_healing["Early Spring + Dryad's Dance"] == pytest.approx(expected, rel=0.01)


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
    assert results.talent_healing["Early Spring + Dryad's Dance"] == 0.0


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
    assert results.talent_healing["Early Spring + Dryad's Dance"] == 0.0


# --- 2-charge (Prosperity) tests ---


def test_two_charge_rapid_then_on_cooldown():
    """With Prosperity: 2 rapid casts, then 3rd pressed on cooldown when 1st charge returns."""
    sotf = SoulOfTheForestAttributor()
    gg = GroveGuardiansAttributor()
    sm_cd = SmCooldownReductionAttributor(downstream_attributors=[sotf, gg])

    # Early Spring + Renewing Surge + Prosperity: reduced_cd = 11075, unreduced_cd = 12075
    events = [
        make_combatant_info(0, talent_nodes=[
            EARLY_SPRING_NODE_ID, RENEWING_SURGE_NODE_ID,
            PROSPERITY_NODE_ID,
            82055, 82043,  # SotF, GG nodes
        ], talent_ids=[EARLY_SPRING_TALENT_ID, PROSPERITY_TALENT_ID]),
        # SM 1 — uses charge 1
        make_cast(1000, SWIFTMEND),
        make_applybuff(1001, SOTF_BUFF),
        make_cast(1002, REJUV, target=3),
        make_applybuff(1003, REJUV, target=3),
        make_removebuff(1004, SOTF_BUFF),
        make_heal(1100, REJUV, 10000, target=3),
        make_heal(1200, GG_NOURISH, 5000, source=99),
        # SM 2 — uses charge 2, gap=1000 (both charges available, NOT on CD)
        make_cast(2000, SWIFTMEND),
        make_applybuff(2001, SOTF_BUFF),
        make_cast(2002, REJUV, target=4),
        make_applybuff(2003, REJUV, target=4),
        make_removebuff(2004, SOTF_BUFF),
        make_heal(2100, REJUV, 10000, target=4),
        make_heal(2200, GG_NOURISH, 5000, source=99),
        # SM 3 — 1st charge recharges at 1000+11075=12075, cast at 12500 (425ms after)
        # Was depleted from t=2000..12075, cast within tolerance → on CD
        make_cast(12500, SWIFTMEND),
        make_applybuff(12501, SOTF_BUFF),
        make_cast(12502, REJUV, target=5),
        make_applybuff(12503, REJUV, target=5),
        make_removebuff(12504, SOTF_BUFF),
        make_heal(12600, REJUV, 10000, target=5),
        make_heal(12700, GG_NOURISH, 5000, source=99),
    ]
    pipeline = Pipeline(attributors=[sotf, gg, sm_cd], player_pet_ids={99})
    results = pipeline.run(events)

    # 1 on-CD cast out of 3 total
    ratio = 1 - 11075.0 / 12075.0
    fraction = ratio / 3
    downstream = 3 * (10000 - 10000 / 1.6) + 3 * 5000  # SotF + GG = 11250 + 15000
    expected = fraction * downstream
    assert results.talent_healing["Early Spring + Dryad's Dance"] == pytest.approx(expected, rel=0.01)


def test_two_charge_not_on_cooldown():
    """With Prosperity: 2 rapid casts, then 3rd cast long after charge returned → not on CD."""
    sotf = SoulOfTheForestAttributor()
    sm_cd = SmCooldownReductionAttributor(downstream_attributors=[sotf])

    events = [
        make_combatant_info(0, talent_nodes=[
            EARLY_SPRING_NODE_ID, RENEWING_SURGE_NODE_ID,
            PROSPERITY_NODE_ID, 82055,
        ], talent_ids=[EARLY_SPRING_TALENT_ID, PROSPERITY_TALENT_ID]),
        make_cast(1000, SWIFTMEND),
        make_applybuff(1001, SOTF_BUFF),
        make_cast(1002, REJUV, target=3),
        make_applybuff(1003, REJUV, target=3),
        make_removebuff(1004, SOTF_BUFF),
        make_heal(1100, REJUV, 10000, target=3),
        # SM 2 — use 2nd charge immediately
        make_cast(2000, SWIFTMEND),
        make_applybuff(2001, SOTF_BUFF),
        make_cast(2002, REJUV, target=4),
        make_applybuff(2003, REJUV, target=4),
        make_removebuff(2004, SOTF_BUFF),
        make_heal(2100, REJUV, 10000, target=4),
        # SM 3 — charge back at 12075, cast at 20000 (7925ms later, way past tolerance)
        make_cast(20000, SWIFTMEND),
        make_applybuff(20001, SOTF_BUFF),
        make_cast(20002, REJUV, target=5),
        make_applybuff(20003, REJUV, target=5),
        make_removebuff(20004, SOTF_BUFF),
        make_heal(20100, REJUV, 10000, target=5),
    ]
    pipeline = Pipeline(attributors=[sotf, sm_cd])
    results = pipeline.run(events)
    assert results.talent_healing["Early Spring + Dryad's Dance"] == 0.0


def test_two_charge_sustained_on_cooldown():
    """With Prosperity: sustained on-CD usage — use both charges, then press on CD repeatedly."""
    sotf = SoulOfTheForestAttributor()
    gg = GroveGuardiansAttributor()
    sm_cd = SmCooldownReductionAttributor(downstream_attributors=[sotf, gg])

    # reduced_cd = 11075, unreduced_cd = 12075
    # Charge timeline:
    #   t=1000: cast, charges 2→1, recharge queued at 1000+11075=12075
    #   t=2000: cast, charges 1→0, recharge queued at 12075+11075=23150
    #   t=12500: 1st charge back at 12075 (425ms ago) → on CD. charges 1→0. recharge at 23150+11075=34225? No...
    # Wait - when cast 3 happens, we consume the charge and schedule recharge.
    # recharge_start = pending[-1][0] if pending else cast_ts
    # After popping 12075 entry and before consuming: pending=[(23150, 11075)]
    # After consuming: schedule from pending[-1][0]=23150 → (23150+11075=34225, 11075)
    #   t=23500: 2nd charge back at 23150 (350ms ago) → on CD.
    events = [
        make_combatant_info(0, talent_nodes=[
            EARLY_SPRING_NODE_ID, RENEWING_SURGE_NODE_ID,
            PROSPERITY_NODE_ID, 82055, 82043,
        ], talent_ids=[EARLY_SPRING_TALENT_ID, PROSPERITY_TALENT_ID]),
        # SM 1
        make_cast(1000, SWIFTMEND),
        make_applybuff(1001, SOTF_BUFF),
        make_cast(1002, REJUV, target=3),
        make_applybuff(1003, REJUV, target=3),
        make_removebuff(1004, SOTF_BUFF),
        make_heal(1100, REJUV, 10000, target=3),
        make_heal(1200, GG_NOURISH, 5000, source=99),
        # SM 2 — both charges used
        make_cast(2000, SWIFTMEND),
        make_applybuff(2001, SOTF_BUFF),
        make_cast(2002, REJUV, target=4),
        make_applybuff(2003, REJUV, target=4),
        make_removebuff(2004, SOTF_BUFF),
        make_heal(2100, REJUV, 10000, target=4),
        make_heal(2200, GG_NOURISH, 5000, source=99),
        # SM 3 — on CD (charge back at 12075, cast at 12500)
        make_cast(12500, SWIFTMEND),
        make_applybuff(12501, SOTF_BUFF),
        make_cast(12502, REJUV, target=5),
        make_applybuff(12503, REJUV, target=5),
        make_removebuff(12504, SOTF_BUFF),
        make_heal(12600, REJUV, 10000, target=5),
        make_heal(12700, GG_NOURISH, 5000, source=99),
        # SM 4 — on CD (charge back at 23150, cast at 23500)
        make_cast(23500, SWIFTMEND),
        make_applybuff(23501, SOTF_BUFF),
        make_cast(23502, REJUV, target=6),
        make_applybuff(23503, REJUV, target=6),
        make_removebuff(23504, SOTF_BUFF),
        make_heal(23600, REJUV, 10000, target=6),
        make_heal(23700, GG_NOURISH, 5000, source=99),
    ]
    pipeline = Pipeline(attributors=[sotf, gg, sm_cd], player_pet_ids={99})
    results = pipeline.run(events)

    # 2 on-CD casts (3rd and 4th), 4 total
    ratio = 1 - 11075.0 / 12075.0
    fraction = (ratio * 2) / 4
    downstream = 4 * (10000 - 10000 / 1.6) + 4 * 5000  # 15000 + 20000
    expected = fraction * downstream
    assert results.talent_healing["Early Spring + Dryad's Dance"] == pytest.approx(expected, rel=0.01)


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
        make_begincast(0, WILD_GROWTH),
        make_begincast(8000, WILD_GROWTH),
        make_begincast(16000, WILD_GROWTH),
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
        make_begincast(1000, WILD_GROWTH),
        make_heal(1100, GG_NOURISH, 5000, source=99),
        # WG 2 — gap=7500 < 7000+1500 → on cooldown
        make_begincast(8500, WILD_GROWTH),
        make_heal(8600, GG_NOURISH, 5000, source=99),
        # WG 3 — gap=7500 → on cooldown
        make_begincast(16000, WILD_GROWTH),
        make_heal(16100, GG_NOURISH, 5000, source=99),
    ]
    pipeline = Pipeline(attributors=[gg, wg_cd], player_pet_ids={99})
    results = pipeline.run(events)

    # GG total: 3 * 5000 = 15000
    # 2 on-CD casts, ratio = 0.125 each
    # fraction = (0.125 * 2) / 3
    # attribution = fraction * 15000
    ratio = 1 - 7000.0 / 8000.0
    fraction = (ratio * 2) / 3
    expected = fraction * 15000
    assert results.talent_healing["Early Spring (WG)"] == pytest.approx(expected, rel=0.01)


def test_wg_no_attribution_when_not_on_cooldown():
    gg = GroveGuardiansAttributor()
    wg_cd = WgCooldownReductionAttributor(downstream_attributors=[gg], has_4pc=True)

    events = [
        make_combatant_info(0, talent_nodes=[
            EARLY_SPRING_NODE_ID, 82043,
        ], talent_ids=[EARLY_SPRING_TALENT_ID]),
        make_begincast(1000, WILD_GROWTH),
        make_heal(1100, GG_NOURISH, 5000, source=99),
        # WG 2 — gap=30000 >> 7000+1500
        make_begincast(31000, WILD_GROWTH),
        make_heal(31100, GG_NOURISH, 5000, source=99),
    ]
    pipeline = Pipeline(attributors=[gg, wg_cd], player_pet_ids={99})
    results = pipeline.run(events)
    assert results.talent_healing["Early Spring (WG)"] == 0.0
