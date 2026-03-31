# Restoration Druid Talent Data (Midnight Season 1 / 12.0.1)

Pulled from Blizzard Game Data API (`static-us` namespace).

Note: The spec tree includes hero talent nodes inline (Wildstalker at cols 24-27, Keeper of the Grove at cols 12-15).
These are also listed separately under Hero Talent Trees below.

## Spec Talent Tree

### Lifebloom — Row 2, Col 20 (node 82049)

- **Type**: ACTIVE
- **Talent ID**: 108105, **Spell**: Lifebloom (ID: 33763)
- **Cast**: Instant
- Heals the target for 1,715 over 15 sec. When Lifebloom expires, is dispelled, or is refreshed near expiration the target is instantly healed for 849.

May be active on one target at a time.

### Thriving Growth — Row 2, Col 26 (node 94626)

- **Type**: PASSIVE
- **Talent ID**: 122238, **Spell**: Thriving Growth (ID: 439528)
- **Cast**: Passive
- Rip and Rake damage has a chance to cause Bloodseeker Vines to grow on the victim, dealing 34 Bleed damage over 6 sec.

Wild Growth, Regrowth, and Efflorescence healing has a chance to cause Symbiotic Blooms to grow on the target, healing for 834 over 6 sec.

Multiple instances of these can overlap.

### Swiftmend — Row 3, Col 19 (node 82047)

- **Type**: ACTIVE
- **Talent ID**: 108103, **Spell**: Swiftmend (ID: 18562)
- **Cast**: Instant
- Consume a Regrowth, Wild Growth, or Rejuvenation effect to instantly heal an ally for 3,976.

### Nature's Swiftness — Row 3, Col 20 (node 82050)

- **Type**: ACTIVE
- **Talent ID**: 108106, **Spell**: Nature's Swiftness (ID: 132158)
- **Cast**: Instant
- **Cooldown**: 1 min cooldown
- Your next Regrowth, Rebirth, or Entangling Roots is instant, free, castable in all forms, and heals for an additional 60%.

### Omen of Clarity — Row 3, Col 21 (node 104125)

- **Type**: PASSIVE
- **Talent ID**: 133509, **Spell**: Omen of Clarity (ID: 113043)
- **Cast**: Passive
- Your healing over time from Lifebloom has a 4% chance to cause a Clearcasting state, making your next Regrowth cost no mana.

### Hunt Beneath the Open Skies — Row 3, Col 24 (node 94629)

- **Type**: PASSIVE
- **Talent ID**: 122243, **Spell**: Hunt Beneath the Open Skies (ID: 439868)
- **Cast**: Passive
- Damage and healing while in Cat Form increased by 3%.

Moonfire and Sunfire damage increased by 10%.

### Strategic Infusion — Row 3, Col 25 (node 94623)

- **Type**: PASSIVE
- **Talent ID**: 122235, **Spell**: Strategic Infusion (ID: 439890)
- **Cast**: Passive
- Attacking from Prowl increases the chance for Shred, Rake, and Swipe to critically strike by 8% for 6 sec.

Your periodic heals have a 4% increased chance to critically heal.

### Wildstalker's Power — Row 3, Col 26 (node 94621)

- **Type**: PASSIVE
- **Talent ID**: 122233, **Spell**: Wildstalker's Power (ID: 439926)
- **Cast**: Passive
- Rip and Ferocious Bite damage increased by 5%.

Rejuvenation healing increased by 10%.

### Green Thumb — Row 3, Col 27 (node 109717)

- **Type**: PASSIVE
- **Talent ID**: 140730, **Spell**: Green Thumb (ID: 1270565)
- **Cast**: Passive
- The rate at which Symbiotic Blooms grow is increased by 20%.

### [Choice] Row 4, Col 18 (node 82079)

**Option: Verdant Infusion** (talent 108142, spell 392410: Verdant Infusion)
- Cast: Passive
- Swiftmend no longer consumes a heal over time effect, and extends the duration of your heal over time effects on the target by 8 sec.

**Option: Prosperity** (talent 108141, spell 200383: Prosperity)
- Cast: Passive
- Swiftmend now has 2 charges.

### [Choice] Row 4, Col 20 (node 82051)

**Option: Nature's Splendor** (talent 108108, spell 392288: Nature's Splendor)
- Cast: Passive
- The healing bonus to Regrowth from Nature's Swiftness is increased by 35%.

**Option: Passing Seasons** (talent 108107, spell 382550: Passing Seasons)
- Cast: Passive
- Nature's Swiftness's cooldown is reduced by 12 sec.

### Improved Regrowth — Row 4, Col 22 (node 82083)

- **Type**: PASSIVE
- **Talent ID**: 108147, **Spell**: Improved Regrowth (ID: 231032)
- **Cast**: Passive
- Regrowth's initial heal has a 40% increased chance for a critical effect if the target is already affected by Regrowth.

### Lethal Preservation — Row 4, Col 24 (node 94624)

- **Type**: PASSIVE
- **Talent ID**: 122236, **Spell**: Lethal Preservation (ID: 455461)
- **Cast**: Passive
- When you remove an effect with Soothe or Remove Corruption, gain a combo point and heal for 4% of your maximum health. If you are at full health an injured party or raid member will be healed instead.

### [Choice] Row 4, Col 25 (node 94622)

