#!/usr/bin/env python3
"""Fetch Restoration Druid talent data from the Blizzard Game Data API.

Writes docs/resto_druid_talents.md with spec tree, class tree, and hero talent trees.
Requires BNET_CLIENT_ID and BNET_CLIENT_SECRET in .env (or environment).

Hero talent descriptions are supplemented from the Wowhead tooltip API because the
Blizzard API returns the "canonical" spec's description (usually Balance/Feral) even
on the spec-specific endpoint — it does not return Resto-specific descriptions.

Usage:
    python scripts/fetch_talents.py
"""

import html
import json
import os
import re
import sys
import time
from pathlib import Path

import requests
from dotenv import load_dotenv

RESTO_SPEC_ID = 105
TALENT_TREE_ID = 793
NAMESPACE = "static-us"
LOCALE = "en_US"
API_BASE = "https://us.api.blizzard.com"
WOWHEAD_TOOLTIP_URL = "https://nether.wowhead.com/tooltip/spell/{spell_id}?dataEnv=1&locale=0"

load_dotenv(Path(__file__).resolve().parent.parent / ".env")


def get_token() -> str:
    resp = requests.post(
        "https://oauth.battle.net/token",
        data={"grant_type": "client_credentials"},
        auth=(os.environ["BNET_CLIENT_ID"], os.environ["BNET_CLIENT_SECRET"]),
    )
    resp.raise_for_status()
    return resp.json()["access_token"]


def api_get(token: str, path: str) -> dict:
    resp = requests.get(
        f"{API_BASE}{path}",
        params={"namespace": NAMESPACE, "locale": LOCALE},
        headers={"Authorization": f"Bearer {token}"},
    )
    resp.raise_for_status()
    return resp.json()


def extract_resto_description(tooltip_html: str) -> str | None:
    """Extract the Restoration-specific description from a Wowhead tooltip.

    Wowhead tooltips use <span class='q2'>... Restoration</span><br/> to mark
    spec-specific sections. Returns the Resto section text, or None if not found.
    """
    # Find Restoration section: after "Restoration</span>" until the next spec header or end
    pattern = r"Restoration</span>\s*<br\s*/?>(.+?)(?:<span class=['\"]q2['\"]|<!--sp|$)"
    match = re.search(pattern, tooltip_html, re.DOTALL)
    if not match:
        return None
    raw = match.group(1)
    # Clean HTML tags
    text = re.sub(r"<br\s*/?>", "\n", raw)
    text = re.sub(r"<[^>]+>", "", text)
    text = html.unescape(text).strip()
    # Remove trailing blank lines
    text = re.sub(r"\n{3,}", "\n\n", text)
    return text if text else None


def fetch_wowhead_resto_description(spell_id: int) -> str | None:
    """Fetch spell tooltip from Wowhead and extract the Restoration description."""
    try:
        resp = requests.get(WOWHEAD_TOOLTIP_URL.format(spell_id=spell_id), timeout=10)
        if not resp.ok:
            return None
        data = resp.json()
        tooltip = data.get("tooltip", "")
        return extract_resto_description(tooltip)
    except Exception as e:
        print(f"  Warning: Wowhead fetch failed for spell {spell_id}: {e}")
        return None


def check_talent_spec(token: str, talent_id: int) -> int | None:
    """Check which spec a talent belongs to. Returns spec ID or None."""
    try:
        data = api_get(token, f"/data/wow/talent/{talent_id}")
        return data.get("playable_specialization", {}).get("id")
    except Exception:
        return None


def fix_spell_description(token: str, spell_tt: dict, talent: dict) -> str:
    """Try to get Resto-specific description for a hero talent spell.
    Returns the source label."""
    talent_id = talent.get("id")
    spell = spell_tt.get("spell", {})
    spell_id = spell.get("id")
    name = talent.get("name", "?")

    if not (talent_id and spell_id):
        return "no-ids"

    # Check if Wowhead has a Resto-specific section
    resto_desc = fetch_wowhead_resto_description(spell_id)
    if resto_desc:
        spell_tt["description"] = resto_desc
        print(f"    Fixed: {name}")
        return "wowhead-resto"

    # If no Resto section on Wowhead, check if it's even a Resto talent
    spec_id = check_talent_spec(token, talent_id)
    if spec_id and spec_id != RESTO_SPEC_ID:
        print(f"    No Resto desc (non-Resto talent): {name}")
        return "blizzard-wrong-spec"

    # Resto-tagged in Blizzard API, no separate Resto section on Wowhead = description is fine
    return "blizzard-ok"


