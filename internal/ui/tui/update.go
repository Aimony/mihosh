package tui

import (
	"fmt"
	"time"

	"github.com/aimony/mihosh/internal/ui/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// connTickMsg 连接页面定时刷新消息
type connTickMsg time.Time

// connRefreshInterval 连接刷新间隔
const connRefreshInterval = 1 * time.Second

// connTick 创建连接页面定时器
func connTick() tea.Cmd {
	return tea.Tick(connRefreshInterval, func(t time.Time) tea.Msg {
		return connTickMsg(t)
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

		// 全局快捷键
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.NextPage):
			m.currentPage = (m.currentPage + 1) % 4
			return m, m.onPageChange()

		case key.Matches(msg, keys.PrevPage):
			m.currentPage = (m.currentPage + 3) % 4
			return m, m.onPageChange()

		case key.Matches(msg, keys.Page1):
			m.currentPage = components.PageNodes
			return m, m.onPageChange()

		case key.Matches(msg, keys.Page2):
			m.currentPage = components.PageConnections
			return m, fetchConnections(m.client)

		case key.Matches(msg, keys.Page3):
			m.currentPage = components.PageSettings
			return m, nil

		case key.Matches(msg, keys.Page4):
			m.currentPage = components.PageHelp
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
		// 更新图表数据
		if msg != nil && m.chartData != nil {
			// 计算速度（当前总量 - 上次总量）
			uploadSpeed := int64(0)
			downloadSpeed := int64(0)
			if m.lastUpload > 0 {
				uploadSpeed = msg.UploadTotal - m.lastUpload
				if uploadSpeed < 0 {
					uploadSpeed = 0
				}
			}
			if m.lastDownload > 0 {
				downloadSpeed = msg.DownloadTotal - m.lastDownload
				if downloadSpeed < 0 {
					downloadSpeed = 0
				}
			}
			m.lastUpload = msg.UploadTotal
			m.lastDownload = msg.DownloadTotal

			// 添加速度数据
			m.chartData.AddSpeedData(uploadSpeed, downloadSpeed)
			// 添加连接数
			m.chartData.AddConnCountData(len(msg.Connections))
		}
		// 如果当前在连接页面，继续定时刷新
		if m.currentPage == components.PageConnections {
			return m, connTick()
		}

	case memoryMsg:
		// 更新内存历史数据
		if m.chartData != nil {
			m.chartData.AddMemoryData(msg.memory)
		}

	case connTickMsg:
		// 定时器触发：仅在连接页面时刷新
		if m.currentPage == components.PageConnections {
			return m, fetchConnectionsAndMemory(m.client)
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
		// 进入连接页面时启动自动刷新（同时获取连接和内存数据）
		return tea.Batch(fetchConnectionsAndMemory(m.client), connTick())
	}
	return nil
}

// refreshCurrentPage 刷新当前页面
func (m Model) refreshCurrentPage() tea.Cmd {
	switch m.currentPage {
	case components.PageNodes:
		return tea.Batch(fetchGroups(m.client), fetchProxies(m.client))
	case components.PageConnections:
		return fetchConnectionsAndMemory(m.client)
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
