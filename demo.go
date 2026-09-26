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
	iconFolder   = ""
	iconBranch   = ""
	iconRobot    = "\U000f06a9"
	iconGauge    = "\U000f029a"
	iconClock    = ""
	iconDollar   = ""
	iconWarn     = ""
)

type demoSegment struct {
	text   string
	fg, bg int // xterm-256 colors
}

func fg256(c int) string { return fmt.Sprintf("\033[38;5;%dm", c) }
func bg256(c int) string { return fmt.Sprintf("\033[48;5;%dm", c) }

// powerline renders filled segments joined by a separator glyph whose
// foreground is the previous background, which creates the arrow effect.
func powerline(segs []demoSegment, sep string) string {
	var b strings.Builder
	for i, s := range segs {
		b.WriteString(bg256(s.bg) + fg256(s.fg) + " " + s.text + " ")
		if i+1 < len(segs) {
			b.WriteString(bg256(segs[i+1].bg) + fg256(s.bg) + sep)
		} else {
			b.WriteString(reset + fg256(s.bg) + sep + reset)
		}
	}
	return b.String()
}

// lean renders p10k's "lean" style: no backgrounds, colored text, thin separators.
func lean(segs []demoSegment, sep string) string {
	parts := make([]string, len(segs))
	for i, s := range segs {
		parts[i] = fg256(s.bg) + s.text + reset
	}
	return strings.Join(parts, " "+fg256(240)+sep+reset+" ")
}

func runDemo() {
	base := func(ctxWarn bool) []demoSegment {
		ctxBg, ctxTxt := 70, iconGauge+" 46k/1M"
		if ctxWarn {
			ctxBg, ctxTxt = 172, iconGauge+" 124k/1M "+iconWarn
		}
		return []demoSegment{
			{iconFolder + " claude-statusline " + iconBranch + " master", 255, 31},
			{iconRobot + " Opus 5.5", 255, 61},
			{ctxTxt, 234, ctxBg},
			{iconClock + " 25% · 3h54", 234, 114},
			{"~" + iconDollar + "0.67", 250, 238},
		}
	}

	variants := []struct {
		name string
		out  func(warn bool) string
	}{
		{"A · powerline classique", func(w bool) string { return powerline(base(w), sepSolid) }},
		{"B · arrondi", func(w bool) string { return powerline(base(w), sepRound) }},
		{"C · biseauté", func(w bool) string { return powerline(base(w), sepSlant) }},
		{"D · lean (sans fond)", func(w bool) string { return lean(base(w), sepThin) }},
		{"E · lean arrondi", func(w bool) string { return lean(base(w), sepRoundThin) }},
	}
	for _, v := range variants {
		fmt.Printf("%s%s%s\n  %s\n  %s\n\n", bold, v.name, reset, v.out(false), v.out(true))
	}
}
