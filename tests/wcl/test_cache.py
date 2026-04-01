import json
from pathlib import Path
from unittest.mock import MagicMock

from rdruid_analyzer.wcl.cache import CachedWCLClient


def test_get_report_caches_to_disk(tmp_path):
    inner = MagicMock()
    inner.get_report.return_value = {"title": "Test Report", "fights": []}
    client = CachedWCLClient(inner, cache_dir=tmp_path)

    result = client.get_report("ABC123")

    assert result["title"] == "Test Report"
    cache_file = tmp_path / "ABC123_report.json"
    assert cache_file.exists()
    assert json.loads(cache_file.read_text())["title"] == "Test Report"


def test_get_report_reads_from_cache(tmp_path):
    inner = MagicMock()
    cached = {"title": "Cached Report", "fights": []}
    cache_file = tmp_path / "ABC123_report.json"
    cache_file.write_text(json.dumps(cached))
    client = CachedWCLClient(inner, cache_dir=tmp_path)

    result = client.get_report("ABC123")

    assert result["title"] == "Cached Report"
    inner.get_report.assert_not_called()


def test_get_events_caches_to_disk(tmp_path):
    inner = MagicMock()
    inner.get_events.return_value = [{"type": "heal", "timestamp": 1}]
    client = CachedWCLClient(inner, cache_dir=tmp_path)

    result = client.get_events("ABC123", 1, 5, 0, 10000)

    assert len(result) == 1
    cache_file = tmp_path / "ABC123_1_5_events.json"
    assert cache_file.exists()


def test_get_events_reads_from_cache(tmp_path):
    inner = MagicMock()
    cached = [{"type": "heal", "timestamp": 1}]
    cache_file = tmp_path / "ABC123_1_5_events.json"
    cache_file.write_text(json.dumps(cached))
    client = CachedWCLClient(inner, cache_dir=tmp_path)

    result = client.get_events("ABC123", 1, 5, 0, 10000)

    assert result == cached
    inner.get_events.assert_not_called()
