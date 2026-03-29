package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNodesState_TestAllStartsWithFirstProxyTarget(t *testing.T) {
	state := NodesState{
		groupNames:     []string{"Auto"},
		selectedGroup:  0,
		currentProxies: []string{"HK-01", "JP-01"},
	}

	next, _ := state.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}, nil, nil, "", 0)

	if !next.testing {
		t.Fatalf("expected testing=true after pressing a")
	}
	if !strings.Contains(next.testingTarget, "HK-01") {
		t.Fatalf("expected testingTarget contains first proxy, got %q", next.testingTarget)
	}
	if !next.testAllActive || next.testAllTotal != 2 || len(next.testAllRunning) != 2 {
		t.Fatalf("expected concurrent batch state initialized, active=%v total=%d running=%d", next.testAllActive, next.testAllTotal, len(next.testAllRunning))
	}
}

func TestNodesState_ApplyTestDone_AdvancesBatchTarget(t *testing.T) {
	state := NodesState{
		testing:        true,
		testingTarget:  "HK-01（已完成 0/2）",
		testAllActive:  true,
		testAllRunning: []string{"HK-01", "JP-01"},
		testAllTotal:   2,
		testAllDone:    0,
	}

	state = state.ApplyTestDone("HK-01", 123, nil)
	if !state.testing || !strings.Contains(state.testingTarget, "JP-01") {
		t.Fatalf("expected batch to continue with JP-01, got testing=%v target=%q", state.testing, state.testingTarget)
	}

	state = state.ApplyTestDone("JP-01", 101, nil)
	if state.testing || state.testingTarget != "" || state.testAllActive {
		t.Fatalf("expected batch to finish and clear status, got testing=%v target=%q active=%v", state.testing, state.testingTarget, state.testAllActive)
	}
}
