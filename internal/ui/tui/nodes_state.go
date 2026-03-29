package tui

import (
	"fmt"
	"sort"
	"strings"

	"github.com/aimony/mihosh/internal/app/service"
	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/infrastructure/api"
	"github.com/aimony/mihosh/internal/ui/tui/pages"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const testFailureCap = 100 // 最多保留最近 100 条测速失败记录

// ProxySortOrder 节点列表排序方式
type ProxySortOrder int

const (
	SortOrderOriginal ProxySortOrder = iota // 原始顺序
	SortOrderAZ                             // A-Z 升序
	SortOrderZA                             // Z-A 降序
)

var sortOrderLabels = []string{"原始顺序", "A-Z 升序", "Z-A 降序"}

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
	// Ring Buffer for test failures
	testFailures [testFailureCap]string
	failHead     int // 写入位置
	failCount    int // 已写入总数（上限 testFailureCap）
	showFailureDetail   bool
	failureScrollTop    int
	// 排序
	proxySortOrder      ProxySortOrder
	originalProxies     []string // 记录原始顺序，便于恢复
	// 搜索
	nodeFilter          string
	nodeFilterMode      bool
	filteredProxyIndices []int // 过滤结果的索引缓存（对应 currentProxies 的下标）
}

// appendTestFailure 向 Ring Buffer 追加一条测速失败记录
func (s *NodesState) appendTestFailure(msg string) {
	s.testFailures[s.failHead] = msg
	s.failHead = (s.failHead + 1) % testFailureCap
	if s.failCount < testFailureCap {
		s.failCount++
	}
}

// TestFailures 返回测速失败列表（最新在前）
func (s NodesState) TestFailures() []string {
	if s.failCount == 0 {
		return nil
	}
	result := make([]string, s.failCount)
	for i := 0; i < s.failCount; i++ {
		idx := (s.failHead - 1 - i + testFailureCap) % testFailureCap
		result[i] = s.testFailures[idx]
	}
	return result
}

// sortedTestFailures 返回按当前排序模式排列的测速失败列表
func (s NodesState) sortedTestFailures() []string {
	result := s.TestFailures()
	if len(result) == 0 {
		return result
	}
	// 格式为 "节点名: 错误信息"，按冒号前的节点名排序
	nodeName := func(entry string) string {
		if i := strings.Index(entry, ": "); i >= 0 {
			return strings.ToLower(entry[:i])
		}
		return strings.ToLower(entry)
	}
	switch s.proxySortOrder {
	case SortOrderAZ:
		sort.Slice(result, func(i, j int) bool {
			return nodeName(result[i]) < nodeName(result[j])
		})
	case SortOrderZA:
		sort.Slice(result, func(i, j int) bool {
			return nodeName(result[i]) > nodeName(result[j])
		})
	// SortOrderOriginal: 保持 TestFailures() 返回的最新在前顺序
	}
	return result
}

// clearTestFailures 清空测速失败记录
func (s *NodesState) clearTestFailures() {
	s.failHead = 0
	s.failCount = 0
}

// displayProxies 返回当前应显示的节点列表（搜索时返回过滤结果，否则返回全部）
func (s NodesState) displayProxies() []string {
	if s.nodeFilter == "" {
		return s.currentProxies
	}
	result := make([]string, len(s.filteredProxyIndices))
	for i, idx := range s.filteredProxyIndices {
		result[i] = s.currentProxies[idx]
	}
	return result
}

// ToPageState 转换为渲染层所需的 NodesPageState
func (s NodesState) ToPageState(width, height int) pages.NodesPageState {
	return pages.NodesPageState{
		Groups:            s.groups,
		Proxies:           s.proxies,
		GroupNames:        s.groupNames,
		SelectedGroup:     s.selectedGroup,
		SelectedProxy:     s.selectedProxy,
		CurrentProxies:    s.displayProxies(),
		Testing:           s.testing,
		TestFailures:      s.sortedTestFailures(),
		ShowFailureDetail: s.showFailureDetail,
		FailureScrollTop:  s.failureScrollTop,
		SortOrderLabels:   sortOrderLabels,
		CurrentSortOrder:  int(s.proxySortOrder),
		Width:             width,
		Height:            height,
		GroupScrollTop:    s.groupScrollTop,
		ProxyScrollTop:    s.proxyScrollTop,
		FilterText:        s.nodeFilter,
		FilterMode:        s.nodeFilterMode,
	}
}

