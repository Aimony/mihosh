package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

// ==================== 节点管理页面 ====================

// updateNodesPage 更新节点页面
func (m Model) updateNodesPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Up):
		if m.selectedProxy > 0 {
			m.selectedProxy--
		}

	case key.Matches(msg, keys.Down):
		if m.selectedProxy < len(m.currentProxies)-1 {
			m.selectedProxy++
		}

	case key.Matches(msg, keys.Left):
		if m.selectedGroup > 0 {
			m.selectedGroup--
			m.updateCurrentProxies()
			m.selectedProxy = 0
		}

	case key.Matches(msg, keys.Right):
		if m.selectedGroup < len(m.groupNames)-1 {
			m.selectedGroup++
			m.updateCurrentProxies()
			m.selectedProxy = 0
		}

	case key.Matches(msg, keys.Enter):
		if len(m.currentProxies) > 0 && m.selectedProxy < len(m.currentProxies) {
			groupName := m.groupNames[m.selectedGroup]
			proxyName := m.currentProxies[m.selectedProxy]
			return m, selectProxy(m.client, groupName, proxyName)
		}

	case key.Matches(msg, keys.Test):
		if len(m.currentProxies) > 0 && m.selectedProxy < len(m.currentProxies) {
			proxyName := m.currentProxies[m.selectedProxy]
			m.testing = true
			return m, testProxy(m.client, proxyName, m.testURL, m.timeout)
		}

	case key.Matches(msg, keys.TestAll):
		if len(m.groupNames) > 0 && len(m.currentProxies) > 0 {
			m.testing = true
			m.testFailures = []string{} // 清空之前的失败记录
			// 逐个测速当前组的所有节点
			return m, testAllProxies(m.client, m.currentProxies, m.testURL, m.timeout)
		}
	}

	return m, nil
}

// renderNodesPage 渲染节点管理页面
func (m Model) renderNodesPage() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("#666"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	activeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1E90FF"))

	// 策略组列表 - 优化表格对齐
	var groupList string
	if len(m.groupNames) == 0 {
		groupList = "  正在加载..."
	} else {
		// 计算最大列宽,向上取整到8的倍数以实现制表符效果
		maxNameLen := 0
		maxTypeLen := 0
		for _, name := range m.groupNames {
			nameLen := len(name)
			// 考虑中文字符占用更多空间
			for _, r := range name {
				if r > 127 {
					nameLen++
				}
			}
			if nameLen > maxNameLen {
				maxNameLen = nameLen
			}
			
			group := m.groups[name]
			if len(group.Type) > maxTypeLen {
				maxTypeLen = len(group.Type)
			}
		}
		
		// 向上取整到8的倍数
		maxNameLen = ((maxNameLen + 7) / 8) * 8
		maxTypeLen = ((maxTypeLen + 7) / 8) * 8

		var lines []string
		for i, name := range m.groupNames {
			group := m.groups[name]
			prefix := "  "
			if i == m.selectedGroup {
				prefix = "► "
			}

			// 计算实际占用宽度
			actualLen := len(name)
			for _, r := range name {
				if r > 127 {
					actualLen++
				}
			}
			
			// 表格对齐
			namePadding := strings.Repeat(" ", maxNameLen-actualLen)
			typePadding := strings.Repeat(" ", maxTypeLen-len(group.Type))
			line := fmt.Sprintf("%s%s%s │ %s%s │ %s", prefix, name, namePadding, group.Type, typePadding, group.Now)

			if i == m.selectedGroup {
				line = selectedStyle.Render(line)
			} else if group.Now != "" {
				line = activeStyle.Render(line)
			}

			lines = append(lines, line)
		}
		groupList = strings.Join(lines, "\n")
	}

	// 节点列表 - 表格对齐
	var proxyList string
	if len(m.currentProxies) == 0 {
		proxyList = "  无可用节点"
	} else {
		var currentNode string
		if len(m.groupNames) > 0 && m.selectedGroup < len(m.groupNames) {
			groupName := m.groupNames[m.selectedGroup]
			if group, ok := m.groups[groupName]; ok {
				currentNode = group.Now
			}
		}

		// 计算最大名称长度,考虑中文字符
		maxNameLen := 0
		for _, name := range m.currentProxies {
			nameLen := len(name)
			for _, r := range name {
				if r > 127 {
					nameLen++
				}
			}
			if nameLen > maxNameLen {
				maxNameLen = nameLen
			}
		}
		
		// 向上取整到8的倍数
		maxNameLen = ((maxNameLen + 7) / 8) * 8

		var lines []string
		for i, name := range m.currentProxies {
			proxy, exists := m.proxies[name]

			prefix := "  "
			if i == m.selectedProxy {
				prefix = "► "
			}

			// 计算实际占用宽度
			actualLen := len(name)
			for _, r := range name {
				if r > 127 {
					actualLen++
				}
			}
			
			// 节点名称对齐
			namePadding := strings.Repeat(" ", maxNameLen-actualLen)

			// 延迟信息 - 固定宽度
			delayText := "      "
			if exists && len(proxy.History) > 0 {
				lastDelay := proxy.History[len(proxy.History)-1].Delay
				if lastDelay > 0 {
					delayColor := m.getDelayColor(lastDelay)
					delayText = lipgloss.NewStyle().
						Foreground(delayColor).
						Render(fmt.Sprintf("%4dms", lastDelay))
				}
			}

			// 状态标记
			status := " "
			if name == currentNode {
				status = "✓"
			}

			line := fmt.Sprintf("%s%s%s │ %s │ %s", prefix, name, namePadding, delayText, status)

			if i == m.selectedProxy {
				line = selectedStyle.Render(line)
			} else if name == currentNode {
				line = activeStyle.Render(line)
			}

			lines = append(lines, line)
		}
		proxyList = strings.Join(lines, "\n")
	}

	// 操作提示
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Render("[↑/↓]选择 [←/→]切组 [Enter]切换 [t]测速 [a]全测 [r]刷新")

	// 测速失败信息
	var failureInfo string
	if len(m.testFailures) > 0 {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))
		
		failureLines := []string{errorStyle.Render("\n⚠ 测速失败的节点:")}
		for _, failure := range m.testFailures {
			failureLines = append(failureLines, "  "+failure)
		}
		failureInfo = strings.Join(failureLines, "\n")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Width(m.width-4).Render("策略组"),
		groupList,
		"",
		headerStyle.Width(m.width-4).Render("节点列表"),
		proxyList,
		failureInfo,
		"",
		helpText,
	)
}

