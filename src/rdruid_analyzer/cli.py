import os

import typer
from dotenv import load_dotenv
from rich.console import Console
from rich.prompt import Prompt

from rdruid_analyzer.wcl.client import WCLClient
from rdruid_analyzer.models.config import load_config
from rdruid_analyzer.analysis.pipeline import Pipeline
from rdruid_analyzer.analysis.talents.soul_of_the_forest import SoulOfTheForestAttributor
from rdruid_analyzer.analysis.talents.direct_spells import (
    EverbloomAttributor,
    GroveGuardiansAttributor,
    DreamSurgeAttributor,
    EfflorescenceAttributor,
    VerdancyAttributor,
    NaturesBountyAttributor,
    RegenerativeHeartwoodAttributor,
    CultivationAttributor,
    YserasGiftAttributor,
    EmbraceOfTheDreamAttributor,
    RampantGrowthAttributor,
    ImprovedSwiftmendAttributor,
    FlourishAttributor,
)
from rdruid_analyzer.analysis.talents.buff_multipliers import (
    WildSynthesisAttributor,
    WildstalkersPowerAttributor,
    PatientCustodianAttributor,
    LifetreadingAttributor,
    HarmonyOfTheGroveAttributor,
    GrovesInspirationAttributor,
    CenariusMightAttributor,
    BountifulBloomAttributor,
)
from rdruid_analyzer.output.table import render_results

app = typer.Typer()
console = Console()

DRUID_CLASS = "Druid"


def get_wcl_client() -> WCLClient:
    load_dotenv()
    client_id = os.environ["WCL_CLIENT_ID"]
    client_secret = os.environ["WCL_CLIENT_SECRET"]
    return WCLClient(client_id, client_secret)


def build_attributors(config: dict) -> list:
    all_attributors = [
        SoulOfTheForestAttributor(),
        EverbloomAttributor(),
        GroveGuardiansAttributor(),
        DreamSurgeAttributor(),
        EfflorescenceAttributor(),
        VerdancyAttributor(),
        NaturesBountyAttributor(),
        RegenerativeHeartwoodAttributor(),
        CultivationAttributor(),
        YserasGiftAttributor(),
        EmbraceOfTheDreamAttributor(),
        RampantGrowthAttributor(),
        ImprovedSwiftmendAttributor(),
        FlourishAttributor(),
        WildSynthesisAttributor(),
        WildstalkersPowerAttributor(),
        PatientCustodianAttributor(),
        LifetreadingAttributor(),
        HarmonyOfTheGroveAttributor(),
        GrovesInspirationAttributor(),
        CenariusMightAttributor(),
        BountifulBloomAttributor(),
    ]
    active = []
    for a in all_attributors:
        key = a.name.lower().replace(" ", "_").replace("'", "")
        cfg = config.get(key)
        if cfg and cfg.skip:
            continue
        active.append(a)
    return active


@app.command()
def analyze(
    report_code: str = typer.Argument(help="WarcraftLogs report code"),
    fight: int | None = typer.Option(None, help="Fight ID"),
    player: str | None = typer.Option(None, help="Player name"),
    config_path: str = typer.Option("config/talents.yaml", help="Talent config path"),
):
    """Analyze a WarcraftLogs report for talent healing attribution."""
    config = load_config(config_path) if os.path.exists(config_path) else {}

    client = get_wcl_client()
    report = client.get_report(report_code)
    console.print(f"[bold]Report:[/] {report['title']}")

    # Select fight
    fights = [f for f in report["fights"] if f.get("encounterID", 0) > 0]
    if fight is None:
        console.print("\n[bold]Fights:[/]")
        for f in fights:
            kill_str = "[green]Kill[/]" if f["kill"] else "[red]Wipe[/]"
            duration = (f["endTime"] - f["startTime"]) / 1000
            console.print(f"  {f['id']:3d}: {f['name']} ({kill_str}, {duration:.0f}s)")
        fight = int(Prompt.ask("Select fight ID"))

    selected_fight = next(f for f in report["fights"] if f["id"] == fight)

    # Select player
    actors = report["masterData"]["actors"]
    druids = [a for a in actors if a.get("subType") == DRUID_CLASS]
    if player is None:
        if len(druids) == 1:
            selected_player = druids[0]
            console.print(f"Auto-selected: [bold]{selected_player['name']}[/]")
        else:
            console.print("\n[bold]Resto Druids:[/]")
            for d in druids:
                console.print(f"  {d['id']:3d}: {d['name']} ({d.get('server', '')})")
            player_id = int(Prompt.ask("Select player ID"))
            selected_player = next(d for d in druids if d["id"] == player_id)
    else:
        selected_player = next(d for d in druids if d["name"].lower() == player.lower())

    # Fetch events
    console.print(
        f"\nFetching events for [bold]{selected_player['name']}[/] "
        f"in [bold]{selected_fight['name']}[/]..."
    )
    raw_events = client.get_events(
        report_code,
        selected_fight["id"],
        selected_player["id"],
        selected_fight["startTime"],
        selected_fight["endTime"],
    )
    console.print(f"Fetched {len(raw_events)} events")

    # Run analysis
    attributors = build_attributors(config)
    pipeline = Pipeline(attributors=attributors)
    results = pipeline.run(raw_events)

    # Output
    render_results(results, fight_name=selected_fight["name"], player_name=selected_player["name"])


if __name__ == "__main__":
    app()
