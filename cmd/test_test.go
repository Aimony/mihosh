package cmd

import (
	"strings"
	"testing"

	"github.com/aimony/mihosh/internal/domain/model"
)

func TestResolveTestAction(t *testing.T) {
	tests := []struct {
		name       string
		args       []string
		wantAction testAction
		wantTarget string
		wantErr    string
	}{
		{
			name:       "no args tests current selected node",
			args:       []string{},
			wantAction: actionCurrent,
		},
		{
			name:       "node action with target",
			args:       []string{"node", "HK"},
			wantAction: actionNode,
			wantTarget: "HK",
		},
		{
			name:       "group action with target",
			args:       []string{"group", "Auto"},
			wantAction: actionGroup,
			wantTarget: "Auto",
		},
		{
			name:    "single arg is invalid",
			args:    []string{"HK"},
			wantErr: "参数格式错误",
		},
		{
			name:    "unknown action is invalid",
			args:    []string{"foo", "bar"},
			wantErr: "参数格式错误",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			action, target, err := resolveTestAction(tc.args)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tc.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if action != tc.wantAction {
				t.Fatalf("expected action %q, got %q", tc.wantAction, action)
			}
			if target != tc.wantTarget {
				t.Fatalf("expected target %q, got %q", tc.wantTarget, target)
			}
		})
	}
}

func TestResolveCurrentSelectedNode(t *testing.T) {
	tests := []struct {
		name      string
		proxies   map[string]model.Proxy
		wantNode  string
		wantFound bool
	}{
		{
			name: "resolve from global chain",
			proxies: map[string]model.Proxy{
				"GLOBAL": {Now: "Proxy", All: []string{"Proxy"}},
				"Proxy":  {Now: "HK", All: []string{"HK", "US"}},
				"HK":     {},
			},
			wantNode:  "HK",
			wantFound: true,
		},
		{
			name: "resolve from proxy root without global",
			proxies: map[string]model.Proxy{
				"Proxy": {Now: "JP", All: []string{"JP"}},
				"JP":    {},
			},
			wantNode:  "JP",
			wantFound: true,
		},
		{
			name: "no selected leaf returns not found",
			proxies: map[string]model.Proxy{
				"GLOBAL": {Now: "Proxy", All: []string{"Proxy"}},
				"Proxy":  {All: []string{"HK", "US"}},
			},
			wantFound: false,
		},
		{
			name: "cycle returns not found",
			proxies: map[string]model.Proxy{
				"GLOBAL": {Now: "Proxy", All: []string{"Proxy"}},
				"Proxy":  {Now: "GLOBAL", All: []string{"GLOBAL"}},
			},
			wantFound: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotNode, gotFound := resolveCurrentSelectedNode(tc.proxies)
			if gotFound != tc.wantFound {
				t.Fatalf("expected found=%v, got %v", tc.wantFound, gotFound)
			}
			if gotNode != tc.wantNode {
				t.Fatalf("expected node %q, got %q", tc.wantNode, gotNode)
			}
		})
	}
}
