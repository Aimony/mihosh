package rules

import (
	"fmt"
	"strings"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/ui/tui/components/common"
	"github.com/charmbracelet/lipgloss"
)

const (
	rulesFixedLines  = 6
	rulesMinHeight   = 5
	rulesScrollWidth = 2
)

// 规则类型颜色
var ruleTypeColors = map[string]lipgloss.Color{
	// 标准格式
	"DOMAIN":         common.CSecondary,
	"DOMAIN-SUFFIX":  common.CSecondary,
	"DOMAIN-KEYWORD": common.CInfo,
	"IP-CIDR":        common.CPurple,
	"IP-CIDR6":       common.CPurple,
	"GEOIP":          common.CDanger,
	"GEOSITE":        common.COrange,
	"RULE-SET":       common.CSuccess,
	"MATCH":          common.CWarning,
	"DIRECT":         common.CGray,
	// Clash Meta 驼峰格式
	"Domain":        common.CSecondary,
	"DomainSuffix":  common.CSecondary,
	"DomainKeyword": common.CInfo,
	"IPCIDR":        common.CPurple,
	"IPCIDR6":       common.CPurple,
	"GeoIP":         common.CDanger,
	"GeoSite":       common.COrange,
	"RuleSet":       common.CSuccess,
	"Match":         common.CWarning,
}

// filteredRule 带原始索引的规则
type filteredRule struct {
	Index int        // 原始索引
	Rule  model.Rule // 规则数据
}

// PageState 规则页面状态
type PageState struct {
	Rules               []model.Rule // 规则列表
	FilteredRuleIndices []int        // 过滤后的规则索引
	FilterText          string       // 搜索关键词
	FilterMode          bool         // 是否处于过滤输入模式
	SelectedRule        int          // 选中的规则索引
	ScrollTop           int          // 滚动偏移
	Width               int          // 页面宽度
	Height              int          // 页面高度
}

// RenderRulesPage 渲染规则页面
func RenderRulesPage(state PageState) string {
	var sections []string

	// 渲染搜索框
	searchBox := renderRuleSearchBox(state.FilterText, state.FilterMode)
	sections = append(sections, searchBox)
	sections = append(sections, "")

	// 过滤规则 (使用缓存的索引)
	var filteredRules []filteredRule
	for _, idx := range state.FilteredRuleIndices {
		if idx >= 0 && idx < len(state.Rules) {
			filteredRules = append(filteredRules, filteredRule{Index: idx, Rule: state.Rules[idx]})
		}
	}

	// 渲染统计信息
	stats := fmt.Sprintf("共 %d 条规则", len(filteredRules))
	if state.FilterText != "" {
		stats += fmt.Sprintf(" (过滤自 %d 条)", len(state.Rules))
	}
	sections = append(sections, common.MutedStyle.Render(stats))
	sections = append(sections, "")

	// 计算可显示的规则行数 (搜索框 + 统计 + 间隔)
	availableHeight := state.Height - rulesFixedLines
	if availableHeight < rulesMinHeight {
		availableHeight = rulesMinHeight
	}

	// 渲染规则列表
	ruleList := renderRuleList(filteredRules, state.SelectedRule, state.ScrollTop, availableHeight, state.Width)
	sections = append(sections, ruleList)

	// 统一底部的提示信息
	helpText := "[↑/↓]选择 [/]搜索 [Esc]清除搜索 [r]刷新"
	mainContent := strings.Join(sections, "\n")
	contentLines := strings.Count(mainContent, "\n") + 1

	footer := common.RenderFooter(state.Width, state.Height, contentLines, helpText)
	return mainContent + footer
}

// renderRuleSearchBox 渲染搜索框
func renderRuleSearchBox(filterText string, filterMode bool) string {
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	if filterMode {
		inputStyle = inputStyle.Background(common.CHighlight)
	}

	label := common.MutedStyle.Render("搜索: ")
	input := inputStyle.Render(filterText)

	if filterMode {
		input += inputStyle.Render("█")
	}

	hint := common.MutedStyle.Render(" (多个关键词用空格分隔)")
	return label + input + hint
}

