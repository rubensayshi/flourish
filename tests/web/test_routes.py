from unittest.mock import patch, MagicMock
from fastapi.testclient import TestClient

from flourish.web.app import create_app


def _mock_wcl_client():
    client = MagicMock()
    client.get_report.return_value = {
        "title": "Test Raid",
        "fights": [
            {"id": 1, "name": "Boss", "kill": True, "startTime": 0,
             "endTime": 60000, "encounterID": 123, "difficulty": 4},
            {"id": 2, "name": "Trash", "kill": True, "startTime": 0,
             "endTime": 30000, "encounterID": 0, "difficulty": 0},
        ],
        "masterData": {
            "actors": [
                {"id": 10, "name": "Saikó", "type": "Player",
                 "subType": "Druid", "server": "Draenor", "petOwner": None},
                {"id": 11, "name": "Warrior", "type": "Player",
                 "subType": "Warrior", "server": "Draenor", "petOwner": None},
            ]
        },
    }
    return client


@patch("flourish.web.routes.get_wcl_client")
def test_report_endpoint_returns_fights_and_druids(mock_get_client):
    mock_get_client.return_value = _mock_wcl_client()
    app = create_app()
    client = TestClient(app)

    resp = client.get("/api/report/ABC123")
    assert resp.status_code == 200
    data = resp.json()
    assert data["title"] == "Test Raid"
    assert len(data["fights"]) == 1
    assert data["fights"][0]["name"] == "Boss"
    assert len(data["druids"]) == 1
    assert data["druids"][0]["name"] == "Saikó"


@patch("flourish.web.routes.get_wcl_client")
def test_report_endpoint_404_on_invalid_code(mock_get_client):
    mock_get_client.return_value = MagicMock(
        get_report=MagicMock(side_effect=Exception("not found"))
    )
    app = create_app()
    client = TestClient(app)

    resp = client.get("/api/report/INVALID")
    assert resp.status_code == 404


def test_health_endpoint():
    app = create_app()
    client = TestClient(app)
    resp = client.get("/api/health")
    assert resp.status_code == 200
