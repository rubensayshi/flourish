import json
from pathlib import Path

from rdruid_analyzer.wcl.client import WCLClient

DEFAULT_CACHE_DIR = Path("data/cache")


class CachedWCLClient:
    def __init__(self, inner: WCLClient, cache_dir: Path = DEFAULT_CACHE_DIR):
        self._inner = inner
        self._cache_dir = cache_dir
        self._cache_dir.mkdir(parents=True, exist_ok=True)

    def _read(self, path: Path):
        if path.exists():
            return json.loads(path.read_text())
        return None

    def _write(self, path: Path, data):
        path.write_text(json.dumps(data))

    def get_report(self, code: str) -> dict:
        path = self._cache_dir / f"{code}_report.json"
        cached = self._read(path)
        if cached is not None:
            return cached
        result = self._inner.get_report(code)
        self._write(path, result)
        return result

    def get_events(
        self,
        code: str,
        fight_id: int,
        source_id: int,
        start_time: float,
        end_time: float,
    ) -> list[dict]:
        path = self._cache_dir / f"{code}_{fight_id}_{source_id}_events.json"
        cached = self._read(path)
        if cached is not None:
            return cached
        result = self._inner.get_events(code, fight_id, source_id, start_time, end_time)
        self._write(path, result)
        return result

    def get_damage_taken(
        self,
        code: str,
        fight_id: int,
        source_id: int,
        start_time: float,
        end_time: float,
        filter_expression: str | None = None,
    ) -> int:
        suffix = f"_{filter_expression}" if filter_expression else ""
        path = self._cache_dir / f"{code}_{fight_id}_{source_id}_dmgtaken{suffix}.json"
        cached = self._read(path)
        if cached is not None:
            return cached
        result = self._inner.get_damage_taken(
            code, fight_id, source_id, start_time, end_time, filter_expression
        )
        self._write(path, result)
        return result
