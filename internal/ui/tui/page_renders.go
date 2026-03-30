package tui

import (
	"github.com/aimony/mihosh/internal/ui/tui/components"
	"github.com/aimony/mihosh/internal/ui/tui/components/common"
	"github.com/aimony/mihosh/internal/ui/tui/pages"
)

// getPageSize 计算页面内容的可用宽度和高度
func (m Model) getPageSize() (pageWidth, pageHeight int) {
	sidebarRenderedWidth := components.SidebarWidth + 1
	mainWidth := m.width - sidebarRenderedWidth
	if mainWidth < common.MinMainWidth {
		mainWidth = common.MinMainWidth
	}
	pageWidth = mainWidth - 2
	if pageWidth < common.MinMainWidth {
		pageWidth = common.MinMainWidth
	}
	pageHeight = m.height - 8
	return pageWidth, pageHeight
}

// renderNodesPage 渲染节点管理页面
func (m Model) renderNodesPage() string {
	pageWidth, pageHeight := m.getPageSize()
	state := m.nodesState.ToPageState(pageWidth, pageHeight)
	return pages.RenderNodesPage(state)
}

// renderConnectionsPage 渲染连接监控页面
func (m Model) renderConnectionsPage() string {
	pageWidth, pageHeight := m.getPageSize()
	state := m.connsState.ToPageState(m.chartData, pageWidth, pageHeight)
	return pages.RenderConnectionsPage(state)
}

// renderSettingsPage 渲染设置页面
func (m Model) renderSettingsPage() string {
	pageWidth, pageHeight := m.getPageSize()
	state := m.settingsState.ToPageState(m.config)
	return pages.RenderSettingsPage(state, pageWidth, pageHeight)
}

// renderHelpPage 渲染帮助页面弹窗
func (m Model) renderHelpPage() string {
	return pages.RenderHelpPage(m.width, m.height)
}

// renderLogsPage 渲染日志页面
func (m Model) renderLogsPage() string {
	pageWidth, pageHeight := m.getPageSize()
	state := m.logsState.ToPageState(pageWidth, pageHeight)
	return pages.RenderLogsPage(state)
}

// renderRulesPage 渲染规则页面
func (m Model) renderRulesPage() string {
	pageWidth, pageHeight := m.getPageSize()
	state := m.rulesState.ToPageState(pageWidth, pageHeight)
	return pages.RenderRulesPage(state)
}