**Option: Entangling Vortex** (talent 122234, spell 439895: Entangling Vortex)
- Cast: Passive
- Enemies pulled into Ursol's Vortex are rooted in place for 3 sec. Damage may cancel the effect.

**Option: Flower Walk** (talent 124755, spell 439901: Flower Walk)
- Cast: Passive
- During Barkskin your movement speed is increased by 10% and every second flowers grow beneath your feet that heal up to 3 nearby injured allies for 76.

### [Choice] Row 4, Col 26 (node 94625)

**Option: Bond with Nature** (talent 122237, spell 439929: Bond with Nature)
- Cast: Passive
- Healing you receive is increased by 4%.

**Option: Harmonious Constitution** (talent 124754, spell 440116: Harmonious Constitution)
- Cast: Passive
- Your Regrowth's healing to yourself is increased by 35%.

### Bursting Growth — Row 4, Col 27 (node 109716)

- **Type**: PASSIVE
- **Talent ID**: 140729, **Spell**: Bursting Growth (ID: 440120)
- **Cast**: Passive
- When Bloodseeker Vines expire or you use Ferocious Bite on their target they explode in thorns, dealing 30 physical damage to nearby enemies. Damage reduced above 5 targets.

When Symbiotic Blooms expire or you cast Rejuvenation on their target flowers grow around their target, healing them and up to 3 nearby allies for 160.

### Soul of the Forest — Row 5, Col 19 (node 82055)

- **Type**: PASSIVE
- **Talent ID**: 108114, **Spell**: Soul of the Forest (ID: 158478)
- **Cast**: Passive
- Swiftmend increases the healing of your next Regrowth or Rejuvenation by 60%.

### Tranquil Mind — Row 5, Col 21 (node 92674)

- **Type**: PASSIVE
- **Talent ID**: 119815, **Spell**: Tranquil Mind (ID: 403521)
- **Cast**: Passive
- Increases Omen of Clarity's chance to activate Clearcasting to 5% and Clearcasting can stack 1 additional time.

### [Choice] Row 5, Col 24 (node 94631)

**Option: Resilient Flourishing** (talent 122246, spell 439880: Resilient Flourishing)
- Cast: Passive
- Bloodseeker Vines and Symbiotic Blooms last 2 additional sec.

When a target afflicted by Bloodseeker Vines dies, the vines jump to a valid nearby target for their remaining duration.

**Option: Root Network** (talent 122245, spell 439882: Root Network)
- Cast: Passive
- Each active Bloodseeker Vine increases the damage your abilities deal by 2%.

Each active Symbiotic Bloom increases the healing of your spells by 2%.

### Patient Custodian — Row 5, Col 25 (node 94630)

- **Type**: PASSIVE
- **Talent ID**: 122244, **Spell**: Patient Custodian (ID: 1270592)
- **Cast**: Passive
- Your heal over time effects are 6% more effective.

### [Choice] Row 5, Col 26 (node 94628)

**Option: Twin Sprouts** (talent 122242, spell 440117: Twin Sprouts)
- Cast: Passive
- When Bloodseeker Vines or Symbiotic Blooms grow, they have a 30% chance to cause another growth of the same type to immediately grow on a valid nearby target.

**Option: Implant** (talent 122241, spell 440118: Implant)
- Cast: Passive
- Casting Swiftmend or Wild Growth causes a Symbiotic Bloom to grow on a target for 6 sec.

### Rampancy — Row 5, Col 27 (node 109715)

- **Type**: PASSIVE
- **Talent ID**: 140728, **Spell**: Rampancy (ID: 1270586)
- **Cast**: Passive
- Symbiotic Blooms have a 20% chance to trigger Bursting Growth every 2 sec at 100% effectiveness.


### Efflorescence — Row 6, Col 18 (node 82057)

- **Type**: ACTIVE
- **Talent ID**: 108116, **Spell**: Efflorescence (ID: 145205)
- **Cast**: Instant
- Grows a healing blossom at the target location, restoring 221 health to three injured allies within 10 yards every 1.8 sec for 30 sec. Limit 1.

### Tranquility — Row 6, Col 20 (node 82054)

- **Type**: ACTIVE
- **Talent ID**: 108113, **Spell**: Tranquility (ID: 740)
- **Cast**: Channeled
- **Cooldown**: 3 min cooldown
- Heals all allies within 40 yards for 19,334 over 5.4 sec.

Healing decreased beyond 5 targets.

### Ironbark — Row 6, Col 22 (node 82082)

- **Type**: ACTIVE
- **Talent ID**: 108146, **Spell**: Ironbark (ID: 102342)
- **Cast**: Instant
- **Cooldown**: 1.5 min cooldown
- The target's skin becomes as tough as Ironwood, reducing damage taken by 20% for 12 sec.

### Vigorous Creepers — Row 6, Col 26 (node 94627)

- **Type**: PASSIVE
- **Talent ID**: 122239, **Spell**: Vigorous Creepers (ID: 440119)
- **Cast**: Passive
- Bloodseeker Vines increase the damage your abilities deal to affected enemies by 4%.

Symbiotic Blooms increase the healing your spells do to affected targets by 20%.

### Dream Surge — Row 7, Col 13 (node 94600)

- **Type**: PASSIVE
- **Talent ID**: 122207, **Spell**: Dream Surge (ID: 433831)
- **Cast**: Passive
- When Grove Guardians are summoned, they grow Dream Petals on your target, healing up to 3 nearby allies for 542.

