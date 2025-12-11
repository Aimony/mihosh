package tui

import (
	"fmt"
	"time"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/infrastructure/api"
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
	)
}

// Update 更新
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// 编辑模式特殊处理
		if m.editMode {
			return m.handleEditMode(msg)
		}

		// 日志过滤模式特殊处理
		if m.logFilterMode {
			return m.handleLogFilterMode(msg)
		}

		// 规则过滤模式特殊处理
		if m.ruleFilterMode {
			return m.handleRuleFilterMode(msg)
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
			return m, fetchConnections(m.client)

		case key.Matches(msg, keys.Page3):
			m.currentPage = components.PageLogs
			return m, m.onPageChange()

		case key.Matches(msg, keys.Page4):
			m.currentPage = components.PageRules
			return m, fetchRules(m.client)

		case key.Matches(msg, keys.Page5):
			m.currentPage = components.PageHelp
			return m, m.onPageChange()

		case key.Matches(msg, keys.Page6):
			m.currentPage = components.PageSettings
			return m, nil

		case key.Matches(msg, keys.Refresh):
			return m, m.refreshCurrentPage()
		}

		// 页面特定快捷键
		switch m.currentPage {
		case components.PageNodes:
			return m.updateNodesPage(msg)
		case components.PageConnections:
			return m.updateConnectionsPage(msg)
		case components.PageSettings:
			return m.updateSettingsPage(msg)
		case components.PageHelp:
			return m.updateHelpPage(msg)
		case components.PageLogs:
			return m.updateLogsPage(msg)
		case components.PageRules:
			return m.updateRulesPage(msg)
		}

	case groupsMsg:
		// 保存当前选中的策略组名称
		var selectedGroupName string
		if len(m.groupNames) > 0 && m.selectedGroup < len(m.groupNames) {
			selectedGroupName = m.groupNames[m.selectedGroup]
		}

		m.groups = msg
		m.groupNames = make([]string, 0, len(msg))
		for name := range msg {
			m.groupNames = append(m.groupNames, name)
		}

		// 恢复之前选中的策略组
		if selectedGroupName != "" {
			for i, name := range m.groupNames {
				if name == selectedGroupName {
					m.selectedGroup = i
					break
				}
			}
		}

		m.updateCurrentProxies()

	case proxiesMsg:
		m.proxies = msg

	case connectionsMsg:
		m.connections = msg
		// 更新图表数据 - 连接数
		if msg != nil && m.chartData != nil {
			m.chartData.AddConnCountData(len(msg.Connections))
		}

	case memoryMsg:
		// 更新内存历史数据
		if m.chartData != nil {
			m.chartData.AddMemoryData(msg.memory)
		}
		// 继续监听WebSocket消息
		if m.currentPage == components.PageConnections && m.wsMsgChan != nil {
			return m, listenWSMessages(m.wsMsgChan)
		}

	case trafficWSMsg:
		// 更新速度图表数据
		if m.chartData != nil {
			m.chartData.AddSpeedData(msg.up, msg.down)
		}
		// 继续监听WebSocket消息
		if m.currentPage == components.PageConnections && m.wsMsgChan != nil {
			return m, listenWSMessages(m.wsMsgChan)
		}

	case connectionsWSMsg:
		// 检测已关闭的连接
		currentConnIDs := make(map[string]model.Connection)
		for _, conn := range msg.data.Connections {
			currentConnIDs[conn.ID] = model.Connection{
				ID:            conn.ID,
				Upload:        conn.Upload,
				Download:      conn.Download,
				Start:         conn.Start,
				Chains:        conn.Chains,
				Rule:          conn.Rule,
				RulePayload:   conn.RulePayload,
				DownloadSpeed: conn.DownloadSpeed,
				UploadSpeed:   conn.UploadSpeed,
				Metadata: model.Metadata{
					Network:         conn.Metadata.Network,
					Type:            conn.Metadata.Type,
					SourceIP:        conn.Metadata.SourceIP,
					DestinationIP:   conn.Metadata.DestinationIP,
					SourcePort:      conn.Metadata.SourcePort,
					DestinationPort: conn.Metadata.DestinationPort,
					Host:            conn.Metadata.Host,
					Process:         conn.Metadata.Process,
					ProcessPath:     conn.Metadata.ProcessPath,
				},
			}
		}

		// 找出已关闭的连接（在上次存在但这次不存在）
		if m.prevConnIDs != nil {
			for id, conn := range m.prevConnIDs {
				if _, exists := currentConnIDs[id]; !exists {
					// 这个连接已关闭，加入历史记录
					m.closedConnections = append([]model.Connection{conn}, m.closedConnections...)
					// 限制历史记录最多1000条
					if len(m.closedConnections) > 1000 {
						m.closedConnections = m.closedConnections[:1000]
					}
				}
			}
		}

		// 更新上次连接ID映射
		m.prevConnIDs = currentConnIDs

		// 更新连接数据（转换为model.ConnectionsResponse）
		m.connections = convertToConnectionsResponse(msg.data)
		// 更新连接数图表
		if m.chartData != nil {
			m.chartData.AddConnCountData(len(msg.data.Connections))
		}
		// 继续监听WebSocket消息
		if m.currentPage == components.PageConnections && m.wsMsgChan != nil {
			return m, listenWSMessages(m.wsMsgChan)
		}

	case connTickMsg:
		// 定时器触发：仅继续定时器（连接数据由WebSocket推送）
		if m.currentPage == components.PageConnections {
			return m, connTick()
		}

	case logsTickMsg:
		// 日志页面定时器触发：继续定时器并确保监听
		if m.currentPage == components.PageLogs && m.wsMsgChan != nil {
			return m, tea.Batch(logsTick(), listenWSMessages(m.wsMsgChan))
		}

	case testDoneMsg:
		m.testing = false

		// 记录测速失败的节点
		if msg.err != nil {
			failureInfo := fmt.Sprintf("%s: %s", msg.name, msg.err.Error())
			m.testFailures = append(m.testFailures, failureInfo)
		}

		return m, fetchProxies(m.client)

	case configSavedMsg:
		m.editMode = false
		m.err = nil

	case connectionClosedMsg:
		// 连接关闭后调整选择索引
		if m.selectedConn > 0 {
			m.selectedConn--
		}

	case allConnectionsClosedMsg:
		// 所有连接关闭后重置状态
		m.selectedConn = 0
		m.connScrollTop = 0

	case ipInfoMsg:
		// 保存IP地理信息
		if msg.info != nil {
			m.connIPInfo = msg.info
		}

	case logsWSMsg:
		// 添加日志到列表
		newLog := model.LogEntry{
			Type:      msg.logType,
			Payload:   msg.payload,
			Timestamp: time.Now(),
		}
		m.logs = append([]model.LogEntry{newLog}, m.logs...)
		// 限制日志最多1000条
		if len(m.logs) > 1000 {
			m.logs = m.logs[:1000]
		}
		// 继续监听WebSocket消息
		if m.wsMsgChan != nil && (m.currentPage == components.PageConnections || m.currentPage == components.PageLogs) {
			return m, listenWSMessages(m.wsMsgChan)
		}

	case rulesMsg:
		// 更新规则列表
		m.rules = msg

	case errMsg:
		m.err = msg
		m.testing = false
	}

	return m, nil
}

