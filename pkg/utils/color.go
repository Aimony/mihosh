package utils

import "github.com/charmbracelet/lipgloss"

// GetDelayColor 根据延迟获取颜色
func GetDelayColor(delay int) lipgloss.Color {
	switch {
	case delay == 0:
		return lipgloss.Color("240") // 灰色 - 未测试
	case delay < 100:
		return lipgloss.Color("10") // 绿色 - 优秀
	case delay < 200:
		return lipgloss.Color("11") // 黄色 - 良好
	case delay < 500:
		return lipgloss.Color("208") // 橙色 - 一般
	default:
		return lipgloss.Color("9") // 红色 - 较差
	}
}
