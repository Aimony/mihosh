package tui

import (
	"time"

	"github.com/aimony/mihosh/internal/ui/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// connTickMsg 连接页面定时刷新消息
type connTickMsg time.Time

// logsTickMsg 日志页面定时刷新消息
type logsTickMsg time.Time

// connRefreshInterval 连接刷新间隔
const connRefreshInterval = 1 * time.Second

// connTick 创建连接页面定时器
func connTick() tea.Cmd {
	return tea.Tick(connRefreshInterval, func(t time.Time) tea.Msg {
		return connTickMsg(t)
	})
}

// logsTick 创建日志页面定时器
func logsTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return logsTickMsg(t)
	})
}

// Init 初始化
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchGroups(m.client),
		fetchProxies(m.client),
		startWSStreams(m.wsClient, m.wsMsgChan),
		listenWSMessages(m.wsCtx, m.wsMsgChan),
	)
}

// Update 消息路由器：全局消息自处理，页面消息分发到子状态
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// ── 全局：窗口大小 ──
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	// ── 全局：鼠标事件 ──
	case tea.MouseMsg:
		switch msg.Type {
		case tea.MouseLeft:
			statusBarHeight := 2
			contentHeight := m.height - statusBarHeight
			if contentHeight < 5 {
				contentHeight = 5
			}
			if msg.X >= 0 && msg.X < components.SidebarWidth && msg.Y >= 0 && msg.Y < contentHeight {
				clickedPage := components.GetClickedPage(msg.X, msg.Y)
				if clickedPage >= 0 && clickedPage < components.PageCount {
					m.currentPage = clickedPage
					return m, m.onPageChange()
				}
			}
		case tea.MouseWheelUp:
			return m.handleMouseScroll(true, msg.X)
		case tea.MouseWheelDown:
			return m.handleMouseScroll(false, msg.X)
		}
		return m, nil

	// ── 全局：键盘 ──
	case tea.KeyMsg:
		// 帮助弹窗拦截
		if m.showHelp {
			switch msg.String() {
			case "esc", "q", "?":
				m.showHelp = false
				return m, nil
			}
			if key.Matches(msg, keys.Quit) {
				return m, tea.Quit
			}
			return m, nil
		}

		// 全局帮助
		if msg.String() == "?" {
			m.showHelp = true
			return m, nil
		}

		// 全局快捷键
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.NextPage):
			m.currentPage = (m.currentPage + 1) % components.PageCount
			return m, m.onPageChange()

		case key.Matches(msg, keys.PrevPage):
			m.currentPage = (m.currentPage + components.PageCount - 1) % components.PageCount
			return m, m.onPageChange()

		case key.Matches(msg, keys.Page1):
			m.currentPage = components.PageNodes
			return m, m.onPageChange()

		case key.Matches(msg, keys.Page2):
			m.currentPage = components.PageConnections
			return m, m.onPageChange()

		case key.Matches(msg, keys.Page3):
			m.currentPage = components.PageLogs
			return m, m.onPageChange()

		case key.Matches(msg, keys.Page4):
			m.currentPage = components.PageRules
			return m, m.onPageChange()

		case key.Matches(msg, keys.Page5):
			m.currentPage = components.PageSettings
			return m, nil

		case key.Matches(msg, keys.Refresh):
			return m, m.refreshCurrentPage()
		}

		// 分发到页面子状态
		return m.dispatchKeyToPage(msg)

	// ── 数据消息：分发到子状态 ──

	case groupsMsg:
		m.nodesState = m.nodesState.ApplyGroups(msg.groups, msg.orderedNames)

	case proxiesMsg:
		m.nodesState = m.nodesState.ApplyProxies(msg)

	case connectionsMsg:
		m.connsState = m.connsState.ApplyConnections(msg)
		if msg != nil && m.chartData != nil {
			m.chartData.AddConnCountData(len(msg.Connections))
		}

	case memoryMsg:
		if m.chartData != nil {
			m.chartData.AddMemoryData(msg.memory)
		}
		if m.wsMsgChan != nil {
			return m, listenWSMessages(m.wsCtx, m.wsMsgChan)
		}

	case trafficWSMsg:
		if m.chartData != nil {
			m.chartData.AddSpeedData(msg.up, msg.down)
		}
		if m.wsMsgChan != nil {
			return m, listenWSMessages(m.wsCtx, m.wsMsgChan)
		}

	case connectionsWSMsg:
		m.connsState = m.connsState.ApplyWSConnections(msg.data)
		if m.chartData != nil {
			m.chartData.AddConnCountData(len(msg.data.Connections))
		}
		if m.wsMsgChan != nil {
			return m, listenWSMessages(m.wsCtx, m.wsMsgChan)
		}

	case logsWSMsg:
		m.logsState = m.logsState.AppendLog(msg.logType, msg.payload)
		if m.wsMsgChan != nil {
			return m, listenWSMessages(m.wsCtx, m.wsMsgChan)
		}

	case rulesMsg:
		m.rulesState = m.rulesState.ApplyRules(msg)

	case siteTestMsg:
		m.connsState = m.connsState.ApplySiteTestResult(msg.name, msg.delay, msg.err)

	case testDoneMsg:
		m.nodesState = m.nodesState.ApplyTestDone(msg.name, msg.delay, msg.err)
		return m, fetchProxies(m.client)

	case testAllDoneMsg:
		m.nodesState = m.nodesState.ApplyTestAllDone(msg.results)
		return m, fetchProxies(m.client)

	case ipInfoMsg:
		if msg.info != nil {
			m.connsState = m.connsState.ApplyIPInfo(msg.info)
		}

	case connectionClosedMsg:
		m.connsState = m.connsState.ApplyConnectionClosed()

	case allConnectionsClosedMsg:
		m.connsState = m.connsState.ApplyAllConnectionsClosed()

	case connTickMsg:
		if m.currentPage == components.PageConnections {
			return m, connTick()
		}

	case logsTickMsg:
		if m.currentPage == components.PageLogs {
			return m, logsTick()
		}

	case errMsg:
		m.err = msg
		m.nodesState.testing = false
		m.nodesState.testPending = 0
	}

	return m, nil
}

