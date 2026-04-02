# Flourish

Analyzes WarcraftLogs data for a Restoration Druid and attributes effective healing to specific talents. Answers the question: **"How much healing would I lose if I dropped this talent?"**

Built for WoW Midnight Season 1 (12.0.1).

## Setup

### Prerequisites

- Python 3.12+
- [uv](https://docs.astral.sh/uv/) package manager
- A [WarcraftLogs](https://www.warcraftlogs.com) API v2 client (create one at [WCL API Clients](https://www.warcraftlogs.com/api/clients))

### Install

```bash
git clone <repo-url>
cd flourish
uv sync --all-extras
```

### Configure

Create a `.env` file:

```
WCL_CLIENT_ID=your-client-id
WCL_CLIENT_SECRET=your-client-secret
```

## Usage

### Basic (interactive)

```bash
uv run flourish <report-code>
```

The report code is the alphanumeric string from a WarcraftLogs URL:
`https://www.warcraftlogs.com/reports/Aq7RXDt8FHNcQwKk` → `Aq7RXDt8FHNcQwKk`

You'll be prompted to select a fight and player.

### Skip prompts

```bash
uv run flourish Aq7RXDt8FHNcQwKk --fight 1 --player Saikó
```

### Custom config

```bash
uv run flourish Aq7RXDt8FHNcQwKk --config-path my-talents.yaml
```

## Output

```
Fight: Windrunner Spire  Player: Saikó
Total effective healing: 80.6M

┏━━━━━━━━━━━━━━━━━━━━━━━━━━━┳━━━━━━━━━━━━┳━━━━━━━━━┳━━━━━━┓
┃ Talent                    ┃ Attributed ┃ % Total ┃  HPS ┃
┡━━━━━━━━━━━━━━━━━━━━━━━━━━━╇━━━━━━━━━━━━╇━━━━━━━━━╇━━━━━━┩
│ Grove Guardians           │       8.1M │   10.0% │ 5.2k │
│ Convoke the Spirits       │       5.9M │    7.3% │ 3.8k │
│ Harmony of the Grove      │       4.8M │    5.9% │ 3.1k │
│ ...                       │            │         │      │
├───────────────────────────┼────────────┼─────────┼──────┤
│ Wasted (>50% OH)          │       1.7M │       — │    — │
│ Unattributed              │       6.7M │       — │    — │
└───────────────────────────┴────────────┴─────────┴──────┘
```

Each talent's value answers: "how much healing would I lose if I untalented this?"

Talents can buff the same heal (e.g., Grove's Inspiration + Wild Synthesis both buff Grove Guardian healing), so attributed totals may sum to more than 100%.

## Configuration

Edit `config/talents.yaml` to customize:

```yaml
# Skip a talent from analysis
germination:
  skip: true
  skip_reason: "always take"

# Override a multiplier (e.g., if patch changes SotF from 60% to 50%)
soul_of_the_forest:
  skip: false
  multiplier: 0.6

# Convoke healing ratio (default 0.7 = 70% of Convoke healing attributed)
convoke_the_spirits:
  skip: false
  multiplier: 0.7
```

## How it works

1. Fetches combat events from WCL v2 GraphQL API (including pet/treant events)
2. Parses events into typed dataclasses
3. Replays events through HoT and buff trackers
4. Each talent attributor claims its portion of healing:
   - **Direct spell** — talent grants a unique spell (e.g., Grove Guardians → Nourish)
   - **Buff multiplier** — talent buffs an existing spell by X% (e.g., SotF → +60% Rejuv)
   - **Stateful** — talent requires tracking game state (e.g., Convoke channel window)
5. Heals where overheal > 50% of raw healing are filtered as "wasted"

## Overheal filter

Any heal tick where more than 50% of the raw healing was overheal gets no talent attribution. This prevents inflating talent values with wasted healing.

## Implemented talents (28)

### Direct spell
Everbloom, Grove Guardians, Dream Surge, Efflorescence, Verdancy, Nature's Bounty, Regenerative Heartwood, Cultivation, Ysera's Gift, Embrace of the Dream, Rampant Growth, Improved Swiftmend, Flourish, Bursting Growth

### Buff multiplier
Wild Synthesis, Wildstalker's Power, Patient Custodian, Lifetreading, Grove's Inspiration, Cenarius' Might, Bounteous Bloom, Unstoppable Growth, Liveliness, Regenesis

### Stateful
Soul of the Forest, Incarnation: Tree of Life, Convoke the Spirits, Improved Wild Growth, Reforestation, Harmony of the Grove, Abundance, Intensity, Photosynthesis, Nurturing Dormancy, Vigorous Creepers, Implant, Root Network

## Development

```bash
# Run tests
uv run pytest

# Run with verbose output
uv run pytest -v
```

## Limitations

- Class tree talents are not analyzed
- Mastery (Harmony) attribution is not yet implemented
- Germination is skipped (always-take)
- Some approximations are used (Regenesis uses flat 15% instead of health-based scaling, Unstoppable Growth uses flat 27.7% multiplier)
