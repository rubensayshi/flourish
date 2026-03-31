from dataclasses import dataclass, field

from rdruid_analyzer.models.events import ApplyBuffEvent, RefreshBuffEvent, RemoveBuffEvent


@dataclass
class HotInstance:
    spell_id: int
    target_id: int
    applied_at: int
    last_refresh: int = 0
    tags: set[str] = field(default_factory=set)


class HotTracker:
    def __init__(self):
        self._hots: dict[tuple[int, int], HotInstance] = {}

    def process(self, event: ApplyBuffEvent | RefreshBuffEvent | RemoveBuffEvent):
        key = (event.target_id, event.ability_id)

        if isinstance(event, ApplyBuffEvent):
            self._hots[key] = HotInstance(
                spell_id=event.ability_id,
                target_id=event.target_id,
                applied_at=event.timestamp,
            )
        elif isinstance(event, RefreshBuffEvent):
            existing = self._hots.get(key)
            if existing:
                existing.last_refresh = event.timestamp
            else:
                self._hots[key] = HotInstance(
                    spell_id=event.ability_id,
                    target_id=event.target_id,
                    applied_at=event.timestamp,
                    last_refresh=event.timestamp,
                )
        elif isinstance(event, RemoveBuffEvent):
            self._hots.pop(key, None)

    def get(self, target_id: int, spell_id: int) -> HotInstance | None:
        return self._hots.get((target_id, spell_id))

    def get_all(self, target_id: int) -> list[HotInstance]:
        return [h for (tid, _), h in self._hots.items() if tid == target_id]

    def get_all_by_spell(self, spell_id: int) -> list[HotInstance]:
        return [h for (_, sid), h in self._hots.items() if sid == spell_id]
