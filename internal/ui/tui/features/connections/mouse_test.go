package connections

import (
	"testing"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/ui/tui/features/connections/components"
)

func TestConnsHandleMouseLeft_DoubleClickConnectionEntersDetail(t *testing.T) {
	state := State{
		Connections: &model.ConnectionsResponse{
			Connections: []model.Connection{
				{
					ID: "conn-1",
					Metadata: model.Metadata{
						Host:          "example.com",
						DestinationIP: "1.1.1.1",
					},
				},
			},
		},
	}

	const width, height = 120, 30
	x, y, ok := findConnMousePoint(state, width, height, MouseTargetConnection, 0)
	if !ok {
		t.Fatalf("failed to locate first connection row mouse point")
	}

	next, cmd := state.HandleMouseLeft(x, y, width, height, nil, 3000)
	if cmd != nil {
		t.Fatalf("expected nil cmd on first click, got non-nil")
	}
	if next.connDetailMode {
		t.Fatalf("expected detail mode disabled on first click")
	}

	next, cmd = next.HandleMouseLeft(x, y, width, height, nil, 3000)
	if cmd == nil {
		t.Fatalf("expected non-nil cmd on connection double click")
	}
	if !next.connDetailMode {
		t.Fatalf("expected detail mode enabled after double click")
	}
	if next.connDetailSnapshot == nil || next.connDetailSnapshot.ID != "conn-1" {
		t.Fatalf("expected detail snapshot for conn-1, got %#v", next.connDetailSnapshot)
	}
}

func TestConnsHandleMouseLeft_ClickHistoryTabSwitchesView(t *testing.T) {
	state := State{
		Connections:   &model.ConnectionsResponse{},
		connViewMode:  ConnViewActive,
		selectedConn:  5,
		connScrollTop: 3,
	}

	const width, height = 120, 30
	x, y, ok := findConnMousePoint(state, width, height, MouseTargetViewHistory, -1)
	if !ok {
		t.Fatalf("failed to locate history tab mouse point")
	}

	next, cmd := state.HandleMouseLeft(x, y, width, height, nil, 3000)
	if cmd != nil {
		t.Fatalf("expected nil cmd when switching view tab")
	}
	if next.connViewMode != ConnViewHistory {
		t.Fatalf("expected history view mode, got %d", next.connViewMode)
	}
	if next.selectedConn != 0 || next.connScrollTop != 0 {
		t.Fatalf("expected selection reset after tab switch, got selected=%d scrollTop=%d", next.selectedConn, next.connScrollTop)
	}
}

func TestConnsHandleMouseLeft_ClickActiveTabSwitchesView(t *testing.T) {
	state := State{
		Connections:  &model.ConnectionsResponse{},
		connViewMode: ConnViewHistory,
	}
	state.appendClosed(model.Connection{ID: "closed-1", Metadata: model.Metadata{Host: "closed.example"}})

	const width, height = 120, 30
	x, y, ok := findConnMousePoint(state, width, height, MouseTargetViewActive, -1)
	if !ok {
		t.Fatalf("failed to locate active tab mouse point")
	}

	next, cmd := state.HandleMouseLeft(x, y, width, height, nil, 3000)
	if cmd != nil {
		t.Fatalf("expected nil cmd when switching view tab")
	}
	if next.connViewMode != ConnViewActive {
		t.Fatalf("expected active view mode, got %d", next.connViewMode)
	}
}

func TestConnsHandleMouseLeft_DoubleClickHistoryConnectionEntersDetail(t *testing.T) {
	state := State{
		Connections:  &model.ConnectionsResponse{},
		connViewMode: ConnViewHistory,
	}
	state.appendClosed(model.Connection{
		ID: "closed-1",
		Metadata: model.Metadata{
			Host:          "closed.example.com",
			DestinationIP: "8.8.8.8",
		},
	})

	const width, height = 120, 30
	x, y, ok := findConnMousePoint(state, width, height, MouseTargetConnection, 0)
	if !ok {
		t.Fatalf("failed to locate first history connection row mouse point")
	}

	next, cmd := state.HandleMouseLeft(x, y, width, height, nil, 3000)
	if cmd != nil {
		t.Fatalf("expected nil cmd on first history click, got non-nil")
	}
	if next.connDetailMode {
		t.Fatalf("expected detail mode disabled on first history click")
	}

	next, cmd = next.HandleMouseLeft(x, y, width, height, nil, 3000)
	if cmd == nil {
		t.Fatalf("expected non-nil cmd on history connection double click")
	}
	if !next.connDetailMode {
		t.Fatalf("expected detail mode enabled after history double click")
	}
	if next.connDetailSnapshot == nil || next.connDetailSnapshot.ID != "closed-1" {
		t.Fatalf("expected detail snapshot for closed-1, got %#v", next.connDetailSnapshot)
	}
}

