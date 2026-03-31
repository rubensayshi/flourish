from unittest.mock import MagicMock

from rdruid_analyzer.wcl.client import WCLClient


def test_authenticate_and_query():
    client = WCLClient("test_id", "test_secret")

    mock_http = MagicMock()
    auth_resp = MagicMock()
    auth_resp.json.return_value = {"access_token": "fake_token"}
    query_resp = MagicMock()
    query_resp.json.return_value = {
        "data": {"reportData": {"report": {"title": "Test"}}}
    }
    mock_http.post.side_effect = [auth_resp, query_resp]
    client._http = mock_http

    result = client.get_report("abc123")
    assert result["title"] == "Test"
    assert mock_http.post.call_count == 2


def test_get_events_paginates():
    client = WCLClient("test_id", "test_secret")
    client._token = "fake_token"

    mock_http = MagicMock()
    page1 = MagicMock()
    page1.json.return_value = {
        "data": {
            "reportData": {
                "report": {
                    "events": {
                        "data": [{"type": "heal", "timestamp": 1}],
                        "nextPageTimestamp": 5000,
                    }
                }
            }
        }
    }
    page2 = MagicMock()
    page2.json.return_value = {
        "data": {
            "reportData": {
                "report": {
                    "events": {
                        "data": [{"type": "heal", "timestamp": 5001}],
                        "nextPageTimestamp": None,
                    }
                }
            }
        }
    }
    mock_http.post.side_effect = [page1, page2]
    client._http = mock_http

    events = client.get_events("abc", 1, 1, 0, 10000)
    assert len(events) == 2
    assert mock_http.post.call_count == 2


def test_get_events_includes_pet_events():
    """WCL returns pet events when querying the owner, no separate pet query needed."""
    client = WCLClient("test_id", "test_secret")
    client._token = "fake_token"

    mock_http = MagicMock()
    resp = MagicMock()
    resp.json.return_value = {
        "data": {
            "reportData": {
                "report": {
                    "events": {
                        "data": [
                            {"type": "heal", "timestamp": 100, "sourceID": 3},
                            {"type": "heal", "timestamp": 200, "sourceID": 10},
                            {"type": "heal", "timestamp": 300, "sourceID": 3},
                        ],
                        "nextPageTimestamp": None,
                    }
                }
            }
        }
    }
    mock_http.post.side_effect = [resp]
    client._http = mock_http

    events = client.get_events("abc", 1, 3, 0, 10000)
    assert len(events) == 3
    assert [e["timestamp"] for e in events] == [100, 200, 300]
    assert mock_http.post.call_count == 1
