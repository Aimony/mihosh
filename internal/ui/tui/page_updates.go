package tui

import (
	"strings"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/ui/tui/pages"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// updateNodesPage 更新节点页面
func (m Model) updateNodesPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Up):
		if m.selectedProxy > 0 {
			m.selectedProxy--
			// 调整滚动位置
			if m.selectedProxy < m.proxyScrollTop {
				m.proxyScrollTop = m.selectedProxy
			}
		}

	case key.Matches(msg, keys.Down):
		if m.selectedProxy < len(m.currentProxies)-1 {
			m.selectedProxy++
			// 滚动位置将在渲染时自动调整
		}

	case key.Matches(msg, keys.Left):
		if m.selectedGroup > 0 {
			m.selectedGroup--
			m.updateCurrentProxies()
			m.selectedProxy = 0
			m.proxyScrollTop = 0
			// 调整策略组滚动位置
			if m.selectedGroup < m.groupScrollTop {
				m.groupScrollTop = m.selectedGroup
			}
		}

	case key.Matches(msg, keys.Right):
		if m.selectedGroup < len(m.groupNames)-1 {
			m.selectedGroup++
			m.updateCurrentProxies()
			m.selectedProxy = 0
			m.proxyScrollTop = 0
			// 滚动位置将在渲染时自动调整
		}

	case key.Matches(msg, keys.Enter):
		if len(m.currentProxies) > 0 && m.selectedProxy < len(m.currentProxies) {
			groupName := m.groupNames[m.selectedGroup]
			proxyName := m.currentProxies[m.selectedProxy]
			return m, selectProxy(m.client, groupName, proxyName)
		}

	case key.Matches(msg, keys.Test):
		if len(m.currentProxies) > 0 && m.selectedProxy < len(m.currentProxies) {
			proxyName := m.currentProxies[m.selectedProxy]
			m.testing = true
			return m, testProxy(m.client, proxyName, m.testURL, m.timeout)
		}

	case key.Matches(msg, keys.TestAll):
		if len(m.groupNames) > 0 && len(m.currentProxies) > 0 {
			m.testing = true
			m.testFailures = []string{} // 清空之前的失败记录
			m.showFailureDetail = false // 重置详情显示
			return m, testAllProxies(m.client, m.currentProxies, m.testURL, m.timeout)
		}

	case msg.String() == "f":
		// 切换测速失败详情显示
		if len(m.testFailures) > 0 {
			m.showFailureDetail = !m.showFailureDetail
		}
	}

	return m, nil
}

// updateConnectionsPage 更新连接页面
func (m Model) updateConnectionsPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 详情模式处理
	if m.connDetailMode {
		switch {
		case key.Matches(msg, keys.Escape), key.Matches(msg, keys.Enter):
			m.connDetailMode = false
			m.connDetailSnapshot = nil // 清除快照
			m.connIPInfo = nil         // 清除IP信息
			m.connDetailScroll = 0     // 重置滚动位置
			return m, nil
		case key.Matches(msg, keys.Up):
			// 向上滚动
			if m.connDetailScroll > 0 {
				m.connDetailScroll--
			}
			return m, nil
		case key.Matches(msg, keys.Down):
			// 向下滚动
			m.connDetailScroll++
			return m, nil
		}
		return m, nil
	}

	// 过滤模式处理
	if m.connFilterMode {
		return m.handleConnFilterMode(msg)
	}

	switch {
	case key.Matches(msg, keys.Refresh):
		return m, fetchConnections(m.client)

	case key.Matches(msg, keys.Up):
		if m.selectedConn > 0 {
			m.selectedConn--
			// 调整滚动位置
			if m.selectedConn < m.connScrollTop {
				m.connScrollTop = m.selectedConn
			}
		}

	case key.Matches(msg, keys.Down):
		connCount := m.getFilteredConnCount()
		if m.selectedConn < connCount-1 {
			m.selectedConn++
		}

	case key.Matches(msg, keys.Enter):
		// 显示连接详情 - 保存快照
		conn := m.getSelectedConnection()
		if conn != nil {
			// 复制连接数据作为快照
			snapshot := *conn
			m.connDetailSnapshot = &snapshot
			m.connDetailMode = true
			m.connIPInfo = nil // 清除旧的IP信息
			// 异步获取目标IP的地理信息
			return m, fetchIPInfo(conn.Metadata.DestinationIP)
		}
		return m, nil

	case msg.String() == "x":
		// 关闭选中的连接
		conn := m.getSelectedConnection()
		if conn != nil {
			return m, tea.Batch(
				closeConnection(m.client, conn.ID),
				fetchConnections(m.client),
			)
		}

	case msg.String() == "X":
		// 关闭所有连接
		return m, tea.Batch(
			closeAllConnections(m.client),
			fetchConnections(m.client),
		)

	case msg.String() == "/":
		// 进入搜索模式
		m.connFilterMode = true
		return m, nil

	case msg.String() == "h":
		// 切换视图模式：活跃连接 <-> 历史连接
		m.connViewMode = (m.connViewMode + 1) % 2
		m.selectedConn = 0
		m.connScrollTop = 0
		return m, nil

	case msg.String() == "s":
		// 测试选中的网站
		if len(m.siteTests) > 0 && m.selectedSiteTest >= 0 && m.selectedSiteTest < len(m.siteTests) {
			site := m.siteTests[m.selectedSiteTest]
			m.siteTests[m.selectedSiteTest].Testing = true
			return m, testSiteDelay(m.proxyAddr, site.Name, site.URL, m.timeout)
		}

	case msg.String() == "S":
		// 测试所有网站
		if len(m.siteTests) > 0 {
			for i := range m.siteTests {
				m.siteTests[i].Testing = true
			}
			return m, testAllSites(m.proxyAddr, m.siteTests, m.timeout)
		}

	case key.Matches(msg, keys.Left):
		// 选择上一个网站
		if m.selectedSiteTest > 0 {
			m.selectedSiteTest--
		}

	case key.Matches(msg, keys.Right):
		// 选择下一个网站
		if m.selectedSiteTest < len(m.siteTests)-1 {
			m.selectedSiteTest++
		}

	case key.Matches(msg, keys.Escape):
		// 清除过滤
		if m.connFilter != "" {
			m.connFilter = ""
			m.selectedConn = 0
			m.connScrollTop = 0
		}
	}

	return m, nil
}

