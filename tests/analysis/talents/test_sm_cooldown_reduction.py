import pytest

from rdruid_analyzer.analysis.pipeline import Pipeline
from rdruid_analyzer.analysis.talents.sm_cooldown_reduction import SmCooldownReductionAttributor

SWIFTMEND = 18562


def make_cast(ts, ability, source=1, target=2):
    return {"timestamp": ts, "type": "cast", "sourceID": source, "targetID": target, "abilityGameID": ability}


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