### Verdancy — Row 7, Col 17 (node 82059)

- **Type**: PASSIVE
- **Talent ID**: 108118, **Spell**: Verdancy (ID: 392325)
- **Cast**: Passive
- When Lifebloom blooms, up to 3 targets within your Efflorescence are healed for 360.

### Lifetreading — Row 7, Col 18 (node 103874)

- **Type**: PASSIVE
- **Talent ID**: 133510, **Spell**: Lifetreading (ID: 1217941)
- **Cast**: Passive
- Efflorescence healing increased by 25%, and it now automatically grows beneath your Lifebloom target's feet.

### Grove Guardians — Row 7, Col 19 (node 82043)

- **Type**: PASSIVE
- **Talent ID**: 122116, **Spell**: Grove Guardians (ID: 1226140)
- **Cast**: Passive
- Casting Swiftmend or Wild Growth summons a Treant that casts Nourish on that target or a nearby ally periodically, healing for 229. Lasts 8 sec.

### [Choice] Row 7, Col 20 (node 82053)

**Option: Inner Peace** (talent 108112, spell 197073: Inner Peace)
- Cast: Passive
- Reduces the cooldown of Tranquility by 30 sec.

While channeling Tranquility, you take 20% reduced damage and are immune to knockbacks.

**Option: Flourish** (talent 108111, spell 197721: Flourish)
- Cast: Passive
- Tranquility extends the duration of all of your heal over time effects by 2 sec every 0.9 sec.

### Cultivation — Row 7, Col 21 (node 82056)

- **Type**: PASSIVE
- **Talent ID**: 108115, **Spell**: Cultivation (ID: 200390)
- **Cast**: Passive
- When Rejuvenation heals a target below 60% health, they are instantly healed for 115.

### Improved Wild Growth — Row 7, Col 22 (node 82045)

- **Type**: PASSIVE
- **Talent ID**: 108101, **Spell**: Improved Wild Growth (ID: 328025)
- **Cast**: Passive
- Wild Growth heals 2 additional targets.

### [Choice] Row 7, Col 23 (node 82081)

**Option: Stonebark** (talent 108145, spell 197061: Stonebark)
- Cast: Passive
- Ironbark increases healing from your heal over time effects by 20%.

**Option: Improved Ironbark** (talent 108144, spell 382552: Improved Ironbark)
- Cast: Passive
- Ironbark's cooldown is reduced by 20 sec.

### Treants of the Moon — Row 8, Col 12 (node 94599)

- **Type**: PASSIVE
- **Talent ID**: 122206, **Spell**: Treants of the Moon (ID: 428544)
- **Cast**: Passive
- Your Grove Guardians cast Moonfire on nearby targets about once every 6 sec.

### Expansiveness — Row 8, Col 13 (node 94602)

- **Type**: PASSIVE
- **Talent ID**: 122209, **Spell**: Expansiveness (ID: 429399)
- **Cast**: Passive
- Your maximum mana is increased by 5%.

### Protective Growth — Row 8, Col 14 (node 94593)

- **Type**: PASSIVE
- **Talent ID**: 122198, **Spell**: Protective Growth (ID: 433748)
- **Cast**: Passive
- Your Regrowth protects you, reducing damage you take by 8% while your Regrowth is on you.

### Sylvan Beckoning — Row 8, Col 15 (node 109714)

- **Type**: PASSIVE
- **Talent ID**: 140727, **Spell**: Sylvan Beckoning (ID: 1264614)
- **Cast**: Passive
- Entering an Eclipse summons a Dryad to assist you for 8 sec, casting Starsurge dealing 565 astral damage and Starfall at 200% effectiveness.

### Renewing Surge — Row 8, Col 16 (node 82060)

- **Type**: PASSIVE
- **Talent ID**: 108119, **Spell**: Renewing Surge (ID: 470562)
- **Cast**: Passive
- Swiftmend cooldown is reduced by 15%, increasing up to 30% on lower health targets.

### Rampant Growth — Row 8, Col 17 (node 82058)

- **Type**: PASSIVE
- **Talent ID**: 108117, **Spell**: Rampant Growth (ID: 404521)
- **Cast**: Passive
- Regrowth's healing over time is increased by 100%, and it also applies to the target of your Lifebloom.

### Regenesis (2 ranks) — Row 8, Col 18 (node 82062)

- **Type**: PASSIVE
- **Talent ID**: 108122, **Spell**: Regenesis (ID: 383191)
- **Cast**: Passive
- Rejuvenation healing is increased by up to 30%, and Tranquility healing is increased by up to 30%, healing for more on low-health targets.

### Wild Synthesis — Row 8, Col 19 (node 94535)

- **Type**: PASSIVE
- **Talent ID**: 122117, **Spell**: Wild Synthesis (ID: 400533)
- **Cast**: Passive
- Grove Guardians, Efflorescence, and your other summons heal for 30% more.

### Power of the Archdruid — Row 8, Col 20 (node 82065)

- **Type**: PASSIVE
- **Talent ID**: 108126, **Spell**: Power of the Archdruid (ID: 392302)
- **Cast**: Passive
- Soul of the Forest now causes your next Rejuvenation or Regrowth to apply to 2 additional allies within 20 yards of the target.

### Unstoppable Growth (2 ranks) — Row 8, Col 22 (node 82080)

