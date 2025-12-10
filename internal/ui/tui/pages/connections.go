package pages

import (
	"fmt"
	"strings"
	"time"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/ui/styles"
	"github.com/aimony/mihosh/pkg/utils"
	"github.com/charmbracelet/lipgloss"
)

// ConnectionsPageState 连接页面状态
type ConnectionsPageState struct {
	Connections   *model.ConnectionsResponse
	Width         int
	Height        int
	SelectedIndex int
	ScrollTop     int
	FilterText    string
	FilterMode    bool
}

// 表格列宽配置
const (
	colWidthClose = 4
	colWidthHost  = 22
	colWidthType  = 12
	colWidthRule  = 16
	colWidthChain = 16
	colWidthDL    = 10
	colWidthUL    = 10
	colWidthTime  = 8
)

// RenderConnectionsPage 渲染连接监控页面
func RenderConnectionsPage(state ConnectionsPageState) string {
	if state.Connections == nil {
		return "正在加载连接信息..."
	}

	// 样式定义
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorPrimary)

	selectedStyle := lipgloss.NewStyle().
		Foreground(styles.ColorPrimary).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFF"))

	dimStyle := lipgloss.NewStyle().
		Foreground(styles.ColorSecondary)

	// 过滤连接
	filteredConns := filterConnections(state.Connections.Connections, state.FilterText)

	// 统计信息
	stats := fmt.Sprintf(
		"活跃: %s | 上传: %s | 下载: %s",
		headerStyle.Render(fmt.Sprintf("%d", len(filteredConns))),
		headerStyle.Render(utils.FormatBytes(state.Connections.UploadTotal)),
		headerStyle.Render(utils.FormatBytes(state.Connections.DownloadTotal)),
	)

	// 过滤输入框
	filterLine := ""
	if state.FilterMode {
		filterLine = fmt.Sprintf("过滤: %s█", state.FilterText)
	} else if state.FilterText != "" {
		filterLine = dimStyle.Render(fmt.Sprintf("过滤: %s (按/编辑, Esc清除)", state.FilterText))
	}

	// 表头
	tableHeader := renderTableHeader(headerStyle)

	// 计算可显示的行数
	maxDisplay := 15
	if state.Height > 0 {
		maxDisplay = (state.Height - 10) // 减去标题、统计、表头、帮助等行
		if maxDisplay < 5 {
			maxDisplay = 5
		}
	}

	// 连接列表
	var rows []string
	if len(filteredConns) == 0 {
		rows = append(rows, dimStyle.Render("  无活跃连接"))
	} else {
		// 确保选中索引在有效范围内
		selectedIdx := state.SelectedIndex
		if selectedIdx >= len(filteredConns) {
			selectedIdx = len(filteredConns) - 1
		}
		if selectedIdx < 0 {
			selectedIdx = 0
		}

		// 计算滚动范围
		scrollTop := state.ScrollTop
		if selectedIdx >= scrollTop+maxDisplay {
			scrollTop = selectedIdx - maxDisplay + 1
		}
		if selectedIdx < scrollTop {
			scrollTop = selectedIdx
		}

		endIdx := scrollTop + maxDisplay
		if endIdx > len(filteredConns) {
			endIdx = len(filteredConns)
		}

		for i := scrollTop; i < endIdx; i++ {
			conn := filteredConns[i]
			isSelected := i == selectedIdx

			rowStyle := normalStyle
			prefix := "  "
			if isSelected {
				rowStyle = selectedStyle
				prefix = "▶ "
			}

			row := renderConnectionRow(conn, rowStyle, prefix, state.Width)
			rows = append(rows, row)
		}

		// 显示滚动提示
		if scrollTop > 0 {
			rows = append([]string{dimStyle.Render(fmt.Sprintf("  ↑ 还有 %d 条", scrollTop))}, rows...)
		}
		if endIdx < len(filteredConns) {
			rows = append(rows, dimStyle.Render(fmt.Sprintf("  ↓ 还有 %d 条", len(filteredConns)-endIdx)))
		}
	}

	// 帮助提示
	helpText := dimStyle.Render("[↑↓]选择 [x]关闭 [X]全部关闭 [/]搜索 [r]刷新")

	// 组装页面
	var content []string
	content = append(content, headerStyle.Render("连接监控"))
	content = append(content, "")
	content = append(content, stats)
	if filterLine != "" {
		content = append(content, filterLine)
	}
	content = append(content, "")
	content = append(content, tableHeader)
	content = append(content, styles.DividerStyle.Render(strings.Repeat("─", min(state.Width-2, 100))))
	content = append(content, strings.Join(rows, "\n"))
	content = append(content, "")
	content = append(content, helpText)

	return strings.Join(content, "\n")
}

