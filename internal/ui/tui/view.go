package tui

import (
	"github.com/aimony/mihosh/internal/ui/styles"
	"github.com/aimony/mihosh/internal/ui/tui/components"
	"github.com/aimony/mihosh/internal/ui/tui/components/common"
	"github.com/charmbracelet/lipgloss"
)

// View 渲染视图
func (m Model) View() string {
	if m.width == 0 {
		return "正在初始化..."
	}

	// 帮助弹窗处理
	if m.showHelp {
		helpView := m.renderHelpPage()
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpView)
	}

	// ── 布局参数 ──
	sidebarRenderedWidth := components.SidebarWidth + 1 // 含右边框 │
	statusBarHeight := common.StatusBarHeight           // 分隔线 + 信息行
	contentHeight := m.height - statusBarHeight
	if contentHeight < common.MinContentHeight {
		contentHeight = common.MinContentHeight
	}
	mainWidth := m.width - sidebarRenderedWidth
	if mainWidth < common.MinMainWidth {
		mainWidth = common.MinMainWidth
	}

	// ── 侧边栏 ──
	sidebar := components.RenderSidebar(m.currentPage, contentHeight)

	// ── 渲染当前页面内容 ──
	var pageContent string
	switch m.currentPage {
	case components.PageNodes:
		pageContent = m.renderNodesPage()
	case components.PageConnections:
		pageContent = m.renderConnectionsPage()
	case components.PageSettings:
		pageContent = m.renderSettingsPage()
	case components.PageLogs:
		pageContent = m.renderLogsPage()
	case components.PageRules:
		pageContent = m.renderRulesPage()
	}

	// ── 主面板 ──
	pageTitle := components.GetPageTitle(m.currentPage)

	titleStyle := lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Bold(true).
		Padding(0, 1)

	mainPaneStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(styles.ColorBorder).
		Width(mainWidth - 2).
		Height(contentHeight - 2)

	mainContent := lipgloss.JoinVertical(lipgloss.Left,
		titleStyle.Render(pageTitle),
		"",
		pageContent,
	)

	mainPane := mainPaneStyle.Render(mainContent)

	// ── 横向拼接：侧边栏 | 主面板 ──
	upper := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, mainPane)

	// ── 底部状态栏 ──
	var uploadTotal, downloadTotal int64
	if m.connsState.connections != nil {
		uploadTotal = m.connsState.connections.UploadTotal
		downloadTotal = m.connsState.connections.DownloadTotal
	}
	statusBar := components.RenderStatusBar(
		m.width,
		m.err,
		m.nodesState.testing,
		m.nodesState.testingTarget,
		m.chartData,
		uploadTotal,
		downloadTotal,
	)

	return lipgloss.JoinVertical(lipgloss.Left, upper, statusBar)
}
