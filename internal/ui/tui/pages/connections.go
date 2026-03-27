package pages

import (
	"fmt"
	"strings"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/ui/styles"
	"github.com/aimony/mihosh/internal/ui/tui/components/common"
	"github.com/aimony/mihosh/internal/ui/tui/components/connections"
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
	// 视图模式
	ViewMode          int                // 0=活跃连接, 1=历史连接
	ClosedConnections []model.Connection // 已关闭的连接历史
	// 网站测速
	SiteTests        []model.SiteTest // 网站测试数据
	SelectedSiteTest int              // 选中的网站索引
}

// RenderConnectionsPage 渲染连接监控页面
func RenderConnectionsPage(state ConnectionsPageState) string {
	// 详情模式：渲染连接详情
	if state.DetailMode && state.SelectedConnection != nil {
		return connections.RenderConnectionDetail(state.SelectedConnection, state.IPInfo, state.Height, state.DetailScroll)
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

	// 根据视图模式选择数据源
	var connList []model.Connection
	var viewModeLabel string
	if state.ViewMode == 0 {
		// 活跃连接
		if state.Connections == nil {
			return "正在加载连接信息..."
		}
		connList = state.Connections.Connections
		viewModeLabel = headerStyle.Render("● 活跃连接") + dimStyle.Render("  历史连接")
	} else {
		// 历史连接
		connList = state.ClosedConnections
		viewModeLabel = dimStyle.Render("活跃连接  ") + headerStyle.Render("● 历史连接")
	}

	// 过滤连接
	filteredConns := filterConnections(connList, state.FilterText)

	// 统计信息
	var stats string
	if state.ViewMode == 0 && state.Connections != nil {
		stats = fmt.Sprintf(
			"活跃: %s | 上传: %s | 下载: %s",
			headerStyle.Render(fmt.Sprintf("%d", len(filteredConns))),
			headerStyle.Render(utils.FormatBytes(state.Connections.UploadTotal)),
			headerStyle.Render(utils.FormatBytes(state.Connections.DownloadTotal)),
		)
	} else {
		stats = fmt.Sprintf(
			"历史: %s 条记录",
			headerStyle.Render(fmt.Sprintf("%d", len(filteredConns))),
		)
	}

	// 过滤输入框
	filterLine := ""
	if state.FilterMode {
		filterLine = fmt.Sprintf("过滤: %s█", state.FilterText)
	} else if state.FilterText != "" {
		filterLine = dimStyle.Render(fmt.Sprintf("过滤: %s (按/编辑, Esc清除)", state.FilterText))
	}

	// 表头
	tableHeader := connections.RenderTableHeader(headerStyle)

	// 计算使用的行数 (Header + Stats + Spacers + TableHeader + Divider + Footer-Space)
	usedLines := 8 // 基础占用行数 (含底部页脚 1 行)

	// 加上图表和测试区域的行数
	if state.ViewMode == 0 {
		if state.ChartData != nil {
			usedLines += 6 // 图表高度(估计) + 间距
		}
		if len(state.SiteTests) > 0 {
			usedLines += 8 // 测试区域高度(Title + Space + Card + Space)
		}
	}

	// 加上过滤器行数
	if filterLine != "" {
		usedLines++
	}

	// 计算列表可显示的行数
	maxDisplay := state.Height - usedLines
	if maxDisplay < 5 {
		maxDisplay = 5
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

			row := connections.RenderConnectionRow(conn, rowStyle, prefix)
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
	var helpText string
	if state.ViewMode == 0 {
		helpText = dimStyle.Render("[↑↓]选择 [x]关闭 [X]全部关闭 [/]搜索 [h]历史 [s]测速 [S]全测 [r]刷新")
	} else {
		helpText = dimStyle.Render("[↑↓]选择 [Enter]详情 [/]搜索 [h]活跃")
	}

	// 组装页面
	var content []string
	content = append(content, headerStyle.Render("连接监控")+"  "+viewModeLabel)
	content = append(content, "")

	// 渲染监控图表区域（仅在活跃连接视图显示）
	if state.ViewMode == 0 {
		chartsSection := connections.RenderChartsSection(state.ChartData, state.Width)
		if chartsSection != "" {
			content = append(content, chartsSection)
			content = append(content, "")
		}
		// 渲染网站测速区域
		if len(state.SiteTests) > 0 {
			siteTestSection := connections.RenderSiteTestSection(state.SiteTests, state.SelectedSiteTest, state.Width)
			content = append(content, siteTestSection)
			content = append(content, "")
		}
	}

	content = append(content, stats)
	if filterLine != "" {
		content = append(content, filterLine)
	}
	content = append(content, "")
	content = append(content, tableHeader)
	content = append(content, styles.DividerStyle.Render(strings.Repeat("─", min(state.Width-2, 100))))
	content = append(content, strings.Join(rows, "\n"))

	// 统一底部的提示信息，固定到底部
	mainContent := strings.Join(content, "\n")
	contentLines := strings.Count(mainContent, "\n") + 1
	
	footer := common.RenderFooter(state.Width, state.Height, contentLines, helpText)
	return mainContent + footer
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
