from rdruid_analyzer.tracking.buff_tracker import BuffTracker
from rdruid_analyzer.models.events import ApplyBuffEvent, RemoveBuffEvent

SOTF_BUFF = 114108


def test_buff_applied():
    tracker = BuffTracker()
    tracker.process(ApplyBuffEvent(timestamp=100, source_id=1, type="applybuff", target_id=1, ability_id=SOTF_BUFF))
    assert tracker.is_active(SOTF_BUFF)


def test_buff_removed():
    tracker = BuffTracker()
    tracker.process(ApplyBuffEvent(timestamp=100, source_id=1, type="applybuff", target_id=1, ability_id=SOTF_BUFF))
    tracker.process(RemoveBuffEvent(timestamp=200, source_id=1, type="removebuff", target_id=1, ability_id=SOTF_BUFF))
    assert not tracker.is_active(SOTF_BUFF)


def test_buff_not_active_by_default():
    tracker = BuffTracker()
    assert not tracker.is_active(SOTF_BUFF)
