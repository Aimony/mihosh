package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/aimony/mihosh/internal/domain/model"
)

func TestRenderGroupListRespectsConfiguredOrder(t *testing.T) {
	groups := map[string]model.Group{
		"Auto": {
			Name: "Auto",
			Type: "Selector",
			Now:  "HK",
			All:  []string{"HK", "US"},
		},
		"Global": {
			Name: "Global",
			Type: "Selector",
			Now:  "Direct",
			All:  []string{"Direct", "Auto"},
		},
	}
	orderedNames := []string{"Global", "Auto"}
	proxies := map[string]model.Proxy{
		"HK":     {History: []model.Delay{{Delay: 10}}},
		"US":     {History: []model.Delay{{Delay: 20}}},
		"Direct": {History: []model.Delay{{Delay: 5}}},
		"Auto":   {History: []model.Delay{{Delay: 30}}},
	}

	var out bytes.Buffer
	if err := renderGroupList(&out, groups, orderedNames, proxies, outputFormatPlain); err != nil {
		t.Fatalf("renderGroupList returned error: %v", err)
	}

	output := out.String()
	globalIdx := strings.Index(output, "[Selector] Global")
	autoIdx := strings.Index(output, "[Selector] Auto")
	if globalIdx == -1 || autoIdx == -1 {
		t.Fatalf("missing group output, got:\n%s", output)
	}
	if globalIdx > autoIdx {
		t.Fatalf("expected Global before Auto, got:\n%s", output)
	}
}

func TestRenderGroupListJSON(t *testing.T) {
	groups := map[string]model.Group{
		"Auto": {
			Name: "Auto",
			Type: "Selector",
			Now:  "HK",
			All:  []string{"HK", "US"},
		},
	}
	orderedNames := []string{"Auto"}
	proxies := map[string]model.Proxy{
		"HK": {History: []model.Delay{{Delay: 10}}},
		"US": {History: []model.Delay{{Delay: 20}}},
	}

	var out bytes.Buffer
	if err := renderGroupList(&out, groups, orderedNames, proxies, outputFormatJSON); err != nil {
		t.Fatalf("renderGroupList returned error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, `"groups"`) {
		t.Fatalf("expected json output to contain groups key, got:\n%s", output)
	}
	if !strings.Contains(output, `"name": "Auto"`) {
		t.Fatalf("expected json output to contain group name, got:\n%s", output)
	}
	if !strings.Contains(output, `"delay_ms": 10`) {
		t.Fatalf("expected json output to contain delay_ms, got:\n%s", output)
	}
}

func TestRenderGroupListTable(t *testing.T) {
	groups := map[string]model.Group{
		"Auto": {
			Name: "Auto",
			Type: "Selector",
			Now:  "HK",
			All:  []string{"HK"},
		},
	}
	orderedNames := []string{"Auto"}
	proxies := map[string]model.Proxy{
		"HK": {History: []model.Delay{{Delay: 10}}},
	}

	var out bytes.Buffer
	if err := renderGroupList(&out, groups, orderedNames, proxies, outputFormatTable); err != nil {
		t.Fatalf("renderGroupList returned error: %v", err)
	}

	output := out.String()
	if !strings.Contains(output, "GROUP") || !strings.Contains(output, "PROXY") {
		t.Fatalf("expected table header in output, got:\n%s", output)
	}
	if !strings.Contains(output, "Auto") || !strings.Contains(output, "HK") {
		t.Fatalf("expected table rows in output, got:\n%s", output)
	}
}
