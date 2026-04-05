/**
 * Mapping of talent display names → Wowhead spell IDs.
 * Used to generate tooltip links in ResultsTable.
 * Spell IDs sourced from docs/resto_druid_talents.md (Blizzard Game Data API).
 */
export const SPELL_IDS = {
  // Spec talents
  'Abundance':                  207383,
  'Convoke the Spirits':        391528,
  'Cultivation':                200390,
  'Efflorescence':              145205,
  'Embrace of the Dream':       392124,
  'Everbloom: Splash':       1244331,  // talent spell (rank-aware), not heal spell 1244341
  'Everbloom: Blooming Frenzy': 1244470,
  'Flourish':                   197721,
  'Grove Guardians':            1226140,
  'Harmonious Blooming':        392256,
  'Improved Swiftmend':         470549,
  'Improved Wild Growth':       328025,
  'Incarnation: Tree of Life':  33891,
  'Intensity':                  1264649,
  'Lifetreading':               1217941,
  'Liveliness':                 426702,
  "Nature's Bounty":            1263879,
  'Nurturing Dormancy':         392099,
  'Photosynthesis':             274902,
  'Power of the Archdruid':     392302,
  'Rampant Growth':             404521,
  'Reforestation':              392356,
  'Regenerative Heartwood':     392116,
  'Regenesis':                  383191,
  'SotF + PotA':                158478,   // Soul of the Forest
  'Thriving Vegetation: Rejuv':  447131,
  'Thriving Vegetation: Regrowth': 447131,
  'Unstoppable Growth':         382559,
  'Verdancy':                   392325,
  'Wild Synthesis':             400533,
  "Ysera's Gift":               145108,

  // Keeper of the Grove
  'Bounteous Bloom':            429215,
  "Cenarius' Might":            455797,
  'Dream Surge':                433831,
  "Grove's Inspiration":        429402,
  'Harmony of the Grove':       428731,
  'Potent Enchantments':        429420,
  'Power of Nature':            428859,
  'Protective Growth':          433748,
  "Early Spring + Dryad's Dance": 428937, // Early Spring
  'Early Spring (WG)':          428937,   // Early Spring
  'Spirit of the Thicket':      1264899,
  'Sylvan Beckoning':           1264614,

  // Wildstalker
  'Bursting Growth':            440120,
  'Implant':                    440118,
  'Patient Custodian':          1270592,
  'Root Network':               439882,
  'Strategic Infusion':         439890,
  'Symbiotic Bloom Mastery':    392256,   // Harmonious Blooming (closest related talent)
  'Thriving Growth':            439528,
  'Vigorous Creepers':          440119,
  "Wildstalker's Power":        439926,
}

export function wowheadUrl(name) {
  const id = SPELL_IDS[name]
  return id ? `https://www.wowhead.com/spell=${id}` : null
}
