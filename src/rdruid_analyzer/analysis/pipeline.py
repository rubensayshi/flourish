from dataclasses import dataclass, field

from rdruid_analyzer.models.events import (
    parse_event,
    CombatantInfoEvent,
    HealEvent,
    ApplyBuffEvent,
    RefreshBuffEvent,
    RemoveBuffEvent,
)
from rdruid_analyzer.tracking.hot_tracker import HotTracker
from rdruid_analyzer.tracking.buff_tracker import BuffTracker
from rdruid_analyzer.analysis.attributor import TalentAttributor


@dataclass
class AnalysisResults:
    total_healing: int = 0
    wasted: int = 0
    talent_healing: dict[str, float] = field(default_factory=dict)
    fight_duration_ms: int = 0
    combatant_info: CombatantInfoEvent | None = None


class Pipeline:
    def __init__(self, attributors: list[TalentAttributor]):
        self.attributors = attributors
        self.hot_tracker = HotTracker()
        self.buff_tracker = BuffTracker()

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

            # Process heals
            if isinstance(event, HealEvent):
                results.total_healing += event.amount

                if event.is_wasted:
                    results.wasted += event.amount
                    continue

                for attr in self.attributors:
                    attributed = attr.process_heal(event, self.hot_tracker, self.buff_tracker)
                    results.talent_healing[attr.name] += attributed

        # Let attributors finalize (for deferred attribution like Photosynthesis)
        for attr in self.attributors:
            results.talent_healing[attr.name] += attr.finalize()

        return results
