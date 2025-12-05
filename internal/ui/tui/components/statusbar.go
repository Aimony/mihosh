package components

import (
	"fmt"
	"strings"

	"github.com/aimony/mihosh/internal/ui/styles"
	"github.com/charmbracelet/lipgloss"
)

// RenderStatusBar 渲染状态栏
func RenderStatusBar(width int, err error, testing bool) string {
	var status string
	if err != nil {
		// 优化错误信息显示
		errText := err.Error()

		// 截断过长的错误信息
		maxErrLen := width - 20
		if len(errText) > maxErrLen {
			errText = errText[:maxErrLen] + "..."
		}

		// 提供更友好的错误提示
		friendlyErr := errText
		if strings.Contains(errText, "context dead") {
			friendlyErr = "测速超时，节点可能不可用"
		} else if strings.Contains(errText, "connection refused") {
			friendlyErr = "无法连接mihomo API，请检查mihomo是否运行"
		} else if strings.Contains(errText, "timeout") {
			friendlyErr = "请求超时，请检查网络或增加超时时间"
		}

		status = styles.ErrorStyle.Render(fmt.Sprintf("❌ %s", friendlyErr))
	} else if testing {
		status = styles.TestingStyle.Render("⏳ 测速中...")
	} else {
		// 显示连接状态
		status = styles.StatusStyle.Render("●连接正常")
	}

	// 帮助提示
	helpHint := styles.StatusStyle.Render(" | 按Tab切换页面 | 按数字键快速跳转 | 按r刷新 | 按q退出")

	divider := styles.DividerStyle.
		Render(strings.Repeat("─", width))

	statusBar := status + helpHint

	return lipgloss.JoinVertical(lipgloss.Left, divider, statusBar)
}
