import os
from functools import lru_cache

from dotenv import load_dotenv
from fastapi import Request

from flourish.wcl.client import WCLClient, WCLUserClient
from flourish.wcl.cache import CachedWCLClient


@lru_cache
def get_wcl_client() -> CachedWCLClient:
    load_dotenv()
    client_id = os.environ.get("WCL_CLIENT_ID", "")
    client_secret = os.environ.get("WCL_CLIENT_SECRET", "")
    if not client_id or not client_secret:
        raise RuntimeError("WCL_CLIENT_ID and WCL_CLIENT_SECRET must be set")
    return CachedWCLClient(WCLClient(client_id, client_secret))


def get_user_token(request: Request) -> str | None:
    """Extract Bearer token from Authorization header."""
    auth = request.headers.get("Authorization", "")
    if auth.startswith("Bearer "):
        return auth[7:]
    return None


def get_wcl_client_for_request(request: Request) -> CachedWCLClient | WCLUserClient:
    """Return a user-token client if authenticated, otherwise the shared app client."""
    token = get_user_token(request)
    if token:
        return WCLUserClient(token)
    return get_wcl_client()