- **Type**: PASSIVE
- **Talent ID**: 108143, **Spell**: Unstoppable Growth (ID: 382559)
- **Cast**: Passive
- Wild Growth's healing falls off 30% less over time.

### Improved Swiftmend — Row 8, Col 23 (node 103873)

- **Type**: PASSIVE
- **Talent ID**: 133081, **Spell**: Improved Swiftmend (ID: 470549)
- **Cast**: Passive
- Swiftmend healing increased by 30%.

### Regenerative Heartwood — Row 8, Col 24 (node 82075)

- **Type**: PASSIVE
- **Talent ID**: 108136, **Spell**: Regenerative Heartwood (ID: 392116)
- **Cast**: Passive
- Allies protected by your Ironbark also receive 75% of the healing from each of your active Rejuvenations and Ironbark's duration is increased by 4 sec.

### [Choice] Row 9, Col 12 (node 94605)

**Option: Power of Nature** (talent 122213, spell 428859: Power of Nature)
- Cast: Passive
- Your Grove Guardians increase the healing of your Rejuvenation, Efflorescence, and Lifebloom by 10% while active.

**Option: Durability of Nature** (talent 122212, spell 429227: Durability of Nature)
- Cast: Passive
- Your Force of Nature treants have 100% increased health.

### Cenarius' Might — Row 9, Col 13 (node 94604)

- **Type**: PASSIVE
- **Talent ID**: 122211, **Spell**: Cenarius' Might (ID: 455797)
- **Cast**: Passive
- Swiftmend healing is increased by 20%.

### [Choice] Row 9, Col 14 (node 94595)

**Option: Grove's Inspiration** (talent 122201, spell 429402: Grove's Inspiration)
- Cast: Passive
- Wrath and Starfire damage increased by 10%. 

Regrowth, Wild Growth, and Swiftmend healing increased by 9%.

**Option: Potent Enchantments** (talent 122200, spell 429420: Potent Enchantments)
- Cast: Passive
- Orbital Strike damage increased by 30%, and damage of Stellar Flares it applies increased by 30%.

Whirling Stars increases the haste you gain during Celestial Alignment by an additional 10%.

### Dryad's Dance — Row 9, Col 15 (node 109713)

- **Type**: PASSIVE
- **Talent ID**: 140726, **Spell**: Dryad's Dance (ID: 1264776)
- **Cast**: Passive
- Dryads cause most of your Astral power generation to be increased by 10%.

### Ysera's Gift — Row 9, Col 17 (node 82048)

- **Type**: PASSIVE
- **Talent ID**: 108104, **Spell**: Ysera's Gift (ID: 145108)
- **Cast**: Passive
- Heals you for 3% of your maximum health every 5 sec. If you are at full health, an injured party or raid member will be healed instead.

### [Choice] Row 9, Col 19 (node 82064)

**Option: Incarnation: Tree of Life** (talent 108125, spell 33891: Incarnation: Tree of Life)
- Cast: Instant
- Cooldown: 3 min cooldown
- Shapeshift into the Tree of Life, increasing healing done by 10%, increasing armor by 120%, and granting protection from Polymorph effects. Functionality of Rejuvenation, Wild Growth, Regrowth, Entangling Roots, and Wrath is enhanced.

Lasts 30 sec. You may shapeshift in and out of this form for its duration.

**Option: Convoke the Spirits** (talent 108124, spell 391528: Convoke the Spirits)
- Cast: Channeled
- Cooldown: 2 min cooldown
- Call upon the spirits for an eruption of energy, channeling a rapid flurry of 16 Druid spells and abilities over 4 sec.

You will cast Wild Growth, Swiftmend, Moonfire, Wrath, Regrowth, Rejuvenation, Rake, and Thrash on appropriate nearby targets, favoring your current shapeshift form.

### Call of the Elder Druid — Row 9, Col 21 (node 82067)

- **Type**: PASSIVE
- **Talent ID**: 108128, **Spell**: Call of the Elder Druid (ID: 426784)
- **Cast**: Passive
- When you cast Starsurge, Rake, Shred, or Frenzied Regeneration you gain Call of the Elder Druid for 15 sec, once every 1 min.

 Call of the Elder Druid
Abilities not associated with your specialization are substantially empowered for 45 sec.

Balance: Cast time of Balance spells reduced by 30% and damage increased by 20%.

Feral: Gain 1 Combo Point every 2 sec while in Cat Form and Physical damage increased by 20%.

Guardian: Bear Form gives an additional 20% Stamina, multiple uses of Ironfur may overlap, and Frenzied Regeneration has 2 charges.

Restoration: Healing increased by 30%, and mana costs reduced by 50%.

### Intensity — Row 9, Col 23 (node 82052)

- **Type**: PASSIVE
- **Talent ID**: 108110, **Spell**: Intensity (ID: 1264649)
- **Cast**: Passive
- When Regrowth critically heals, it is 260% effective instead of the usual 200%.

### [Choice] Row 10, Col 12 (node 94591)

**Option: Bounteous Bloom** (talent 122196, spell 429215: Bounteous Bloom)
- Cast: Passive
- Your Grove Guardians' healing is increased by 30%.

**Option: Early Spring** (talent 122907, spell 428937: Early Spring)
- Cast: Passive
- Swiftmend and Wild Growth cooldowns reduced by 1 sec.

