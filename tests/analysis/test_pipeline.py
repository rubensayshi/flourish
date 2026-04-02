from flourish.analysis.pipeline import Pipeline
from flourish.analysis.attributor import TalentAttributor
from flourish.models.events import HealEvent


class FakeAttributor(TalentAttributor):
    name = "Fake Talent"

    def process_heal(self, event, hot_tracker, buff_tracker) -> float:
        if event.ability_id == 774:
            return event.amount * 0.5
        return 0.0


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


def test_pipeline_attributes_healing():
    raw_events = [make_heal(100, 774, 10000)]
    pipeline = Pipeline(attributors=[FakeAttributor()])
    results = pipeline.run(raw_events)
    assert results.talent_healing["Fake Talent"] == 5000


def test_pipeline_skips_wasted_heals():
    raw_events = [make_heal(100, 774, 2000, overheal=3000)]  # 60% OH
    pipeline = Pipeline(attributors=[FakeAttributor()])
    results = pipeline.run(raw_events)
    assert results.talent_healing["Fake Talent"] == 0
    assert results.wasted > 0


def test_pipeline_tracks_total_healing():
    raw_events = [
        make_heal(100, 774, 10000),
        make_heal(200, 48438, 5000),  # Wild Growth, fake doesn't claim
    ]
    pipeline = Pipeline(attributors=[FakeAttributor()])
    results = pipeline.run(raw_events)
    assert results.total_healing == 15000
    assert results.talent_healing["Fake Talent"] == 5000
