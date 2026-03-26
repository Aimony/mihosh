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
	PageSettings
	PageCount // 页面总数，必须放在最后
)

// 侧边栏项目
var sidebarItems = []struct {
	Label string
}{
	{"节点"},
	{"连接"},
	{"日志"},
	{"规则"},
	{"设置"},
}

// SidebarWidth 侧边栏渲染宽度（含右边框）
const SidebarWidth = 6

// RenderSidebar 渲染侧边栏
func RenderSidebar(currentPage PageType, height int) string {
	activeStyle := lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Bold(true).
		Width(SidebarWidth).
		Align(lipgloss.Center)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(styles.ColorGray).
		Width(SidebarWidth).
		Align(lipgloss.Center)

	var items []string

	for i, item := range sidebarItems {
		var label string
		if PageType(i) == currentPage {
			label = activeStyle.Render(item.Label)
		} else {
			label = inactiveStyle.Render(item.Label)
		}
		items = append(items, label)
		if i < len(sidebarItems)-1 {
			items = append(items, "")
		}
	}

	content := strings.Join(items, "\n")

	// 用右侧边框分隔
	barStyle := lipgloss.NewStyle().
		Width(SidebarWidth).
		Height(height).
		BorderStyle(lipgloss.Border{Right: "│"}).
		BorderRight(true).
		BorderForeground(styles.ColorBorder)

	return barStyle.Render(content)
}

// GetPageTitle 获取页面标题
func GetPageTitle(page PageType) string {
	titles := []string{"节点管理", "连接监控", "系统日志", "规则列表", "设置"}
	if int(page) < len(titles) {
		return titles[page]
	}
	return ""
}
