#!/usr/bin/env python3
"""
Parse SimulationCraft druid spell dump and compute effective healing coefficients.

Fetches druid.txt from SimC's GitHub, parses raw SP coefficients and passive
aura modifiers, then computes the effective coefficient for each healing spell
as played by a Restoration Druid.

Usage:
    python scripts/parse_simc_spells.py [--dump druid.txt] [--output config/spell_coefficients.yaml]
"""

import re
import sys
import argparse
import urllib.request
from dataclasses import dataclass, field
from typing import Optional

SIMC_URL = "https://raw.githubusercontent.com/simulationcraft/simc/refs/heads/midnight/SpellDataDump/druid.txt"

# Healing spells we care about (spell ID -> name)
HEALING_SPELLS = {
    774: "Rejuvenation",
    155777: "Rejuvenation (Germination)",
    8936: "Regrowth",
    48438: "Wild Growth",
    18562: "Swiftmend",
    33763: "Lifebloom",
    33778: "Lifebloom (Bloom)",
    81269: "Efflorescence",
    102352: "Cenarion Ward",  # SimC uses 102352, WCL uses 157982
    157982: "Cenarion Ward (trigger)",
    200389: "Cultivation",
    200390: "Cultivation",
    # Talent-created spells (direct attribution)
    434141: "Dream Surge",
    439530: "Symbiotic Bloom",
    440121: "Bursting Growth",
    474760: "Thriving Growth",
    1264376: "Nature's Bounty",
    1244341: "Everbloom",
    392329: "Verdancy",
    392117: "Regenerative Heartwood",
    392124: "Embrace of the Dream",
    447132: "Thriving Vegetation",
    # Tranquility
    44203: "Tranquility (tick)",
    740: "Tranquility (channel)",
}

# Always-on passives for Restoration Druid that modify healing coefficients.
# These are spec passives that are always active, not optional talents.
ALWAYS_ON_PASSIVES = {
    137012,  # Restoration Druid (spec passive)
    # Add more if discovered (e.g., learned-by-default passives)
}


@dataclass
class SpellEffect:
    effect_id: int
    effect_type: str  # "Direct Heal", "Periodic Heal", "Add Percent Modifier", etc.
    subtype: str = ""  # "Spell Direct Amount", "Spell Periodic Amount", etc.
    sp_coefficient: float = 0.0
    base_value: int = 0
    period_ms: int = 0
    affected_spell_ids: list = field(default_factory=list)


@dataclass
class Spell:
    spell_id: int
    name: str
    duration_ms: int = 0
    effects: list = field(default_factory=list)


def parse_duration(text: str) -> int:
    """Parse duration string like '12 seconds' or '15 seconds' to ms."""
    m = re.search(r"([\d.]+)\s*seconds?", text)
    if m:
        return int(float(m.group(1)) * 1000)
    return 0


def parse_period(text: str) -> int:
    """Parse period from effect line like 'every 3 seconds'."""
    m = re.search(r"every ([\d.]+) seconds?", text)
    if m:
        return int(float(m.group(1)) * 1000)
    return 0


def parse_sp_coeff(text: str) -> float:
    """Extract SP Coefficient from effect detail line."""
    m = re.search(r"SP Coefficient:\s*([\d.]+)", text)
    if m:
        return float(m.group(1))
    return 0.0


def parse_base_value(text: str) -> int:
    """Extract Base Value from effect detail line."""
    m = re.search(r"Base Value:\s*(-?\d+)", text)
    if m:
        return int(m.group(1))
    return 0


def parse_affected_spells(text: str) -> list[int]:
    """Extract spell IDs from Affected Spells line."""
    return [int(x) for x in re.findall(r"\((\d+)\)", text) if x.isdigit()]


def parse_spells(lines: list[str]) -> dict[int, Spell]:
    """Parse all spell entries from SimC dump."""
    spells = {}
    current_spell = None
    current_effect = None
    in_affected_spells = False

    i = 0
    while i < len(lines):
        line = lines[i]

        # New spell entry
        m = re.match(r"^Name\s+:\s+(.+?)\s+\(id=(\d+)\)", line)
        if m:
            if current_spell and current_effect:
                current_spell.effects.append(current_effect)
            if current_spell:
                spells[current_spell.spell_id] = current_spell

            spell_name = m.group(1)
            spell_id = int(m.group(2))
            current_spell = Spell(spell_id=spell_id, name=spell_name)
            current_effect = None
            in_affected_spells = False
            i += 1
            continue

        if not current_spell:
            i += 1
            continue

        # Duration
        m = re.match(r"^Duration\s+:\s+(.+)", line)
        if m:
            current_spell.duration_ms = parse_duration(m.group(1))
            i += 1
            continue

        # Effect header
        m = re.match(r"^#(\d+)\s+\(id=(\d+)\)\s+:\s+(.+)", line)
        if m:
            if current_effect:
                current_spell.effects.append(current_effect)

            effect_id = int(m.group(2))
            effect_desc = m.group(3)

            effect_type = ""
            subtype = ""
            period = 0

            if "Periodic Heal" in effect_desc:
                effect_type = "Periodic Heal"
                period = parse_period(effect_desc)
            elif "Direct Heal" in effect_desc:
                effect_type = "Direct Heal"
            elif "Add Percent Modifier" in effect_desc:
                effect_type = "Add Percent Modifier"
                if "Spell Direct Amount" in effect_desc:
                    subtype = "direct"
                elif "Spell Periodic Amount" in effect_desc:
                    subtype = "periodic"
                elif "Spell Resource Cost" in effect_desc:
                    subtype = "resource_cost"
                elif "Spell Tick Time" in effect_desc:
                    subtype = "tick_time"

            current_effect = SpellEffect(
                effect_id=effect_id,
                effect_type=effect_type,
                subtype=subtype,
                period_ms=period,
            )
            in_affected_spells = False
            i += 1
            continue

        # Effect detail line (indented, contains Base Value / SP Coefficient)
        if current_effect and re.match(r"^\s+Base Value:", line):
            current_effect.sp_coefficient = parse_sp_coeff(line)
            current_effect.base_value = parse_base_value(line)
            i += 1
            continue

        # Affected Spells line
        if current_effect and "Affected Spells:" in line:
            # May span multiple lines via Family Flags continuation
            affected_text = line
            current_effect.affected_spell_ids = parse_affected_spells(affected_text)
            in_affected_spells = True
            i += 1
            continue

        # Continuation of affected spells (indented, before next section)
        if in_affected_spells and line.startswith("                   ") and "(" in line and not line.strip().startswith("Family"):
            current_effect.affected_spell_ids.extend(parse_affected_spells(line))
            i += 1
            continue
        else:
            in_affected_spells = False

        i += 1

    # Don't forget the last spell
    if current_spell and current_effect:
        current_spell.effects.append(current_effect)
    if current_spell:
        spells[current_spell.spell_id] = current_spell

    return spells