### [Choice] Row 10, Col 13 (node 94592)

**Option: Power of the Dream** (talent 122197, spell 434220: Power of the Dream)
- Cast: Passive
- Dream Surge heals 1 additional ally.

**Option: Control of the Dream** (talent 122906, spell 434249: Control of the Dream)
- Cast: Passive
- Time elapsed while your major abilities are available to be used or at maximum charges is subtracted from that ability's cooldown after the next time you use it, up to 15 seconds.

Affects Force of Nature, Celestial Alignment, and Convoke the Spirits.

### Blooming Infusion — Row 10, Col 14 (node 94601)

- **Type**: PASSIVE
- **Talent ID**: 122208, **Spell**: Blooming Infusion (ID: 429433)
- **Cast**: Passive
- Every 5 Regrowths you cast makes your next Wrath, Starfire, or Entangling Roots instant and increases damage it deals by 100%.

Every 5 Starsurges you cast makes your next Regrowth or Entangling roots instant.

### Spirit of the Thicket — Row 10, Col 15 (node 109712)

- **Type**: PASSIVE
- **Talent ID**: 140725, **Spell**: Spirit of the Thicket (ID: 1264899)
- **Cast**: Passive
- Your Starfall damage is increased by 12% and your Starsurge damage is increased by 8%.

### [Choice] Row 10, Col 16 (node 82074)

**Option: Liveliness** (talent 108135, spell 426702: Liveliness)
- Cast: Passive
- Your damage over time effects deal their damage 25% faster, and your healing over time effects heal 5% faster.

**Option: Master Shapeshifter** (talent 119816, spell 289237: Master Shapeshifter)
- Cast: Passive
- Your abilities are amplified based on your current shapeshift form, granting an additional effect.

Wrath, Starfire, and Starsurge deal 30% additional damage and generate 324 Mana.

Bear Form
Ironfur grants 30% additional armor and generates 375 Mana.

 Cat Form
Rip, Ferocious Bite, and Maim deal 60% additional damage and generate 1,500 Mana when cast with 5 combo points.

### Waking Dream — Row 10, Col 17 (node 82046)

- **Type**: PASSIVE
- **Talent ID**: 108102, **Spell**: Waking Dream (ID: 392221)
- **Cast**: Passive
- Ysera's Gift now heals every 4 sec and its healing is increased by 8% for each of your active Rejuvenations.

### Embrace of the Dream — Row 10, Col 18 (node 82070)

- **Type**: PASSIVE
- **Talent ID**: 108131, **Spell**: Embrace of the Dream (ID: 392124)
- **Cast**: Passive
- Wild Growth momentarily shifts your mind into the Emerald Dream, instantly healing all allies affected by your Rejuvenation or Regrowth for 820.

### Cenarius' Guidance — Row 10, Col 19 (node 82063)

- **Type**: PASSIVE
- **Talent ID**: 108123, **Spell**: Cenarius' Guidance (ID: 393371)
- **Cast**: Passive
-  Incarnation: Tree of Life
During Incarnation: Tree of Life, you summon a Grove Guardian every 10 sec. The cooldown of Incarnation: Tree of Life is reduced by 5.0 sec when Grove Guardians fade.

 Convoke the Spirits
Convoke the Spirits' cooldown is reduced by 50% and its duration and number of spells cast is reduced by 25%. Convoke the Spirits has an increased chance to use an exceptional spell or ability.

### Nature's Bounty — Row 10, Col 20 (node 82072)

- **Type**: PASSIVE
- **Talent ID**: 108133, **Spell**: Nature's Bounty (ID: 1263879)
- **Cast**: Passive
- Regrowth heals all other allies with Regrowth for 20% of its healing.

### Dream of Cenarius — Row 10, Col 21 (node 82066)

- **Type**: PASSIVE
- **Talent ID**: 108127, **Spell**: Dream of Cenarius (ID: 158504)
- **Cast**: Passive
- Wrath and Shred transfer 100% of their damage and Starfire and Swipe transfer 50% of their damage into healing onto a nearby ally. 

This effect is increased by 200% while Call of the Elder Druid is active.

### Thriving Vegetation (2 ranks) — Row 10, Col 22 (node 82068)

- **Type**: PASSIVE
- **Talent ID**: 108129, **Spell**: Thriving Vegetation (ID: 447131)
- **Cast**: Passive
- Rejuvenation instantly heals your target for 15% of its total periodic effect and Regrowth's duration is increased by 3 sec.

### Abundance — Row 10, Col 23 (node 103876)

- **Type**: PASSIVE
- **Talent ID**: 133084, **Spell**: Abundance (ID: 207383)
- **Cast**: Passive
- For each Rejuvenation you have active, Regrowth's cost is reduced by 8% and critical effect chance is increased by 8%, up to a maximum of 96%.

### Nurturing Dormancy — Row 10, Col 24 (node 82076)

- **Type**: PASSIVE
- **Talent ID**: 108137, **Spell**: Nurturing Dormancy (ID: 392099)
- **Cast**: Passive
- When your Rejuvenation heals a full health target, its duration is increased by 2 sec, up to a maximum total increase of 6 sec per cast.

### Harmony of the Grove — Row 11, Col 13 (node 94606)

- **Type**: PASSIVE
- **Talent ID**: 122215, **Spell**: Harmony of the Grove (ID: 428731)
- **Cast**: Passive
- Each of your Grove Guardians increases your healing done by 5% while active.

