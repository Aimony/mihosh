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
	PageLogs
	PageRules
	PageHelp
	PageSettings
	PageCount // 页面总数，必须放在最后
)

// RenderTabs 渲染标签栏
func RenderTabs(currentPage PageType, width int) string {
	tabs := []string{
		"[1] 节点管理",
		"[2] 连接监控",
		"[3] 日志",
		"[4] 规则",
		"[5] 帮助",
		"[6] 设置",
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
