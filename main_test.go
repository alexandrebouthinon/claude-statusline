package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestHumanTokens(t *testing.T) {
	cases := map[int]string{0: "0", 999: "999", 46_000: "46k", 1_000_000: "1M", 1_500_000: "1.5M"}
	for in, want := range cases {
		if got := humanTokens(in); got != want {
			t.Errorf("humanTokens(%d) = %q, want %q", in, got, want)
		}
	}
}

func TestRenderSample(t *testing.T) {
	data, err := os.ReadFile("testdata/sample.json")
	if err != nil {
		t.Fatal(err)
	}
	var p payload
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatal(err)
	}
	out := render(p, time.Unix(p.RateLimits.FiveHour.ResetsAt-3600, 0))
	for _, want := range []string{"Opus 5.5", "42k", "/1M", "22%", "1h00", "0.22"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q: %q", want, out)
		}
	}
	if strings.Contains(out, "/compact") {
		t.Errorf("compact warning under threshold: %q", out)
	}
}

func TestCompactWarning(t *testing.T) {
	var p payload
	_ = json.Unmarshal([]byte(`{"context_window":{"context_window_size":1000000,"current_usage":{"cache_read_input_tokens":120000}}}`), &p)
	if out := contextSegment(p); !strings.Contains(out.text, "/compact") {
		t.Errorf("expected compact warning: %q", out.text)
	}
}

func TestShortDuration(t *testing.T) {
	cases := map[time.Duration]string{
		42 * time.Minute:              "42m",
		3*time.Hour + 5*time.Minute:   "3h05",
		76*time.Hour + 10*time.Minute: "3j04h",
	}
	for in, want := range cases {
		if got := shortDuration(in); got != want {
			t.Errorf("shortDuration(%v) = %q, want %q", in, got, want)
		}
	}
}

func TestModelFamily(t *testing.T) {
	cases := []struct{ model, parent, want string }{
		{"haiku", "Opus 5.5", "haiku"},
		{"claude-sonnet-5", "", "sonnet"},
		{"inherit", "Opus 5.5 (1M context)", "opus"},
		{"", "Fable 5.1", "fable"},
		{"", "", "other"},
	}
	for _, c := range cases {
		if got := modelFamily(c.model, c.parent); got != c.want {
			t.Errorf("modelFamily(%q, %q) = %q, want %q", c.model, c.parent, got, c.want)
		}
	}
}

func TestRunningAgents(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	projects := t.TempDir()
	p := payload{SessionID: "s1", TranscriptPath: filepath.Join(projects, "s1.jsonl")}
	p.Model.DisplayName = "Opus 5.5"

	sub := filepath.Join(projects, "s1", "subagents")
	os.MkdirAll(sub, 0o700)
	os.WriteFile(filepath.Join(sub, "agent-a1.meta.json"), []byte(`{"model":"haiku"}`), 0o600)
	os.WriteFile(filepath.Join(sub, "agent-a2.meta.json"), []byte(`{}`), 0o600)

	for _, id := range []string{"a1", "a2"} {
		hook(t, `{"hook_event_name":"SubagentStart","session_id":"s1","agent_id":"`+id+`"}`)
	}
	got := runningAgents(p, time.Now())
	if got["haiku"] != 1 || got["opus"] != 1 {
		t.Fatalf("after start: %v", got)
	}

	hook(t, `{"hook_event_name":"SubagentStop","session_id":"s1","agent_id":"a1"}`)
	if got := runningAgents(p, time.Now()); got["haiku"] != 0 || got["opus"] != 1 {
		t.Fatalf("after stop: %v", got)
	}
	hook(t, `{"hook_event_name":"PostToolUse","session_id":"s1","tool_name":"TaskStop","tool_input":{"task_id":"a2"}}`)
	if got := runningAgents(p, time.Now()); len(got) != 0 {
		t.Fatalf("after TaskStop: %v", got)
	}
	if got := runningAgents(p, time.Now().Add(time.Hour)); len(got) != 0 {
		t.Fatalf("stale markers kept: %v", got)
	}
}

func hook(t *testing.T, input string) {
	t.Helper()
	r, w, _ := os.Pipe()
	w.WriteString(input)
	w.Close()
	stdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = stdin }()
	runHook()
}

func TestLayoutTeamRow(t *testing.T) {
	segs := []segment{{"core", theme.light, theme.dir}}
	if out := layout(segs, nil); strings.Contains(out, "\n") {
		t.Errorf("team row without agents: %q", out)
	}
	out := layout(segs, map[string]int{"haiku": 3, "sonnet": 1})
	lines := strings.Split(out, "\n")
	if len(lines) != 2 {
		t.Fatalf("want 2 rows, got %d: %q", len(lines), out)
	}
	for _, want := range []string{"team", "●●● Haiku", "● Sonnet"} {
		if !strings.Contains(lines[1], want) {
			t.Errorf("team row missing %q: %q", want, lines[1])
		}
	}
	if strings.Index(lines[1], "Sonnet") > strings.Index(lines[1], "Haiku") {
		t.Errorf("team row not ordered by capability: %q", lines[1])
	}
}
