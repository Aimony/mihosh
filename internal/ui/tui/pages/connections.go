package pages

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/ui/styles"
	"github.com/aimony/mihosh/internal/ui/tui/components"
	"github.com/aimony/mihosh/pkg/utils"
	"github.com/charmbracelet/lipgloss"
)

// ConnectionsPageState 连接页面状态
type ConnectionsPageState struct {
	Connections        *model.ConnectionsResponse
	Width              int
	Height             int
	SelectedIndex      int
	ScrollTop          int
	FilterText         string
	FilterMode         bool
	DetailMode         bool              // 是否显示详情
	SelectedConnection *model.Connection // 选中的连接
	IPInfo             *model.IPInfo     // 目标IP地理信息
	DetailScroll       int               // 详情页面滚动偏移
	// 图表数据
	ChartData *model.ChartData
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

	// 详情模式：渲染连接详情
	if state.DetailMode && state.SelectedConnection != nil {
		return renderConnectionDetail(state.SelectedConnection, state.IPInfo, state.Width, state.Height, state.DetailScroll)
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

	// 渲染监控图表区域
	chartsSection := renderChartsSection(state, state.Width)
	if chartsSection != "" {
		content = append(content, chartsSection)
		content = append(content, "")
	}

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

// renderConnectionDetail 渲染连接详情（JSON格式，支持滚动）
func renderConnectionDetail(conn *model.Connection, ipInfo *model.IPInfo, width, height, scrollTop int) string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(styles.ColorPrimary)

	dimStyle := lipgloss.NewStyle().
		Foreground(styles.ColorSecondary)

	jsonStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))

	ipInfoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00BFFF"))

	// 将连接信息格式化为JSON
	jsonBytes, err := json.MarshalIndent(conn, "", "  ")
	if err != nil {
		return fmt.Sprintf("无法解析连接信息: %v", err)
	}

	// 构建完整内容
	var allLines []string
	allLines = append(allLines, headerStyle.Render("连接详情"))
	allLines = append(allLines, "")

	// 基本信息摘要
	host := conn.Metadata.Host
	if host == "" {
		host = conn.Metadata.DestinationIP
	}
	allLines = append(allLines, fmt.Sprintf("主机: %s", headerStyle.Render(host)))
	allLines = append(allLines, fmt.Sprintf("规则: %s → %s", conn.Rule, strings.Join(conn.Chains, " → ")))
	allLines = append(allLines, fmt.Sprintf("流量: ↓%s  ↑%s", utils.FormatBytes(conn.Download), utils.FormatBytes(conn.Upload)))
	allLines = append(allLines, "")
	allLines = append(allLines, dimStyle.Render("─── JSON 详情 ───"))
	allLines = append(allLines, "")

	// 添加JSON内容（每行分开）
	jsonLines := strings.Split(jsonStyle.Render(string(jsonBytes)), "\n")
	allLines = append(allLines, jsonLines...)
	allLines = append(allLines, "")

	// 添加IP地理信息
	ipInfoLines := strings.Split(renderIPInfoSection(ipInfo, ipInfoStyle, dimStyle), "\n")
	allLines = append(allLines, ipInfoLines...)
	allLines = append(allLines, "")

	// 计算可显示的行数
	maxDisplay := height - 4 // 留出空间给帮助提示和边框
	if maxDisplay < 10 {
		maxDisplay = 10
	}

	totalLines := len(allLines)

	// 限制滚动范围
	if scrollTop > totalLines-maxDisplay {
		scrollTop = totalLines - maxDisplay
	}
	if scrollTop < 0 {
		scrollTop = 0
	}

	// 截取显示内容
	endIdx := scrollTop + maxDisplay
	if endIdx > totalLines {
		endIdx = totalLines
	}

	visibleLines := allLines[scrollTop:endIdx]

	// 构建输出
	var output []string

	// 滚动提示（上方）
	if scrollTop > 0 {
		output = append(output, dimStyle.Render(fmt.Sprintf("↑ 还有 %d 行", scrollTop)))
	}

	output = append(output, visibleLines...)

	// 滚动提示（下方）
	if endIdx < totalLines {
		output = append(output, dimStyle.Render(fmt.Sprintf("↓ 还有 %d 行", totalLines-endIdx)))
	}

	// 帮助提示
	output = append(output, "")
	output = append(output, dimStyle.Render("[↑↓] 滚动 [Esc/Enter] 返回列表"))

	return strings.Join(output, "\n")
}