// ==================== 连接监控页面 ====================

// updateConnectionsPage 更新连接页面
func (m Model) updateConnectionsPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Refresh):
		return m, fetchConnections(m.client)
	}
	return m, nil
}

// renderConnectionsPage 渲染连接监控页面
func (m Model) renderConnectionsPage() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700"))

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00BFFF"))

	if m.connections == nil {
		return "正在加载连接信息..."
	}

	// 统计信息
	stats := fmt.Sprintf(
		"总连接: %s | 上传: %s | 下载: %s",
		infoStyle.Render(fmt.Sprintf("%d", len(m.connections.Connections))),
		infoStyle.Render(formatBytes(m.connections.UploadTotal)),
		infoStyle.Render(formatBytes(m.connections.DownloadTotal)),
	)

	// 连接列表
	var connLines []string
	if len(m.connections.Connections) == 0 {
		connLines = append(connLines, "  无活跃连接")
	} else {
		maxDisplay := 15 // 最多显示15个连接
		for i, conn := range m.connections.Connections {
			if i >= maxDisplay {
				remaining := len(m.connections.Connections) - maxDisplay
				connLines = append(connLines, fmt.Sprintf("  ... 还有 %d 个连接", remaining))
				break
			}

			proxy := "DIRECT"
			if len(conn.Chains) > 0 {
				proxy = conn.Chains[len(conn.Chains)-1]
			}

			line := fmt.Sprintf(
				"  %s:%s → %s:%s via %s",
				conn.Metadata.SourceIP,
				conn.Metadata.SourcePort,
				conn.Metadata.Host,
				conn.Metadata.DestinationPort,
				infoStyle.Render(proxy),
			)
			connLines = append(connLines, line)
		}
	}

	// 操作提示
	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Render("[r]刷新")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("连接监控"),
		"",
		stats,
		"",
		strings.Join(connLines, "\n"),
		"",
		helpText,
	)
}

// formatBytes 格式化字节数
func formatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
