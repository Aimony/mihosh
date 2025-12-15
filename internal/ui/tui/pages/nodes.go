package pages

import (
	"fmt"
	"strings"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/pkg/utils"
	"github.com/charmbracelet/lipgloss"
)

// NodesPageState 节点页面状态（由 Model 传入）
type NodesPageState struct {
	Groups            map[string]model.Group
	Proxies           map[string]model.Proxy
	GroupNames        []string
	SelectedGroup     int
	SelectedProxy     int
	CurrentProxies    []string
	Testing           bool
	TestFailures      []string
	ShowFailureDetail bool // 是否显示测速失败详情
	Width             int
	Height            int // 终端高度
	GroupScrollTop    int // 策略组列表滚动偏移
	ProxyScrollTop    int // 节点列表滚动偏移
}

// displayWidth 计算字符串的显示宽度（中文占2个单位，英文占1个单位）
func displayWidth(s string) int {
	width := 0
	for _, r := range s {
		if r > 127 {
			width += 2
		} else {
			width++
		}
	}
	return width
}

// padString 将字符串填充到指定显示宽度
func padString(s string, targetWidth int) string {
	currentWidth := displayWidth(s)
	if currentWidth >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-currentWidth)
}

// RenderNodesPage 渲染节点管理页面
func RenderNodesPage(state NodesPageState) string {
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

	dimStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666"))

	// 计算可用高度（减去标签栏、状态栏、标题、帮助提示等固定区域）
	// 标签栏2行 + 状态栏2行 + 策略组标题2行 + 节点标题2行 + 帮助提示1行 + 间隔3行 = 约12行
	fixedLines := 12
	availableHeight := state.Height - fixedLines
	if availableHeight < 10 {
		availableHeight = 10
	}

	// 将可用空间分配给策略组和节点列表（比例约为 1:2）
	groupMaxLines := availableHeight / 3
	if groupMaxLines < 3 {
		groupMaxLines = 3
	}
	proxyMaxLines := availableHeight - groupMaxLines
	if proxyMaxLines < 5 {
		proxyMaxLines = 5
	}

	// 策略组列表
	var groupList string
	if len(state.GroupNames) == 0 {
		groupList = "  正在加载..."
	} else {
		// 计算各列的最大宽度
		maxNameLen := 8 // 最小宽度
		maxTypeLen := 8
		maxNowLen := 8

		for _, name := range state.GroupNames {
			nameWidth := displayWidth(name)
			if nameWidth > maxNameLen {
				maxNameLen = nameWidth
			}

			group := state.Groups[name]
			typeWidth := displayWidth(group.Type)
			if typeWidth > maxTypeLen {
				maxTypeLen = typeWidth
			}

			nowWidth := displayWidth(group.Now)
			if nowWidth > maxNowLen {
				maxNowLen = nowWidth
			}
		}

		// 渲染表头
		tableHeaderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888")).
			Bold(true)

		header := fmt.Sprintf("  %s │ %s │ %s",
			padString("名称", maxNameLen),
			padString("类型", maxTypeLen),
			padString("当前节点", maxNowLen),
		)
		groupList = tableHeaderStyle.Render(header) + "\n"

		// 调整滚动位置确保选中项可见
		groupScrollTop := state.GroupScrollTop
		if state.SelectedGroup < groupScrollTop {
			groupScrollTop = state.SelectedGroup
		}
		if state.SelectedGroup >= groupScrollTop+groupMaxLines {
			groupScrollTop = state.SelectedGroup - groupMaxLines + 1
		}

		// 显示滚动指示（上方）
		if groupScrollTop > 0 {
			groupList += dimStyle.Render(fmt.Sprintf("  ↑ 还有 %d 项\n", groupScrollTop))
		}

		// 渲染可见的策略组列表
		endIdx := groupScrollTop + groupMaxLines
		if endIdx > len(state.GroupNames) {
			endIdx = len(state.GroupNames)
		}

		var lines []string
		for i := groupScrollTop; i < endIdx; i++ {
			name := state.GroupNames[i]
			group := state.Groups[name]
			prefix := "  "
			if i == state.SelectedGroup {
				prefix = "► "
			}

			line := fmt.Sprintf("%s%s │ %s │ %s",
				prefix,
				padString(name, maxNameLen),
				padString(group.Type, maxTypeLen),
				padString(group.Now, maxNowLen),
			)

			if i == state.SelectedGroup {
				line = selectedStyle.Render(line)
			} else if group.Now != "" {
				line = activeStyle.Render(line)
			}

			lines = append(lines, line)
		}
		groupList += strings.Join(lines, "\n")

		// 显示滚动指示（下方）
		if endIdx < len(state.GroupNames) {
			groupList += "\n" + dimStyle.Render(fmt.Sprintf("  ↓ 还有 %d 项", len(state.GroupNames)-endIdx))
		}
	}

	// 节点列表
	var proxyList string
	if len(state.CurrentProxies) == 0 {
		proxyList = "  无可用节点"
	} else {
		var currentNode string
		if len(state.GroupNames) > 0 && state.SelectedGroup < len(state.GroupNames) {
			groupName := state.GroupNames[state.SelectedGroup]
			if group, ok := state.Groups[groupName]; ok {
				currentNode = group.Now
			}
		}

		// 计算节点名称的最大宽度
		maxNameLen := 8 // 最小宽度
		for _, name := range state.CurrentProxies {
			nameWidth := displayWidth(name)
			if nameWidth > maxNameLen {
				maxNameLen = nameWidth
			}
		}

		// 固定延迟和状态列的宽度
		delayColWidth := 6  // "999ms" 或 "     "
		statusColWidth := 2 // "✓ " 或 "  "

		// 渲染表头
		tableHeaderStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888")).
			Bold(true)

		header := fmt.Sprintf("  %s │ %s │ %s",
			padString("名称", maxNameLen),
			padString("延迟", delayColWidth),
			padString("状态", statusColWidth),
		)
		proxyList = tableHeaderStyle.Render(header) + "\n"

		// 调整滚动位置确保选中项可见
		proxyScrollTop := state.ProxyScrollTop
		if state.SelectedProxy < proxyScrollTop {
			proxyScrollTop = state.SelectedProxy
		}
		if state.SelectedProxy >= proxyScrollTop+proxyMaxLines {
			proxyScrollTop = state.SelectedProxy - proxyMaxLines + 1
		}

		// 显示滚动指示（上方）
		if proxyScrollTop > 0 {
			proxyList += dimStyle.Render(fmt.Sprintf("  ↑ 还有 %d 项\n", proxyScrollTop))
		}

		// 渲染可见的节点列表
		endIdx := proxyScrollTop + proxyMaxLines
		if endIdx > len(state.CurrentProxies) {
			endIdx = len(state.CurrentProxies)
		}

		var lines []string
		for i := proxyScrollTop; i < endIdx; i++ {
			name := state.CurrentProxies[i]
			proxy, exists := state.Proxies[name]

			prefix := "  "
			if i == state.SelectedProxy {
				prefix = selectedStyle.Render("► ")
			}

			// 名称部分
			namePart := padString(name, maxNameLen)
			if i == state.SelectedProxy {
				namePart = selectedStyle.Render(namePart)
			} else if name == currentNode {
				namePart = activeStyle.Render(namePart)
			}

			// 延迟部分
			delayStr := "      "
			if exists && len(proxy.History) > 0 {
				lastDelay := proxy.History[len(proxy.History)-1].Delay
				if lastDelay > 0 {
					delayStr = fmt.Sprintf("%4dms", lastDelay)
					if i != state.SelectedProxy && name != currentNode {
						// 只有非选中和非当前节点才单独着色
						delayColor := utils.GetDelayColor(lastDelay)
						delayStr = lipgloss.NewStyle().Foreground(delayColor).Render(delayStr)
					} else if i == state.SelectedProxy {
						delayStr = selectedStyle.Render(delayStr)
					} else if name == currentNode {
						delayStr = activeStyle.Render(delayStr)
					}
				}
			}

			// 状态部分
			status := "  "
			if name == currentNode {
				status = "✓ "
			}
			if i == state.SelectedProxy {
				status = selectedStyle.Render(status)
			} else if name == currentNode {
				status = activeStyle.Render(status)
			}

			line := prefix + namePart + " │ " + delayStr + " │ " + status
			lines = append(lines, line)
		}
		proxyList += strings.Join(lines, "\n")

		// 显示滚动指示（下方）
		if endIdx < len(state.CurrentProxies) {
			proxyList += "\n" + dimStyle.Render(fmt.Sprintf("  ↓ 还有 %d 项", len(state.CurrentProxies)-endIdx))
		}
	}

	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Render("[↑/↓]选择 [←/→]切组 [Enter]切换 [t]测速 [a]全测 [r]刷新")

	var failureInfo string
	if len(state.TestFailures) > 0 {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))
		hintStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888"))

		if state.ShowFailureDetail {
			// 详情模式：显示所有失败节点及错误信息
			failureLines := []string{errorStyle.Render("⚠ 测速失败的节点:")}
			for _, failure := range state.TestFailures {
				failureLines = append(failureLines, "  "+failure)
			}
			failureLines = append(failureLines, hintStyle.Render("  [f]收起详情"))
			failureInfo = strings.Join(failureLines, "\n")
		} else {
			// 简略模式：只显示失败数量
			failureInfo = errorStyle.Render(fmt.Sprintf("⚠ %d 个节点测速失败", len(state.TestFailures))) +
				" " + hintStyle.Render("[f]查看详情")
		}
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Width(state.Width-4).Render("策略组"),
		groupList,
		"",
		headerStyle.Width(state.Width-4).Render("节点列表"),
		proxyList,
		"",
		helpText,
		failureInfo,
	)
}
