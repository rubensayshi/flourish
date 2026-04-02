import httpx

from flourish.wcl.queries import DAMAGE_TAKEN_TABLE_QUERY, EVENTS_QUERY, FIGHTS_QUERY

WCL_OAUTH_URL = "https://www.warcraftlogs.com/oauth/token"
WCL_AUTHORIZE_URL = "https://www.warcraftlogs.com/oauth/authorize"
WCL_API_URL = "https://www.warcraftlogs.com/api/v2/client"
WCL_USER_API_URL = "https://www.warcraftlogs.com/api/v2/user"


class WCLClient:
    def __init__(self, client_id: str, client_secret: str):
        self.client_id = client_id
        self.client_secret = client_secret
        self._token: str | None = None
        self._http = httpx.Client(timeout=30)

    def _authenticate(self):
        resp = self._http.post(
            WCL_OAUTH_URL,
            auth=(self.client_id, self.client_secret),
            data={"grant_type": "client_credentials"},
        )
        resp.raise_for_status()
        self._token = resp.json()["access_token"]

    def _query(self, query: str, variables: dict) -> dict:
        if not self._token:
            self._authenticate()
        resp = self._http.post(
            WCL_API_URL,
            headers={"Authorization": f"Bearer {self._token}"},
            json={"query": query, "variables": variables},
        )
        resp.raise_for_status()
        data = resp.json()
        if "errors" in data:
            raise RuntimeError(f"WCL API errors: {data['errors']}")
        return data["data"]

    def get_report(self, code: str) -> dict:
        data = self._query(FIGHTS_QUERY, {"code": code})
        return data["reportData"]["report"]

    def get_events(
        self,
        code: str,
        fight_id: int,
        source_id: int,
        start_time: float,
        end_time: float,
    ) -> list[dict]:
        """Fetch all events for a source. WCL includes pet events automatically."""
        all_events = []
        current_start = start_time

        while current_start is not None:
            data = self._query(
                EVENTS_QUERY,
                {
                    "code": code,
                    "startTime": current_start,
                    "endTime": end_time,
                    "sourceID": source_id,
                    "fightIDs": [fight_id],
                },
            )
            events_data = data["reportData"]["report"]["events"]
            all_events.extend(events_data["data"])
            current_start = events_data.get("nextPageTimestamp")

        all_events.sort(key=lambda e: e["timestamp"])
        return all_events

    def get_damage_taken(
        self,
        code: str,
        fight_id: int,
        source_id: int,
        start_time: float,
        end_time: float,
        filter_expression: str | None = None,
    ) -> int:
        """Fetch total damage taken using the table endpoint."""
        data = self._query(
            DAMAGE_TAKEN_TABLE_QUERY,
            {
                "code": code,
                "startTime": start_time,
                "endTime": end_time,
                "sourceID": source_id,
                "fightIDs": [fight_id],
                "filterExpression": filter_expression,
            },
        )
        table_data = data["reportData"]["report"]["table"]["data"]
        return sum(entry.get("total", 0) for entry in table_data.get("entries", []))


class WCLUserClient:
    """WCL client using a user's OAuth access token instead of app credentials."""

    def __init__(self, access_token: str):
        self._token = access_token
        self._http = httpx.Client(timeout=30)

    def _query(self, query: str, variables: dict) -> dict:
        resp = self._http.post(
            WCL_USER_API_URL,
            headers={"Authorization": f"Bearer {self._token}"},
            json={"query": query, "variables": variables},
        )
        resp.raise_for_status()
        data = resp.json()
        if "errors" in data:
            raise RuntimeError(f"WCL API errors: {data['errors']}")
        return data["data"]

    def get_report(self, code: str) -> dict:
        data = self._query(FIGHTS_QUERY, {"code": code})
        return data["reportData"]["report"]

    def get_events(
        self,
        code: str,
        fight_id: int,
        source_id: int,
        start_time: float,
        end_time: float,
    ) -> list[dict]:
        all_events = []
        current_start = start_time
        while current_start is not None:
            data = self._query(
                EVENTS_QUERY,
                {
                    "code": code,
                    "startTime": current_start,
                    "endTime": end_time,
                    "sourceID": source_id,
                    "fightIDs": [fight_id],
                },
            )
            events_data = data["reportData"]["report"]["events"]
            all_events.extend(events_data["data"])
            current_start = events_data.get("nextPageTimestamp")
        all_events.sort(key=lambda e: e["timestamp"])
        return all_events

    def get_damage_taken(
        self,
        code: str,
        fight_id: int,
        source_id: int,
        start_time: float,
        end_time: float,
        filter_expression: str | None = None,
    ) -> int:
        data = self._query(
            DAMAGE_TAKEN_TABLE_QUERY,
            {
                "code": code,
                "startTime": start_time,
                "endTime": end_time,
                "sourceID": source_id,
                "fightIDs": [fight_id],
                "filterExpression": filter_expression,
            },
        )
        table_data = data["reportData"]["report"]["table"]["data"]
        return sum(entry.get("total", 0) for entry in table_data.get("entries", []))
