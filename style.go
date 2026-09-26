package main

import (
	"fmt"
	"strings"
)

// Nerd Font glyphs, as used by Powerlevel9k/10k.
const (
	sepSolid     = ""
	sepThin      = ""
	sepRound     = ""
	sepRoundThin = ""
	sepSlant     = ""
	iconFolder   = "\uf115"
	iconBranch   = "\ue725"
	iconModel    = "\U000f0674"
	iconContext  = "\uf1c0"
	iconLimit    = "\uf252"
	iconDollar   = ""
	iconWarn     = ""
)

const (
	reset = "\033[0m"
	bold  = "\033[1m"
)

// palette holds hex truecolors for every segment role.
type palette struct {
	name                    string
	light, dark             string
	dir, model              string
	ok, limitOK, warn, crit string
	costBg, costFg          string
}

var palettes = []palette{
	{"p10k", "#eeeeee", "#1c1c1c", "#0087af", "#5f5faf", "#5faf00", "#87d787", "#d78700", "#d70000", "#444444", "#bcbcbc"},
	{"Catppuccin Mocha", "#cdd6f4", "#1e1e2e", "#89b4fa", "#cba6f7", "#a6e3a1", "#94e2d5", "#fab387", "#f38ba8", "#45475a", "#cdd6f4"},
	{"Tokyo Night", "#c0caf5", "#1a1b26", "#7aa2f7", "#bb9af7", "#9ece6a", "#7dcfff", "#e0af68", "#f7768e", "#414868", "#c0caf5"},
	{"Nord", "#eceff4", "#2e3440", "#5e81ac", "#b48ead", "#a3be8c", "#88c0d0", "#ebcb8b", "#bf616a", "#3b4252", "#d8dee9"},
	{"Gruvbox", "#fbf1c7", "#282828", "#458588", "#b16286", "#98971a", "#689d6a", "#d79921", "#cc241d", "#504945", "#ebdbb2"},
}

var theme = palettes[2] // Tokyo Night

type segment struct {
	text   string
	fg, bg string
}

func rgb(hex string) (r, g, b int) {
	fmt.Sscanf(strings.TrimPrefix(hex, "#"), "%02x%02x%02x", &r, &g, &b)
	return
}

func fgHex(hex string) string {
	r, g, b := rgb(hex)
	return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

func bgHex(hex string) string {
	r, g, b := rgb(hex)
	return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}

// powerline renders filled segments joined by a separator glyph whose
// foreground is the previous background, which creates the arrow effect.
func powerline(segs []segment, sep string) string {
	var b strings.Builder
	for i, s := range segs {
		b.WriteString(bgHex(s.bg) + fgHex(textOn(s.bg)) + " " + s.text + " ")
		switch {
		case i+1 < len(segs) && segs[i+1].bg == s.bg:
			// Same background on both sides would hide the arrow: fall back to a thin divider.
			b.WriteString(fgHex(textOn(s.bg)) + thinOf(sep))
		case i+1 < len(segs):
			b.WriteString(bgHex(segs[i+1].bg) + fgHex(s.bg) + sep)
		default:
			b.WriteString(reset + fgHex(s.bg) + sep + reset)
		}
	}
	return b.String()
}

func thinOf(sep string) string {
	switch sep {
	case sepSolid:
		return sepThin
	case sepRound:
		return sepRoundThin
	}
	return sepThin
}

// textOn picks the palette's dark or light text for readable contrast on bg.
func textOn(bg string) string {
	r, g, b := rgb(bg)
	if 0.299*float64(r)+0.587*float64(g)+0.114*float64(b) > 140 {
		return theme.dark
	}
	return theme.light
}
