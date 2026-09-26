package main

import (
	"encoding/json"
	"os"
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
	for _, want := range []string{"Opus 5.5", "42k", "/1M", "22%", "1h00", "~$0.22"} {
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
	if out := contextSegment(p); !strings.Contains(out, "/compact") {
		t.Errorf("expected compact warning: %q", out)
	}
}
