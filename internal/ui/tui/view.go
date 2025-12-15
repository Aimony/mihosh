package tui

import (
	"github.com/aimony/mihosh/internal/ui/tui/components"
	"github.com/charmbracelet/lipgloss"
)

// View 渲染视图
func (m Model) View() string {
	if m.width == 0 {
		return "初始化中..."
	}

	// 帮助弹窗处理
	if m.showHelp {
		helpView := m.renderHelpPage()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpView)
	}

	// 渲染标签栏
	tabs := components.RenderTabs(m.currentPage, m.width)

	// 渲染内容区域
	var content string
	switch m.currentPage {
	case components.PageNodes:
		content = m.renderNodesPage()
	case components.PageConnections:
		content = m.renderConnectionsPage()
	case components.PageSettings:
		content = m.renderSettingsPage()
	case components.PageLogs:
		content = m.renderLogsPage()
	case components.PageRules:
		content = m.renderRulesPage()
	}

	// 渲染状态栏
	statusBar := components.RenderStatusBar(m.width, m.err, m.testing)

	// 组合布局
	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabs,
		"",
		content,
		"",
		statusBar,
	)
}
