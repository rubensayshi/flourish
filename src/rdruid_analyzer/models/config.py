from dataclasses import dataclass

import yaml


@dataclass
class TalentConfig:
    skip: bool = False
    skip_reason: str = ""
    multiplier: float | None = None
    mastery_pct: float | None = None
    mastery_base_stacks: int | None = None
    mastery_dr_table: list[float] | None = None


def load_config(path: str) -> dict[str, TalentConfig]:
    with open(path) as f:
        raw = yaml.safe_load(f) or {}
    result = {}
    for name, values in raw.items():
        values = values or {}
        result[name] = TalentConfig(
            skip=values.get("skip", False),
            skip_reason=values.get("skip_reason", ""),
            multiplier=values.get("multiplier"),
            mastery_pct=values.get("mastery_pct"),
            mastery_base_stacks=values.get("mastery_base_stacks"),
            mastery_dr_table=values.get("mastery_dr_table"),
        )
    return result
