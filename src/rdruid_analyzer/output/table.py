from rich.console import Console
from rich.table import Table
from rdruid_analyzer.analysis.pipeline import AnalysisResults


def format_healing(amount: float) -> str:
    if amount >= 1_000_000:
        return f"{amount / 1_000_000:.1f}M"
    if amount >= 1_000:
        return f"{amount / 1_000:.1f}k"
    return str(int(amount))


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

    sorted_talents = sorted(results.talent_healing.items(), key=lambda x: x[1], reverse=True)
    for name, amount in sorted_talents:
        if amount <= 0:
            continue
        pct = (amount / results.total_healing * 100) if results.total_healing > 0 else 0
        hps = amount / duration_sec
        table.add_row(name, format_healing(amount), f"{pct:.1f}%", format_healing(hps))

    table.add_section()
    table.add_row("Wasted (>50% OH)", format_healing(results.wasted), "—", "—", style="dim")
    unattributed = results.total_healing - sum(a for a in results.talent_healing.values() if a > 0) - results.wasted
    table.add_row("Unattributed", format_healing(max(0, unattributed)), "—", "—", style="dim")

    console.print(table)
    return console.export_text()