### Photosynthesis — Row 11, Col 17 (node 82073)

- **Type**: PASSIVE
- **Talent ID**: 108134, **Spell**: Photosynthesis (ID: 274902)
- **Cast**: Passive
- Your periodic heals on targets with Lifebloom have a 8% chance to cause it to bloom.

### Harmonious Blooming — Row 11, Col 19 (node 82077)

- **Type**: PASSIVE
- **Talent ID**: 108139, **Spell**: Harmonious Blooming (ID: 392256)
- **Cast**: Passive
- Lifebloom counts for 3 stacks of Mastery: Harmony.

### Reforestation — Row 11, Col 21 (node 82069)

- **Type**: PASSIVE
- **Talent ID**: 108130, **Spell**: Reforestation (ID: 392356)
- **Cast**: Passive
- Every 4 casts of Swiftmend grants you Incarnation: Tree of Life for 10 sec.

### Germination — Row 11, Col 23 (node 82071)

- **Type**: PASSIVE
- **Talent ID**: 108132, **Spell**: Germination (ID: 155675)
- **Cast**: Passive
- You can apply Rejuvenation twice to the same target. Rejuvenation's duration is increased by 2 sec.

### Everbloom — Row 12, Col 20 (node 110424)

- **Type**: ACTIVE
- **Talent ID**: 141803, **Spell**: Everbloom (ID: 392167)
- **Cast**: Passive
- Lifebloom stacks every 5 sec, stacking up to 3 times.

## Hero Talent Trees

### Druid of the Claw (tree ID: 21) — NOT relevant for Resto

#### Ravage (node 94609, talent 122218, spell 441583)
- Your auto-attacks have a chance to make your next Ferocious Bite become Ravage.

Ravage
Finishing move that slashes through your target in a wide arc, dealing Physical damage per combo point to your target and consuming up to 25 additional Energy to increase that damage by up to 100%. Hits all other enemies in front of you for reduced damage per combo point spent. 

1 point: 13 damage, 6 in an arc
2 points: 25 damage, 12 in an arc
3 points: 39 damage, 17 in an arc
4 points: 52 damage, 23 in an arc
5 points: 66 damage, 30 in an arc

#### Wildshape Mastery (node 94610, talent 122219, spell 441678)
- Ironfur and Frenzied Regeneration persist in Cat Form.

When transforming from Bear to Cat Form, you retain 80% of your Bear Form armor and health for 6 sec.

For 6 sec after entering Bear Form, you heal for 10% of damage taken over 8 sec. 

#### Bestial Strength (node 94611, talent 122220, spell 441841)
- Maul and Raze damage increased by 10%.

#### [Choice] (node 94612)

**Option: Empowered Shapeshifting** (talent 122222, spell 441689)
- Frenzied Regeneration can be cast in Cat Form for 40 Energy.

Bear Form reduces magic damage you take by 6%.

Shred and Swipe damage increased by 10%. Mangle damage increased by 25%.

**Option: Wildpower Surge** (talent 122221, spell 441691)
- Mangle grants Feline Potential. When you have 6 stacks, the next time you transform into Cat Form, gain 5 combo points and your next Ferocious Bite or Rip deals 50% increased damage for its full duration.

#### Claw Rampage (node 94613, talent 122223, spell 441835)
- During Berserk, Shred, and Swipe have a 20% chance to make your next Ferocious Bite become Ravage.

#### [Choice] (node 94614)

**Option: Strike for the Heart** (talent 122226, spell 441845)
- Mangle damage increased by 10% and its critical strike chance is increased by 10%.



**Option: Tear Down the Mighty** (talent 122225, spell 441846)
- The cooldown of Sundering Roar is reduced by 15 sec.

#### Pack's Endurance (node 94615, talent 122227, spell 441844)
- Stampeding Roar's duration is increased by 25%.

#### Aggravate Wounds (node 94616, talent 122228, spell 441829)
- Every attack with an Energy cost that you cast extends the duration of your Dreadful Wounds by 0.6 sec, up to 8 additional sec.

#### Fount of Strength (node 94618, talent 122230, spell 441675)
- Your maximum Energy and Rage are increased by 20.

Frenzied Regeneration also increases your maximum health by 10%.

#### Exacerbating Wounds (node 94619, talent 122231, spell 1271839)
- Your Dreadful Wounds increase the damage afflicted enemies take from your Bleed damage over time effects by 15%.

#### Dreadful Wound (node 94620, talent 122232, spell 441809)
- Ravage also inflicts a Bleed that causes 35 damage over 6 sec and saps its victims' strength, reducing damage they deal to you by 15%.

Dreadful Wound is not affected by Circle of Life and Death. 

#### Twin Claw (node 109721, talent 140734, spell 1271635)
- You have a 18% chance to follow up any single target melee ability with a Twin Claw, dealing 118 Physical damage and generating 5 Rage.



#### Limb from Limb (node 109722, talent 140735, spell 1271540)
- Your auto-attacks are 30% more likely to make your next Maul become Ravage.

#### [Choice] (node 109723)

**Option: Ruthless Aggression** (talent 140736, spell 441814)
- Ravage increases your auto-attack speed by 35% for 6 sec.

**Option: Killing Strikes** (talent 141396, spell 441824)
- Ravage increases your Agility by 8% and the armor granted by Ironfur by 20% for 8 sec.

