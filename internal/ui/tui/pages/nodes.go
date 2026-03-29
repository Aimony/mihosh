package pages

import (
	"fmt"
	"strings"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/ui/tui/components/common"
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
	Groups            map[string]model.Group
	Proxies           map[string]model.Proxy
	GroupNames        []string
	SelectedGroup     int
	SelectedProxy     int
	CurrentProxies    []string
	Testing           bool
	TestingTarget     string
	TestFailures      []string
	ShowFailureDetail bool     // 是否显示测速失败弹窗
	FailureScrollTop  int      // 测速失败弹窗滚动偏移
	SortOrderLabels   []string // 排序选项文本
	CurrentSortOrder  int      // 当前排序模式
	Width             int
	Height            int    // 终端高度
	GroupScrollTop    int    // 策略组列表滚动偏移
	ProxyScrollTop    int    // 节点列表滚动偏移
	FilterText        string // 节点搜索关键词
	FilterMode        bool   // 是否处于搜索输入模式
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

// RenderNodesPage 渲染节点管理页面
func RenderNodesPage(state NodesPageState) string {
	groupMaxLines, proxyMaxLines := calcNodesListMaxLines(state.Height)
	groupList := renderGroupListComponent(state, groupMaxLines)
	proxyList := renderProxyListComponent(state, proxyMaxLines)

	sortLabel := ""
	if len(state.SortOrderLabels) > 0 && state.CurrentSortOrder < len(state.SortOrderLabels) {
		sortLabel = state.SortOrderLabels[state.CurrentSortOrder]
	}

	// 搜索状态提示行
	var searchLine string
	if state.FilterMode {
		searchLine = common.TableHeaderStyle.Render(fmt.Sprintf("搜索: %s▌", state.FilterText))
	} else if state.FilterText != "" {
		searchLine = common.MutedStyle.Render(fmt.Sprintf("搜索: %s  [Esc]清除", state.FilterText))
	}

	helpText := common.MutedStyle.Render(fmt.Sprintf("[↑/↓]选择 [←/→]切组 [Enter]切换 [t]测速 [a]全测 [s]排序:%s [/]搜索 [r]刷新", sortLabel))

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
		searchLine,
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
