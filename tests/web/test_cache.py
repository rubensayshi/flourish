from flourish.web.cache import ResultCache


def test_get_returns_none_when_missing(tmp_path):
    cache = ResultCache(cache_dir=tmp_path)
    assert cache.get("ABC123", 1, "Player") is None


def test_set_then_get_roundtrips(tmp_path):
    cache = ResultCache(cache_dir=tmp_path)
    data = {"total_healing": 1000, "talents": [{"name": "SotF", "attributed": 500}]}
    cache.set("ABC123", 1, "Player", data)
    assert cache.get("ABC123", 1, "Player") == data


def test_cache_key_is_case_insensitive_for_player(tmp_path):
    cache = ResultCache(cache_dir=tmp_path)
    data = {"total_healing": 1000}
    cache.set("ABC123", 1, "Saikó", data)
    assert cache.get("ABC123", 1, "saikó") == data
