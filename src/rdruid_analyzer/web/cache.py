import json
from pathlib import Path


class ResultCache:
    def __init__(self, cache_dir: Path = Path("data/results_cache")):
        self._dir = cache_dir
        self._dir.mkdir(parents=True, exist_ok=True)

    def _key_path(self, code: str, fight_id: int, player: str) -> Path:
        safe_player = player.lower().replace(" ", "_")
        return self._dir / f"{code}_{fight_id}_{safe_player}.json"

    def get(self, code: str, fight_id: int, player: str) -> dict | None:
        path = self._key_path(code, fight_id, player)
        if not path.exists():
            return None
        return json.loads(path.read_text())

    def set(self, code: str, fight_id: int, player: str, data: dict) -> None:
        path = self._key_path(code, fight_id, player)
        path.write_text(json.dumps(data))
