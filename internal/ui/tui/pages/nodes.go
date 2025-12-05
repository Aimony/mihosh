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
	Groups         map[string]model.Group
	Proxies        map[string]model.Proxy
	GroupNames     []string
	SelectedGroup  int
	SelectedProxy  int
	CurrentProxies []string
	Testing        bool
	TestFailures   []string
	Width          int
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
		headerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888")).
			Bold(true)

		header := fmt.Sprintf("  %s │ %s │ %s",
			padString("名称", maxNameLen),
			padString("类型", maxTypeLen),
			padString("当前节点", maxNowLen),
		)
		groupList = headerStyle.Render(header) + "\n"

		// 渲染策略组列表
		var lines []string
		for i, name := range state.GroupNames {
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
		headerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888")).
			Bold(true)

		header := fmt.Sprintf("  %s │ %s │ %s",
			padString("名称", maxNameLen),
			padString("延迟", delayColWidth),
			padString("状态", statusColWidth),
		)
		proxyList = headerStyle.Render(header) + "\n"

		// 渲染节点列表
		var lines []string
		for i, name := range state.CurrentProxies {
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
	}

	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Render("[↑/↓]选择 [←/→]切组 [Enter]切换 [t]测速 [a]全测 [r]刷新")

	var failureInfo string
	if len(state.TestFailures) > 0 {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000"))

		failureLines := []string{errorStyle.Render("\n⚠ 测速失败的节点:")}
		for _, failure := range state.TestFailures {
			failureLines = append(failureLines, "  "+failure)
		}
		failureInfo = strings.Join(failureLines, "\n")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Width(state.Width-4).Render("策略组"),
		groupList,
		"",
		headerStyle.Width(state.Width-4).Render("节点列表"),
		proxyList,
		failureInfo,
		"",
		helpText,
	)
}