def fix_hero_descriptions(token: str, data: dict) -> dict:
    """For all hero talent nodes, try to get Resto-specific descriptions from Wowhead.
    The Blizzard API has two issues: (1) some talents return the wrong spec's description
    entirely, (2) some return multi-spec text with Feral/Balance mixed in."""
    RESTO_HERO_TREES = {"Wildstalker", "Keeper of the Grove"}
    fixed = 0
    checked = 0
    for ht in data.get("hero_talent_trees", []):
        tree_name = ht.get("name", "")
        if tree_name not in RESTO_HERO_TREES:
            print(f"  Skipping non-Resto hero tree: {tree_name}")
            continue
        print(f"  Checking hero tree: {tree_name}...")
        for node in ht.get("hero_talent_nodes", []):
            for rank in node.get("ranks", []):
                tooltip = rank.get("tooltip", {})
                spell_tt = tooltip.get("spell_tooltip", {})
                talent = tooltip.get("talent", {})

                if talent.get("id") and spell_tt.get("spell", {}).get("id"):
                    checked += 1
                    source = fix_spell_description(token, spell_tt, talent)
                    spell_tt["_source"] = source
                    if source == "wowhead-resto":
                        fixed += 1
                    time.sleep(0.05)

                for ct in rank.get("choice_of_tooltips", []):
                    ct_talent = ct.get("talent", {})
                    ct_spell_tt = ct.get("spell_tooltip", {})

                    if ct_talent.get("id") and ct_spell_tt.get("spell", {}).get("id"):
                        checked += 1
                        source = fix_spell_description(token, ct_spell_tt, ct_talent)
                        ct_spell_tt["_source"] = source
                        if source == "wowhead-resto":
                            fixed += 1
                        time.sleep(0.05)

    print(f"  Checked {checked} hero talents, fixed {fixed} descriptions")
    return data


def format_node(node: dict) -> str:
    """Format a single talent node as markdown."""
    lines = []
    node_id = node["id"]
    node_type = node["node_type"]["type"]
    row = node.get("display_row", "?")
    col = node.get("display_col", "?")

    ranks = node.get("ranks", [])
    if not ranks:
        return ""

    rank = ranks[0]
    tooltip = rank.get("tooltip", {})
    talent = tooltip.get("talent", {})
    spell_tt = tooltip.get("spell_tooltip", {})
    spell = spell_tt.get("spell", {})
    talent_name = talent.get("name", "Unknown")
    talent_id = talent.get("id", "?")
    spell_id = spell.get("id", "?")
    spell_name = spell.get("name", talent_name)
    description = spell_tt.get("description", "").replace("\r\n", "\n").strip()
    cast_time = spell_tt.get("cast_time", "")
    cooldown = spell_tt.get("cooldown", "")

    # Handle choice nodes
    choice_tooltips = rank.get("choice_of_tooltips", [])
    if choice_tooltips:
        talent_name = " / ".join(
            ct.get("talent", {}).get("name", "?") for ct in choice_tooltips
        )

    num_ranks = len(ranks)
    rank_str = f" ({num_ranks} ranks)" if num_ranks > 1 else ""

    lines.append(f"### {talent_name}{rank_str} — Row {row}, Col {col} (node {node_id})")
    lines.append("")

    if choice_tooltips:
        for ct in choice_tooltips:
            ct_talent = ct.get("talent", {})
            ct_spell_tt = ct.get("spell_tooltip", {})
            ct_spell = ct_spell_tt.get("spell", {})
            ct_desc = ct_spell_tt.get("description", "").replace("\r\n", "\n").strip()
            ct_cast = ct_spell_tt.get("cast_time", "")
            ct_cd = ct_spell_tt.get("cooldown", "")

            lines.append(f"**Choice: {ct_talent.get('name', '?')}**")
            lines.append("")
            lines.append(f"- **Definition ID**: {ct_talent.get('id', '?')}, **Spell**: {ct_spell.get('name', '?')} (ID: {ct_spell.get('id', '?')})")
            lines.append(f"- **Cast**: {ct_cast}")
            if ct_cd:
                lines.append(f"- **Cooldown**: {ct_cd}")
            lines.append(f"- {ct_desc}")
            lines.append("")
    else:
        lines.append(f"- **Type**: {node_type}")
        lines.append(f"- **Definition ID**: {talent_id}, **Spell**: {spell_name} (ID: {spell_id})")
        lines.append(f"- **Cast**: {cast_time}")
        if cooldown:
            lines.append(f"- **Cooldown**: {cooldown}")
        lines.append(f"- {description}")
        lines.append("")

    return "\n".join(lines)


