package pages

import (
	"fmt"
	"strings"

	"github.com/aimony/mihomo-cli/internal/domain/model"
	"github.com/charmbracelet/lipgloss"
	"github.com/aimony/mihomo-cli/pkg/utils"
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
		maxNameLen, maxTypeLen := 0, 0
		for _, name := range state.GroupNames {
			nameLen := len(name)
			for _, r := range name {
				if r > 127 {
					nameLen++
				}
			}
			if nameLen > maxNameLen {
				maxNameLen = nameLen
			}

			group := state.Groups[name]
			if len(group.Type) > maxTypeLen {
				maxTypeLen = len(group.Type)
			}
		}

		maxNameLen = ((maxNameLen + 7) / 8) * 8
		maxTypeLen = ((maxTypeLen + 7) / 8) * 8

		var lines []string
		for i, name := range state.GroupNames {
			group := state.Groups[name]
			prefix := "  "
			if i == state.SelectedGroup {
				prefix = "► "
			}

			actualLen := len(name)
			for _, r := range name {
				if r > 127 {
					actualLen++
				}
			}

			namePadding := strings.Repeat(" ", maxNameLen-actualLen)
			typePadding := strings.Repeat(" ", maxTypeLen-len(group.Type))
			line := fmt.Sprintf("%s%s%s │ %s%s │ %s", prefix, name, namePadding, group.Type, typePadding, group.Now)

			if i == state.SelectedGroup {
				line = selectedStyle.Render(line)
			} else if group.Now != "" {
				line = activeStyle.Render(line)
			}

			lines = append(lines, line)
		}
		groupList = strings.Join(lines, "\n")
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

		maxNameLen := 0
		for _, name := range state.CurrentProxies {
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

		maxNameLen = ((maxNameLen + 7) / 8) * 8

		var lines []string
		for i, name := range state.CurrentProxies {
			proxy, exists := state.Proxies[name]

			prefix := "  "
			if i == state.SelectedProxy {
				prefix = "► "
			}

			actualLen := len(name)
			for _, r := range name {
				if r > 127 {
					actualLen++
				}
			}

			namePadding := strings.Repeat(" ", maxNameLen-actualLen)

			delayText := "      "
			if exists && len(proxy.History) > 0 {
				lastDelay := proxy.History[len(proxy.History)-1].Delay
				if lastDelay > 0 {
					delayColor := utils.GetDelayColor(lastDelay)
					delayText = lipgloss.NewStyle().
						Foreground(delayColor).
						Render(fmt.Sprintf("%4dms", lastDelay))
				}
			}

			status := " "
			if name == currentNode {
				status = "✓"
			}

			line := fmt.Sprintf("%s%s%s │ %s │ %s", prefix, name, namePadding, delayText, status)

			if i == state.SelectedProxy {
				line = selectedStyle.Render(line)
			} else if name == currentNode {
				line = activeStyle.Render(line)
			}

			lines = append(lines, line)
		}
		proxyList = strings.Join(lines, "\n")
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
