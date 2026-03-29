package components

import (
	"strings"
	"testing"
)

func TestRenderStatusBar_TestingWithTarget(t *testing.T) {
	bar := RenderStatusBar(120, nil, true, "HK-01", nil)
	if !strings.Contains(bar, "正在测速: HK-01") {
		t.Fatalf("expected testing target in status bar, got: %q", bar)
	}
}

func TestRenderStatusBar_TestingWithoutTarget(t *testing.T) {
	bar := RenderStatusBar(120, nil, true, "", nil)
	if !strings.Contains(bar, "正在测速...") {
		t.Fatalf("expected generic testing text in status bar, got: %q", bar)
	}
}
