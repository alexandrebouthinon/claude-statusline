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

// A marker older than this whose transcript stopped moving is treated as a
// leftover from a crashed session rather than a running agent.
const staleAfter = 15 * time.Minute

type hookInput struct {
	Event     string `json:"hook_event_name"`
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id"`
}

func stateDir(sessionID string) string {
	base := os.Getenv("XDG_RUNTIME_DIR")
	if base == "" {
		base = os.TempDir()
	}
	return filepath.Join(base, "claude-statusline", sessionID)
}

// runHook tracks running subagents: SubagentStart drops a marker, SubagentStop removes it.
func runHook() {
	var in hookInput
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil || in.SessionID == "" || in.AgentID == "" {
		return
	}
	dir := stateDir(in.SessionID)
	marker := filepath.Join(dir, filepath.Base(in.AgentID))
	switch in.Event {
	case "SubagentStart":
		os.MkdirAll(dir, 0o700)
		os.WriteFile(marker, nil, 0o600)
	case "SubagentStop":
		os.Remove(marker)
	}
}

// runningAgents counts live subagents per model family (haiku, sonnet, opus…).
func runningAgents(p payload, now time.Time) map[string]int {
	if p.SessionID == "" {
		return nil
	}
	entries, err := os.ReadDir(stateDir(p.SessionID))
	if err != nil {
		return nil
	}
	subDir := filepath.Join(strings.TrimSuffix(p.TranscriptPath, ".jsonl"), "subagents")
	counts := map[string]int{}
	for _, e := range entries {
		id := e.Name()
		base := filepath.Join(subDir, "agent-"+id)
		if isStale(filepath.Join(stateDir(p.SessionID), id), base+".jsonl", now) {
			os.Remove(filepath.Join(stateDir(p.SessionID), id))
			continue
		}
		counts[modelFamily(agentModel(base), p.Model.DisplayName)]++
	}
	return counts
}

func isStale(marker, transcript string, now time.Time) bool {
	last := time.Time{}
	for _, f := range []string{marker, transcript} {
		if info, err := os.Stat(f); err == nil && info.ModTime().After(last) {
			last = info.ModTime()
		}
	}
	return now.Sub(last) > staleAfter
}

// agentModel reads the model from the agent's meta file, falling back to the
// first assistant message of its transcript.
func agentModel(base string) string {
	var meta struct {
		Model string `json:"model"`
	}
	if data, err := os.ReadFile(base + ".meta.json"); err == nil && json.Unmarshal(data, &meta) == nil && meta.Model != "" {
		return meta.Model
	}
	f, err := os.Open(base + ".jsonl")
	if err != nil {
		return ""
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for sc.Scan() {
		var entry struct {
			Type    string `json:"type"`
			Message struct {
				Model string `json:"model"`
			} `json:"message"`
		}
		if json.Unmarshal(sc.Bytes(), &entry) == nil && entry.Type == "assistant" && entry.Message.Model != "" {
			return entry.Message.Model
		}
	}
	return ""
}

var families = []string{"haiku", "sonnet", "opus", "fable"}

// modelFamily maps an alias or model ID to its family; "inherit" and unknown
// values fall back to the parent session's model.
func modelFamily(model, parent string) string {
	for _, src := range []string{model, parent} {
		low := strings.ToLower(src)
		for _, f := range families {
			if strings.Contains(low, f) {
				return f
			}
		}
	}
	return "other"
}

// agentSegments renders one pill per model family, colored like the model pill.
func agentSegments(counts map[string]int) []segment {
	var segs []segment
	for _, f := range append(families, "other") {
		if n := counts[f]; n > 0 {
			segs = append(segs, segment{text: fmt.Sprintf("%s %s %d", iconAgents, strings.ToUpper(f[:1])+f[1:], n), bg: familyColor(f)})
		}
	}
	return segs
}
