package tui

import (
	"testing"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/ui/tui/pages"
)

func TestConnsHandleMouseLeft_DoubleClickConnectionEntersDetail(t *testing.T) {
	state := ConnsState{
		connections: &model.ConnectionsResponse{
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
	x, y, ok := findConnMousePoint(state, width, height, pages.ConnectionsMouseTargetConnection, 0)
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

func TestConnsHandleMouseLeft_DoubleClickSiteTestTriggersSiteProbe(t *testing.T) {
	state := ConnsState{
		connections: &model.ConnectionsResponse{},
		siteTests: []model.SiteTest{
			{Name: "A", URL: "http://a.test"},
			{Name: "B", URL: "http://b.test"},
			{Name: "C", URL: "http://c.test"},
			{Name: "D", URL: "http://d.test"},
		},
	}

	const width, height = 120, 30
	x, y, ok := findConnMousePoint(state, width, height, pages.ConnectionsMouseTargetSiteTest, 1)
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

func findConnMousePoint(
	state ConnsState,
	width int,
	height int,
	target pages.ConnectionsMouseTarget,
	index int,
) (int, int, bool) {
	pageState := state.ToPageState(nil, width, height)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			hit := pages.ResolveConnectionsMouseHit(pageState, x, y)
			if hit.Target == target && hit.Index == index {
				return x, y, true
			}
		}
	}
	return 0, 0, false
}
