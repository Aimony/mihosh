package tui

import (
	"testing"

	"github.com/aimony/mihosh/internal/domain/model"
)

func TestHandleMouseLeft_GroupSelectionKeepsLinkage(t *testing.T) {
	state := NodesState{
		groups: map[string]model.Group{
			"g1": {All: []string{"p1", "p2"}},
			"g2": {All: []string{"p3"}},
		},
		groupNames:    []string{"g1", "g2"},
		selectedGroup: 0,
	}
	state.updateCurrentProxies()

	next, cmd := state.HandleMouseLeft(4, 120, 24, nil)
	if cmd != nil {
		t.Fatalf("expected nil cmd for single group click")
	}
	if next.selectedGroup != 1 {
		t.Fatalf("expected selectedGroup=1, got %d", next.selectedGroup)
	}
	if len(next.currentProxies) != 1 || next.currentProxies[0] != "p3" {
		t.Fatalf("expected currentProxies linked to g2, got %#v", next.currentProxies)
	}
}

func TestHandleMouseLeft_ProxyDoubleClickExecutesSwitch(t *testing.T) {
	state := NodesState{
		groups: map[string]model.Group{
			"g1": {All: []string{"p1", "p2"}},
		},
		groupNames:    []string{"g1"},
		selectedGroup: 0,
	}
	state.updateCurrentProxies()

	next, cmd := state.HandleMouseLeft(8, 120, 24, nil)
	if cmd != nil {
		t.Fatalf("expected nil cmd for first click")
	}

	next, cmd = next.HandleMouseLeft(8, 120, 24, nil)
	if cmd == nil {
		t.Fatalf("expected non-nil cmd for proxy double click")
	}
}

func TestHandleMouseScroll_RespectsMouseFocus(t *testing.T) {
	state := NodesState{
		currentProxies: []string{"p1", "p2", "p3"},
		selectedProxy:  1,
		mouseFocus:     nodesMouseFocusGroup,
	}

	groupFocused := state.HandleMouseScroll(true)
	if groupFocused.selectedProxy != 1 {
		t.Fatalf("expected selectedProxy unchanged when group focused, got %d", groupFocused.selectedProxy)
	}

	state.mouseFocus = nodesMouseFocusProxy
	proxyFocused := state.HandleMouseScroll(true)
	if proxyFocused.selectedProxy != 0 {
		t.Fatalf("expected selectedProxy moved when proxy focused, got %d", proxyFocused.selectedProxy)
	}
}
