package output

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/rdruid-talent-analyzer/go-backend/internal/analysis"
	"github.com/rdruid-talent-analyzer/go-backend/internal/web"
)

const barMax = 20

var (
	green  = lipgloss.Color("#4ade80")
	dim    = lipgloss.Color("#555555")
	yellow = lipgloss.Color("#facc15")
	purple = lipgloss.Color("#c084fc")
	cyan   = lipgloss.Color("#22d3ee")
	orange = lipgloss.Color("#fb923c")
	white  = lipgloss.Color("#e4e4e7")

	headerStyle   = lipgloss.NewStyle().Bold(true).Foreground(white)
	nameStyle     = lipgloss.NewStyle().Foreground(white)
	numStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#a1a1aa"))
	summaryName   = lipgloss.NewStyle().Foreground(yellow).Italic(true)
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(purple).MarginBottom(1)
	blue = lipgloss.Color("#60a5fa")

	sectionColors = map[string]lipgloss.Color{
		"Keeper of the Grove": cyan,
		"Wildstalker":         orange,
		"Base Healing":        dim,
		"Stats":               blue,
	}
)

func formatNum(n float64) string {
	neg := ""
	if n < 0 {
		neg = "-"
		n = -n
	}
	s := fmt.Sprintf("%.0f", n)
	if len(s) <= 3 {
		return neg + s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return neg + strings.Join(parts, ",")
}

func bar(pct, maxPct float64) string {
	ratio := 0.0
	if maxPct > 0 {
		ratio = pct / maxPct
	}
	filled := int(math.Round(ratio * barMax))
	if filled > barMax {
		filled = barMax
	}
	if filled < 0 {
		filled = 0
	}
	filledStr := lipgloss.NewStyle().Foreground(green).Render(strings.Repeat("█", filled))
	emptyStr := lipgloss.NewStyle().Foreground(dim).Render(strings.Repeat("░", barMax-filled))
	return filledStr + emptyStr
}

type rowKind int

const (
	rowData rowKind = iota
	rowSection
	rowBlank
	rowSummary
)

type tableRow struct {
	kind    rowKind
	section string // for section headers
	cols    []string
}

func RenderResults(results *analysis.AnalysisResults, fightName, playerName string) string {
	var sb strings.Builder

	if fightName != "" || playerName != "" {
		sb.WriteString(titleStyle.Render(fmt.Sprintf("%s — %s", fightName, playerName)))
		sb.WriteString("\n")
	}

	durationSec := math.Max(float64(results.FightDurationMs)/1000.0, 1.0)
	total := float64(results.TotalHealing)

	type entry struct {
		name   string
		amount float64
		group  string // "", "Keeper of the Grove", "Wildstalker"
	}

	var specEntries, heroEntries []entry
	allAttributed := 0.0
	detectedHeroTree := ""
	for name, amount := range results.TalentHealing {
		if amount <= 0 {
			continue
		}
		allAttributed += amount
		tree := ""
		for treeName, talents := range web.HeroTrees {
			if talents[name] {
				tree = treeName
				detectedHeroTree = treeName
				break
			}
		}
		e := entry{name, amount, tree}
		if tree != "" {
			heroEntries = append(heroEntries, e)
		} else {
			specEntries = append(specEntries, e)
		}
	}
	sort.Slice(specEntries, func(i, j int) bool { return specEntries[i].amount > specEntries[j].amount })
	sort.Slice(heroEntries, func(i, j int) bool { return heroEntries[i].amount > heroEntries[j].amount })

	// Max pct across all entries for bar scaling
	maxPct := 0.0
	for _, e := range specEntries {
		if pct := e.amount / math.Max(total, 1) * 100; pct > maxPct {
			maxPct = pct
		}
	}
	for _, e := range heroEntries {
		if pct := e.amount / math.Max(total, 1) * 100; pct > maxPct {
			maxPct = pct
		}
	}

	makeRow := func(e entry) tableRow {
		pct := 0.0
		if total > 0 {
			pct = e.amount / total * 100
		}
		hps := e.amount / durationSec
		return tableRow{kind: rowData, cols: []string{
			e.name,
			formatNum(e.amount),
			fmt.Sprintf("%.1f%%", pct),
			formatNum(hps),
			bar(pct, maxPct),
		}}
	}

	var tableRows []tableRow

	// Spec talents
	tableRows = append(tableRows, tableRow{kind: rowSection, section: "Spec Talents"})
	for _, e := range specEntries {
		tableRows = append(tableRows, makeRow(e))
	}

	// Hero talents
	if len(heroEntries) > 0 {
		heroSum := 0.0
		for _, e := range heroEntries {
			heroSum += e.amount
		}
		heroPct := 0.0
		if total > 0 {
			heroPct = heroSum / total * 100
		}
		heroHps := heroSum / durationSec
		tableRows = append(tableRows, tableRow{kind: rowSection, section: detectedHeroTree, cols: []string{
			detectedHeroTree,
			formatNum(heroSum),
			fmt.Sprintf("%.1f%%", heroPct),
			formatNum(heroHps),
			"",
		}})
		for _, e := range heroEntries {
			tableRows = append(tableRows, makeRow(e))
		}
	}

	// Stats section
	if len(results.StatHealing) > 0 {
		type statEntry struct {
			name   string
			amount float64
		}
		// Fixed order for stats
		statOrder := []string{"Versatility", "Mastery: Harmony", "Critical Strike"}
		var statEntries []statEntry
		statTotal := 0.0
		for _, name := range statOrder {
			if amount, ok := results.StatHealing[name]; ok && amount > 0 {
				statEntries = append(statEntries, statEntry{name, amount})
				statTotal += amount
			}
		}
		if len(statEntries) > 0 {
			statPct := 0.0
			if total > 0 {
				statPct = statTotal / total * 100
			}
			statHps := statTotal / durationSec
			tableRows = append(tableRows, tableRow{kind: rowSection, section: "Stats", cols: []string{
				"Stats",
				formatNum(statTotal),
				fmt.Sprintf("%.1f%%", statPct),
				formatNum(statHps),
				"",
			}})
			for _, e := range statEntries {
				pct := 0.0
				if total > 0 {
					pct = e.amount / total * 100
				}
				hps := e.amount / durationSec
				tableRows = append(tableRows, tableRow{kind: rowData, cols: []string{
					e.name,
					formatNum(e.amount),
					fmt.Sprintf("%.1f%%", pct),
					formatNum(hps),
					bar(pct, maxPct),
				}})
			}
		}
	}

	// Base spell breakdown
	if len(results.BaseSpellHealing) > 0 {
		type spellEntry struct {
			name   string
			amount float64
		}
		var baseEntries []spellEntry
		for name, amount := range results.BaseSpellHealing {
			if amount > 0 {
				baseEntries = append(baseEntries, spellEntry{name, amount})
			}
		}
		sort.Slice(baseEntries, func(i, j int) bool { return baseEntries[i].amount > baseEntries[j].amount })

		baseTotal := 0.0
		for _, e := range baseEntries {
			baseTotal += e.amount
		}
		basePct := 0.0
		if total > 0 {
			basePct = baseTotal / total * 100
		}
		baseHps := baseTotal / durationSec
		tableRows = append(tableRows, tableRow{kind: rowSection, section: "Base Healing", cols: []string{
			"Base Healing",
			formatNum(baseTotal),
			fmt.Sprintf("%.1f%%", basePct),
			formatNum(baseHps),
			"",
		}})
		for _, e := range baseEntries {
			pct := 0.0
			if total > 0 {
				pct = e.amount / total * 100
			}
			hps := e.amount / durationSec
			tableRows = append(tableRows, tableRow{kind: rowData, cols: []string{
				e.name,
				formatNum(e.amount),
				fmt.Sprintf("%.1f%%", pct),
				formatNum(hps),
				bar(pct, maxPct),
			}})
		}
	}

	// Summary
	tableRows = append(tableRows, tableRow{kind: rowBlank})
	if results.Wasted > 0 {
		tableRows = append(tableRows, tableRow{kind: rowSummary, cols: []string{"Wasted (>50% OH)", formatNum(float64(results.Wasted)), "", "", ""}})
	}
	if results.HighHealthHealing > 0 {
		label := fmt.Sprintf("High Health (>%.0f%% HP)", results.HighHealthThreshold*100)
		tableRows = append(tableRows, tableRow{kind: rowSummary, cols: []string{label, formatNum(float64(results.HighHealthHealing)), "", "", ""}})
	}
	unattributed := total - allAttributed - float64(results.Wasted) - float64(results.HighHealthHealing)
	if unattributed < 0 {
		tableRows = append(tableRows, tableRow{kind: rowSummary, cols: []string{"Unattributed", "—", "", "", ""}})
	} else {
		tableRows = append(tableRows, tableRow{kind: rowSummary, cols: []string{"Unattributed", formatNum(unattributed), "", "", ""}})
	}

	// Convert to raw string rows for lipgloss table
	var rawRows [][]string
	for _, r := range tableRows {
		switch r.kind {
		case rowSection:
			if len(r.cols) > 0 {
				rawRows = append(rawRows, r.cols)
			} else {
				rawRows = append(rawRows, []string{r.section, "", "", "", ""})
			}
		case rowBlank:
			rawRows = append(rawRows, []string{"", "", "", "", ""})
		case rowData, rowSummary:
			rawRows = append(rawRows, r.cols)
		}
	}

	t := table.New().
		Headers("Talent", "Healing", "%", "HPS", "").
		Rows(rawRows...).
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(dim)).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle.Padding(0, 1)
			}
			if row < 0 || row >= len(tableRows) {
				return lipgloss.NewStyle().Padding(0, 1)
			}
			tr := tableRows[row]
			base := lipgloss.NewStyle().Padding(0, 1)

			switch tr.kind {
			case rowSection:
				color := lipgloss.Color("#818cf8") // indigo for spec
				for prefix, c := range sectionColors {
					if strings.HasPrefix(tr.section, prefix) {
						color = c
						break
					}
				}
				s := base.Bold(true).Foreground(color).PaddingTop(1)
				if col >= 1 && col <= 3 {
					return s.Align(lipgloss.Right)
				}
				return s
			case rowBlank:
				return base.Height(0)
			case rowSummary:
				if col == 0 {
					return summaryName.Padding(0, 1)
				}
				return numStyle.Padding(0, 1)
			default: // rowData
				switch col {
				case 0:
					return nameStyle.Padding(0, 1)
				case 4:
					return base
				default:
					return numStyle.Padding(0, 1).Align(lipgloss.Right)
				}
			}
		})

	sb.WriteString(t.Render())
	sb.WriteString("\n")

	if allAttributed > total {
		note := lipgloss.NewStyle().Foreground(dim).Italic(true)
		sb.WriteString(note.Render("  Talents can overlap — total may exceed 100%."))
		sb.WriteString("\n")
	}

	return sb.String()
}