Your first Mangle after entering combat makes your next Maul become Ravage.

---

### Wildstalker (tree ID: 22)

#### Wildstalker's Power (node 94621, talent 122233, spell 439926)
- Rip and Ferocious Bite damage increased by 5%.

Rejuvenation healing increased by 10%.

#### [Choice] (node 94622)

**Option: Entangling Vortex** (talent 122234, spell 439895)
- Enemies pulled into Ursol's Vortex are rooted in place for 3 sec. Damage may cancel the effect.

**Option: Flower Walk** (talent 124755, spell 439901)
- During Barkskin your movement speed is increased by 10% and every second flowers grow beneath your feet that heal up to 3 nearby injured allies for 76.

#### Strategic Infusion (node 94623, talent 122235, spell 439890)
- Attacking from Prowl increases the chance for Shred, Rake, and Swipe to critically strike by 8% for 6 sec.

Your periodic heals have a 4% increased chance to critically heal.

#### Lethal Preservation (node 94624, talent 122236, spell 455461)
- When you remove an effect with Soothe or Remove Corruption, gain a combo point and heal for 4% of your maximum health. If you are at full health an injured party or raid member will be healed instead.

#### [Choice] (node 94625)

**Option: Bond with Nature** (talent 122237, spell 439929)
- Healing you receive is increased by 4%.

**Option: Harmonious Constitution** (talent 124754, spell 440116)
- Your Regrowth's healing to yourself is increased by 35%.

#### Thriving Growth (node 94626, talent 122238, spell 439528)
- Rip and Rake damage has a chance to cause Bloodseeker Vines to grow on the victim, dealing 34 Bleed damage over 6 sec.

Wild Growth, Regrowth, and Efflorescence healing has a chance to cause Symbiotic Blooms to grow on the target, healing for 834 over 6 sec.

Multiple instances of these can overlap.

#### Vigorous Creepers (node 94627, talent 122239, spell 440119)
- Bloodseeker Vines increase the damage your abilities deal to affected enemies by 4%.

Symbiotic Blooms increase the healing your spells do to affected targets by 20%.

#### [Choice] (node 94628)

**Option: Twin Sprouts** (talent 122242, spell 440117)
- When Bloodseeker Vines or Symbiotic Blooms grow, they have a 30% chance to cause another growth of the same type to immediately grow on a valid nearby target.

**Option: Implant** (talent 122241, spell 440118)
- Casting Swiftmend or Wild Growth causes a Symbiotic Bloom to grow on a target for 6 sec.

#### Hunt Beneath the Open Skies (node 94629, talent 122243, spell 439868)
- Damage and healing while in Cat Form increased by 3%.

Moonfire and Sunfire damage increased by 10%.

#### Patient Custodian (node 94630, talent 122244, spell 1270592)
- Your heal over time effects are 6% more effective.

#### [Choice] (node 94631)

**Option: Resilient Flourishing** (talent 122246, spell 439880)
- Bloodseeker Vines and Symbiotic Blooms last 2 additional sec.

When a target afflicted by Bloodseeker Vines dies, the vines jump to a valid nearby target for their remaining duration.

**Option: Root Network** (talent 122245, spell 439882)
- Each active Bloodseeker Vine increases the damage your abilities deal by 2%.

Each active Symbiotic Bloom increases the healing of your spells by 2%.

#### Rampancy (node 109715, talent 140728, spell 1270586)
- Symbiotic Blooms have a 20% chance to trigger Bursting Growth every 2 sec at 100% effectiveness.


#### Bursting Growth (node 109716, talent 140729, spell 440120)
- When Bloodseeker Vines expire or you use Ferocious Bite on their target they explode in thorns, dealing 30 physical damage to nearby enemies. Damage reduced above 5 targets.

When Symbiotic Blooms expire or you cast Rejuvenation on their target flowers grow around their target, healing them and up to 3 nearby allies for 161.

#### Green Thumb (node 109717, talent 140730, spell 1270565)
- The rate at which Symbiotic Blooms grow is increased by 20%.

---

### Keeper of the Grove (tree ID: 23)

#### [Choice] (node 94591)

**Option: Bounteous Bloom** (talent 122196, spell 429215)
- Your Grove Guardians' healing is increased by 30%.

**Option: Early Spring** (talent 122907, spell 428937)
- Swiftmend and Wild Growth cooldowns reduced by 1 sec.

#### [Choice] (node 94592)

**Option: Power of the Dream** (talent 122197, spell 434220)
- Dream Surge heals 1 additional ally.

**Option: Control of the Dream** (talent 122906, spell 434249)
- Time elapsed while your major abilities are available to be used or at maximum charges is subtracted from that ability's cooldown after the next time you use it, up to 15 seconds.

Affects Force of Nature, Celestial Alignment, and Convoke the Spirits.

#### Protective Growth (node 94593, talent 122198, spell 433748)
- Your Regrowth protects you, reducing damage you take by 8% while your Regrowth is on you.

#### [Choice] (node 94595)

**Option: Grove's Inspiration** (talent 122201, spell 429402)
- Wrath and Starfire damage increased by 10%. 

Regrowth, Wild Growth, and Swiftmend healing increased by 9%.

**Option: Potent Enchantments** (talent 122200, spell 429420)
- Orbital Strike damage increased by 30%, and damage of Stellar Flares it applies increased by 30%.

