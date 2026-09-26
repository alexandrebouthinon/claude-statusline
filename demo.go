package main

import (
	"fmt"
	"os"
	"time"
)

// runDemo prints the current theme in a few typical session states.
func runDemo() {
	now := time.Unix(1_790_000_000, 0)
	state := func(model string, ctx int, five, week float64, agents map[string]int, cost float64) string {
		var p payload
		p.Workspace.CurrentDir, _ = os.Getwd()
		p.Model.DisplayName = model
		p.Cost.TotalCostUSD = cost
		p.ContextWindow.Size = 1_000_000
		p.ContextWindow.CurrentUsage = &usage{CacheRead: ctx}
		p.RateLimits.FiveHour = &rateLimit{UsedPct: five, ResetsAt: now.Add(3*time.Hour + 54*time.Minute).Unix()}
		p.RateLimits.SevenDay = &rateLimit{UsedPct: week, ResetsAt: now.Add(76 * time.Hour).Unix()}
		segs := segments(p, now)
		// Agent pills normally come from hook markers; inject them before the cost pill.
		last := len(segs) - 1
		segs = append(segs[:last], append(agentSegments(agents), segs[last])...)
		return pills(segs, " ")
	}
	fmt.Println(state("Opus 5.5 (1M context)", 46_000, 25, 12, nil, 0.67))
	fmt.Println(state("Sonnet 5", 88_000, 62, 41, map[string]int{"haiku": 3, "sonnet": 1}, 2.14))
	fmt.Println(state("Opus 5.5 (1M context)", 124_000, 71, 58, map[string]int{"opus": 1, "fable": 1}, 4.80))
	fmt.Println(state("Fable 5.1", 230_000, 91, 83, nil, 9.35))
}

// runGlyphDemo prints candidate glyphs for each segment role, lettered for picking.
func runGlyphDemo() {
	roles := []struct {
		name, bg string
		glyphs   []string
	}{
		{"Dossier", theme.dir, []string{"", "", "", "\U000f024b"}},
		{"Branche", theme.dir, []string{"", "", "", "\U000f062c"}},
		{"Modèle", theme.model, []string{"\U000f06a9", "\U000f09d1", "\U000f0674", ""}},
		{"Contexte", theme.ok, []string{"\U000f029a", "", "\U000f035b", ""}},
		{"Limite 5h", theme.limitOK, []string{"", "", "\U000f051f", ""}},
		{"Coût", theme.costBg, []string{"", "\U000f01c1", ""}},
		{"Alerte", theme.warn, []string{"", "\U000f0026", ""}},
	}
	for _, r := range roles {
		fmt.Printf("%-10s", r.name)
		for i, g := range r.glyphs {
			pill := powerline([]segment{{text: g, bg: r.bg}}, sepRound)
			fmt.Printf("  %c %s", 'A'+i, pill)
		}
		fmt.Println()
	}
}

// runCapsDemo prints separator and cap variants, lettered for picking.
func runCapsDemo() {
	now := time.Unix(1_790_000_000, 0)
	var p payload
	p.Workspace.CurrentDir, _ = os.Getwd()
	p.Model.DisplayName = "Opus 5.5 (1M context)"
	p.Cost.TotalCostUSD = 0.67
	p.ContextWindow.Size = 1_000_000
	p.ContextWindow.CurrentUsage = &usage{CacheRead: 46_000}
	p.RateLimits.FiveHour = &rateLimit{UsedPct: 25, ResetsAt: now.Add(3*time.Hour + 54*time.Minute).Unix()}
	s := segments(p, now)

	variants := []struct{ name, out string }{
		{"actuel, arrondi à droite", powerline(s, sepRound)},
		{"arrondi des deux côtés", capLeft(s, capRound) + powerline(s, sepRound)},
		{"pilules séparées", pills(s, " ")},
		{"pilules serrées", pills(s, "")},
		{"flèches p10k", capLeft(s, capSolid) + powerline(s, sepSolid)},
		{"biseaux", capLeft(s, capSlant) + powerline(s, sepSlant)},
	}
	for i, v := range variants {
		fmt.Printf("%c  %s\n   %s\n\n", 'A'+i, v.name, v.out)
	}
}