// handleConnFilterMode 处理连接过滤输入
func (m Model) handleConnFilterMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.connFilterMode = false
		return m, nil

	case key.Matches(msg, keys.Enter):
		m.connFilterMode = false
		m.selectedConn = 0
		m.connScrollTop = 0
		return m, nil

	case key.Matches(msg, keys.Backspace):
		if len(m.connFilter) > 0 {
			m.connFilter = m.connFilter[:len(m.connFilter)-1]
		}

	default:
		input := msg.String()
		if len(input) == 1 && input[0] >= 32 && input[0] < 127 {
			m.connFilter += input
		}
	}

	return m, nil
}

// getFilteredConnCount 获取过滤后的连接数量
func (m Model) getFilteredConnCount() int {
	// 根据视图模式选择数据源
	var connections []model.Connection
	if m.connViewMode == 0 {
		// 活跃连接
		if m.connections == nil {
			return 0
		}
		connections = m.connections.Connections
	} else {
		// 历史连接
		connections = m.closedConnections
	}

	if m.connFilter == "" {
		return len(connections)
	}
	// 简单计数
	count := 0
	filter := m.connFilter
	for _, conn := range connections {
		if containsFilter(conn, filter) {
			count++
		}
	}
	return count
}

// getSelectedConnection 获取当前选中的连接
func (m Model) getSelectedConnection() *model.Connection {
	// 根据视图模式选择数据源
	var connections []model.Connection
	if m.connViewMode == 0 {
		// 活跃连接
		if m.connections == nil || len(m.connections.Connections) == 0 {
			return nil
		}
		connections = m.connections.Connections
	} else {
		// 历史连接
		if len(m.closedConnections) == 0 {
			return nil
		}
		connections = m.closedConnections
	}

	if m.connFilter == "" {
		if m.selectedConn >= 0 && m.selectedConn < len(connections) {
			return &connections[m.selectedConn]
		}
		return nil
	}

	// 过滤后的列表中查找
	idx := 0
	for i := range connections {
		if containsFilter(connections[i], m.connFilter) {
			if idx == m.selectedConn {
				return &connections[i]
			}
			idx++
		}
	}
	return nil
}

// containsFilter 检查连接是否匹配过滤条件
func containsFilter(conn model.Connection, filter string) bool {
	if filter == "" {
		return true
	}
	filter = strings.ToLower(filter)
	if strings.Contains(strings.ToLower(conn.Metadata.Host), filter) {
		return true
	}
	if strings.Contains(strings.ToLower(conn.Rule), filter) {
		return true
	}
	if strings.Contains(strings.ToLower(conn.Metadata.DestinationIP), filter) {
		return true
	}
	for _, chain := range conn.Chains {
		if strings.Contains(strings.ToLower(chain), filter) {
			return true
		}
	}
	return false
}

