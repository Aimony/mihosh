package pages

import (
	"github.com/charmbracelet/lipgloss"
)

// RenderHelpPage 渲染帮助页面（支持宽度自适应）
func RenderHelpPage(width int) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700"))

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00BFFF")).
		MarginBottom(1)

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Width(12)

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC"))

	// 帮助卡片样式
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#444444")).
		Padding(1, 2).
		MarginRight(2)

	// 渲染键值对
	renderKey := func(key, desc string) string {
		return keyStyle.Render(key) + descStyle.Render(desc)
	}

	// 全局快捷键卡片
	globalKeys := lipgloss.JoinVertical(lipgloss.Left,
		sectionStyle.Render("🌐 全局快捷键"),
		renderKey("1-6", "快速跳转页面"),
		renderKey("Tab", "下一页"),
		renderKey("Shift+Tab", "上一页"),
		renderKey("r", "刷新当前页面"),
		renderKey("q", "退出程序"),
	)

	// 节点管理卡片
	nodesKeys := lipgloss.JoinVertical(lipgloss.Left,
		sectionStyle.Render("📡 节点管理 [1]"),
		renderKey("↑/↓ k/j", "选择节点"),
		renderKey("←/→ h/l", "切换策略组"),
		renderKey("Enter", "切换到选中节点"),
		renderKey("t", "测速当前节点"),
		renderKey("a", "测速当前组所有节点"),
	)

	// 连接监控卡片
	connKeys := lipgloss.JoinVertical(lipgloss.Left,
		sectionStyle.Render("🔗 连接监控 [2]"),
		renderKey("↑/↓ k/j", "选择连接"),
		renderKey("Enter", "查看连接详情"),
		renderKey("x", "关闭选中连接"),
		renderKey("X", "关闭所有连接"),
		renderKey("/", "搜索过滤"),
		renderKey("Esc", "清除过滤/返回"),
		renderKey("Tab", "切换活跃/历史"),
	)

	// 日志页面卡片
	logsKeys := lipgloss.JoinVertical(lipgloss.Left,
		sectionStyle.Render("📜 日志 [3]"),
		renderKey("↑/↓ k/j", "选择日志"),
		renderKey("Tab", "切换日志级别"),
		renderKey("/", "搜索过滤"),
		renderKey("c", "清空日志"),
		renderKey("Esc", "清除搜索"),
	)

	// 规则页面卡片
	rulesKeys := lipgloss.JoinVertical(lipgloss.Left,
		sectionStyle.Render("📋 规则 [4]"),
		renderKey("↑/↓ k/j", "选择规则"),
		renderKey("/", "搜索过滤"),
		renderKey("Esc", "清除搜索"),
	)

	// 设置页面卡片
	settingsKeys := lipgloss.JoinVertical(lipgloss.Left,
		sectionStyle.Render("⚙️  设置 [6]"),
		renderKey("↑/↓", "选择配置项"),
		renderKey("Enter", "编辑配置项"),
		renderKey("Esc", "取消编辑"),
	)

	// 延迟颜色说明卡片
	latencyInfo := lipgloss.JoinVertical(lipgloss.Left,
		sectionStyle.Render("🎨 延迟颜色说明"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render("●")+" "+descStyle.Render("绿色 - 小于200ms"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render("●")+" "+descStyle.Render("黄色 - 200-500ms"),
		lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render("●")+" "+descStyle.Render("红色 - 大于500ms"),
	)

	// 应用卡片样式
	globalCard := cardStyle.Render(globalKeys)
	nodesCard := cardStyle.Render(nodesKeys)
	connCard := cardStyle.Render(connKeys)
	logsCard := cardStyle.Render(logsKeys)
	rulesCard := cardStyle.Render(rulesKeys)
	settingsCard := cardStyle.Render(settingsKeys)
	latencyCard := cardStyle.Render(latencyInfo)

	// 根据宽度决定布局
	var content string
	if width >= 100 {
		// 宽屏：三列布局
		col1 := lipgloss.JoinVertical(lipgloss.Left, globalCard, latencyCard)
		col2 := lipgloss.JoinVertical(lipgloss.Left, nodesCard, logsCard)
		col3 := lipgloss.JoinVertical(lipgloss.Left, connCard, rulesCard, settingsCard)
		content = lipgloss.JoinHorizontal(lipgloss.Top, col1, col2, col3)
	} else if width >= 70 {
		// 中等宽度：两列布局
		col1 := lipgloss.JoinVertical(lipgloss.Left, globalCard, nodesCard, logsCard)
		col2 := lipgloss.JoinVertical(lipgloss.Left, connCard, rulesCard, settingsCard, latencyCard)
		content = lipgloss.JoinHorizontal(lipgloss.Top, col1, col2)
	} else {
		// 窄屏：单列布局
		content = lipgloss.JoinVertical(lipgloss.Left,
			globalCard, nodesCard, connCard, logsCard, rulesCard, settingsCard, latencyCard,
		)
	}

	// 标题
	title := titleStyle.Render("Mihosh 使用帮助")

	// 提示
	tip := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		MarginTop(1).
		Render("💡 提示: 所有命令行功能都可以在这个TUI界面中完成！")

	return lipgloss.JoinVertical(lipgloss.Left, title, "", content, tip)
}
