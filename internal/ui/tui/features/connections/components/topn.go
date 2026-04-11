package components

import (
	"fmt"
	"strings"

	"github.com/aimony/mihosh/internal/ui/tui/components/common"
	"github.com/charmbracelet/lipgloss"
)

// TopNItem 上行+下行流量统计项
type TopNItem struct {
	Name       string
	TotalBytes int64
}

// RenderTopNSection 渲染吞吐量 Top N 排行榜
func RenderTopNSection(items []TopNItem, width int) string {
	if len(items) == 0 {
		return ""
	}

	titleStyle := lipgloss.NewStyle().Foreground(common.CPrimary).Bold(true)
	nameStyle := lipgloss.NewStyle().Foreground(common.CWhite)
	bytesStyle := lipgloss.NewStyle().Foreground(common.CSecondary)
	barColor := lipgloss.Color("#9370DB") // 紫色进度条

	var lines []string
	title := titleStyle.Render("Top 5 吞吐量排行 (过去5分钟)")
	lines = append(lines, title)

	maxBytes := items[0].TotalBytes // 第一项是最大的

	// 预留名字宽度（最长20字符）
	nameWidth := 20
	barsWidth := width - nameWidth - 15 - 4 // 15是数值留宽, 4 是边距
	if barsWidth < 10 {
		barsWidth = 10
	}

	for _, item := range items {
		name := item.Name
		nameRunes := []rune(name)
		if len(nameRunes) > nameWidth {
			name = string(nameRunes[:nameWidth-3]) + "..."
			nameRunes = []rune(name)
		}

		// 格式化名字固定宽度
		padLen := nameWidth - len(nameRunes)
		if padLen < 0 {
			padLen = 0
		}
		nameStr := nameStyle.Render(name + strings.Repeat(" ", padLen))

		// 格式化数值，固定宽度 10
		bytesStr := FormatMemory(item.TotalBytes)
		padBytes := 12 - len(bytesStr)
		if padBytes < 0 {
			padBytes = 0
		}
		bytesStrRendered := bytesStyle.Render(strings.Repeat(" ", padBytes) + bytesStr)

		// 计算进度条
		ratio := float64(0)
		if maxBytes > 0 {
			ratio = float64(item.TotalBytes) / float64(maxBytes)
		}

		barLen := int(ratio * float64(barsWidth))
		if barLen < 1 && item.TotalBytes > 0 {
			barLen = 1
		}

		bar := lipgloss.NewStyle().Foreground(barColor).Render(strings.Repeat("█", barLen))
		emptyBar := lipgloss.NewStyle().Foreground(common.CMuted).Render(strings.Repeat("░", barsWidth-barLen))

		line := fmt.Sprintf("%s │ %s%s %s", nameStr, bar, emptyBar, bytesStrRendered)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}
