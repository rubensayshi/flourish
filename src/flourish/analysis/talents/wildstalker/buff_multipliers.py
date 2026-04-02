from flourish.analysis.talents.buff_multipliers import StaticBuffAttributor


class WildstalkersPowerAttributor(StaticBuffAttributor):
    """Wildstalker's Power: Rejuvenation, Efflorescence, and Lifebloom healing increased by 10%."""
    name = "Wildstalker's Power"
    talent_node_id = 94621
    spell_ids = {774, 155777, 81269, 33763, 33778}  # Rejuv, Germination Rejuv, Efflorescence, LB tick, LB bloom
    multiplier = 0.1


class PatientCustodianAttributor(StaticBuffAttributor):
    """Patient Custodian: HoT effects are 6% more effective.
    Applies to all HoTs: Rejuv, Regrowth, Wild Growth, Lifebloom, etc."""
    name = "Patient Custodian"
    talent_node_id = 94630
    spell_ids = {774, 155777, 8936, 48438, 33763, 33778, 1244341, 1264664}
    # Rejuv, Germination, Regrowth, Wild Growth, Lifebloom tick, Lifebloom bloom, Everbloom, Rampant Growth Regrowth
    multiplier = 0.06
