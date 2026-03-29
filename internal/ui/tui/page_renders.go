package tui

import (
	"github.com/aimony/mihosh/internal/ui/tui/components"
	"github.com/aimony/mihosh/internal/ui/tui/components/common"
	"github.com/aimony/mihosh/internal/ui/tui/pages"
)

// renderNodesPage 渲染节点管理页面
func (m Model) renderNodesPage() string {
	sidebarRenderedWidth := components.SidebarWidth + 1
	mainWidth := m.width - sidebarRenderedWidth
	if mainWidth < common.MinMainWidth {
		mainWidth = common.MinMainWidth
	}
	pageWidth := mainWidth - 2
	if pageWidth < common.MinMainWidth {
		pageWidth = common.MinMainWidth
	}

	pageHeight := m.height - 8
	state := m.nodesState.ToPageState(pageWidth, pageHeight)
	return pages.RenderNodesPage(state)
}

// renderConnectionsPage 渲染连接监控页面
func (m Model) renderConnectionsPage() string {
	pageHeight := m.height - 8
	state := m.connsState.ToPageState(m.chartData, m.width, pageHeight)
	return pages.RenderConnectionsPage(state)
}

// renderSettingsPage 渲染设置页面
func (m Model) renderSettingsPage() string {
	pageHeight := m.height - 8
	state := m.settingsState.ToPageState(m.config)
	return pages.RenderSettingsPage(state, m.width, pageHeight)
}

// renderHelpPage 渲染帮助页面弹窗
func (m Model) renderHelpPage() string {
	return pages.RenderHelpPage(m.width, m.height)
}

// renderLogsPage 渲染日志页面
func (m Model) renderLogsPage() string {
	pageHeight := m.height - 8
	state := m.logsState.ToPageState(m.width, pageHeight)
	return pages.RenderLogsPage(state)
}

// renderRulesPage 渲染规则页面
func (m Model) renderRulesPage() string {
	pageHeight := m.height - 8
	state := m.rulesState.ToPageState(m.width, pageHeight)
	return pages.RenderRulesPage(state)
}
