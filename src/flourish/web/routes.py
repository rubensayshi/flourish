import os

from fastapi import APIRouter, HTTPException, Request
from limits import parse as parse_limit
from slowapi import Limiter
from slowapi.errors import RateLimitExceeded
from slowapi.util import get_remote_address

from flourish.web.dependencies import get_wcl_client_for_request, get_user_token
from flourish.web.auth import check_anon_limit, record_anon_usage, get_anon_remaining
from flourish.web.cache import ResultCache
from flourish.models.config import load_config, Config, MasteryConfig
from flourish.analysis.pipeline import Pipeline
from flourish.cli import build_attributors
from flourish.output.table import HERO_TREES

router = APIRouter(prefix="/api")
limiter = Limiter(key_func=get_remote_address)
result_cache = ResultCache()

DRUID_CLASS = "Druid"
_ANALYZE_LIMIT = parse_limit("10/minute")
_REPORT_LIMIT = parse_limit("15/minute")


def _check_rate_limit(request: Request, rate_limit):
    key = get_remote_address(request)
    if not limiter.limiter.hit(rate_limit, key):
        raise RateLimitExceeded(rate_limit)


def _check_anon_analyze_limit(request: Request):
    """Raise 403 if anonymous user has used all free analyses."""
    if get_user_token(request):
        return
    ip = get_remote_address(request)
    if not check_anon_limit(ip):
        raise HTTPException(
            status_code=403,
            detail=(
                "You've used your 2 free analyses. Log in with WarcraftLogs to continue. "
                "This helps us stay within API rate limits — we only use your login to "
                "analyze logs on your behalf, nothing else."
            ),
        )


@router.get("/health")
def health():
    return {"status": "ok"}


@router.get("/report/{code}")
def get_report(request: Request, code: str):
    _check_rate_limit(request, _REPORT_LIMIT)

    client = get_wcl_client_for_request(request)
    try:
        report = client.get_report(code)
    except Exception:
        raise HTTPException(status_code=404, detail="Report not found")

    fights = [
        {
            "id": f["id"],
            "name": f["name"],
            "kill": f["kill"],
            "duration": round((f["endTime"] - f["startTime"]) / 1000),
        }
        for f in report["fights"]
        if f.get("encounterID", 0) > 0
    ]

    druids = [
        {"id": a["id"], "name": a["name"], "server": a.get("server", "")}
        for a in report["masterData"]["actors"]
        if a.get("subType") == DRUID_CLASS
    ]

    return {"title": report["title"], "fights": fights, "druids": druids}


@router.get("/analyze/{code}/{fight_id}/{player_name}")
def analyze(request: Request, code: str, fight_id: int, player_name: str, base_stacks: int | None = None):
    # Serve cached results without counting against anon limit
    if base_stacks is None:
        cached = result_cache.get(code, fight_id, player_name)
        if cached:
            return cached

    _check_rate_limit(request, _ANALYZE_LIMIT)
    _check_anon_analyze_limit(request)
    # Only count against limit for fresh (non-cached) analyses

    client = get_wcl_client_for_request(request)
    try:
        report = client.get_report(code)
    except Exception:
        raise HTTPException(status_code=404, detail="Report not found")

    # Record anonymous analysis
    if not get_user_token(request):
        record_anon_usage(get_remote_address(request))

    selected_fight = next(
        (f for f in report["fights"] if f["id"] == fight_id), None
    )
    if not selected_fight:
        raise HTTPException(status_code=404, detail="Fight not found")

    all_actors = report["masterData"]["actors"]
    selected_player = next(
        (a for a in all_actors
         if a.get("subType") == DRUID_CLASS
         and a["name"].lower() == player_name.lower()),
        None,
    )
    if not selected_player:
        raise HTTPException(status_code=404, detail="Player not found")

    raw_events = client.get_events(
        code, fight_id, selected_player["id"],
        selected_fight["startTime"], selected_fight["endTime"],
    )

    REGROWTH_BUFF_ID = 8936
    regrowth_filter = (
        f'IN RANGE FROM (type = "applybuff" OR type = "refreshbuff") '
        f"AND ability.id = {REGROWTH_BUFF_ID} "
        f'TO type = "removebuff" AND ability.id = {REGROWTH_BUFF_ID} '
        f"GROUP BY target ON target END"
    )
    damage_taken_with_regrowth = client.get_damage_taken(
        code, fight_id, selected_player["id"],
        selected_fight["startTime"], selected_fight["endTime"],
        filter_expression=regrowth_filter,
    )

    config_path = "config/talents.yaml"
    config = (
        load_config(config_path)
        if os.path.exists(config_path)
        else Config(mastery=MasteryConfig(), talents={})
    )

    if base_stacks is not None:
        config.mastery.base_stacks = max(1, min(base_stacks, 5))

    pet_ids = {a["id"] for a in all_actors if a.get("petOwner")}
    attributors = build_attributors(config, damage_taken_with_regrowth=damage_taken_with_regrowth)
    pipeline = Pipeline(attributors=attributors, pet_ids=pet_ids)
    results = pipeline.run(raw_events)

    duration_sec = max(results.fight_duration_ms / 1000, 1)
    total = results.total_healing

    def _talent_entry(name, amount):
        entry = {
            "name": name,
            "attributed": round(amount),
            "pct": round(amount / total * 100, 1) if total > 0 else 0,
            "hps": round(amount / duration_sec),
        }
        if name in results.talent_ranks:
            entry["rank"] = results.talent_ranks[name]
        return entry

    def _hero_tree_for(name):
        for tree, talents in HERO_TREES.items():
            if name in talents:
                return tree
        return None

    non_hero = []
    hero_groups = {}  # tree_name -> list of (name, amount)
    for name, amount in results.talent_healing.items():
        if amount <= 0:
            continue
        tree = _hero_tree_for(name)
        if tree:
            hero_groups.setdefault(tree, []).append((name, amount))
        else:
            non_hero.append((name, amount))

    talents = [_talent_entry(n, a) for n, a in sorted(non_hero, key=lambda x: x[1], reverse=True)]

    hero_trees = []
    for tree_name, entries in sorted(hero_groups.items(), key=lambda x: sum(a for _, a in x[1]), reverse=True):
        tree_total = sum(a for _, a in entries)
        hero_trees.append({
            "name": tree_name,
            "attributed": round(tree_total),
            "pct": round(tree_total / total * 100, 1) if total > 0 else 0,
            "hps": round(tree_total / duration_sec),
            "talents": [_talent_entry(n, a) for n, a in sorted(entries, key=lambda x: x[1], reverse=True)],
        })

    all_attributed = sum(a for _, a in non_hero) + sum(a for g in hero_groups.values() for _, a in g)
    unattributed = max(0, total - round(all_attributed) - results.wasted)

    response = {
        "fight_name": selected_fight["name"],
        "player_name": selected_player["name"],
        "total_healing": total,
        "duration_sec": round(duration_sec),
        "talents": talents,
        "hero_trees": hero_trees,
        "wasted": results.wasted,
        "unattributed": unattributed,
    }

    result_cache.set(code, fight_id, player_name, response)
    return response
