from dataclasses import dataclass, field

from flourish.models.events import (
    parse_event,
    CombatantInfoEvent,
    HealEvent,
    ApplyBuffEvent,
    RefreshBuffEvent,
    RemoveBuffEvent,
)
from flourish.tracking.hot_tracker import HotTracker
from flourish.tracking.buff_tracker import BuffTracker
from flourish.analysis.attributor import TalentAttributor


@dataclass
class AnalysisResults:
    total_healing: int = 0
    wasted: int = 0
    talent_healing: dict[str, float] = field(default_factory=dict)
    fight_duration_ms: int = 0
    combatant_info: CombatantInfoEvent | None = None


class Pipeline:
    def __init__(self, attributors: list[TalentAttributor], pet_ids: set[int] | None = None):
        self.attributors = attributors
        self.hot_tracker = HotTracker()
        self.buff_tracker = BuffTracker()
        self.pet_ids = pet_ids or set()

    def run(self, raw_events: list[dict]) -> AnalysisResults:
        results = AnalysisResults()
        for attr in self.attributors:
            results.talent_healing[attr.name] = 0.0

        events = [parse_event(e) for e in raw_events]
        events = [e for e in events if e is not None]

        if events:
            results.fight_duration_ms = events[-1].timestamp - events[0].timestamp

        for event in events:
            # Store combatant info, notify attributors, filter by talent selection
            if isinstance(event, CombatantInfoEvent):
                if results.combatant_info is None:
                    results.combatant_info = event
                    for attr in self.attributors:
                        attr.set_combatant_info(event)
                    if event.talent_nodes:
                        self.attributors = [a for a in self.attributors if a.is_selected()]
                        results.talent_healing = {a.name: 0.0 for a in self.attributors}
                continue

            # Update trackers
            if isinstance(event, (ApplyBuffEvent, RefreshBuffEvent, RemoveBuffEvent)):
                self.hot_tracker.process(event)
                self.buff_tracker.process(event)

            # Let attributors see every event for state tracking
            for attr in self.attributors:
                attr.process_event(event, self.hot_tracker, self.buff_tracker)

            # Process heals (skip healing to pets — not counted by WCL)
            if isinstance(event, HealEvent):
                if event.target_id in self.pet_ids:
                    continue
                results.total_healing += event.amount

                if event.is_wasted:
                    results.wasted += event.amount
                    continue

                for attr in self.attributors:
                    attributed = attr.process_heal(event, self.hot_tracker, self.buff_tracker)
                    results.talent_healing[attr.name] += attributed
                    attr.total_attributed += attributed

        # Let attributors finalize (for deferred attribution like Photosynthesis)
        for attr in self.attributors:
            finalized = attr.finalize()
            results.talent_healing[attr.name] += finalized
            attr.total_attributed += finalized

        return results
