package pages

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// RenderHelpPage 渲染帮助页面
func RenderHelpPage() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700"))

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00BFFF"))

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC"))

	helpContent := []string{
		titleStyle.Render("Mihosh 使用帮助"),
		"",
		"",
		sectionStyle.Render("全局快捷键:"),
		"  " + keyStyle.Render("Tab      ") + " - " + descStyle.Render("下一页"),
		"  " + keyStyle.Render("Shift+Tab") + " - " + descStyle.Render("上一页"),
		"  " + keyStyle.Render("1-4      ") + " - " + descStyle.Render("快速跳转页面"),
		"  " + keyStyle.Render("r        ") + " - " + descStyle.Render("刷新当前页面"),
		"  " + keyStyle.Render("q        ") + " - " + descStyle.Render("退出程序"),
		"",
		"",
		sectionStyle.Render("节点管理页面 [1]:"),
		"  " + keyStyle.Render("↑/↓ 或 k/j") + " - " + descStyle.Render("选择节点"),
		"  " + keyStyle.Render("←/→ 或 h/l") + " - " + descStyle.Render("切换策略组"),
		"  " + keyStyle.Render("Enter     ") + " - " + descStyle.Render("切换到选中节点"),
		"  " + keyStyle.Render("t         ") + " - " + descStyle.Render("测速当前节点"),
		"  " + keyStyle.Render("a         ") + " - " + descStyle.Render("测速当前组所有节点"),
		"",
		"",
		sectionStyle.Render("连接监控页面 [2]:"),
		"  " + keyStyle.Render("r         ") + " - " + descStyle.Render("刷新连接列表"),
		"",
		"",
		sectionStyle.Render("设置页面 [3]:"),
		"  " + keyStyle.Render("↑/↓      ") + " - " + descStyle.Render("选择配置项"),
		"  " + keyStyle.Render("Enter     ") + " - " + descStyle.Render("编辑配置项"),
		"  " + keyStyle.Render("Esc       ") + " - " + descStyle.Render("取消编辑"),
		"",
		"",
		sectionStyle.Render("延迟颜色说明:"),
		"  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render("●") + " " + descStyle.Render("绿色 - 小于200ms"),
		"  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render("●") + " " + descStyle.Render("黄色 - 200-500ms"),
		"  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render("●") + " " + descStyle.Render("红色 - 大于500ms"),
		"",
		"",
		descStyle.Render("💡 提示: 所有命令行功能都可以在这个TUI界面中完成！"),
	}

	return strings.Join(helpContent, "\n")
}