// renderTableHeader 渲染表头
func renderTableHeader(style lipgloss.Style) string {
	header := fmt.Sprintf(
		"%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s",
		colWidthClose, "",
		colWidthHost, "主机",
		colWidthType, "类型",
		colWidthRule, "规则",
		colWidthChain, "代理链",
		colWidthDL, "↓下载",
		colWidthUL, "↑上传",
		colWidthTime, "时长",
	)
	return style.Render(header)
}

// renderConnectionRow 渲染单行连接
func renderConnectionRow(conn model.Connection, style lipgloss.Style, prefix string, width int) string {
	// 主机
	host := conn.Metadata.Host
	if host == "" {
		host = conn.Metadata.DestinationIP
	}
	host = truncateString(host, colWidthHost)

	// 类型
	connType := fmt.Sprintf("%s/%s", conn.Metadata.Network, conn.Metadata.Type)
	connType = truncateString(connType, colWidthType)

	// 规则
	rule := conn.Rule
	if conn.RulePayload != "" && displayWidth(rule)+displayWidth(conn.RulePayload)+1 <= colWidthRule {
		rule = fmt.Sprintf("%s:%s", rule, conn.RulePayload)
	}
	rule = truncateString(rule, colWidthRule)

	// 代理链
	chain := "DIRECT"
	if len(conn.Chains) > 0 {
		chain = conn.Chains[len(conn.Chains)-1]
	}
	chain = truncateString(chain, colWidthChain)

	// 流量
	download := utils.FormatBytes(conn.Download)
	upload := utils.FormatBytes(conn.Upload)

	// 连接时长
	duration := formatDuration(conn.Start)

	row := fmt.Sprintf(
		"%s%s %s %s %s %s %s %s %s",
		prefix,
		padString("×", colWidthClose-2),
		padString(host, colWidthHost),
		padString(connType, colWidthType),
		padString(rule, colWidthRule),
		padString(chain, colWidthChain),
		padString(download, colWidthDL),
		padString(upload, colWidthUL),
		padString(duration, colWidthTime),
	)

	return style.Render(row)
}

// truncateString 根据显示宽度截断字符串（支持中文）
func truncateString(s string, maxWidth int) string {
	if displayWidth(s) <= maxWidth {
		return s
	}
	// 逐字符截断直到符合宽度
	result := ""
	currentWidth := 0
	for _, r := range s {
		var rw int
		if r > 127 {
			rw = 2
		} else {
			rw = 1
		}
		if currentWidth+rw > maxWidth-2 {
			break
		}
		result += string(r)
		currentWidth += rw
	}
	return result + ".."
}

// filterConnections 过滤连接
func filterConnections(connections []model.Connection, filter string) []model.Connection {
	if filter == "" {
		return connections
	}

	filter = strings.ToLower(filter)
	var filtered []model.Connection
	for _, conn := range connections {
		// 搜索主机、规则、代理链
		if strings.Contains(strings.ToLower(conn.Metadata.Host), filter) ||
			strings.Contains(strings.ToLower(conn.Rule), filter) ||
			containsAnyLower(conn.Chains, filter) ||
			strings.Contains(strings.ToLower(conn.Metadata.DestinationIP), filter) {
			filtered = append(filtered, conn)
		}
	}
	return filtered
}

// containsAnyLower 检查字符串切片中是否有包含子串的元素
func containsAnyLower(slice []string, sub string) bool {
	for _, s := range slice {
		if strings.Contains(strings.ToLower(s), sub) {
			return true
		}
	}
	return false
}

// formatDuration 格式化连接时长
func formatDuration(startStr string) string {
	start, err := time.Parse(time.RFC3339, startStr)
	if err != nil {
		return "-"
	}

	duration := time.Since(start)
	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}
	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	}
	return fmt.Sprintf("%dh%dm", int(duration.Hours()), int(duration.Minutes())%60)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