Whirling Stars increases the haste you gain during Celestial Alignment by an additional 10%.

#### Treants of the Moon (node 94599, talent 122206, spell 428544)
- Your Grove Guardians cast Moonfire on nearby targets about once every 6 sec.

#### Dream Surge (node 94600, talent 122207, spell 433831)
- When Grove Guardians are summoned, they grow Dream Petals on your target, healing up to 3 nearby allies for 565.

#### Blooming Infusion (node 94601, talent 122208, spell 429433)
- Every 5 Regrowths you cast makes your next Wrath, Starfire, or Entangling Roots instant and increases damage it deals by 100%.

Every 5 Starsurges you cast makes your next Regrowth or Entangling roots instant.

#### Expansiveness (node 94602, talent 122209, spell 429399)
- Your maximum mana is increased by 5%.

#### Cenarius' Might (node 94604, talent 122211, spell 455797)
- Swiftmend healing is increased by 20%.

#### [Choice] (node 94605)

**Option: Power of Nature** (talent 122213, spell 428859)
- Your Grove Guardians increase the healing of your Rejuvenation, Efflorescence, and Lifebloom by 10% while active.

**Option: Durability of Nature** (talent 122212, spell 429227)
- Your Force of Nature treants have 100% increased health.

#### Harmony of the Grove (node 94606, talent 122215, spell 428731)
- Each of your Grove Guardians increases your healing done by 5% while active.

#### Spirit of the Thicket (node 109712, talent 140725, spell 1264899)
- Your Starfall damage is increased by 12% and your Starsurge damage is increased by 8%.

#### Dryad's Dance (node 109713, talent 140726, spell 1264776)
- Dryads cause most of your Astral power generation to be increased by 10%.

#### Sylvan Beckoning (node 109714, talent 140727, spell 1264614)
- Entering an Eclipse summons a Dryad to assist you for 8 sec, casting Starsurge dealing 560 astral damage and Starfall at 200% effectiveness.

---

### Elune's Chosen (tree ID: 24) — NOT relevant for Resto

#### [Choice] (node 94585)

**Option: The Light of Elune** (talent 122188, spell 428655)
- Moonfire damage has a chance to call down a Fury of Elune to follow your target for 3 sec.

 Fury of Elune
Calls down a beam of pure celestial energy, dealing 425 Astral damage over 3 sec within its area.

Generates 15 Astral Power over its duration.

**Option: Astral Insight** (talent 122784, spell 429536)
- Incarnation: Guardian of Ursoc increases Arcane damage from spells and abilities by 10% while active.

Increases the duration and number of spells cast by Convoke the Spirits by 25%.

#### [Choice] (node 94586)

**Option: Arcane Affinity** (talent 122190, spell 429540)
- All Arcane damage from your spells and abilities is increased by 3%.

**Option: Lunation** (talent 122189, spell 429539)
- Your Arcane abilities reduce the cooldown of Lunar Beam by 3.0 sec.


#### The Eternal Moon (node 94587, talent 122191, spell 424113)
- Further increases the power of Boundless Moonlight.

 Fury of Elune
The flash of energy now generates 6 Astral Power and its damage is increased by 50%.

 Full Moon
New Moon and Half Moon now also call down 1 Minor Moon.

#### Lunar Insight (node 94588, talent 122193, spell 429530)
- Moonfire deals 20% additional damage.

#### Lunar Calling (node 94590, talent 122195, spell 429523)
- Thrash now deals Arcane damage and its damage is increased by 12%.

#### Glistening Fur (node 94594, talent 122781, spell 429533)
- Bear Form and Moonkin Form reduce Arcane damage taken by 6% and all other magic damage taken by 3%.

#### Stellar Command (node 94596, talent 122202, spell 429668)
- Increases the damage of Lunar Beam by 30% and Fury of Elune by 15%.

#### [Choice] (node 94597)

**Option: Moondust** (talent 122204, spell 429538)
- Enemies affected by Moonfire are slowed by 20%.

**Option: Elune's Grace** (talent 128177, spell 443046)
- Using Wild Charge while in Bear Form or Moonkin Form incurs a 3 sec shorter cooldown.

#### Moon Guardian (node 94598, talent 122205, spell 429520)
- Free automatic Moonfires from Galactic Guardian generate 5 Rage.

#### Atmospheric Exposure (node 94607, talent 122216, spell 429532)
- Enemies damaged by Lunar Beam or Fury of Elune take 6% increased damage from you for 6 sec.

#### Boundless Moonlight (node 94608, talent 122217, spell 424058)
-  Fury of Elune
Fury of Elune now ends with a flash of energy, blasting nearby enemies for 822 Astral damage.

 Full Moon
Full Moon calls down 2 Minor Moons that deal 715 Astral damage and generate 3 Astral Power.

#### Bask in Moonlight (node 109718, talent 140731, spell 1271305)
- Starsurge damage increased by 10%. 
Starfall damage increased by 10%. 


#### Penumbral Swell (node 109719, talent 140732, spell 1271261)
- Lunar Eclipse increases Arcane damage by an additional 3%.

#### Star Cascade (node 109720, talent 140733, spell 1271206)
- Gaining Astral Power with Wrath or Starfire has a 40% chance to launch a Starsurge at a victim at 70% effectiveness.


---
