from dataclasses import dataclass

OVERHEAL_WASTE_THRESHOLD = 0.5


@dataclass
class BaseEvent:
    timestamp: int
    source_id: int
    type: str


@dataclass
class CastEvent(BaseEvent):
    target_id: int
    ability_id: int


@dataclass
class HealEvent(BaseEvent):
    target_id: int
    ability_id: int
    amount: int
    overheal: int
    absorb: int  # healing absorbed by shields
    hit_type: int  # 1=normal, 2=crit

    @property
    def raw_heal(self) -> int:
        return self.amount + self.overheal + self.absorb

    @property
    def overheal_pct(self) -> float:
        raw = self.raw_heal
        return self.overheal / raw if raw > 0 else 0.0

    @property
    def is_wasted(self) -> bool:
        return self.overheal_pct > OVERHEAL_WASTE_THRESHOLD


@dataclass
class CombatantInfoEvent(BaseEvent):
    talent_nodes: set[int]  # set of nodeIDs from talentTree
    crit_spell: float
    haste_spell: float
    mastery: float
    spec_id: int


@dataclass
class ApplyBuffEvent(BaseEvent):
    target_id: int
    ability_id: int


@dataclass
class RefreshBuffEvent(BaseEvent):
    target_id: int
    ability_id: int


@dataclass
class RemoveBuffEvent(BaseEvent):
    target_id: int
    ability_id: int


EVENT_TYPE_MAP = {
    "cast": CastEvent,
    "heal": HealEvent,
    "applybuff": ApplyBuffEvent,
    "refreshbuff": RefreshBuffEvent,
    "removebuff": RemoveBuffEvent,
}


def parse_event(raw: dict) -> BaseEvent | None:
    event_type = raw.get("type")

    if event_type == "combatantinfo":
        talent_tree = raw.get("talentTree", [])
        talent_nodes = {t["nodeID"] for t in talent_tree}
        return CombatantInfoEvent(
            timestamp=raw["timestamp"],
            source_id=raw.get("sourceID", 0),
            type="combatantinfo",
            talent_nodes=talent_nodes,
            crit_spell=raw.get("critSpell", 0),
            haste_spell=raw.get("hasteSpell", 0),
            mastery=raw.get("mastery", 0),
            spec_id=raw.get("specID", 0),
        )

    cls = EVENT_TYPE_MAP.get(event_type)
    if cls is None:
        return None

    base = {
        "timestamp": raw["timestamp"],
        "source_id": raw.get("sourceID", 0),
        "type": event_type,
    }

    if cls is HealEvent:
        return HealEvent(
            **base,
            target_id=raw.get("targetID", 0),
            ability_id=raw.get("abilityGameID", 0),
            amount=raw.get("amount", 0),
            overheal=raw.get("overheal", 0),
            absorb=raw.get("absorb", 0),
            hit_type=raw.get("hitType", 1),
        )

    return cls(**base, target_id=raw.get("targetID", 0), ability_id=raw.get("abilityGameID", 0))
