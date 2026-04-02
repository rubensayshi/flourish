from flourish.analysis.pipeline import Pipeline
from flourish.analysis.talents.direct_spells import (
    EverbloomSplashAttributor,
    EfflorescenceAttributor,
    VerdancyAttributor,
)
from flourish.analysis.talents.keeper.direct_spells import (
    GroveGuardiansAttributor,
    DreamSurgeAttributor,
)


def make_heal(ts, ability, amount, overheal=0):
    return {
        "timestamp": ts,
        "type": "heal",
        "sourceID": 1,
        "targetID": 2,
        "abilityGameID": ability,
        "amount": amount,
        "overheal": overheal,
        "hitType": 1,
    }


def test_everbloom_attributes_all_healing():
    events = [make_heal(100, 1244341, 5000)]
    pipeline = Pipeline(attributors=[EverbloomSplashAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Everbloom: Splash"] == 5000.0


def test_grove_guardians_attributes_nourish():
    events = [make_heal(100, 422090, 3000)]
    pipeline = Pipeline(attributors=[GroveGuardiansAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Grove Guardians"] == 3000.0


def test_dream_surge_attributes_dream_bloom():
    events = [make_heal(100, 434141, 2000)]
    pipeline = Pipeline(attributors=[DreamSurgeAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Dream Surge"] == 2000.0


def test_direct_spell_ignores_unrelated_spells():
    events = [make_heal(100, 774, 10000)]  # Rejuvenation
    pipeline = Pipeline(attributors=[EverbloomSplashAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Everbloom: Splash"] == 0.0


def test_direct_spell_skips_wasted_heals():
    events = [make_heal(100, 1244341, 2000, overheal=3000)]  # 60% OH
    pipeline = Pipeline(attributors=[EverbloomSplashAttributor()])
    results = pipeline.run(events)
    assert results.talent_healing["Everbloom: Splash"] == 0.0


def test_multiple_direct_attributors():
    events = [
        make_heal(100, 1244341, 5000),  # Everbloom
        make_heal(200, 422090, 3000),  # Nourish
        make_heal(300, 434141, 2000),  # Dream Bloom
        make_heal(400, 81269, 1000),  # Efflorescence
        make_heal(500, 774, 8000),  # Rejuvenation (unattributed)
    ]
    pipeline = Pipeline(
        attributors=[
            EverbloomSplashAttributor(),
            GroveGuardiansAttributor(),
            DreamSurgeAttributor(),
            EfflorescenceAttributor(),
        ]
    )
    results = pipeline.run(events)
    assert results.talent_healing["Everbloom: Splash"] == 5000.0
    assert results.talent_healing["Grove Guardians"] == 3000.0
    assert results.talent_healing["Dream Surge"] == 2000.0
    assert results.talent_healing["Efflorescence"] == 1000.0
    assert results.total_healing == 19000
