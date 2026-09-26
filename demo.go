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
		fmt.Println()
	}
}
