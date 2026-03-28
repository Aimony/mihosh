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
	renderGroupList(&out, groups, orderedNames, proxies)

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