// Update 处理节点页面按键
func (s NodesState) Update(msg tea.KeyMsg, client *api.Client, proxySvc *service.ProxyService, testURL string, timeout int) (NodesState, tea.Cmd) {
	// 搜索输入模式：拦截所有按键用于输入
	if s.nodeFilterMode {
		return s.handleNodeFilterMode(msg)
	}

	// 失败详情弹窗打开时，↑/↓ 控制弹窗滚动，f/Esc 关闭弹窗
	if s.showFailureDetail {
		switch {
		case key.Matches(msg, keys.Up):
			if s.failureScrollTop > 0 {
				s.failureScrollTop--
			}
		case key.Matches(msg, keys.Down):
			s.failureScrollTop++
		case msg.String() == "f", msg.String() == "esc":
			s.showFailureDetail = false
			s.failureScrollTop = 0
		}
		return s, nil
	}

	display := s.displayProxies()

	switch {
	case key.Matches(msg, keys.Up):
		if s.selectedProxy > 0 {
			s.selectedProxy--
			if s.selectedProxy < s.proxyScrollTop {
				s.proxyScrollTop = s.selectedProxy
			}
		}

	case key.Matches(msg, keys.Down):
		if s.selectedProxy < len(display)-1 {
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
			len(display) > 0 && s.selectedProxy < len(display) {
			groupName := s.groupNames[s.selectedGroup]
			proxyName := display[s.selectedProxy]
			return s, selectProxy(client, groupName, proxyName)
		}

	case key.Matches(msg, keys.Test):
		if len(display) > 0 && s.selectedProxy < len(display) {
			proxyName := display[s.selectedProxy]
			s.testing = true
			return s, testProxy(client, proxyName, testURL, timeout)
		}

	case key.Matches(msg, keys.TestAll):
		if len(s.groupNames) > 0 && len(s.currentProxies) > 0 {
			s.testing = true
			s.clearTestFailures()
			s.showFailureDetail = false
			return s, testAllProxies(proxySvc, s.currentProxies)
		}

	case msg.String() == "f":
		if s.failCount > 0 {
			s.showFailureDetail = true
			s.failureScrollTop = 0
		}

	case msg.String() == "s":
		s.proxySortOrder = (s.proxySortOrder + 1) % ProxySortOrder(len(sortOrderLabels))
		s.applySortOrder()
		s.updateFilteredProxies()
		s.selectedProxy = 0
		s.proxyScrollTop = 0

	case msg.String() == "/":
		s.nodeFilterMode = true

	case key.Matches(msg, keys.Escape):
		if s.nodeFilter != "" {
			s.nodeFilter = ""
			s.selectedProxy = 0
			s.proxyScrollTop = 0
			s.updateFilteredProxies()
		}
	}

	return s, nil
}

// handleNodeFilterMode 搜索输入模式处理
func (s NodesState) handleNodeFilterMode(msg tea.KeyMsg) (NodesState, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		s.nodeFilterMode = false
		s.nodeFilter = ""
		s.selectedProxy = 0
		s.proxyScrollTop = 0
		s.updateFilteredProxies()
	case key.Matches(msg, keys.Enter):
		s.nodeFilterMode = false
		s.selectedProxy = 0
		s.proxyScrollTop = 0
	case key.Matches(msg, keys.Backspace):
		if len(s.nodeFilter) > 0 {
			// 正确截断多字节字符
			runes := []rune(s.nodeFilter)
			s.nodeFilter = string(runes[:len(runes)-1])
			s.updateFilteredProxies()
			s.selectedProxy = 0
			s.proxyScrollTop = 0
		}
	default:
		input := msg.String()
		// 接受可打印的单字符（ASCII）或多字节字符（如中文）
		runes := []rune(input)
		if len(runes) == 1 && runes[0] >= 32 {
			s.nodeFilter += input
			s.updateFilteredProxies()
			s.selectedProxy = 0
			s.proxyScrollTop = 0
		}
	}
	return s, nil
}

// updateFilteredProxies 重建搜索过滤索引缓存
func (s *NodesState) updateFilteredProxies() {
	if s.filteredProxyIndices == nil {
		s.filteredProxyIndices = make([]int, 0, 64)
	} else {
		s.filteredProxyIndices = s.filteredProxyIndices[:0]
	}
	if s.nodeFilter == "" {
		return
	}
	filter := strings.ToLower(s.nodeFilter)
	for i, name := range s.currentProxies {
		if strings.Contains(strings.ToLower(name), filter) {
			s.filteredProxyIndices = append(s.filteredProxyIndices, i)
		}
	}
}

// HandleMouseScroll 处理鼠标滚轮（弹窗打开时控制弹窗滚动，否则控制节点列表）
func (s NodesState) HandleMouseScroll(up bool) NodesState {
	if s.showFailureDetail {
		if up {
			if s.failureScrollTop > 0 {
				s.failureScrollTop--
			}
		} else {
			s.failureScrollTop++
		}
		return s
	}

	displayCount := len(s.displayProxies())
	if up {
		if s.selectedProxy > 0 {
			s.selectedProxy--
			if s.selectedProxy < s.proxyScrollTop {
				s.proxyScrollTop = s.selectedProxy
			}
		}
	} else {
		if s.selectedProxy < displayCount-1 {
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
		s.appendTestFailure(fmt.Sprintf("%s: %s", name, err.Error()))
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
			s.appendTestFailure(fmt.Sprintf("%s: timeout or error", name))
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
			// 保存原始顺序副本
			original := make([]string, len(group.All))
			copy(original, group.All)
			s.originalProxies = original
			// 应用当前排序
			s.currentProxies = original
			s.applySortOrder()
			s.updateFilteredProxies()
		}
	}
}

// applySortOrder 根据当前排序模式重新排列 currentProxies
func (s *NodesState) applySortOrder() {
	switch s.proxySortOrder {
	case SortOrderOriginal:
		// 恢复原始顺序
		if len(s.originalProxies) > 0 {
			copied := make([]string, len(s.originalProxies))
			copy(copied, s.originalProxies)
			s.currentProxies = copied
		}
	case SortOrderAZ:
		copied := make([]string, len(s.originalProxies))
		copy(copied, s.originalProxies)
		sort.Slice(copied, func(i, j int) bool {
			return strings.ToLower(copied[i]) < strings.ToLower(copied[j])
		})
		s.currentProxies = copied
	case SortOrderZA:
		copied := make([]string, len(s.originalProxies))
		copy(copied, s.originalProxies)
		sort.Slice(copied, func(i, j int) bool {
			return strings.ToLower(copied[i]) > strings.ToLower(copied[j])
		})
		s.currentProxies = copied
	}
}