// updateSettingsPage 更新设置页面
func (m Model) updateSettingsPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 如果在编辑模式,交给编辑处理器
	if m.editMode {
		return m.handleEditMode(msg)
	}

	switch {
	case key.Matches(msg, keys.Up):
		if m.selectedSetting > 0 {
			m.selectedSetting--
		}

	case key.Matches(msg, keys.Down):
		if m.selectedSetting < len(pages.SettingKeys)-1 {
			m.selectedSetting++
		}

	case key.Matches(msg, keys.Enter):
		// 进入编辑模式
		m.editMode = true
		m.editValue = pages.GetSettingValue(m.config, m.selectedSetting)
		m.editCursor = len(m.editValue) // 光标初始化到末尾
		return m, nil
	}

	return m, nil
}

// handleEditMode 处理编辑模式的按键
func (m Model) handleEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.editMode = false
		m.editValue = ""
		m.editCursor = 0
		return m, nil

	case key.Matches(msg, keys.Enter):
		// 保存配置
		key := pages.SettingKeys[m.selectedSetting]
		if err := m.configSvc.SetConfigValue(key, m.editValue); err != nil {
			m.err = err
		} else {
			// 重新加载配置
			cfg, _ := m.configSvc.LoadConfig()
			m.config = cfg
			m.editMode = false
			m.editValue = ""
			m.editCursor = 0
		}
		return m, nil

	case msg.String() == "left":
		// 光标左移（仅响应方向键，不响应 h 键以允许输入）
		if m.editCursor > 0 {
			m.editCursor--
		}

	case msg.String() == "right":
		// 光标右移（仅响应方向键，不响应 l 键以允许输入）
		if m.editCursor < len(m.editValue) {
			m.editCursor++
		}

	case key.Matches(msg, keys.Home):
		// 光标移到开头
		m.editCursor = 0

	case key.Matches(msg, keys.End):
		// 光标移到末尾
		m.editCursor = len(m.editValue)

	case key.Matches(msg, keys.Backspace):
		// 删除光标前一个字符
		if m.editCursor > 0 {
			m.editValue = m.editValue[:m.editCursor-1] + m.editValue[m.editCursor:]
			m.editCursor--
		}

	case key.Matches(msg, keys.Delete):
		// 删除光标后一个字符
		if m.editCursor < len(m.editValue) {
			m.editValue = m.editValue[:m.editCursor] + m.editValue[m.editCursor+1:]
		}

	default:
		// 处理字符输入和粘贴
		input := msg.String()
		// 过滤掉控制字符，但允许多字符粘贴
		if len(input) > 0 && (len(input) > 1 || (input[0] >= 32 && input[0] < 127)) {
			// 在光标位置插入
			m.editValue = m.editValue[:m.editCursor] + input + m.editValue[m.editCursor:]
			m.editCursor += len(input)
		}
	}

	return m, nil
}

// updateHelpPage 更新帮助页面
func (m Model) updateHelpPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 帮助页面没有特殊交互，只响应全局快捷键
	return m, nil
}

// updateLogsPage 更新日志页面
func (m Model) updateLogsPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Up):
		if m.selectedLog > 0 {
			m.selectedLog--
			// 调整滚动位置
			if m.selectedLog < m.logScrollTop {
				m.logScrollTop = m.selectedLog
			}
		}

	case key.Matches(msg, keys.Down):
		logCount := m.getFilteredLogCount()
		if m.selectedLog < logCount-1 {
			m.selectedLog++
		}

	case key.Matches(msg, keys.Left):
		// 水平向左滚动
		if m.logHScrollOffset > 0 {
			m.logHScrollOffset -= 10
			if m.logHScrollOffset < 0 {
				m.logHScrollOffset = 0
			}
		}

	case key.Matches(msg, keys.Right):
		// 水平向右滚动
		m.logHScrollOffset += 10

	case msg.String() == "<" || msg.String() == ",":
		// 切换到上一个日志级别
		if m.logLevel > 0 {
			m.logLevel--
		}
		m.selectedLog = 0
		m.logScrollTop = 0

	case msg.String() == ">" || msg.String() == ".":
		// 切换到下一个日志级别
		if m.logLevel < 4 {
			m.logLevel++
		}
		m.selectedLog = 0
		m.logScrollTop = 0

	case msg.String() == "/":
		// 进入搜索模式
		m.logFilterMode = true
		return m, nil

	case key.Matches(msg, keys.Clear):
		// 清空日志
		m.logs = nil
		m.selectedLog = 0
		m.logScrollTop = 0
		return m, nil

	case key.Matches(msg, keys.Escape):
		// 清除过滤
		if m.logFilter != "" {
			m.logFilter = ""
			m.selectedLog = 0
			m.logScrollTop = 0
		}
	}

	return m, nil
}