func TestConnsHandleMouseLeft_DoubleClickSiteTestTriggersSiteProbe(t *testing.T) {
	state := State{
		Connections: &model.ConnectionsResponse{},
		siteTests: []model.SiteTest{
			{Name: "A", URL: "http://a.test"},
			{Name: "B", URL: "http://b.test"},
			{Name: "C", URL: "http://c.test"},
			{Name: "D", URL: "http://d.test"},
		},
	}

	const width, height = 120, 30
	x, y, ok := findConnMousePoint(state, width, height, MouseTargetSiteTest, 1)
	if !ok {
		t.Fatalf("failed to locate second site-test card mouse point")
	}

	next, cmd := state.HandleMouseLeft(x, y, width, height, nil, 3000)
	if cmd != nil {
		t.Fatalf("expected nil cmd on first site click, got non-nil")
	}
	if next.selectedSiteTest != 1 {
		t.Fatalf("expected selectedSiteTest=1, got %d", next.selectedSiteTest)
	}
	if next.siteTests[1].Testing {
		t.Fatalf("expected selected site not testing on first click")
	}

	next, cmd = next.HandleMouseLeft(x, y, width, height, nil, 3000)
	if cmd == nil {
		t.Fatalf("expected non-nil cmd on site card double click")
	}
	if !next.siteTests[1].Testing {
		t.Fatalf("expected selected site testing state true after double click")
	}
}

func TestConnsHandleMouseLeft_ClickOutsideDetailClosesModal(t *testing.T) {
	state := State{
		connDetailMode: true,
		connDetailSnapshot: &model.Connection{
			ID: "conn-1",
			Metadata: model.Metadata{
				Host:          "example.com",
				DestinationIP: "1.1.1.1",
			},
		},
		connIPInfo:            &model.IPInfo{Country: "TW"},
		connDetailLeftScroll:  2,
		connDetailRightScroll: 3,
		connDetailFocusPanel:  1,
	}

	const width, height = 120, 30
	next, cmd := state.HandleMouseLeft(0, height-1, width, height, nil, 3000)
	if cmd != nil {
		t.Fatalf("expected nil cmd when closing detail by outside click")
	}
	if next.connDetailMode {
		t.Fatalf("expected detail mode closed by outside click")
	}
	if next.connDetailSnapshot != nil {
		t.Fatalf("expected detail snapshot cleared after close")
	}
	if next.connIPInfo != nil {
		t.Fatalf("expected ip info cleared after close")
	}
	if next.connDetailLeftScroll != 0 || next.connDetailRightScroll != 0 {
		t.Fatalf("expected detail scroll reset after close, got left=%d right=%d", next.connDetailLeftScroll, next.connDetailRightScroll)
	}
	if next.connDetailFocusPanel != 0 {
		t.Fatalf("expected detail focus reset after close, got %d", next.connDetailFocusPanel)
	}
}

func TestConnsHandleMouseLeft_ClickInsideDetailKeepsModal(t *testing.T) {
	state := State{
		connDetailMode: true,
		connDetailSnapshot: &model.Connection{
			ID: "conn-1",
			Metadata: model.Metadata{
				Host:          "example.com",
				DestinationIP: "1.1.1.1",
			},
		},
	}

	const width, height = 120, 30
	left, top, right, bottom := components.ResolveConnectionDetailModalBounds(
		state.connDetailSnapshot,
		state.connIPInfo,
		width,
		height,
		state.connDetailLeftScroll,
		state.connDetailRightScroll,
		state.connDetailFocusPanel,
	)
	if right <= left || bottom <= top {
		t.Fatalf("invalid modal bounds: left=%d top=%d right=%d bottom=%d", left, top, right, bottom)
	}
	clickX := left
	clickY := top

	next, cmd := state.HandleMouseLeft(clickX, clickY, width, height, nil, 3000)
	if cmd != nil {
		t.Fatalf("expected nil cmd when clicking inside detail modal")
	}
	if !next.connDetailMode {
		t.Fatalf("expected detail mode keep open when clicking inside modal")
	}
	if next.connDetailSnapshot == nil || next.connDetailSnapshot.ID != "conn-1" {
		t.Fatalf("expected detail snapshot retained, got %#v", next.connDetailSnapshot)
	}
}

func findConnMousePoint(
	state State,
	width int,
	height int,
	target MouseTarget,
	index int,
) (int, int, bool) {
	pageState := state.ToPageState(nil, width, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			hit := ResolveMouseHit(pageState, x, y)
			if hit.Target != target {
				continue
			}
			if index >= 0 && hit.Index != index {
				continue
			}
			return x, y, true
		}
	}
	return 0, 0, false
}
