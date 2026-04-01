import os

import typer
from dotenv import load_dotenv
from rich.console import Console
from rich.prompt import Prompt

from rdruid_analyzer.wcl.client import WCLClient
from rdruid_analyzer.wcl.cache import CachedWCLClient
from rdruid_analyzer.models.config import load_config, Config, MasteryConfig
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
    FlourishAttributor,
    BurstingGrowthAttributor,
    ThrivingGrowthAttributor,
    SpiritOfTheThicketAttributor,
    ThrivingVegetationAttributor,
)
from rdruid_analyzer.analysis.talents.buff_multipliers import (
    ImprovedSwiftmendAttributor,
    WildSynthesisAttributor,
    WildstalkersPowerAttributor,
    PatientCustodianAttributor,
    LifetreadingAttributor,
    HarmonyOfTheGroveAttributor,
    GrovesInspirationAttributor,
    CenariusMightAttributor,
    BountifulBloomAttributor,
    UnstoppableGrowthAttributor,
    IntensityAttributor,
    LivelinessAttributor,
    RegenesisAttributor,
    PowerOfNatureAttributor,
)
from rdruid_analyzer.analysis.talents.tree_of_life import TreeOfLifeAttributor
from rdruid_analyzer.analysis.talents.convoke import ConvokeAttributor
from rdruid_analyzer.analysis.talents.improved_wild_growth import ImprovedWildGrowthAttributor
from rdruid_analyzer.analysis.talents.reforestation import ReforestationAttributor
from rdruid_analyzer.analysis.talents.wildstalker import (
    VigorousCreepersAttributor,
    ImplantAttributor,
    RootNetworkAttributor,
    StrategicInfusionAttributor,
)
from rdruid_analyzer.analysis.talents.symbiotic_bloom_mastery import SymbioticBloomMasteryAttributor
from rdruid_analyzer.analysis.talents.harmonious_blooming import HarmoniousBloomingAttributor
from rdruid_analyzer.analysis.talents.sm_cooldown_reduction import (
    SmCooldownReductionAttributor,
    WgCooldownReductionAttributor,
)
from rdruid_analyzer.analysis.talents.sylvan_beckoning import SylvanBeckoningAttributor
from rdruid_analyzer.analysis.talents.abundance import AbundanceAttributor
from rdruid_analyzer.analysis.talents.photosynthesis import PhotosynthesisAttributor
from rdruid_analyzer.analysis.talents.nurturing_dormancy import NurturingDormancyAttributor
from rdruid_analyzer.analysis.talents.protective_growth import ProtectiveGrowthAttributor
from rdruid_analyzer.output.table import render_results

app = typer.Typer()
console = Console()

DRUID_CLASS = "Druid"


def get_wcl_client() -> WCLClient:
    load_dotenv()
    client_id = os.environ.get("WCL_CLIENT_ID")
    client_secret = os.environ.get("WCL_CLIENT_SECRET")
    if not client_id or not client_secret:
        console.print("[red]Error:[/] WCL_CLIENT_ID and WCL_CLIENT_SECRET must be set in .env")
        raise typer.Exit(1)
    return WCLClient(client_id, client_secret)


def build_attributors(config: Config, damage_taken_with_regrowth: int | None = None) -> list:
    talents = config.talents
    mastery = config.mastery
    mastery_kwargs = {"mastery_pct": mastery.pct, "base_stacks": mastery.base_stacks, "dr_table": mastery.dr_table}

    convoke_cfg = talents.get("convoke_the_spirits")
    convoke_ratio = convoke_cfg.multiplier if convoke_cfg and convoke_cfg.multiplier is not None else 0.7

    sotf = SoulOfTheForestAttributor()
    gg = GroveGuardiansAttributor()
    sm_cd = SmCooldownReductionAttributor(downstream_attributors=[sotf, gg])
    wg_cd = WgCooldownReductionAttributor(downstream_attributors=[gg], has_4pc=False)

    all_attributors = [
        sotf,
        EverbloomAttributor(),
        gg,
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
        UnstoppableGrowthAttributor(),
        IntensityAttributor(),
        LivelinessAttributor(),
        RegenesisAttributor(),
        BurstingGrowthAttributor(),
        TreeOfLifeAttributor(),
        ConvokeAttributor(healing_ratio=convoke_ratio),
        ImprovedWildGrowthAttributor(),
        ReforestationAttributor(),
        VigorousCreepersAttributor(),
        ImplantAttributor(),
        RootNetworkAttributor(),
        StrategicInfusionAttributor(),
        SymbioticBloomMasteryAttributor(**mastery_kwargs),
        HarmoniousBloomingAttributor(**mastery_kwargs),
        AbundanceAttributor(),
        PhotosynthesisAttributor(),
        NurturingDormancyAttributor(),
        ProtectiveGrowthAttributor(damage_taken_with_regrowth=damage_taken_with_regrowth),
        SylvanBeckoningAttributor(),
        ThrivingGrowthAttributor(),
        SpiritOfTheThicketAttributor(),
        PowerOfNatureAttributor(),
        ThrivingVegetationAttributor(),
        sm_cd,
        wg_cd,
    ]
    active = []
    for a in all_attributors:
        key = a.name.lower().replace(" ", "_").replace("'", "")
        cfg = talents.get(key)
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
    config = load_config(config_path) if os.path.exists(config_path) else Config(mastery=MasteryConfig(), talents={})

    client = get_wcl_client()
    client = CachedWCLClient(client)
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
    all_actors = report["masterData"]["actors"]
    druids = [a for a in all_actors if a.get("subType") == DRUID_CLASS]
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

    # Fetch events (WCL includes pet events automatically when querying owner)
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

    REGROWTH_BUFF_ID = 8936
    regrowth_filter = (
        f'IN RANGE FROM (type = "applybuff" OR type = "refreshbuff") AND ability.id = {REGROWTH_BUFF_ID} '
        f'TO type = "removebuff" AND ability.id = {REGROWTH_BUFF_ID} '
        f"GROUP BY target ON target END"
    )
    damage_taken_with_regrowth = client.get_damage_taken(
        report_code,
        selected_fight["id"],
        selected_player["id"],
        selected_fight["startTime"],
        selected_fight["endTime"],
        filter_expression=regrowth_filter,
    )


    # Run analysis
    pet_ids = {a["id"] for a in all_actors if a.get("petOwner")}
    attributors = build_attributors(config, damage_taken_with_regrowth=damage_taken_with_regrowth)
    pipeline = Pipeline(attributors=attributors, pet_ids=pet_ids)
    results = pipeline.run(raw_events)

    # Output
    render_results(results, fight_name=selected_fight["name"], player_name=selected_player["name"])


if __name__ == "__main__":
    app()
