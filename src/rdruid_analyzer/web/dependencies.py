import os
from functools import lru_cache

from dotenv import load_dotenv

from rdruid_analyzer.wcl.client import WCLClient
from rdruid_analyzer.wcl.cache import CachedWCLClient


@lru_cache
def get_wcl_client() -> CachedWCLClient:
    load_dotenv()
    client_id = os.environ.get("WCL_CLIENT_ID", "")
    client_secret = os.environ.get("WCL_CLIENT_SECRET", "")
    if not client_id or not client_secret:
        raise RuntimeError("WCL_CLIENT_ID and WCL_CLIENT_SECRET must be set")
    return CachedWCLClient(WCLClient(client_id, client_secret))
