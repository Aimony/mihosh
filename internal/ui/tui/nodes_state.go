package tui

import (
	"fmt"

	"github.com/aimony/mihosh/internal/app/service"
	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/infrastructure/api"
	"github.com/aimony/mihosh/internal/ui/tui/pages"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// NodesState 节点页面完整状态
type NodesState struct {
	groups            map[string]model.Group
	proxies           map[string]model.Proxy
	groupNames        []string
	currentProxies    []string
	selectedGroup     int
	selectedProxy     int
	groupScrollTop    int
	proxyScrollTop    int
	testPending       int
	testing           bool
	testFailures      []string
	showFailureDetail bool
}

// ToPageState 转换为渲染层所需的 NodesPageState
func (s NodesState) ToPageState(width, height int) pages.NodesPageState {
	return pages.NodesPageState{
		Groups:            s.groups,
		Proxies:           s.proxies,
		GroupNames:        s.groupNames,
		SelectedGroup:     s.selectedGroup,
		SelectedProxy:     s.selectedProxy,
		CurrentProxies:    s.currentProxies,
		Testing:           s.testing,
		TestFailures:      s.testFailures,
		ShowFailureDetail: s.showFailureDetail,
		Width:             width,
		Height:            height,
		GroupScrollTop:    s.groupScrollTop,
		ProxyScrollTop:    s.proxyScrollTop,
	}
}

// Update 处理节点页面按键
func (s NodesState) Update(msg tea.KeyMsg, client *api.Client, proxySvc *service.ProxyService, testURL string, timeout int) (NodesState, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Up):
		if s.selectedProxy > 0 {
			s.selectedProxy--
			if s.selectedProxy < s.proxyScrollTop {
				s.proxyScrollTop = s.selectedProxy
			}
		}

	case key.Matches(msg, keys.Down):
		if s.selectedProxy < len(s.currentProxies)-1 {
			s.selectedProxy++
		}

	case key.Matches(msg, keys.Left):
		if s.selectedGroup > 0 {
			s.selectedGroup--
			s.updateCurrentProxies()
			s.selectedProxy = 0
			s.proxyScrollTop = 0
			if s.selectedGroup < s.groupScrollTop {
				s.groupScrollTop = s.selectedGroup
			}
		}

	case key.Matches(msg, keys.Right):
		if s.selectedGroup < len(s.groupNames)-1 {
			s.selectedGroup++
			s.updateCurrentProxies()
			s.selectedProxy = 0
			s.proxyScrollTop = 0
		}

	case key.Matches(msg, keys.Enter):
		if len(s.groupNames) > 0 && s.selectedGroup < len(s.groupNames) &&
			len(s.currentProxies) > 0 && s.selectedProxy < len(s.currentProxies) {
			groupName := s.groupNames[s.selectedGroup]
			proxyName := s.currentProxies[s.selectedProxy]
			return s, selectProxy(client, groupName, proxyName)
		}

	case key.Matches(msg, keys.Test):
		if len(s.currentProxies) > 0 && s.selectedProxy < len(s.currentProxies) {
			proxyName := s.currentProxies[s.selectedProxy]
			s.testing = true
			return s, testProxy(client, proxyName, testURL, timeout)
		}

	case key.Matches(msg, keys.TestAll):
		if len(s.groupNames) > 0 && len(s.currentProxies) > 0 {
			s.testing = true
			s.testFailures = nil
			s.showFailureDetail = false
			return s, testAllProxies(proxySvc, s.currentProxies)
		}

	case msg.String() == "f":
		if len(s.testFailures) > 0 {
			s.showFailureDetail = !s.showFailureDetail
		}
	}

	return s, nil
}

// HandleMouseScroll 处理鼠标滚轮（节点列表区域）
func (s NodesState) HandleMouseScroll(up bool) NodesState {
	if up {
		if s.selectedProxy > 0 {
			s.selectedProxy--
			if s.selectedProxy < s.proxyScrollTop {
				s.proxyScrollTop = s.selectedProxy
			}
		}
	} else {
		if s.selectedProxy < len(s.currentProxies)-1 {
			s.selectedProxy++
		}
	}
	return s
}

// ApplyGroups 应用策略组刷新结果（保留选中项）
func (s NodesState) ApplyGroups(groups map[string]model.Group, orderedNames []string) NodesState {
	var selectedGroupName, selectedProxyName string
	if len(s.groupNames) > 0 && s.selectedGroup < len(s.groupNames) {
		selectedGroupName = s.groupNames[s.selectedGroup]
	}
	if len(s.currentProxies) > 0 && s.selectedProxy < len(s.currentProxies) {
		selectedProxyName = s.currentProxies[s.selectedProxy]
	}

	s.groups = groups
	s.groupNames = orderedNames

	if selectedGroupName != "" {
		for i, name := range s.groupNames {
			if name == selectedGroupName {
				s.selectedGroup = i
				break
			}
		}
	}
	s.updateCurrentProxies()
	if selectedProxyName != "" {
		for i, name := range s.currentProxies {
			if name == selectedProxyName {
				s.selectedProxy = i
				break
			}
		}
	}
	return s
}

// ApplyProxies 应用节点数据
func (s NodesState) ApplyProxies(proxies map[string]model.Proxy) NodesState {
	s.proxies = proxies
	return s
}

// ApplyTestDone 单节点测速完成
func (s NodesState) ApplyTestDone(name string, delay int, err error) NodesState {
	if err != nil {
		s.testFailures = append(s.testFailures, fmt.Sprintf("%s: %s", name, err.Error()))
	}
	if s.testPending > 0 {
		s.testPending--
		if s.testPending == 0 {
			s.testing = false
		}
	} else {
		s.testing = false
	}
	return s
}

// ApplyTestAllDone 批量测速完成
func (s NodesState) ApplyTestAllDone(results map[string]int) NodesState {
	for name, delay := range results {
		if delay == -1 {
			s.testFailures = append(s.testFailures, fmt.Sprintf("%s: timeout or error", name))
		}
	}
	s.testing = false
	return s
}

// updateCurrentProxies 更新当前策略组的节点列表（指针接收者，修改自身）
func (s *NodesState) updateCurrentProxies() {
	if len(s.groupNames) > 0 && s.selectedGroup < len(s.groupNames) {
		groupName := s.groupNames[s.selectedGroup]
		if group, ok := s.groups[groupName]; ok {
			s.currentProxies = group.All
		}
	}
}