// renderRuleList 渲染规则列表（含整体垂直滚动条）
func renderRuleList(rules []filteredRule, selectedIdx, scrollTop, maxLines, width int) string {
	if len(rules) == 0 {
		return common.MutedStyle.Render("暂无规则")
	}

	// 调整滚动位置确保选中项可见
	if selectedIdx < scrollTop {
		scrollTop = selectedIdx
	}
	if selectedIdx >= scrollTop+maxLines {
		scrollTop = selectedIdx - maxLines + 1
	}

	endIdx := scrollTop + maxLines
	if endIdx > len(rules) {
		endIdx = len(rules)
	}

	// 渲染规则行（预留滚动条宽度）
	listWidth := width - rulesScrollWidth
	var lines []string
	for i := scrollTop; i < endIdx; i++ {
		fr := rules[i]
		line := renderRuleEntry(fr.Rule, fr.Index, i == selectedIdx, listWidth)
		lines = append(lines, line)
	}
	listStr := strings.Join(lines, "\n")

	// 构建整体垂直滚动条
	scrollbarStr := buildScrollbar(maxLines, len(rules), scrollTop)

	// 计算滚动块的起止行
	thumbStart, thumbEnd := calcThumbRange(maxLines, len(rules), scrollTop)

	var barLines []string
	for i, ch := range strings.Split(scrollbarStr, "\n") {
		if i >= thumbStart && i < thumbEnd {
			barLines = append(barLines, common.MutedStyle.Foreground(lipgloss.Color("#AAAAAA")).Render(ch))
		} else {
			barLines = append(barLines, common.DimStyle.Render(ch))
		}
	}
	barStr := strings.Join(barLines, "\n")

	// 仅在内容超出可视区域时显示滚动条
	if len(rules) <= maxLines {
		return listStr
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, listStr, " "+barStr)
}

// buildScrollbar 构建高度为 viewHeight 的滚动条字符串（每行一个字符，换行连接）
func buildScrollbar(viewHeight, total, scrollTop int) string {
	lines := make([]string, viewHeight)
	for i := range lines {
		lines[i] = common.SymbolScrollbarTrack
	}
	// 用实心块覆盖滑块区域
	start, end := calcThumbRange(viewHeight, total, scrollTop)
	for i := start; i < end; i++ {
		if i < viewHeight {
			lines[i] = common.SymbolScrollbarThumb
		}
	}
	return strings.Join(lines, "\n")
}

// calcThumbRange 计算滑块在滚动条中的起止行（左闭右开）
func calcThumbRange(viewHeight, total, scrollTop int) (start, end int) {
	if total <= 0 {
		return 0, viewHeight
	}
	thumbSize := float64(viewHeight) * float64(viewHeight) / float64(total)
	if thumbSize < 1 {
		thumbSize = 1
	}
	thumbStart := float64(scrollTop) * float64(viewHeight) / float64(total)
	start = int(thumbStart)
	end = start + int(thumbSize+0.5)
	if end > viewHeight {
		end = viewHeight
	}
	if start >= end {
		end = start + 1
	}
	return
}

// renderRuleEntry 渲染单条规则
func renderRuleEntry(rule model.Rule, index int, selected bool, width int) string {
	// 获取规则类型颜色
	color := ruleTypeColors[rule.Type]
	if color == "" {
		color = lipgloss.Color("#CCCCCC")
	}

	// 序号
	indexStyle := lipgloss.NewStyle().Foreground(common.CSuccess).Width(6)
	indexStr := indexStyle.Render(fmt.Sprintf("%d.", index+1))

	// 类型标签
	typeStyle := lipgloss.NewStyle().Foreground(color).Bold(true).Width(16)
	typeStr := typeStyle.Render(rule.Type)

	// Payload
	payloadStyle := lipgloss.NewStyle().Foreground(common.CSecondary)
	payloadWidth := width - 50
	if payloadWidth < 20 {
		payloadWidth = 20
	}
	payload := rule.Payload
	if len(payload) > payloadWidth {
		payload = payload[:payloadWidth-3] + "..."
	}
	payloadStr := payloadStyle.Width(payloadWidth).Render(payload)

	// 代理
	proxyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))
	proxyStr := proxyStyle.Render(rule.Proxy)

	// 构建行
	line := fmt.Sprintf("%s %s %s %s", indexStr, typeStr, payloadStr, proxyStr)

	// 选中样式
	if selected {
		line = lipgloss.NewStyle().
			Background(common.CHighlight).
			Render(common.SymbolSelectActive + line)
	} else {
		line = common.SymbolSelectInactive + line
	}

	return line
}
