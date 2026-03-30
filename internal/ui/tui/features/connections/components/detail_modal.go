package components

import (
	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/ui/tui/components/common"
	"github.com/charmbracelet/lipgloss"
)

// RenderConnectionDetailModal 渲染连接详情模态弹窗
func RenderConnectionDetailModal(
	conn *model.Connection,
	ipInfo *model.IPInfo,
	width, height, leftScroll, rightScroll, focusPanel int,
) string {
	// 模态框边框和标题样式
	modalStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(common.CSecondary).
		Padding(1, 2)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(common.CWarning)

	host := firstNonEmpty(conn.Metadata.Host, conn.Metadata.SniffHost, conn.Metadata.DestinationIP, "-")
	title := titleStyle.Render("🔗 连接详情 - " + host)

	// 计算内部可用尺寸
	innerW := width - 6  // 减去模态框边框(2)和内边距(4)
	innerH := height - 6 // 减去标题行(1)、提示行(1)、模态框边框(2)和内边距(2)

	if innerW < 40 {
		innerW = 40 // 最小宽度保证
	}
	if innerH < 15 {
		innerH = 15 // 最小高度保证
	}

	// 布局计算：三列 vs 两列 vs 单列
	var content string

	if innerW >= 100 {
		// 宽屏：左右并排
		leftW := innerW * 4 / 10
		rightW := innerW - leftW - 4 // 减去列间距

		leftPanel := lipgloss.NewStyle().
			Width(leftW).
			Render(RenderDetailModalLeft(conn, ipInfo, leftW, innerH, leftScroll, focusPanel == 0))

		rightPanel := lipgloss.NewStyle().
			Width(rightW).
			Render(RenderDetailModalRight(conn, rightW, innerH, rightScroll, focusPanel == 1))

		content = lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "    ", rightPanel)

	} else {
		// 窄屏：上下堆叠，焦点所在的面板占据主要空间
		// 动态分配高度
		var topH, bottomH int
		if focusPanel == 0 {
			topH = innerH * 2 / 3
			bottomH = innerH - topH - 2
		} else {
			bottomH = innerH * 2 / 3
			topH = innerH - bottomH - 2
		}

		leftPanel := lipgloss.NewStyle().
			Width(innerW).
			Render(RenderDetailModalLeft(conn, ipInfo, innerW, topH, leftScroll, focusPanel == 0))

		rightPanel := lipgloss.NewStyle().
			Width(innerW).
			Render(RenderDetailModalRight(conn, innerW, bottomH, rightScroll, focusPanel == 1))

		content = lipgloss.JoinVertical(lipgloss.Left, leftPanel, "", rightPanel)
	}

	// 组装模态框
	helpText := common.DimStyle.Render("[←/→/h/l] 切换焦点  [↑/↓/k/j] 滚动  [q/Esc/Enter] 关闭")

	modalContent := lipgloss.JoinVertical(lipgloss.Left,
		title,
		"",
		content,
	)

	// 居中渲染模态框
	modal := modalStyle.Render(modalContent)

	// 在整个屏幕中居中显示
	centeredModal := lipgloss.Place(
		width,
		height-2, // 为底部帮助留出空间
		lipgloss.Center,
		lipgloss.Center,
		modal,
	)

	return lipgloss.JoinVertical(lipgloss.Left, centeredModal, helpText)
}
