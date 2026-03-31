from dataclasses import dataclass

import yaml


@dataclass
class TalentConfig:
    skip: bool = False
    skip_reason: str = ""
    multiplier: float | None = None


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
        )
    return result
