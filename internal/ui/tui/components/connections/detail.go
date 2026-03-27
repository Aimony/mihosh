package connections

import (
	"encoding/json"
	"fmt"
	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/ui/styles"
	"github.com/aimony/mihosh/pkg/utils"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

// RenderConnectionDetail 渲染连接详情（JSON格式，支持滚动）
func RenderConnectionDetail(conn *model.Connection, ipInfo *model.IPInfo, height, scrollTop int) string {
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
	ipInfoLines := strings.Split(RenderIPInfoSection(ipInfo, ipInfoStyle, dimStyle), "\n")
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

// RenderIPInfoSection 渲染IP地理信息部分
func RenderIPInfoSection(ipInfo *model.IPInfo, infoStyle, dimStyle lipgloss.Style) string {
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
