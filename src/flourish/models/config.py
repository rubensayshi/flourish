from dataclasses import dataclass, field

import yaml


@dataclass
class MasteryConfig:
    base_stacks: int = 2
    dr_table: list[float] = field(default_factory=lambda: [1.0, 1.7, 2.3, 2.8, 3.2])


@dataclass
class TalentConfig:
    skip: bool = False
    skip_reason: str = ""
    multiplier: float | None = None


@dataclass
class Config:
    mastery: MasteryConfig
    talents: dict[str, TalentConfig]


def load_config(path: str) -> Config:
    with open(path) as f:
        raw = yaml.safe_load(f) or {}

    mastery_raw = raw.pop("mastery", {}) or {}
    mastery = MasteryConfig(
        base_stacks=mastery_raw.get("base_stacks", 2),
        dr_table=mastery_raw.get("dr_table", [1.0, 1.7, 2.3, 2.8, 3.2]),
    )

    talents = {}
    for name, values in raw.items():
        values = values or {}
        talents[name] = TalentConfig(
            skip=values.get("skip", False),
            skip_reason=values.get("skip_reason", ""),
            multiplier=values.get("multiplier"),
        )
    return Config(mastery=mastery, talents=talents)
