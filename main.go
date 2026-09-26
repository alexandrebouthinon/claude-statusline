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
		CurrentUsage *struct {
			Input       int `json:"input_tokens"`
			CacheCreate int `json:"cache_creation_input_tokens"`
			CacheRead   int `json:"cache_read_input_tokens"`
		} `json:"current_usage"`
	} `json:"context_window"`
	RateLimits struct {
		FiveHour *rateLimit `json:"five_hour"`
		SevenDay *rateLimit `json:"seven_day"`
	} `json:"rate_limits"`
}

type rateLimit struct {
	UsedPct  float64 `json:"used_percentage"`
	ResetsAt int64   `json:"resets_at"`
}

const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	dim    = "\033[2m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	cyan   = "\033[36m"
)

func main() {
	var p payload
	if err := json.NewDecoder(bufio.NewReader(os.Stdin)).Decode(&p); err != nil {
		fmt.Print("statusline: bad input")
		return
	}
	fmt.Print(render(p, time.Now()))
}

func render(p payload, now time.Time) string {
	dir := p.Workspace.CurrentDir
	if dir == "" {
		dir = p.Cwd
	}

	parts := []string{bold + filepath.Base(dir) + reset}
	if b := gitBranch(dir); b != "" {
		parts[0] += " " + dim + b + reset
	}
	parts = append(parts, cyan+p.Model.DisplayName+reset, contextSegment(p))
	if rl := p.RateLimits.FiveHour; rl != nil {
		parts = append(parts, "5h "+limitSegment(*rl, now))
	}
	if p.Cost.TotalCostUSD > 0 {
		parts = append(parts, fmt.Sprintf("%s~$%.2f%s", dim, p.Cost.TotalCostUSD, reset))
	}
	return strings.Join(parts, " │ ")
}

func contextSegment(p payload) string {
	cw := p.ContextWindow
	if cw.CurrentUsage == nil {
		return "ctx n/a"
	}
	used := cw.CurrentUsage.Input + cw.CurrentUsage.CacheCreate + cw.CurrentUsage.CacheRead
	color := green
	switch {
	case used >= compactThreshold*2:
		color = red
	case used >= compactThreshold:
		color = yellow
	}
	s := fmt.Sprintf("ctx %s%s%s/%s", color, humanTokens(used), reset, humanTokens(cw.Size))
	if used >= compactThreshold {
		s += " " + color + "⚠ /compact" + reset
	}
	return s
}

func limitSegment(rl rateLimit, now time.Time) string {
	color := pctColor(rl.UsedPct)
	s := fmt.Sprintf("%s%.0f%%%s", color, rl.UsedPct, reset)
	if rl.ResetsAt > 0 {
		if left := time.Unix(rl.ResetsAt, 0).Sub(now); left > 0 {
			s += " " + dim + shortDuration(left) + reset
		}
	}
	return s
}

func pctColor(pct float64) string {
	switch {
	case pct >= 80:
		return red
	case pct >= 50:
		return yellow
	}
	return green
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
