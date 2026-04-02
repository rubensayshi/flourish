import logging
import os
import secrets
from urllib.parse import urlencode

import httpx
from fastapi import APIRouter, HTTPException, Request
from fastapi.responses import RedirectResponse

from flourish.wcl.client import WCL_AUTHORIZE_URL, WCL_OAUTH_URL

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/auth")

# In-memory CSRF state tokens (short-lived, cleared after use)
_pending_states: set[str] = set()

# Track anonymous usage: {ip: set(report_codes)}
_anon_usage: dict[str, set[str]] = {}
ANON_REPORT_LIMIT = 2


def _get_redirect_uri(request: Request) -> str:
    """Build callback URI from env or derive from the incoming request."""
    override = os.environ.get("WCL_REDIRECT_URI")
    if override:
        return override
    return str(request.base_url).rstrip("/") + "/api/auth/callback"


def _get_frontend_url(request: Request) -> str:
    """Frontend URL from env or derive from the incoming request origin."""
    override = os.environ.get("FRONTEND_URL")
    if override:
        return override
    return str(request.base_url).rstrip("/")


def check_anon_limit(ip: str, report_code: str) -> bool:
    """Return True if this anonymous IP is allowed to access this report."""
    used = _anon_usage.get(ip, set())
    if report_code in used:
        return True  # Already accessed this report
    return len(used) < ANON_REPORT_LIMIT


def record_anon_usage(ip: str, report_code: str):
    """Record that an anonymous IP accessed a report."""
    _anon_usage.setdefault(ip, set()).add(report_code)


@router.get("/login")
def login(request: Request):
    state = secrets.token_urlsafe(32)
    _pending_states.add(state)
    params = urlencode({
        "client_id": os.environ.get("WCL_CLIENT_ID", ""),
        "redirect_uri": _get_redirect_uri(request),
        "response_type": "code",
        "state": state,
    })
    return RedirectResponse(f"{WCL_AUTHORIZE_URL}?{params}")


@router.get("/callback")
def callback(request: Request, code: str | None = None, state: str | None = None, error: str | None = None):
    frontend_url = _get_frontend_url(request)

    if error:
        return RedirectResponse(f"{frontend_url}/?auth_error={error}")

    if not code or not state:
        raise HTTPException(status_code=400, detail="Missing code or state")

    if state not in _pending_states:
        logger.warning("Unknown state token — may have been lost to server reload")
        raise HTTPException(status_code=400, detail="Invalid state parameter. Try logging in again.")
    _pending_states.discard(state)

    # Exchange code for token
    client_id = os.environ.get("WCL_CLIENT_ID", "")
    client_secret = os.environ.get("WCL_CLIENT_SECRET", "")
    redirect_uri = _get_redirect_uri(request)
    logger.info("Token exchange: redirect_uri=%s", redirect_uri)

    resp = httpx.post(
        WCL_OAUTH_URL,
        auth=(client_id, client_secret),
        data={
            "grant_type": "authorization_code",
            "code": code,
            "redirect_uri": redirect_uri,
        },
    )
    if resp.status_code != 200:
        logger.error("Token exchange failed: %s %s", resp.status_code, resp.text)
        return RedirectResponse(f"{frontend_url}/?auth_error=token_exchange_failed")

    token_data = resp.json()
    access_token = token_data.get("access_token", "")

    return RedirectResponse(f"{frontend_url}/?wcl_token={access_token}")
