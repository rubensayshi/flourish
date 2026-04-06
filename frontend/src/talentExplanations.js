/**
 * Explanations of how complex talent values are calculated.
 * Keys must match talent display names used in ResultsTable.
 * Values are HTML strings rendered via v-html.
 */
export const TALENT_EXPLANATIONS = {
  'SotF + PotA': `SotF grants a 60% bonus to the next Rejuv, Regrowth, or Wild Growth after Swiftmend. We detect SotF consumption by watching for the buff removal, then retroactively tag the HoT that consumed it.
<br><br><b>Primary target:</b> the 60% bonus portion of each heal.<br><b>PotA spreads:</b> when PotA copies a SotF-buffed Rejuv to nearby targets, the copies are attributed at 100% (they wouldn't exist without PotA).`,

  'Abundance': `+8% crit to Regrowth per active Rejuvenation (incl. Germination), capped at 96%.
<br><br>On Regrowth crits, we calculate Abundance's share: <code>abundance_crit / total_crit \u00d7 crit_bonus_healing</code>.`,

  'Convoke the Spirits': `During Convoke's channel (4s, or 3s with Cenarius' Guidance):
<br><br><b>Direct heals</b> during channel: attributed at 100%.<br><b>HoTs applied</b> during channel: tagged, subsequent ticks attributed at 100%.<br>Pre-existing HoT ticks during the channel are excluded.
<br><br><b>Opportunity cost:</b> first 3 Rejuv/Regrowth casts are skipped (you could have cast those yourself).`,

  'Incarnation: Tree of Life': `While Tree of Life is active:
<br><br><b>Rejuvenation:</b> +50% healing bonus.<br><b>Wild Growth:</b> +10% base bonus, plus extra-target healing calculated as <code>(actual_targets \u2212 base_targets) / actual_targets \u00d7 total_healing</code>.`,

  'Reforestation': `Every 4th Swiftmend triggers a mini Tree of Life (10s base).
<br><br>Same multipliers as real ToL (Rejuv +50%, others +10%), but only when real ToL is not already active to avoid double-counting.
<br><br>With Potent Enchantments, duration extends to 16s \u2014 healing in the extra 6s is attributed to Potent Enchantments separately.`,

  'Potent Enchantments': `Extends Reforestation\u2019s mini Tree of Life by 6s (10s \u2192 16s).
<br><br>Healing during the extra 6s window uses the same ToL multipliers (Rejuv +50%, others +10%) and is split out from Reforestation\u2019s base 10s attribution.`,

  'Photosynthesis': `Attributes extra Lifebloom blooms caused by Photosynthesis. Every bloom is classified as "explained" or not:
<br><br><b>Explained</b> (not Photosynthesis): natural expiry, refresh near expiry, manual recast, or Everbloom proc.<br><b>Unexplained</b> blooms \u2192 attributed to Photosynthesis.`,

  'Harmonious Blooming': `Lifebloom counts as 3 HoT stacks for Mastery instead of 1. Calculates the marginal mastery bonus from those 2 extra stacks using the diminishing returns table.
<br><br><code>fraction = 1 \u2212 (mult_at_base_stacks / mult_at_base+2_stacks)</code>`,

  'Improved Wild Growth': `WG normally hits 5 targets; this adds 2 more. Attributed = 2/7 of total WG healing.
<br><br>Skipped during Tree of Life to avoid double-counting with ToL's own calculation.`,

  'Nurturing Dormancy': `Attributes Rejuv healing from ticks beyond 17s after application (12s base + 3s Lingering Healing + 2s Germination).
<br><br>Any healing past that threshold is from Nurturing Dormancy's damage-triggered extensions.`,

  'Protective Growth': `Regrowth applies 8% DR to the target. Converted to equivalent healing:
<br><br><code>damage_taken \u00d7 0.08 / (1 \u2212 0.08)</code>`,

  'Strategic Infusion': `+4% crit on periodic heals. On HoT crits:
<br><br><code>attributed = 0.04 / total_crit \u00d7 crit_bonus_healing</code>`,

  'Vigorous Creepers': `+20% healing on targets with active Symbiotic Bloom (excluding Symbiotic Bloom's own ticks).
<br><br><code>attributed = heal \u2212 heal / 1.2</code>`,

  'Root Network': `+2% healing per active Symbiotic Bloom.
<br><br>Scales dynamically: <code>heal \u00d7 0.02 \u00d7 active_bloom_count</code>.`,

  'Implant': `When Swiftmend or Wild Growth is cast, any Symbiotic Bloom applied within 500ms is tagged as an Implant proc. Full healing of tagged blooms is attributed (they wouldn't exist without the talent).`,

  'Twin Sprouts': `Detects Twin Sprouts procs by finding Symbiotic Bloom applications within 50ms of another bloom on a different target. Full healing of proc'd blooms is attributed. Note: may slightly over-count if two natural blooms proc simultaneously (e.g. from WG ticks).`,

  'Symbiotic Bloom Mastery': `Each Symbiotic Bloom adds 1 extra HoT stack for Mastery. Same DR-table calculation as Harmonious Blooming but per bloom:
<br><br><code>fraction = 1 \u2212 (mult_at_base / mult_at_base+bloom_count)</code>`,

  'Grove Guardians': `100% of treant healing (Nourish + direct heals), minus portions claimed by Wild Synthesis (+30%) and Bounteous Bloom (+30%) to avoid double-counting.
<br><br>If both are talented: <code>heal / 1.3 / 1.3</code>.`,

  'Wild Synthesis': `+30% to Grove Guardian, Efflorescence, and Dream Bloom healing. Grove Guardians reduces its own value accordingly.`,

  'Bounteous Bloom': `+30% to Grove Guardian healing. Grove Guardians reduces its own value accordingly.`,

  'Harmony of the Grove': `+5% healing per active Grove Guardian. Tracks guardian summons and despawn timers (extended 20% with Durability of Nature).
<br><br><code>heal \u00d7 0.05 \u00d7 active_guardian_count</code>`,

  'Power of Nature': `+10% healing per active Grove Guardian, but only to Rejuvenation, Efflorescence, and Lifebloom. Same guardian lifecycle tracking as Harmony of the Grove.`,

  "Early Spring + Dryad's Dance": `Tracks cast timestamps and detects on-cooldown usage for Swiftmend (casts sooner than unreduced CD allows).
<br><br><code>ratio = 1 \u2212 (reduced_cd / unreduced_cd)</code>. Sum of ratios \u00d7 downstream healing from those extra casts.
<br><br>Simulates charges (2 with Prosperity) and accounts for Dryad's Dance 1.25\u00d7 CD speed.`,

  'Early Spring (WG)': `Tracks Wild Growth cast timestamps and detects on-cooldown usage.
<br><br><code>ratio = 1 \u2212 (reduced_cd / unreduced_cd)</code>. Sum of ratios \u00d7 downstream healing (Efflorescence uptime from extra WG casts).`,

  'Sylvan Beckoning': `Identifies healing done by the Keeper's Dryad pet by checking <code>source_id \u2260 player_source_id</code>. The Dryad's spells share IDs with other sources, so pet source discrimination is essential.`,

  'Thriving Vegetation: Regrowth': `Regrowth HoT duration is increased by 3 sec per rank. Tracks where the HoT would expire without this bonus (base 12s + pandemic), and attributes ticks beyond that point.
<br><br>On refresh: if the non-TV HoT would have already expired, the refresh is treated as a fresh application (no pandemic carry-over).`,
}

export function hasExplanation(name) {
  return name in TALENT_EXPLANATIONS
}

export function getExplanation(name) {
  return TALENT_EXPLANATIONS[name] || null
}
