import pytest

from rdruid_analyzer.analysis.pipeline import Pipeline
from rdruid_analyzer.analysis.talents.abundance import AbundanceAttributor
from rdruid_analyzer.analysis.talents.photosynthesis import PhotosynthesisAttributor
from rdruid_analyzer.analysis.talents.nurturing_dormancy import NurturingDormancyAttributor

TARGET = 10
TARGET2 = 20
REJUV = 774
GERMINATION_REJUV = 155777
REGROWTH = 8936
LIFEBLOOM = 33763
LIFEBLOOM_BLOOM = 33778


def make_apply(ts, ability, target=TARGET):
    return {"timestamp": ts, "type": "applybuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_remove(ts, ability, target=TARGET):
    return {"timestamp": ts, "type": "removebuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_refresh(ts, ability, target=TARGET):
    return {"timestamp": ts, "type": "refreshbuff", "sourceID": 1, "targetID": target, "abilityGameID": ability}


def make_heal(ts, ability, amount, target=TARGET, overheal=0, hit_type=1):
    return {
        "timestamp": ts, "type": "heal", "sourceID": 1, "targetID": target,
        "abilityGameID": ability, "amount": amount, "overheal": overheal, "hitType": hit_type,
    }


def make_combatant_info(ts=0, crit_spell=350.0):
    return {
        "timestamp": ts, "type": "combatantinfo", "sourceID": 1,
        "talentTree": [], "critSpell": crit_spell, "hasteSpell": 0, "mastery": 0, "specID": 105,
    }


# --- Abundance ---

class TestAbundance:
    def test_regrowth_crit_with_rejuvs(self):
        events = [
            make_combatant_info(0, crit_spell=350),  # base crit = 350/700 = 0.5
            make_apply(100, REJUV, target=TARGET),
            make_apply(110, REJUV, target=TARGET2),
            make_heal(200, REGROWTH, 20000, hit_type=2),  # crit
        ]
        pipeline = Pipeline(attributors=[AbundanceAttributor()])
        results = pipeline.run(events)
        # 2 rejuvs * 0.08 = 0.16 abundance crit
        # base crit = 0.5, total = 0.66
        # abundance share = 0.16 / 0.66
        # crit bonus = 20000 / 2 = 10000
        # attributed = 10000 * 0.16 / 0.66
        expected = 10000 * (0.16 / 0.66)
        assert results.talent_healing["Abundance"] == pytest.approx(expected, rel=0.01)

    def test_non_crit_regrowth_not_attributed(self):
        events = [
            make_apply(100, REJUV),
            make_heal(200, REGROWTH, 10000, hit_type=1),  # normal hit
        ]
        pipeline = Pipeline(attributors=[AbundanceAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Abundance"] == 0.0

    def test_no_rejuvs_not_attributed(self):
        events = [
            make_heal(200, REGROWTH, 10000, hit_type=2),
        ]
        pipeline = Pipeline(attributors=[AbundanceAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Abundance"] == 0.0

    def test_non_regrowth_not_attributed(self):
        events = [
            make_apply(100, REJUV),
            make_heal(200, REJUV, 10000, hit_type=2),  # Rejuv crit, not Regrowth
        ]
        pipeline = Pipeline(attributors=[AbundanceAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Abundance"] == 0.0

    def test_germination_rejuv_counted(self):
        events = [
            make_combatant_info(0, crit_spell=350),
            make_apply(100, GERMINATION_REJUV, target=TARGET),
            make_heal(200, REGROWTH, 20000, hit_type=2),
        ]
        pipeline = Pipeline(attributors=[AbundanceAttributor()])
        results = pipeline.run(events)
        # 1 rejuv * 0.08 = 0.08, base 0.5, total 0.58
        expected = 10000 * (0.08 / 0.58)
        assert results.talent_healing["Abundance"] == pytest.approx(expected, rel=0.01)


# --- Photosynthesis ---

class TestPhotosynthesis:
    def test_unexplained_bloom_attributed(self):
        """Bloom without subsequent remove/refresh = Photosynthesis proc."""
        events = [
            make_apply(100, LIFEBLOOM),
            make_heal(500, LIFEBLOOM_BLOOM, 15000),
            # No removebuff follows — this is a photo proc
        ]
        pipeline = Pipeline(attributors=[PhotosynthesisAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Photosynthesis"] == pytest.approx(15000.0)

    def test_bloom_from_expiry_not_attributed(self):
        """Bloom followed by removebuff = natural expiry, not Photosynthesis."""
        events = [
            make_apply(100, LIFEBLOOM),
            make_heal(500, LIFEBLOOM_BLOOM, 15000),
            make_remove(550, LIFEBLOOM),  # within 200ms
        ]
        pipeline = Pipeline(attributors=[PhotosynthesisAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Photosynthesis"] == 0.0

    def test_bloom_from_refresh_not_attributed(self):
        """Bloom followed by refreshbuff = pandemic refresh, not Photosynthesis."""
        events = [
            make_apply(100, LIFEBLOOM),
            make_heal(500, LIFEBLOOM_BLOOM, 15000),
            make_refresh(550, LIFEBLOOM),  # within 200ms
        ]
        pipeline = Pipeline(attributors=[PhotosynthesisAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Photosynthesis"] == 0.0

    def test_mixed_blooms(self):
        """Mix of photo procs and natural expiry."""
        events = [
            make_apply(100, LIFEBLOOM),
            make_heal(500, LIFEBLOOM_BLOOM, 10000),    # photo proc (no remove follows)
            make_heal(1000, LIFEBLOOM_BLOOM, 10000),   # natural expiry
            make_remove(1050, LIFEBLOOM),
        ]
        pipeline = Pipeline(attributors=[PhotosynthesisAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Photosynthesis"] == pytest.approx(10000.0)


# --- Nurturing Dormancy ---

class TestNurturingDormancy:
    def test_tick_past_base_duration_attributed(self):
        events = [
            make_apply(0, REJUV),
            make_heal(13000, REJUV, 5000),  # 13s > 12s base
        ]
        pipeline = Pipeline(attributors=[NurturingDormancyAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Nurturing Dormancy"] == pytest.approx(5000.0)

    def test_tick_within_base_duration_not_attributed(self):
        events = [
            make_apply(0, REJUV),
            make_heal(5000, REJUV, 5000),  # 5s < 12s base
        ]
        pipeline = Pipeline(attributors=[NurturingDormancyAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Nurturing Dormancy"] == 0.0

    def test_no_hot_tracked_not_attributed(self):
        events = [
            make_heal(13000, REJUV, 5000),  # no applybuff
        ]
        pipeline = Pipeline(attributors=[NurturingDormancyAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Nurturing Dormancy"] == 0.0

    def test_non_rejuv_not_attributed(self):
        events = [
            make_apply(0, REGROWTH),
            make_heal(13000, REGROWTH, 5000),
        ]
        pipeline = Pipeline(attributors=[NurturingDormancyAttributor()])
        results = pipeline.run(events)
        assert results.talent_healing["Nurturing Dormancy"] == 0.0