def sort_nodes(nodes: list[dict]) -> list[dict]:
    """Sort nodes by display_row then display_col."""
    return sorted(nodes, key=lambda n: (n.get("display_row", 0), n.get("display_col", 0)))


def main():
    token = get_token()

    print(f"Fetching talent tree {TALENT_TREE_ID} for spec {RESTO_SPEC_ID}...")
    data = api_get(token, f"/data/wow/talent-tree/{TALENT_TREE_ID}/playable-specialization/{RESTO_SPEC_ID}")

    print("Fixing hero talent descriptions from Wowhead...")
    data = fix_hero_descriptions(token, data)

    # Save raw JSON (with fixes applied)
    raw_path = Path(__file__).resolve().parent.parent / "data" / "resto_talent_tree_raw.json"
    raw_path.parent.mkdir(exist_ok=True)
    with open(raw_path, "w") as f:
        json.dump(data, f, indent=2)
    print(f"Saved raw JSON to {raw_path}")

    # Collect hero node IDs to exclude from spec tree (they appear in both)
    hero_node_ids = set()
    for ht in data.get("hero_talent_trees", []):
        for node in ht.get("hero_talent_nodes", []):
            hero_node_ids.add(node["id"])

    # Build markdown
    md = []
    md.append("# Restoration Druid Talent Data (Midnight Season 1 / 12.0.1)")
    md.append("")
    md.append("Auto-generated from Blizzard Game Data API (`static-us` namespace, spec ID 105).")
    md.append("Hero talent descriptions sourced from Wowhead tooltip API (Blizzard API bug: returns wrong spec).")
    md.append(f"Regenerate with: `python scripts/fetch_talents.py`")
    md.append("")

    # Spec talent tree (excluding hero talent nodes that appear inline)
    md.append("## Spec Talent Tree")
    md.append("")
    for node in sort_nodes(data.get("spec_talent_nodes", [])):
        if node["id"] in hero_node_ids:
            continue
        section = format_node(node)
        if section:
            md.append(section)

    # Class talent tree
    md.append("## Class Talent Tree")
    md.append("")
    for node in sort_nodes(data.get("class_talent_nodes", [])):
        section = format_node(node)
        if section:
            md.append(section)

    # Hero talent trees (only Resto-available: Wildstalker and Keeper of the Grove)
    RESTO_HERO_TREES = {"Wildstalker", "Keeper of the Grove"}
    md.append("## Hero Talent Trees")
    md.append("")
    for hero_tree in data.get("hero_talent_trees", []):
        tree_name = hero_tree.get("name", "Unknown")
        if tree_name not in RESTO_HERO_TREES:
            continue
        nodes = hero_tree.get("hero_talent_nodes", [])
        md.append(f"### _{tree_name}_ ({len(nodes)} talents)")
        md.append("")
        for node in sort_nodes(nodes):
            section = format_node(node)
            if section:
                section = section.replace("### ", "#### ", 1)
                md.append(section)

    out_path = Path(__file__).resolve().parent.parent / "docs" / "resto_druid_talents.md"
    with open(out_path, "w") as f:
        f.write("\n".join(md))

    spec_count = len(data.get("spec_talent_nodes", []))
    class_count = len(data.get("class_talent_nodes", []))
    hero_count = sum(len(ht.get("hero_talent_nodes", [])) for ht in data.get("hero_talent_trees", []))
    print(f"Wrote {out_path}: {spec_count} spec + {class_count} class + {hero_count} hero talents")


if __name__ == "__main__":
    main()
