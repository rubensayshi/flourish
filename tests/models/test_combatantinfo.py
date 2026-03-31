from rdruid_analyzer.models.events import parse_event, CombatantInfoEvent


def test_parse_combatantinfo():
    raw = {
        "timestamp": 1000,
        "type": "combatantinfo",
        "sourceID": 3,
        "talentTree": [
            {"id": 103098, "rank": 1, "nodeID": 82047},
            {"id": 103100, "rank": 1, "nodeID": 82049},
        ],
        "critSpell": 256,
        "hasteSpell": 564,
        "mastery": 893,
        "specID": 105,
    }
    event = parse_event(raw)
    assert isinstance(event, CombatantInfoEvent)
    assert 82047 in event.talent_nodes
    assert 82049 in event.talent_nodes
    assert event.crit_spell == 256
    assert event.spec_id == 105


def test_combatantinfo_passed_to_attributors():
    from rdruid_analyzer.analysis.pipeline import Pipeline
    from rdruid_analyzer.analysis.attributor import TalentAttributor

    class TestAttributor(TalentAttributor):
        name = "Test"
        saw_info = False

        def set_combatant_info(self, info):
            super().set_combatant_info(info)
            self.saw_info = True

    raw_events = [
        {
            "timestamp": 1000,
            "type": "combatantinfo",
            "sourceID": 3,
            "talentTree": [{"id": 1, "rank": 1, "nodeID": 82047}],
            "critSpell": 256,
            "hasteSpell": 564,
            "mastery": 893,
            "specID": 105,
        },
        {
            "timestamp": 2000,
            "type": "heal",
            "sourceID": 3,
            "targetID": 4,
            "abilityGameID": 774,
            "amount": 1000,
            "overheal": 0,
            "hitType": 1,
        },
    ]
    attr = TestAttributor()
    pipeline = Pipeline(attributors=[attr])
    results = pipeline.run(raw_events)
    assert attr.saw_info
    assert attr.has_talent(82047)
    assert not attr.has_talent(99999)
