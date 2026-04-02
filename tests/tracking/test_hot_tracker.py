from flourish.tracking.hot_tracker import HotTracker, HotInstance
from flourish.models.events import ApplyBuffEvent, RefreshBuffEvent, RemoveBuffEvent

REJUV_ID = 774
TARGET_A = 10


def make_apply(ts, target, ability):
    return ApplyBuffEvent(timestamp=ts, source_id=1, type="applybuff", target_id=target, ability_id=ability)


def make_remove(ts, target, ability):
    return RemoveBuffEvent(timestamp=ts, source_id=1, type="removebuff", target_id=target, ability_id=ability)


def make_refresh(ts, target, ability):
    return RefreshBuffEvent(timestamp=ts, source_id=1, type="refreshbuff", target_id=target, ability_id=ability)


def test_apply_creates_hot():
    tracker = HotTracker()
    tracker.process(make_apply(100, TARGET_A, REJUV_ID))
    hot = tracker.get(TARGET_A, REJUV_ID)
    assert hot is not None
    assert hot.spell_id == REJUV_ID
    assert hot.applied_at == 100


def test_remove_clears_hot():
    tracker = HotTracker()
    tracker.process(make_apply(100, TARGET_A, REJUV_ID))
    tracker.process(make_remove(200, TARGET_A, REJUV_ID))
    assert tracker.get(TARGET_A, REJUV_ID) is None


def test_refresh_updates_timestamp():
    tracker = HotTracker()
    tracker.process(make_apply(100, TARGET_A, REJUV_ID))
    tracker.process(make_refresh(200, TARGET_A, REJUV_ID))
    hot = tracker.get(TARGET_A, REJUV_ID)
    assert hot.applied_at == 100  # original apply time preserved
    assert hot.last_refresh == 200


def test_tags_cleared_on_refresh():
    """Refresh = new HoT application, so old tags should be cleared."""
    tracker = HotTracker()
    tracker.process(make_apply(100, TARGET_A, REJUV_ID))
    hot = tracker.get(TARGET_A, REJUV_ID)
    hot.tags.add("sotf")
    tracker.process(make_refresh(200, TARGET_A, REJUV_ID))
    assert "sotf" not in tracker.get(TARGET_A, REJUV_ID).tags


def test_get_all_hots_on_target():
    tracker = HotTracker()
    tracker.process(make_apply(100, TARGET_A, 774))
    tracker.process(make_apply(100, TARGET_A, 33763))  # Lifebloom
    hots = tracker.get_all(TARGET_A)
    assert len(hots) == 2
