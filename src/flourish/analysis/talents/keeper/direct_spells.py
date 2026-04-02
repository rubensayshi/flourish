from flourish.analysis.talents.direct_spell import DirectSpellAttributor
from flourish.models.events import CombatantInfoEvent, HealEvent
from flourish.tracking.hot_tracker import HotTracker
from flourish.tracking.buff_tracker import BuffTracker

WILD_SYNTHESIS_NODE = 94535
BOUNTEOUS_BLOOM_NODE = 94591
BOUNTEOUS_BLOOM_ENTRY = 117184  # WCL entryId (choice node vs Early Spring)


class GroveGuardiansAttributor(DirectSpellAttributor):
    """Grove Guardians: Treant pets cast Nourish and direct heals.
    Divides out Wild Synthesis (+30%) and Bounteous Bloom (+30%) multipliers
    so those talents can claim their portions without double-counting."""
    name = "Grove Guardians"
    talent_node_id = 82043
    spell_ids = {422090, 142421}  # Nourish + direct heal (from treants)
    allow_pet_source = True

    def __init__(self):
        super().__init__()
        self._divisor = 1.0

    def set_combatant_info(self, info: CombatantInfoEvent):
        super().set_combatant_info(info)
        self._divisor = 1.0
        if self.has_talent(WILD_SYNTHESIS_NODE):
            self._divisor *= 1.3
        if (BOUNTEOUS_BLOOM_NODE in info.talent_nodes
                and BOUNTEOUS_BLOOM_ENTRY in info.talent_ids):
            self._divisor *= 1.3

    def process_heal(self, event: HealEvent, hot_tracker: HotTracker, buff_tracker: BuffTracker) -> float:
        base = super().process_heal(event, hot_tracker, buff_tracker)
        if base > 0 and self._divisor > 1.0:
            return base / self._divisor
        return base


class DreamSurgeAttributor(DirectSpellAttributor):
    name = "Dream Surge"
    talent_node_id = 94600
    spell_ids = {434141}  # Dream Bloom


class SpiritOfTheThicketAttributor(DirectSpellAttributor):
    """Spirit of the Thicket: Ironbark summons a Dryad that channels a healing beam.
    The heal comes from a pet source, so we override the pet-source guard."""
    name = "Spirit of the Thicket"
    talent_node_id = 109712
    spell_ids = {1264905}
    allow_pet_source = True
