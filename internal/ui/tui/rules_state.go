package tui

import (
	"strings"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/infrastructure/api"
	"github.com/aimony/mihosh/internal/ui/tui/pages"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// RulesState 规则页面完整状态
type RulesState struct {
	rules               []model.Rule
	filteredRuleIndices []int
	ruleFilter          string
	ruleFilterMode      bool
	selectedRule        int
	ruleScrollTop       int
}

// ToPageState 转换为渲染层所需的 RulesPageState
func (s RulesState) ToPageState(width, height int) pages.RulesPageState {
	return pages.RulesPageState{
		Rules:               s.rules,
		FilteredRuleIndices: s.filteredRuleIndices,
		FilterText:          s.ruleFilter,
		FilterMode:          s.ruleFilterMode,
		SelectedRule:        s.selectedRule,
		ScrollTop:           s.ruleScrollTop,
		Width:               width,
		Height:              height,
	}
}

// Update 处理规则页面按键
func (s RulesState) Update(msg tea.KeyMsg, client *api.Client) (RulesState, tea.Cmd) {
	if s.ruleFilterMode {
		return s.handleRuleFilterMode(msg)
	}

	switch {
	case key.Matches(msg, keys.Up):
		if s.selectedRule > 0 {
			s.selectedRule--
			if s.selectedRule < s.ruleScrollTop {
				s.ruleScrollTop = s.selectedRule
			}
		}

	case key.Matches(msg, keys.Down):
		if s.selectedRule < len(s.filteredRuleIndices)-1 {
			s.selectedRule++
		}

	case msg.String() == "/":
		s.ruleFilterMode = true

	case key.Matches(msg, keys.Refresh):
		return s, fetchRules(client)

	case key.Matches(msg, keys.Escape):
		if s.ruleFilter != "" {
			s.ruleFilter = ""
			s.selectedRule = 0
			s.ruleScrollTop = 0
			s.updateFilteredRules()
		}
	}

	return s, nil
}

// HandleMouseScroll 鼠标滚轮处理
func (s RulesState) HandleMouseScroll(up bool) RulesState {
	count := len(s.filteredRuleIndices)
	if up {
		if s.selectedRule > 0 {
			s.selectedRule--
			if s.selectedRule < s.ruleScrollTop {
				s.ruleScrollTop = s.selectedRule
			}
		}
	} else {
		if s.selectedRule < count-1 {
			s.selectedRule++
		}
	}
	return s
}

// ApplyRules 应用新规则列表并重建过滤缓存
func (s RulesState) ApplyRules(rules []model.Rule) RulesState {
	s.rules = rules
	s.updateFilteredRules()
	return s
}

// handleRuleFilterMode 规则过滤输入模式
func (s RulesState) handleRuleFilterMode(msg tea.KeyMsg) (RulesState, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		s.ruleFilterMode = false
	case key.Matches(msg, keys.Enter):
		s.ruleFilterMode = false
		s.selectedRule = 0
		s.ruleScrollTop = 0
	case key.Matches(msg, keys.Backspace):
		if len(s.ruleFilter) > 0 {
			s.ruleFilter = s.ruleFilter[:len(s.ruleFilter)-1]
			s.updateFilteredRules()
		}
	default:
		input := msg.String()
		if len(input) == 1 && input[0] >= 32 && input[0] < 127 {
			s.ruleFilter += input
			s.updateFilteredRules()
		}
	}
	return s, nil
}

// updateFilteredRules 重建规则过滤索引缓存
func (s *RulesState) updateFilteredRules() {
	if len(s.rules) == 0 {
		s.filteredRuleIndices = nil
		return
	}
	if s.ruleFilter == "" {
		indices := make([]int, len(s.rules))
		for i := range s.rules {
			indices[i] = i
		}
		s.filteredRuleIndices = indices
		return
	}

	keywords := strings.Fields(strings.ToLower(s.ruleFilter))
	if len(keywords) == 0 {
		indices := make([]int, len(s.rules))
		for i := range s.rules {
			indices[i] = i
		}
		s.filteredRuleIndices = indices
		return
	}

	var indices []int
	for i, rule := range s.rules {
		searchText := strings.ToLower(rule.Type + " " + rule.Payload + " " + rule.Proxy)
		allMatch := true
		for _, kw := range keywords {
			if !strings.Contains(searchText, kw) {
				allMatch = false
				break
			}
		}
		if allMatch {
			indices = append(indices, i)
		}
	}
	s.filteredRuleIndices = indices
}
