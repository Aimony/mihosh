package pages

import (
	"fmt"
	"strings"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/ui/tui/components/common"
	"github.com/aimony/mihosh/pkg/utils"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const (
	nodesFixedLines     = 8
	nodesMinHeight      = 10
	nodesDefaultNameLen = 8
)

// NodesPageState 节点页面状态（由 Model 传入）
type NodesPageState struct {
	Groups             map[string]model.Group
	Proxies            map[string]model.Proxy
	GroupNames         []string
	SelectedGroup      int
	SelectedProxy      int
	CurrentProxies     []string
	Testing            bool
	TestFailures       []string
	ShowFailureDetail  bool // 是否显示测速失败弹窗
	FailureScrollTop   int  // 测速失败弹窗滚动偏移
	Width              int
	Height             int // 终端高度
	GroupScrollTop     int // 策略组列表滚动偏移
	ProxyScrollTop     int // 节点列表滚动偏移
}

// displayWidth 计算字符串的显示宽度（使用 runewidth 库精确计算）
func displayWidth(s string) int {
	return runewidth.StringWidth(s)
}

// padString 将字符串填充到指定显示宽度
func padString(s string, targetWidth int) string {
	currentWidth := displayWidth(s)
	if currentWidth >= targetWidth {
		return s
	}
	return s + strings.Repeat(" ", targetWidth-currentWidth)
}

// renderScrollbar 渲染垂直滚动条
func renderScrollbar(height, total, scrollTop, currentIdx int) string {
	if total <= height {
		return " "
	}

	// 计算滚动块占据的行数比例
	barHeight := float64(height) * float64(height) / float64(total)
	if barHeight < 1 {
		barHeight = 1
	}

	// 计算滚动块起始位置比例
	barStart := float64(scrollTop) * float64(height) / float64(total)

	// 判断当前行 (currentIdx) 是否在渲染块内
	// currentIdx 是相对于列表可见区域的索引 (0 到 height-1)
	if float64(currentIdx) >= barStart && float64(currentIdx) < barStart+barHeight {
		return common.SymbolScrollbarThumb
	}
	return common.SymbolScrollbarTrack
}

// RenderNodesPage 渲染节点管理页面
func RenderNodesPage(state NodesPageState) string {
	// 计算可用高度（减去标签栏、状态栏、标题、帮助提示等固定区域）
	// 策略组标题1行 + 节点标题1行 + 列表内表头2行 + 间隔2行 + 底部提示1行 = 约7行
	availableHeight := state.Height - nodesFixedLines
	if availableHeight < nodesMinHeight {
		availableHeight = nodesMinHeight
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
		maxNameLen := nodesDefaultNameLen // 最小宽度
		maxTypeLen := nodesDefaultNameLen
		maxNowLen := nodesDefaultNameLen

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

		// 调整滚动位置确保选中项可见
		groupScrollTop := state.GroupScrollTop
		if state.SelectedGroup < groupScrollTop {
			groupScrollTop = state.SelectedGroup
		}
		if state.SelectedGroup >= groupScrollTop+groupMaxLines {
			groupScrollTop = state.SelectedGroup - groupMaxLines + 1
		}

		// 渲染表头
		header := fmt.Sprintf("  %s │ %s │ %s",
			padString("名称", maxNameLen),
			padString("类型", maxTypeLen),
			padString("当前节点", maxNowLen),
		)
		groupList += common.TableHeaderStyle.Render(header) + "\n"

		// 渲染可见的策略组列表
		endIdx := groupScrollTop + groupMaxLines
		if endIdx > len(state.GroupNames) {
			endIdx = len(state.GroupNames)
		}

		var lines []string
		for i := groupScrollTop; i < endIdx; i++ {
			name := state.GroupNames[i]
			group := state.Groups[name]
			prefix := common.SymbolSelectInactive
			if i == state.SelectedGroup {
				prefix = common.SymbolSelectActive
			}

			// 计算这一行的内容
			content := fmt.Sprintf("%s%s │ %s │ %s",
				prefix,
				padString(name, maxNameLen),
				padString(group.Type, maxTypeLen),
				padString(group.Now, maxNowLen),
			)

			if i == state.SelectedGroup {
				content = common.SelectedStyle.Render(content)
			} else if group.Now != "" {
				content = common.InactiveStyle.Render(content)
			}

			// 追加滚动条
			bar := renderScrollbar(groupMaxLines, len(state.GroupNames), groupScrollTop, i-groupScrollTop)
			lines = append(lines, content+" "+common.DimStyle.Render(bar))
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
		maxNameLen := nodesDefaultNameLen // 最小宽度
		for _, name := range state.CurrentProxies {
			nameWidth := displayWidth(name)
			if nameWidth > maxNameLen {
				maxNameLen = nameWidth
			}
		}

		// 固定延迟和状态列的宽度
		delayColWidth := 6  // "999ms" 或 "     "
		statusColWidth := 2 // "✓ " 或 "  "

		// 调整滚动位置确保选中项可见
		proxyScrollTop := state.ProxyScrollTop
		if state.SelectedProxy < proxyScrollTop {
			proxyScrollTop = state.SelectedProxy
		}
		if state.SelectedProxy >= proxyScrollTop+proxyMaxLines {
			proxyScrollTop = state.SelectedProxy - proxyMaxLines + 1
		}

		// 渲染表头
		header := fmt.Sprintf("  %s │ %s │ %s",
			padString("名称", maxNameLen),
			padString("延迟", delayColWidth),
			padString("状态", statusColWidth),
		)
		proxyList += common.TableHeaderStyle.Render(header) + "\n"

		// 渲染可见的节点列表
		endIdx := proxyScrollTop + proxyMaxLines
		if endIdx > len(state.CurrentProxies) {
			endIdx = len(state.CurrentProxies)
		}

		var lines []string
		for i := proxyScrollTop; i < endIdx; i++ {
			name := state.CurrentProxies[i]
			proxy, exists := state.Proxies[name]

			prefix := common.SymbolSelectInactive
			if i == state.SelectedProxy {
				prefix = common.SelectedStyle.Render(common.SymbolSelectActive)
			}

			// 名称部分
			namePart := padString(name, maxNameLen)
			if i == state.SelectedProxy {
				namePart = common.SelectedStyle.Render(namePart)
			} else if name == currentNode {
				namePart = common.InactiveStyle.Render(namePart)
			}

			// 延迟部分
			delayStr := "      "
			if exists && len(proxy.History) > 0 {
				lastEntry := proxy.History[len(proxy.History)-1]
				lastDelay := lastEntry.Delay
				if lastEntry.Error != "" || lastDelay < 0 {
					// -1 或者显式 Error 表示测试失败
					delayStr = " Error"
					if i != state.SelectedProxy && name != currentNode {
						delayStr = common.ErrorStyle.Render(delayStr)
					} else if i == state.SelectedProxy {
						delayStr = common.SelectedStyle.Render(delayStr)
					} else if name == currentNode {
						delayStr = common.InactiveStyle.Render(delayStr)
					}
				} else if lastDelay >= 0 {
					delayStr = fmt.Sprintf("%4dms", lastDelay)
					if i != state.SelectedProxy && name != currentNode {
						// 只有非选中和非当前节点才单独着色
						delayColor := utils.GetDelayColor(lastDelay)
						delayStr = lipgloss.NewStyle().Foreground(delayColor).Render(delayStr)
					} else if i == state.SelectedProxy {
						delayStr = common.SelectedStyle.Render(delayStr)
					} else if name == currentNode {
						delayStr = common.InactiveStyle.Render(delayStr)
					}
				}
			}

			// 状态部分
			status := common.SymbolSelectInactive
			if name == currentNode {
				status = common.SymbolCheck
			}
			if i == state.SelectedProxy {
				status = common.SelectedStyle.Render(status)
			} else if name == currentNode {
				status = common.InactiveStyle.Render(status)
			}

			line := prefix + namePart + " │ " + delayStr + " │ " + status
			// 追加滚动条
			bar := renderScrollbar(proxyMaxLines, len(state.CurrentProxies), proxyScrollTop, i-proxyScrollTop)
			lines = append(lines, line+" "+common.DimStyle.Render(bar))
		}
		proxyList += strings.Join(lines, "\n")
	}

	helpText := common.MutedStyle.Render("[↑/↓]选择 [←/→]切组 [Enter]切换 [t]测速 [a]全测 [r]刷新")

	var failureBadge string
	if len(state.TestFailures) > 0 {
		failureBadge = common.ErrorStyle.Render(fmt.Sprintf("⚠ %d 个节点测速失败", len(state.TestFailures))) +
			" " + common.MutedStyle.Render("[f]查看详情")
	}

	mainContent := lipgloss.JoinVertical(
		lipgloss.Left,
		common.PageHeaderStyle.Width(state.Width-4).Render(fmt.Sprintf("策略组 [%d/%d]", state.SelectedGroup+1, len(state.GroupNames))),
		groupList,
		"",
		common.PageHeaderStyle.Width(state.Width-4).Render(fmt.Sprintf("节点列表 [%d/%d]", state.SelectedProxy+1, len(state.CurrentProxies))),
		proxyList,
		"",
		failureBadge,
	)

	contentLines := strings.Count(mainContent, "\n") + 1
	footer := common.RenderFooter(state.Width, state.Height, contentLines, helpText)
	fullPage := mainContent + footer

	if state.ShowFailureDetail {
		modal := buildFailureModal(state)
		return overlayCenter(fullPage, modal, state.Width, state.Height)
	}
	return fullPage
}

// buildFailureModal 构建测速失败详情弹窗字符串
func buildFailureModal(state NodesPageState) string {
	failures := state.TestFailures

	// 弹窗内容区宽度（去掉左右边框各1 + 内边距各1 = 4）
	modalWidth := state.Width - 10
	if modalWidth < 50 {
		modalWidth = 50
	}
	if modalWidth > 100 {
		modalWidth = 100
	}
	innerWidth := modalWidth - 4

	// 可显示的最大行数（去掉标题、分隔线、空行、帮助行 = 4行）
	modalHeight := state.Height - 8
	if modalHeight < 6 {
		modalHeight = 6
	}
	maxDisplay := modalHeight - 4
	if maxDisplay < 1 {
		maxDisplay = 1
	}

	// 截断过长的失败信息
	var allLines []string
	for _, f := range failures {
		if displayWidth(f) > innerWidth {
			runes := []rune(f)
			w, end := 0, 0
			for end < len(runes) {
				cw := runewidth.RuneWidth(runes[end])
				if w+cw > innerWidth-1 {
					break
				}
				w += cw
				end++
			}
			allLines = append(allLines, string(runes[:end])+"…")
		} else {
			allLines = append(allLines, f)
		}
	}

	total := len(allLines)

	// 限制滚动范围
	scrollTop := state.FailureScrollTop
	if scrollTop > total-maxDisplay {
		scrollTop = total - maxDisplay
	}
	if scrollTop < 0 {
		scrollTop = 0
	}
	endIdx := scrollTop + maxDisplay
	if endIdx > total {
		endIdx = total
	}

	// 构建内容行
	var bodyLines []string
	if scrollTop > 0 {
		bodyLines = append(bodyLines, common.DimStyle.Render(fmt.Sprintf("↑ 还有 %d 行", scrollTop)))
	}
	for _, line := range allLines[scrollTop:endIdx] {
		bodyLines = append(bodyLines, line)
	}
	if endIdx < total {
		bodyLines = append(bodyLines, common.DimStyle.Render(fmt.Sprintf("↓ 还有 %d 行", total-endIdx)))
	}
	bodyLines = append(bodyLines, "")
	bodyLines = append(bodyLines, common.MutedStyle.Render("[↑/↓] 滚动  [f/Esc] 关闭"))

	body := strings.Join(bodyLines, "\n")

	title := common.ErrorStyle.Render(fmt.Sprintf("⚠ 测速失败节点列表  共 %d 条", total))
	separator := common.DimStyle.Render(strings.Repeat("─", innerWidth))
	content := lipgloss.JoinVertical(lipgloss.Left, title, separator, body)

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#E74C3C")).
		Padding(0, 1).
		Width(modalWidth).
		Render(content)
}

// overlayCenter 将弹窗字符串居中叠加在底层页面上
func overlayCenter(base, overlay string, width, height int) string {
	baseLines := strings.Split(base, "\n")
	overlayLines := strings.Split(overlay, "\n")

	// 补齐底层行数到 height
	for len(baseLines) < height {
		baseLines = append(baseLines, "")
	}

	overlayH := len(overlayLines)
	overlayW := 0
	for _, l := range overlayLines {
		if w := displayWidth(l); w > overlayW {
			overlayW = w
		}
	}

	// 计算叠加起始位置（居中）
	startRow := (height - overlayH) / 2
	if startRow < 0 {
		startRow = 0
	}
	startCol := (width - overlayW) / 2
	if startCol < 0 {
		startCol = 0
	}

	result := make([]string, len(baseLines))
	copy(result, baseLines)

	for i, ol := range overlayLines {
		row := startRow + i
		if row >= len(result) {
			break
		}
		bl := result[row]
		// 将底层行按显示宽度截断到 startCol，然后拼上弹窗行
		blRunes := []rune(bl)
		w, col := 0, 0
		for col < len(blRunes) {
			cw := runewidth.RuneWidth(blRunes[col])
			if w+cw > startCol {
				break
			}
			w += cw
			col++
		}
		prefix := string(blRunes[:col])
		// 补齐空格到 startCol
		if w < startCol {
			prefix += strings.Repeat(" ", startCol-w)
		}
		result[row] = prefix + ol
	}

	return strings.Join(result, "\n")
}
