# Talent Calculation Methods

How Flourish calculates the healing value of complex talents. Simple talents that just sum up healing from their own spell (e.g., Efflorescence, Ysera's Gift) are not listed here.

---

## Soul of the Forest (+ Power of the Archdruid)

SotF grants a 60% bonus to the next Rejuvenation, Regrowth, or Wild Growth after Swiftmend. We detect SotF consumption by watching for the SotF buff removal, then retroactively tag the HoT that consumed it.

- **Primary target:** attributed healing = `heal * (1 - 1/1.6)` (the 60% bonus portion)
- **Power of the Archdruid spreads:** when PotA copies a SotF-buffed Rejuv to nearby targets within 500ms, the spread copies are attributed at 100% (they wouldn't exist without PotA)
- **Regrowth direct heal:** the initial direct portion also gets the 60% bonus attribution

---

## Abundance

Abundance grants +8% crit chance to Regrowth per active Rejuvenation (including Germination), capped at 96%.

When a Regrowth **crits**, we calculate Abundance's share of the crit:

1. Count active Rejuvs at time of heal
2. `abundance_crit = rejuv_count * 0.08` (capped at 0.96)
3. `total_crit = base_crit + abundance_crit + other_sources`
4. Attributed healing = `(abundance_crit / total_crit) * crit_bonus_healing`

Where `crit_bonus_healing` is the extra healing from critting (heal amount minus what a non-crit would have done).

---

## Convoke the Spirits

During Convoke's channel (4s, or 3s with Cenarius' Guidance), both direct heals and HoTs applied are tracked.

- **Direct heals** during the channel: attributed at 70% (configurable multiplier)
- **HoTs applied** during the channel: tagged with `"convoke"`, and subsequent ticks are attributed at 70%
- **Pre-existing HoT ticks** during the channel are excluded (they'd happen without Convoke)

---

## Incarnation: Tree of Life

While the Tree of Life buff is active, different spells receive different bonuses:

- **Rejuvenation:** +50% healing bonus
- **Wild Growth:** +10% base bonus, plus extra targets. Extra-target healing is calculated as `(actual_targets - base_targets) / actual_targets * total_wg_healing`

Wild Growth healing is buffered per tick window and attributed during `finalize()` to accurately count targets per tick.

---

## Reforestation

Every 4th Swiftmend triggers a 10-second mini Tree of Life (16s with Potent Enchantments).

During this synthetic ToL window, the same multipliers as real ToL apply (Rejuv +50%, others +10%) — but **only when real ToL is not already active**, to avoid double-counting.

---

## Photosynthesis

Photosynthesis causes Lifebloom to bloom (trigger its final heal) more frequently. To isolate these extra blooms, we check every Lifebloom bloom event and classify it as "explained" or "unexplained":

**Explained blooms** (not from Photosynthesis):
- Natural expiry: `RemoveBuffEvent` within 200ms
- Refresh near expiry: `RefreshBuffEvent` within 200ms
- Manual recast: `CastEvent` within 200ms
- Everbloom proc: SotF consumption within 1200ms before bloom

**Unexplained blooms** are attributed to Photosynthesis.

---

## Harmonious Blooming

Lifebloom counts as 3 HoT stacks for Mastery instead of 1. We calculate the marginal mastery bonus from those 2 extra stacks using the diminishing returns table:

1. Look up mastery multiplier at `base_stacks` (without extra stacks) from DR table
2. Look up mastery multiplier at `base_stacks + 2` (with Harmonious Blooming)
3. `fraction = 1 - (mult_base / mult_with)`
4. Attributed healing = `heal * fraction` for each heal on targets with active Lifebloom

The DR table means each additional stack gives progressively less benefit.

---

## Improved Wild Growth

Wild Growth normally hits 5 targets; this talent adds 2 more (7 total). Attributed healing = `2/7 * total_wg_healing`.

**Exception:** during Tree of Life (which also adds targets), attribution is skipped to avoid double-counting with ToL's own calculation.

---

## Nurturing Dormancy

Extends Rejuvenation when the target takes damage. We attribute healing from Rejuv ticks that occur **beyond 17 seconds** after application (12s base + 3s Lingering Healing + 2s Germination = 17s normal max duration).

---

## Protective Growth

Regrowth applies 8% damage reduction to the target. We convert this DR into equivalent healing:

`attributed_healing = damage_taken_with_regrowth * 0.08 / (1 - 0.08)`

This represents the extra damage that would have been taken without the DR.

---

## Strategic Infusion (Wildstalker)

+4% crit chance on periodic (HoT) heals. When a HoT tick crits:

`attributed_healing = (0.04 / total_crit) * crit_bonus_healing`

Same logic as Abundance but with a flat 4% contribution instead of a scaling one.

---

## Vigorous Creepers (Wildstalker)

+20% healing on targets with active Symbiotic Bloom (excluding Symbiotic Bloom's own ticks).

`attributed_healing = heal - heal / 1.2`

---

## Root Network (Wildstalker)

+2% healing per active Symbiotic Bloom on any target.

`attributed_healing = heal * (0.02 * active_bloom_count)`

Scales dynamically with how many Symbiotic Blooms are out at the time of each heal.

---

## Implant (Wildstalker)

When Swiftmend or Wild Growth is cast, any Symbiotic Bloom applied within 500ms is tagged as an Implant proc. The full healing of these tagged blooms is attributed to Implant (they wouldn't exist without the talent).

---

## Symbiotic Bloom Mastery (Wildstalker)

Each Symbiotic Bloom adds 1 extra HoT stack for Mastery. Same DR-table calculation as Harmonious Blooming but for 1 extra stack per bloom:

1. Look up mastery multiplier at `base_stacks`
2. Look up mastery multiplier at `base_stacks + bloom_count`
3. `fraction = 1 - (mult_base / mult_with)`

---

## Harmony of the Grove (Keeper)

+5% healing per active Grove Guardian. Tracks guardian summons and maintains despawn timers (extended by 20% with Durability of Nature).

`attributed_healing = heal * (0.05 * active_guardian_count)`

---

## Power of Nature (Keeper)

+10% healing per active Grove Guardian, but **only** to Rejuvenation, Efflorescence, and Lifebloom. Uses the same guardian lifecycle tracking as Harmony of the Grove.

---

## SM / WG Cooldown Reduction (Keeper)

These talents reduce the cooldown of Swiftmend and Wild Growth respectively (via various sources like Renewing Surge, Early Spring, Dryad's Dance).

The calculation tracks cast timestamps and detects **on-cooldown usage** (casts that happen sooner than the unreduced cooldown would allow). For each such cast:

`ratio = 1 - (reduced_cd / unreduced_cd)`

The sum of ratios is multiplied by the downstream healing those extra casts enabled (e.g., Grove Guardians from extra Swiftmends, Efflorescence from extra Wild Growths).

SM cooldown reduction also simulates charges (2 with Prosperity) and accounts for Dryad's Dance 1.25x cooldown speed during Dryad overlap windows.

---

## Sylvan Beckoning (Keeper)

Identifies healing done by the Keeper's Dryad pet by checking `source_id != player_source_id`. The Dryad's healing spells share spell IDs with other sources, so pet source discrimination is essential for accurate attribution.
