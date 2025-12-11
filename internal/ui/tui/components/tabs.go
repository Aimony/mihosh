package components

import (
	"strings"

	"github.com/aimony/mihosh/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// PageType 页面类型
type PageType int

const (
	PageNodes PageType = iota
	PageConnections
	PageSettings
	PageLogs
	PageRules
	PageHelp
)

// RenderTabs 渲染标签栏
func RenderTabs(currentPage PageType, width int) string {
	tabs := []string{
		"[1] 节点管理",
		"[2] 连接监控",
		"[3] 设置",
		"[4] 日志",
		"[5] 规则",
		"[6] 帮助",
	}

	var rendered []string
	for i, tab := range tabs {
		if PageType(i) == currentPage {
			rendered = append(rendered, styles.ActiveTabStyle.Render("● "+tab))
		} else {
			rendered = append(rendered, styles.InactiveTabStyle.Render("  "+tab))
		}
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)

	divider := styles.DividerStyle.
		Render(strings.Repeat("─", width))

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, divider)
}
