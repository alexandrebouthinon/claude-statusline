package main

import (
	"fmt"
	"os"
	"time"
)

// runDemo prints every palette in normal, warning and critical states.
func runDemo() {
	now := time.Unix(1_790_000_000, 0)
	state := func(ctx int, pct float64) payload {
		var p payload
		p.Workspace.CurrentDir, _ = os.Getwd()
		p.Model.DisplayName = "Opus 5.5 (1M context)"
		p.Cost.TotalCostUSD = 0.67
		p.ContextWindow.Size = 1_000_000
		p.ContextWindow.CurrentUsage = &usage{CacheRead: ctx}
		p.RateLimits.FiveHour = &rateLimit{UsedPct: pct, ResetsAt: now.Add(3*time.Hour + 54*time.Minute).Unix()}
		return p
	}
	states := []payload{state(46_000, 25), state(124_000, 62), state(230_000, 91)}

	for i, pal := range palettes {
		theme = pal
		fmt.Printf("%s%d · %s%s\n", bold, i+1, pal.name, reset)
		for _, p := range states {
			fmt.Printf("  %s\n", render(p, now))
		}
		week := limitSegment(rateLimit{UsedPct: 41, ResetsAt: now.Add(76 * time.Hour).Unix()}, now, iconWeek, theme.week)
		agents := agentSegments(map[string]int{"haiku": 3, "sonnet": 2, "opus": 1, "fable": 1})
		fmt.Printf("  %s\n", pills(append([]segment{week}, agents...), " "))
		fmt.Println()
	}
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