def compute_effective_coefficients(spells: dict[int, Spell]) -> dict:
    """Compute effective SP coefficients for healing spells.

    In Midnight (12.0+), Blizzard folded old passive multipliers directly into
    the raw coefficients. The spec passive 137012 only applies -6% (PvP-only).
    So raw SimC coefficients = effective PvE coefficients.
    """
    results = {}
    for spell_id, spell_name in HEALING_SPELLS.items():
        if spell_id not in spells:
            continue

        spell = spells[spell_id]
        spell_result = {
            "name": spell_name,
            "spell_id": spell_id,
            "duration_ms": spell.duration_ms,
            "effects": [],
        }

        for eff in spell.effects:
            if eff.effect_type not in ("Direct Heal", "Periodic Heal"):
                continue
            if eff.sp_coefficient == 0:
                continue

            eff_type = "periodic" if eff.effect_type == "Periodic Heal" else "direct"
            effect_data = {
                "type": eff_type,
                "coefficient": round(eff.sp_coefficient, 6),
            }
            if eff.period_ms:
                effect_data["period_ms"] = eff.period_ms

            spell_result["effects"].append(effect_data)

        if spell_result["effects"]:
            results[spell_id] = spell_result

    return results


def format_yaml(results: dict) -> str:
    """Format results as YAML for config/spell_coefficients.yaml."""
    lines = [
        "# Auto-generated from SimulationCraft spell data dump (Midnight 12.0+).",
        "# Re-generate with: python scripts/parse_simc_spells.py",
        "#",
        "# In Midnight, raw SimC coefficients = effective PvE coefficients.",
        "# coefficient is the fraction of spell power per tick/cast.",
        "# Example: Rejuv coefficient 0.803 means each tick heals for 80.3% of SP.",
        "",
        "spells:",
    ]

    for spell_id in sorted(results.keys()):
        entry = results[spell_id]
        lines.append(f"  {spell_id}:  # {entry['name']}")
        lines.append(f"    name: \"{entry['name']}\"")
        if entry["duration_ms"]:
            lines.append(f"    duration_ms: {entry['duration_ms']}")
        lines.append("    effects:")
        for eff in entry["effects"]:
            lines.append(f"      - type: {eff['type']}")
            lines.append(f"        coefficient: {eff['coefficient']}")
            if "period_ms" in eff:
                lines.append(f"        period_ms: {eff['period_ms']}")
    lines.append("")
    return "\n".join(lines)


def print_validation_table(results: dict):
    """Print a comparison table for manual validation against Wowhead."""
    print("\n=== Spell Coefficients (validate against Wowhead tooltips) ===")
    print(f"{'Spell':<30} {'Type':<10} {'Coefficient':<12} {'Tooltip%':<10}")
    print("-" * 62)

    for spell_id in sorted(results.keys()):
        entry = results[spell_id]
        for eff in entry["effects"]:
            tooltip_pct = f"{eff['coefficient'] * 100:.1f}%"
            print(f"{entry['name']:<30} {eff['type']:<10} {eff['coefficient']:<12.6f} {tooltip_pct:<10}")


def main():
    parser = argparse.ArgumentParser(description="Parse SimC druid spell data for healing coefficients")
    parser.add_argument("--dump", help="Path to local druid.txt (fetches from GitHub if not provided)")
    parser.add_argument("--output", default="config/spell_coefficients.yaml", help="Output YAML path")
    parser.add_argument("--validate", action="store_true", help="Print validation table")
    args = parser.parse_args()

    if args.dump:
        with open(args.dump) as f:
            text = f.read()
    else:
        print(f"Fetching {SIMC_URL}...", file=sys.stderr)
        with urllib.request.urlopen(SIMC_URL) as resp:
            text = resp.read().decode("utf-8")

    lines = text.splitlines()
    print(f"Parsing {len(lines)} lines...", file=sys.stderr)

    spells = parse_spells(lines)
    print(f"Found {len(spells)} spells", file=sys.stderr)

    results = compute_effective_coefficients(spells)
    print(f"Computed coefficients for {len(results)} healing spells", file=sys.stderr)

    if args.validate:
        print_validation_table(results)

    yaml_output = format_yaml(results)

    with open(args.output, "w") as f:
        f.write(yaml_output)
    print(f"Written to {args.output}", file=sys.stderr)


if __name__ == "__main__":
    main()