// handleLogFilterMode 处理日志过滤输入
func (m Model) handleLogFilterMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.logFilterMode = false
		return m, nil

	case key.Matches(msg, keys.Enter):
		m.logFilterMode = false
		m.selectedLog = 0
		m.logScrollTop = 0
		return m, nil

	case key.Matches(msg, keys.Backspace):
		if len(m.logFilter) > 0 {
			m.logFilter = m.logFilter[:len(m.logFilter)-1]
		}

	default:
		input := msg.String()
		if len(input) == 1 && input[0] >= 32 && input[0] < 127 {
			m.logFilter += input
		}
	}

	return m, nil
}

// getFilteredLogCount 获取过滤后的日志数量
func (m Model) getFilteredLogCount() int {
	if len(m.logs) == 0 {
		return 0
	}

	count := 0
	levelIndex := m.logLevel

	for _, log := range m.logs {
		logLevelIndex := getLogLevelIndex(log.Type)

		// 只统计当前级别及更高级别的日志
		if logLevelIndex < levelIndex {
			continue
		}

		// 关键词过滤
		if m.logFilter != "" && !strings.Contains(strings.ToLower(log.Payload), strings.ToLower(m.logFilter)) {
			continue
		}

		count++
	}

	return count
}

// getLogLevelIndex 获取日志级别索引
func getLogLevelIndex(level string) int {
	levels := []string{"debug", "info", "warning", "error", "silent"}
	for i, l := range levels {
		if l == level {
			return i
		}
	}
	return 1 // 默认info
}

// updateRulesPage 更新规则页面
func (m Model) updateRulesPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 过滤模式处理
	if m.ruleFilterMode {
		return m.handleRuleFilterMode(msg)
	}

	switch {
	case key.Matches(msg, keys.Up):
		if m.selectedRule > 0 {
			m.selectedRule--
			// 调整滚动位置
			if m.selectedRule < m.ruleScrollTop {
				m.ruleScrollTop = m.selectedRule
			}
		}

	case key.Matches(msg, keys.Down):
		ruleCount := m.getFilteredRuleCount()
		if m.selectedRule < ruleCount-1 {
			m.selectedRule++
		}

	case msg.String() == "/":
		// 进入搜索模式
		m.ruleFilterMode = true
		return m, nil

	case key.Matches(msg, keys.Refresh):
		// 刷新规则列表
		return m, fetchRules(m.client)

	case key.Matches(msg, keys.Escape):
		// 清除过滤
		if m.ruleFilter != "" {
			m.ruleFilter = ""
			m.selectedRule = 0
			m.ruleScrollTop = 0
		}
	}

	return m, nil
}

// handleRuleFilterMode 处理规则过滤输入
func (m Model) handleRuleFilterMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.ruleFilterMode = false
		return m, nil

	case key.Matches(msg, keys.Enter):
		m.ruleFilterMode = false
		m.selectedRule = 0
		m.ruleScrollTop = 0
		return m, nil

	case key.Matches(msg, keys.Backspace):
		if len(m.ruleFilter) > 0 {
			m.ruleFilter = m.ruleFilter[:len(m.ruleFilter)-1]
		}

	default:
		input := msg.String()
		if len(input) == 1 && input[0] >= 32 && input[0] < 127 {
			m.ruleFilter += input
		}
	}

	return m, nil
}

// getFilteredRuleCount 获取过滤后的规则数量
func (m Model) getFilteredRuleCount() int {
	if len(m.rules) == 0 {
		return 0
	}

	if m.ruleFilter == "" {
		return len(m.rules)
	}

	// 分割关键词
	keywords := strings.Fields(strings.ToLower(m.ruleFilter))
	if len(keywords) == 0 {
		return len(m.rules)
	}

	count := 0
	for _, rule := range m.rules {
		// 构建搜索文本
		searchText := strings.ToLower(rule.Type + " " + rule.Payload + " " + rule.Proxy)

		// 所有关键词都必须匹配
		allMatch := true
		for _, keyword := range keywords {
			if !strings.Contains(searchText, keyword) {
				allMatch = false
				break
			}
		}

		if allMatch {
			count++
		}
	}

	return count
}
