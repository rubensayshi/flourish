import os

from fastapi import APIRouter, HTTPException, Request
from slowapi import Limiter
from slowapi.util import get_remote_address

from rdruid_analyzer.web.dependencies import get_wcl_client
from rdruid_analyzer.web.cache import ResultCache
from rdruid_analyzer.models.config import load_config, Config, MasteryConfig
from rdruid_analyzer.analysis.pipeline import Pipeline
from rdruid_analyzer.cli import build_attributors

router = APIRouter(prefix="/api")
limiter = Limiter(key_func=get_remote_address)
result_cache = ResultCache()

DRUID_CLASS = "Druid"


@router.get("/health")
def health():
    return {"status": "ok"}


@router.get("/report/{code}")
def get_report(code: str):
    client = get_wcl_client()
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
@limiter.limit("10/minute")
def analyze(request: Request, code: str, fight_id: int, player_name: str):
    cached = result_cache.get(code, fight_id, player_name)
    if cached:
        return cached

    client = get_wcl_client()
    try:
        report = client.get_report(code)
    except Exception:
        raise HTTPException(status_code=404, detail="Report not found")

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

    pet_ids = {a["id"] for a in all_actors if a.get("petOwner")}
    attributors = build_attributors(config, damage_taken_with_regrowth=damage_taken_with_regrowth)
    pipeline = Pipeline(attributors=attributors, pet_ids=pet_ids)
    results = pipeline.run(raw_events)

    duration_sec = max(results.fight_duration_ms / 1000, 1)
    talents = [
        {
            "name": name,
            "attributed": round(amount),
            "pct": round(amount / results.total_healing * 100, 1) if results.total_healing > 0 else 0,
            "hps": round(amount / duration_sec),
        }
        for name, amount in sorted(
            results.talent_healing.items(), key=lambda x: x[1], reverse=True
        )
        if amount > 0
    ]

    total_attributed = sum(t["attributed"] for t in talents)
    unattributed = max(0, results.total_healing - total_attributed - results.wasted)

    response = {
        "fight_name": selected_fight["name"],
        "player_name": selected_player["name"],
        "total_healing": results.total_healing,
        "duration_sec": round(duration_sec),
        "talents": talents,
        "wasted": results.wasted,
        "unattributed": unattributed,
    }

    result_cache.set(code, fight_id, player_name, response)
    return response
