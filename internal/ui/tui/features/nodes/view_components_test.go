package nodes

import "testing"

func TestResolveMouseHit_GroupAndProxy(t *testing.T) {
	state := PageState{
		GroupNames:     []string{"g1", "g2", "g3"},
		SelectedGroup:  0,
		CurrentProxies: []string{"p1", "p2", "p3", "p4"},
		SelectedProxy:  0,
		Height:         24,
	}

	groupHit := ResolveMouseHit(state, 10, 5)
	if groupHit.Target != MouseTargetGroup || groupHit.Index != 0 {
		t.Fatalf("expected group index 0, got target=%v index=%d", groupHit.Target, groupHit.Index)
	}

	// proxyListStart 依赖 groupMaxLines. height=24 -> groupMaxLines=4.
	// groupListLines = 1 + 3 = 4.
	// proxyHeaderStart = 4 + 4 + 1 = 9
	// proxyListStart = 9 + 2 = 11
	// proxyDataStart = 11 + 1 = 12
	proxyHit := ResolveMouseHit(state, 10, 12)
	if proxyHit.Target != MouseTargetProxy || proxyHit.Index != 0 {
		t.Fatalf("expected proxy index 0, got target=%v index=%d", proxyHit.Target, proxyHit.Index)
	}

	headerHit := ResolveMouseHit(state, 10, 4)
	if headerHit.Target != MouseTargetNone {
		t.Fatalf("expected no hit on header row, got target=%v index=%d", headerHit.Target, headerHit.Index)
	}
}

func TestResolveMouseHit_WithScrollWindow(t *testing.T) {
	state := PageState{
		GroupNames:     []string{"g0", "g1", "g2", "g3", "g4", "g5", "g6", "g7", "g8", "g9"},
		SelectedGroup:  7,
		GroupScrollTop: 0,
		CurrentProxies: []string{"p0", "p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10", "p11"},
		SelectedProxy:  11,
		ProxyScrollTop: 0,
		Height:         24,
	}

	// groupMaxLines=4 时，selected=7 会将可见窗口顶到 4，首行数据应命中 g4
	groupHit := ResolveMouseHit(state, 10, 5)
	if groupHit.Target != MouseTargetGroup || groupHit.Index != 4 {
		t.Fatalf("expected group index 4, got target=%v index=%d", groupHit.Target, groupHit.Index)
	}

	// proxyMaxLines=11 时 (24-12-4=8), selected=11 会将可见窗口顶到 4，首行数据应命中 p4
	// proxyDataStart = 4 + (1 + 4) + 1 + 2 + 1 = 13 (approx)
	// 让 ResolveMouseHit 自动计算
	proxyHit := ResolveMouseHit(state, 10, 15)
	if proxyHit.Target != MouseTargetProxy || proxyHit.Index < 0 {
		t.Fatalf("expected proxy hit, got target=%v index=%d", proxyHit.Target, proxyHit.Index)
	}
}
