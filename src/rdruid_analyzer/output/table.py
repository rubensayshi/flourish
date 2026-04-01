from rich.console import Console
from rich.table import Table
from rdruid_analyzer.analysis.pipeline import AnalysisResults

# Hero tree talent names (from Blizzard Game Data API)
HERO_TREES: dict[str, set[str]] = {
    "Wildstalker": {
        "Wildstalker's Power", "Patient Custodian", "Vigorous Creepers",
        "Bursting Growth", "Rampancy", "Resilient Flourishing", "Root Network",
        "Twin Sprouts", "Implant", "Green Thumb", "Thriving Growth",
        "Symbiotic Bloom Mastery",
        "Entangling Vortex", "Flower Walk", "Strategic Infusion",
        "Lethal Preservation", "Bond with Nature", "Harmonious Constitution",
        "Hunt Beneath the Open Skies",
    },
    "Keeper of the Grove": {
        "Dream Surge", "Treants of the Moon", "Blooming Infusion",
        "Harmony of the Grove", "Power of Nature", "Durability of Nature",
        "Bounteous Bloom", "Early Spring", "Power of the Dream",
        "Control of the Dream", "Grove's Inspiration", "Cenarius' Might",
        "Protective Growth", "Expansiveness", "Potent Enchantments",
        "Spirit of the Thicket", "Dryad's Dance", "Sylvan Beckoning",
        "SM Cooldown Reduction",
        "WG Cooldown Reduction",
    },
}


def format_healing(amount: float) -> str:
    if amount >= 1_000_000:
        return f"{amount / 1_000_000:.1f}M"
    if amount >= 1_000:
        return f"{amount / 1_000:.1f}k"
    return str(int(amount))


def _hero_tree_for(talent_name: str) -> str | None:
    for tree, talents in HERO_TREES.items():
        if talent_name in talents:
            return tree
    return None


def render_results(results: AnalysisResults, fight_name: str = "", player_name: str = "") -> str:
    console = Console(record=True, width=80)
    duration_sec = max(results.fight_duration_ms / 1000, 1)

    if fight_name or player_name:
        console.print(f"\n[bold]Fight:[/] {fight_name}  [bold]Player:[/] {player_name}")
    console.print(f"[bold]Total effective healing:[/] {format_healing(results.total_healing)}\n")

    table = Table(show_header=True, header_style="bold cyan")
    table.add_column("Talent", style="white", min_width=25)
    table.add_column("Attributed", justify="right")
    table.add_column("% Total", justify="right")
    table.add_column("HPS", justify="right")

    # Split talents into non-hero and hero groups
    hero_totals: dict[str, float] = {}
    hero_talents: dict[str, list[tuple[str, float]]] = {}
    non_hero: list[tuple[str, float]] = []

    for name, amount in results.talent_healing.items():
        if amount <= 0:
            continue
        tree = _hero_tree_for(name)
        if tree:
            hero_totals[tree] = hero_totals.get(tree, 0) + amount
            hero_talents.setdefault(tree, []).append((name, amount))
        else:
            non_hero.append((name, amount))

    # Non-hero talents sorted by healing
    for name, amount in sorted(non_hero, key=lambda x: x[1], reverse=True):
        pct = (amount / results.total_healing * 100) if results.total_healing > 0 else 0
        hps = amount / duration_sec
        table.add_row(name, format_healing(amount), f"{pct:.1f}%", format_healing(hps))

    # Hero tree groups: aggregate header + individual talents
    for tree, total in sorted(hero_totals.items(), key=lambda x: x[1], reverse=True):
        table.add_section()
        pct = (total / results.total_healing * 100) if results.total_healing > 0 else 0
        hps = total / duration_sec
        table.add_row(
            f"{tree}", format_healing(total), f"{pct:.1f}%", format_healing(hps),
            style="bold",
        )
        for name, amount in sorted(hero_talents[tree], key=lambda x: x[1], reverse=True):
            pct = (amount / results.total_healing * 100) if results.total_healing > 0 else 0
            hps = amount / duration_sec
            table.add_row(f"  {name}", format_healing(amount), f"{pct:.1f}%", format_healing(hps))

    table.add_section()
    table.add_row("Wasted (>50% OH)", format_healing(results.wasted), "—", "—", style="dim")
    total_attributed = sum(a for a in results.talent_healing.values() if a > 0)
    unattributed = results.total_healing - total_attributed - results.wasted
    if unattributed > 0:
        table.add_row("Unattributed", format_healing(unattributed), "—", "—", style="dim")
    else:
        table.add_row("Unattributed", "—", "—", "—", style="dim")

    console.print(table)

    if total_attributed > results.total_healing:
        console.print(
            "\n[dim]Note: Talents can overlap (multiple talents buff the same heal).[/]"
            "\n[dim]Total attributed may exceed total healing. Each value answers:[/]"
            '\n[dim]"How much healing would I lose dropping this talent?"[/]'
        )

    return console.export_text()