// onPageChange 页面切换时的处理
func (m Model) onPageChange() tea.Cmd {
	m.err = nil
	switch m.currentPage {
	case components.PageConnections:
		// 进入连接页面时启动WebSocket流
		return tea.Batch(
			startWSStreams(m.wsClient, m.wsMsgChan),
			listenWSMessages(m.wsMsgChan),
			connTick(),
		)
	case components.PageLogs:
		// 进入日志页面时启动WebSocket流和定时器
		return tea.Batch(
			startWSStreams(m.wsClient, m.wsMsgChan),
			listenWSMessages(m.wsMsgChan),
			logsTick(),
		)
	case components.PageRules:
		// 进入规则页面时获取规则列表
		return fetchRules(m.client)
	default:
		// 离开连接/日志页面时停止WebSocket流
		if m.wsClient != nil && m.wsClient.IsRunning() {
			return stopWSStreams(m.wsClient)
		}
	}
	return nil
}

// refreshCurrentPage 刷新当前页面
func (m Model) refreshCurrentPage() tea.Cmd {
	switch m.currentPage {
	case components.PageNodes:
		return tea.Batch(fetchGroups(m.client), fetchProxies(m.client))
	case components.PageConnections:
		// connections数据由WebSocket推送，无需手动刷新
		return nil
	case components.PageSettings:
		cfg, _ := m.configSvc.LoadConfig()
		m.config = cfg
		return nil
	}
	return nil
}

// updateCurrentProxies 更新当前显示的代理列表
func (m *Model) updateCurrentProxies() {
	if len(m.groupNames) > 0 && m.selectedGroup < len(m.groupNames) {
		groupName := m.groupNames[m.selectedGroup]
		if group, ok := m.groups[groupName]; ok {
			m.currentProxies = group.All
		}
	}
}

// convertToConnectionsResponse 将api.ConnectionsData转换为model.ConnectionsResponse
func convertToConnectionsResponse(data api.ConnectionsData) *model.ConnectionsResponse {
	connections := make([]model.Connection, len(data.Connections))
	for i, conn := range data.Connections {
		connections[i] = model.Connection{
			ID:            conn.ID,
			Upload:        conn.Upload,
			Download:      conn.Download,
			Start:         conn.Start,
			Chains:        conn.Chains,
			Rule:          conn.Rule,
			RulePayload:   conn.RulePayload,
			DownloadSpeed: conn.DownloadSpeed,
			UploadSpeed:   conn.UploadSpeed,
			Metadata: model.Metadata{
				Network:         conn.Metadata.Network,
				Type:            conn.Metadata.Type,
				SourceIP:        conn.Metadata.SourceIP,
				DestinationIP:   conn.Metadata.DestinationIP,
				SourcePort:      conn.Metadata.SourcePort,
				DestinationPort: conn.Metadata.DestinationPort,
				Host:            conn.Metadata.Host,
				Process:         conn.Metadata.Process,
				ProcessPath:     conn.Metadata.ProcessPath,
			},
		}
	}
	return &model.ConnectionsResponse{
		DownloadTotal: data.DownloadTotal,
		UploadTotal:   data.UploadTotal,
		Connections:   connections,
	}
}
