package tui

import (
	"fmt"

	"github.com/aimony/mihosh/internal/ui/tui/components"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

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
		return fetchConnections(m.client)
	}
	return nil
}

// refreshCurrentPage 刷新当前页面
func (m Model) refreshCurrentPage() tea.Cmd {
	switch m.currentPage {
	case components.PageNodes:
		return tea.Batch(fetchGroups(m.client), fetchProxies(m.client))
	case components.PageConnections:
		return fetchConnections(m.client)
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