// dispatchKeyToPage 将按键分发到当前页面子状态
func (m Model) dispatchKeyToPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch m.currentPage {
	case components.PageNodes:
		m.nodesState, cmd = m.nodesState.Update(msg, m.client, m.proxySvc, m.testURL, m.timeout)

	case components.PageConnections:
		m.connsState, cmd = m.connsState.Update(msg, m.client, m.timeout)

	case components.PageLogs:
		m.logsState, cmd = m.logsState.Update(msg)

	case components.PageRules:
		m.rulesState, cmd = m.rulesState.Update(msg, m.client)

	case components.PageSettings:
		var newCfg, proxyAddr = m.config, ""
		m.settingsState, newCfg, proxyAddr, cmd = m.settingsState.Update(msg, m.config, m.configSvc)
		m.config = newCfg
		if proxyAddr != "" {
			m.connsState = m.connsState.UpdateProxyAddr(proxyAddr)
		}
	}
	return m, cmd
}

// onPageChange 页面切换处理
func (m *Model) onPageChange() tea.Cmd {
	m.err = nil
	switch m.currentPage {
	case components.PageConnections:
		m.connsState = m.connsState.ResetPrevConnIDs()
		return tea.Batch(fetchConnections(m.client), connTick())
	case components.PageLogs:
		return logsTick()
	case components.PageRules:
		return fetchRules(m.client)
	}
	return nil
}

// refreshCurrentPage 刷新当前页面
func (m *Model) refreshCurrentPage() tea.Cmd {
	switch m.currentPage {
	case components.PageNodes:
		return tea.Batch(fetchGroups(m.client), fetchProxies(m.client))
	case components.PageSettings:
		cfg, _ := m.configSvc.LoadConfig()
		m.config = cfg
	}
	return nil
}

// handleMouseScroll 处理鼠标滚轮滚动
func (m Model) handleMouseScroll(up bool, x int) (tea.Model, tea.Cmd) {
	if x >= 0 && x < components.SidebarWidth {
		if up {
			m.currentPage = (m.currentPage + components.PageCount - 1) % components.PageCount
		} else {
			m.currentPage = (m.currentPage + 1) % components.PageCount
		}
		return m, m.onPageChange()
	}

	switch m.currentPage {
	case components.PageNodes:
		m.nodesState = m.nodesState.HandleMouseScroll(up)
	case components.PageConnections:
		m.connsState = m.connsState.HandleMouseScroll(up)
	case components.PageLogs:
		m.logsState = m.logsState.HandleMouseScroll(up)
	case components.PageRules:
		m.rulesState = m.rulesState.HandleMouseScroll(up)
	case components.PageSettings:
		m.settingsState = m.settingsState.HandleMouseScroll(up)
	}

	return m, nil
}