// renderIPInfoSection 渲染IP地理信息部分
func renderIPInfoSection(ipInfo *model.IPInfo, infoStyle, dimStyle lipgloss.Style) string {
	var lines []string
	lines = append(lines, dimStyle.Render("─── 目标 IP 地理信息 ───"))
	lines = append(lines, "")

	if ipInfo == nil {
		lines = append(lines, dimStyle.Render("正在加载 IP 信息..."))
		return strings.Join(lines, "\n")
	}

	// 主要信息行：IP (ASN)
	ipLine := fmt.Sprintf("⊕ %s ( AS%d )", ipInfo.IP, ipInfo.ASN)
	lines = append(lines, infoStyle.Render(ipLine))

	// 地区和ISP信息行
	locationLine := fmt.Sprintf("⊕ %s  ☐ %s", ipInfo.Country, ipInfo.ISP)
	lines = append(lines, infoStyle.Render(locationLine))

	// 详细信息（可选显示）
	if ipInfo.Organization != "" && ipInfo.Organization != ipInfo.ISP {
		lines = append(lines, dimStyle.Render(fmt.Sprintf("  组织: %s", ipInfo.Organization)))
	}
	if ipInfo.ASNOrganization != "" && ipInfo.ASNOrganization != ipInfo.ISP {
		lines = append(lines, dimStyle.Render(fmt.Sprintf("  ASN组织: %s", ipInfo.ASNOrganization)))
	}
	if ipInfo.Timezone != "" {
		lines = append(lines, dimStyle.Render(fmt.Sprintf("  时区: %s", ipInfo.Timezone)))
	}
	if ipInfo.Latitude != 0 || ipInfo.Longitude != 0 {
		lines = append(lines, dimStyle.Render(fmt.Sprintf("  坐标: %.3f, %.3f", ipInfo.Latitude, ipInfo.Longitude)))
	}

	return strings.Join(lines, "\n")
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

// renderChartsSection 渲染监控图表区域
func renderChartsSection(state ConnectionsPageState, width int) string {
	if state.ChartData == nil {
		return ""
	}

	// 计算每个图表的宽度（三个并排显示）
	chartWidth := (width - 8) / 3
	if chartWidth < 25 {
		chartWidth = 25
	}
	if chartWidth > 45 {
		chartWidth = 45
	}

	// 速度图表配置
	speedConfig := components.SparklineConfig{
		Title:      "上传/下载速度",
		Width:      chartWidth,
		Height:     4,
		Color1:     lipgloss.Color("#00BFFF"), // 蓝色 - 上传
		Color2:     lipgloss.Color("#9370DB"), // 紫色 - 下载
		Label1:     "上传速度",
		Label2:     "下载速度",
		MinValue:   0, // Y轴完全自适应
		ShowXAxis:  true,
		MaxSeconds: 60,
		FormatFunc: func(v int64) string {
			return formatSpeed(v)
		},
	}

	// 内存图表配置
	memoryConfig := components.SparklineConfig{
		Title:      "内存使用",
		Width:      chartWidth,
		Height:     4,
		Color1:     lipgloss.Color("#00FF7F"), // 绿色
		Label1:     "内存使用",
		MinValue:   0, // Y轴完全自适应
		ShowXAxis:  true,
		MaxSeconds: 60,
		FormatFunc: func(v int64) string {
			return formatMemory(v)
		},
	}

	// 连接数图表配置
	connConfig := components.SparklineConfig{
		Title:      "连接",
		Width:      chartWidth,
		Height:     4,
		Color1:     lipgloss.Color("#FFD700"), // 金色
		Label1:     "连接",
		MinValue:   0, // 连接数Y轴自适应
		ShowXAxis:  true,
		MaxSeconds: 60,
		FormatFunc: func(v int64) string {
			return fmt.Sprintf("%d", v)
		},
	}

	// 渲染三个图表
	speedChart := components.RenderDualSparkline(
		state.ChartData.SpeedUpHistory,
		state.ChartData.SpeedDownHistory,
		speedConfig,
	)
	memoryChart := components.RenderSparkline(state.ChartData.MemoryHistory, memoryConfig)
	connChart := components.RenderIntSparkline(state.ChartData.ConnCountHistory, connConfig)

	// 横向拼接三个图表
	return lipgloss.JoinHorizontal(lipgloss.Top, speedChart, "  ", memoryChart, "  ", connChart)
}

// formatSpeed 格式化速度
func formatSpeed(bytesPerSec int64) string {
	if bytesPerSec < 1024 {
		return fmt.Sprintf("%d B/s", bytesPerSec)
	} else if bytesPerSec < 1024*1024 {
		return fmt.Sprintf("%.1f KB/s", float64(bytesPerSec)/1024)
	} else {
		return fmt.Sprintf("%.1f MB/s", float64(bytesPerSec)/(1024*1024))
	}
}

// formatMemory 格式化内存
func formatMemory(bytes int64) string {
	if bytes == 0 {
		return "0 B"
	} else if bytes < 1024*1024 {
		return fmt.Sprintf("%.0f KB", float64(bytes)/1024)
	} else if bytes < 1024*1024*1024 {
		return fmt.Sprintf("%.0f MB", float64(bytes)/(1024*1024))
	} else {
		return fmt.Sprintf("%.1f GB", float64(bytes)/(1024*1024*1024))
	}
}
