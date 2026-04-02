from flourish.models.events import ApplyBuffEvent, RemoveBuffEvent, RefreshBuffEvent


class BuffTracker:
    def __init__(self):
        self._active: dict[int, int] = {}  # buff_id -> timestamp applied

    def process(self, event: ApplyBuffEvent | RemoveBuffEvent | RefreshBuffEvent):
        if isinstance(event, (ApplyBuffEvent, RefreshBuffEvent)):
            self._active[event.ability_id] = event.timestamp
        elif isinstance(event, RemoveBuffEvent):
            self._active.pop(event.ability_id, None)

    def is_active(self, buff_id: int) -> bool:
        return buff_id in self._active

    def get_applied_at(self, buff_id: int) -> int | None:
        return self._active.get(buff_id)
