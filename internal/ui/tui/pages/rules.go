package pages

import (
	"fmt"
	"strings"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/charmbracelet/lipgloss"
)

// 规则类型颜色
var ruleTypeColors = map[string]lipgloss.Color{
	// 标准格式
	"DOMAIN":         lipgloss.Color("#00BFFF"), // 蓝色
	"DOMAIN-SUFFIX":  lipgloss.Color("#00BFFF"), // 蓝色
	"DOMAIN-KEYWORD": lipgloss.Color("#00CED1"), // 青色
	"IP-CIDR":        lipgloss.Color("#9B59B6"), // 紫色
	"IP-CIDR6":       lipgloss.Color("#9B59B6"), // 紫色
	"GEOIP":          lipgloss.Color("#E74C3C"), // 红色
	"GEOSITE":        lipgloss.Color("#E67E22"), // 橙色
	"RULE-SET":       lipgloss.Color("#2ECC71"), // 绿色
	"MATCH":          lipgloss.Color("#FFD700"), // 黄色
	"DIRECT":         lipgloss.Color("#95A5A6"), // 灰色
	// Clash Meta 驼峰格式
	"Domain":        lipgloss.Color("#00BFFF"), // 蓝色
	"DomainSuffix":  lipgloss.Color("#00BFFF"), // 蓝色
	"DomainKeyword": lipgloss.Color("#00CED1"), // 青色
	"IPCIDR":        lipgloss.Color("#9B59B6"), // 紫色
	"IPCIDR6":       lipgloss.Color("#9B59B6"), // 紫色
	"GeoIP":         lipgloss.Color("#E74C3C"), // 红色
	"GeoSite":       lipgloss.Color("#E67E22"), // 橙色
	"RuleSet":       lipgloss.Color("#2ECC71"), // 绿色
	"Match":         lipgloss.Color("#FFD700"), // 黄色
}

// filteredRule 带原始索引的规则
type filteredRule struct {
	Index int        // 原始索引
	Rule  model.Rule // 规则数据
}

// RulesPageState 规则页面状态
type RulesPageState struct {
	Rules        []model.Rule // 规则列表
	FilterText   string       // 搜索关键词
	FilterMode   bool         // 是否处于过滤输入模式
	SelectedRule int          // 选中的规则索引
	ScrollTop    int          // 滚动偏移
	Width        int          // 页面宽度
	Height       int          // 页面高度
}

// RenderRulesPage 渲染规则页面
func RenderRulesPage(state RulesPageState) string {
	var sections []string

	// 渲染搜索框
	searchBox := renderRuleSearchBox(state.FilterText, state.FilterMode)
	sections = append(sections, searchBox)
	sections = append(sections, "")

	// 过滤规则
	filteredRules := filterRules(state.Rules, state.FilterText)

	// 渲染统计信息
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	stats := fmt.Sprintf("共 %d 条规则", len(filteredRules))
	if state.FilterText != "" {
		stats += fmt.Sprintf(" (过滤自 %d 条)", len(state.Rules))
	}
	sections = append(sections, statsStyle.Render(stats))
	sections = append(sections, "")

	// 计算可显示的规则行数
	availableHeight := state.Height - 10
	if availableHeight < 5 {
		availableHeight = 5
	}

	// 渲染规则列表
	ruleList := renderRuleList(filteredRules, state.SelectedRule, state.ScrollTop, availableHeight, state.Width)
	sections = append(sections, ruleList)

	return strings.Join(sections, "\n")
}

// renderRuleSearchBox 渲染搜索框
func renderRuleSearchBox(filterText string, filterMode bool) string {
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF"))

	if filterMode {
		inputStyle = inputStyle.Background(lipgloss.Color("#333333"))
	}

	label := labelStyle.Render("搜索: ")
	input := inputStyle.Render(filterText)

	if filterMode {
		input += inputStyle.Render("█")
	}

	hint := labelStyle.Render(" (多个关键词用空格分隔)")
	return label + input + hint
}

// filterRules 过滤规则 - 支持多关键词空格分隔，保留原始索引
func filterRules(rules []model.Rule, filter string) []filteredRule {
	var result []filteredRule

	if filter == "" {
		// 无过滤时返回所有规则并保留原始索引
		for i, rule := range rules {
			result = append(result, filteredRule{Index: i, Rule: rule})
		}
		return result
	}

	// 分割关键词
	keywords := strings.Fields(strings.ToLower(filter))
	if len(keywords) == 0 {
		for i, rule := range rules {
			result = append(result, filteredRule{Index: i, Rule: rule})
		}
		return result
	}

	for i, rule := range rules {
		// 构建搜索文本
		searchText := strings.ToLower(rule.Type + " " + rule.Payload + " " + rule.Proxy)

		// 所有关键词都必须匹配 (AND 逻辑)
		allMatch := true
		for _, keyword := range keywords {
			if !strings.Contains(searchText, keyword) {
				allMatch = false
				break
			}
		}

		if allMatch {
			result = append(result, filteredRule{Index: i, Rule: rule})
		}
	}

	return result
}

// renderRuleList 渲染规则列表
func renderRuleList(rules []filteredRule, selectedIdx, scrollTop, maxLines, width int) string {
	if len(rules) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		return emptyStyle.Render("暂无规则")
	}

	var lines []string

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

	for i := scrollTop; i < endIdx; i++ {
		fr := rules[i]
		// 使用原始索引显示序号
		line := renderRuleEntry(fr.Rule, fr.Index, i == selectedIdx, width)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderRuleEntry 渲染单条规则
func renderRuleEntry(rule model.Rule, index int, selected bool, width int) string {
	// 获取规则类型颜色
	color := ruleTypeColors[rule.Type]
	if color == "" {
		color = lipgloss.Color("#CCCCCC")
	}

	// 序号
	indexStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#2ECC71")).Width(6)
	indexStr := indexStyle.Render(fmt.Sprintf("%d.", index+1))

	// 类型标签
	typeStyle := lipgloss.NewStyle().Foreground(color).Bold(true).Width(16)
	typeStr := typeStyle.Render(rule.Type)

	// Payload
	payloadStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00BFFF"))
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
			Background(lipgloss.Color("#333333")).
			Render("> " + line)
	} else {
		line = "  " + line
	}

	return line
}
