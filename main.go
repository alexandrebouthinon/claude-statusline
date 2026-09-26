// Command claude-statusline renders the Claude Code status line from the
// JSON payload Claude Code pipes on stdin.
package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// compactThreshold matches ~/.claude/hooks/compact-reminder.py.
const compactThreshold = 100_000

type payload struct {
	Model struct {
		DisplayName string `json:"display_name"`
	} `json:"model"`
	Workspace struct {
		CurrentDir string `json:"current_dir"`
	} `json:"workspace"`
	Cwd  string `json:"cwd"`
	Cost struct {
		TotalCostUSD float64 `json:"total_cost_usd"`
	} `json:"cost"`
	ContextWindow struct {
		Size         int      `json:"context_window_size"`
		UsedPct      *float64 `json:"used_percentage"`
		CurrentUsage *usage   `json:"current_usage"`
	} `json:"context_window"`
	RateLimits struct {
		FiveHour *rateLimit `json:"five_hour"`
		SevenDay *rateLimit `json:"seven_day"`
	} `json:"rate_limits"`
}

type usage struct {
	Input       int `json:"input_tokens"`
	CacheCreate int `json:"cache_creation_input_tokens"`
	CacheRead   int `json:"cache_read_input_tokens"`
}

type rateLimit struct {
	UsedPct  float64 `json:"used_percentage"`
	ResetsAt int64   `json:"resets_at"`
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--glyphs" {
		runGlyphDemo()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--caps" {
		runCapsDemo()
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--demo" {
		runDemo()
		return
	}
	var p payload
	if err := json.NewDecoder(bufio.NewReader(os.Stdin)).Decode(&p); err != nil {
		fmt.Print("statusline: bad input")
		return
	}
	fmt.Print(render(p, time.Now()))
}

func render(p payload, now time.Time) string {
	return pills(segments(p, now), " ")
}

func segments(p payload, now time.Time) []segment {
	dir := p.Workspace.CurrentDir
	if dir == "" {
		dir = p.Cwd
	}

	loc := iconFolder + " " + filepath.Base(dir)
	if b := gitBranch(dir); b != "" {
		loc += " " + iconBranch + " " + b
	}
	segs := []segment{
		{loc, theme.light, theme.dir},
		{iconModel + " " + shortModel(p.Model.DisplayName), theme.light, theme.model},
		contextSegment(p),
	}
	if rl := p.RateLimits.FiveHour; rl != nil {
		segs = append(segs, limitSegment(*rl, now))
	}
	if p.Cost.TotalCostUSD > 0 {
		segs = append(segs, segment{fmt.Sprintf("~%s%.2f", iconDollar, p.Cost.TotalCostUSD), theme.costFg, theme.costBg})
	}
	return segs
}

// shortModel drops the "(1M context)" suffix, already shown by the context segment.
func shortModel(name string) string {
	if i := strings.Index(name, " ("); i > 0 {
		return name[:i]
	}
	return name
}

func contextSegment(p payload) segment {
	cw := p.ContextWindow
	if cw.CurrentUsage == nil {
		return segment{iconContext + " n/a", theme.dark, theme.ok}
	}
	used := cw.CurrentUsage.Input + cw.CurrentUsage.CacheCreate + cw.CurrentUsage.CacheRead
	text := iconContext + " " + humanTokens(used) + "/" + humanTokens(cw.Size)
	bg := theme.ok
	switch {
	case used >= compactThreshold*2:
		bg = theme.crit
	case used >= compactThreshold:
		bg = theme.warn
	}
	if used >= compactThreshold {
		text += " " + iconWarn + " /compact"
	}
	return segment{text, theme.dark, bg}
}

func limitSegment(rl rateLimit, now time.Time) segment {
	text := fmt.Sprintf("%s %.0f%%", iconLimit, rl.UsedPct)
	if rl.ResetsAt > 0 {
		if left := time.Unix(rl.ResetsAt, 0).Sub(now); left > 0 {
			text += " · " + shortDuration(left)
		}
	}
	bg := theme.limitOK
	switch {
	case rl.UsedPct >= 80:
		bg = theme.crit
	case rl.UsedPct >= 50:
		bg = theme.warn
	}
	return segment{text, theme.dark, bg}
}

func humanTokens(n int) string {
	switch {
	case n >= 1_000_000:
		return strings.TrimSuffix(fmt.Sprintf("%.1f", float64(n)/1e6), ".0") + "M"
	case n >= 1_000:
		return fmt.Sprintf("%dk", n/1000)
	}
	return fmt.Sprint(n)
}

func shortDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%02d", h, m)
	}
	return fmt.Sprintf("%dm", m)
}

// gitBranch reads .git/HEAD directly to avoid forking git on every refresh.
func gitBranch(dir string) string {
	for d := dir; d != "" && d != "/"; d = filepath.Dir(d) {
		gitPath := filepath.Join(d, ".git")
		info, err := os.Stat(gitPath)
		if err != nil {
			continue
		}
		if !info.IsDir() {
			// Worktree or submodule: .git is a file pointing to the real gitdir.
			data, err := os.ReadFile(gitPath)
			if err != nil {
				return ""
			}
			gd := strings.TrimSpace(strings.TrimPrefix(string(data), "gitdir:"))
			if !filepath.IsAbs(gd) {
				gd = filepath.Join(d, gd)
			}
			gitPath = gd
		}
		head, err := os.ReadFile(filepath.Join(gitPath, "HEAD"))
		if err != nil {
			return ""
		}
		ref := strings.TrimSpace(string(head))
		if b, ok := strings.CutPrefix(ref, "ref: refs/heads/"); ok {
			return b
		}
		if len(ref) >= 7 {
			return ref[:7]
		}
		return ""
	}
	return ""
}
