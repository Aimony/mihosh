package utils

import "github.com/charmbracelet/lipgloss"

// GetDelayColor 根据延迟获取颜色 (Tokyo Night Deep 语义色)
func GetDelayColor(delay int) lipgloss.Color {
	switch {
	case delay == 0:
		return lipgloss.Color("#565f89") // 灰色 — 未测试
	case delay < 100:
		return lipgloss.Color("#9ECE6A") // 绿色 — 优秀
	case delay < 300:
		return lipgloss.Color("#E0AF68") // 黄色 — 良好
	default:
		return lipgloss.Color("#F7768E") // 红色 — 较差/超时
	}
}
