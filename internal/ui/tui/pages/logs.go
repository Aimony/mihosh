package pages

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/charmbracelet/lipgloss"
)

// 日志级别列表
var logLevels = []string{"debug", "info", "warning", "error", "silent"}

// 日志级别颜色
var logLevelColors = map[string]lipgloss.Color{
	"debug":   lipgloss.Color("#888888"), // 灰色
	"info":    lipgloss.Color("#00BFFF"), // 蓝色
	"warning": lipgloss.Color("#FFD700"), // 黄色
	"error":   lipgloss.Color("#FF4444"), // 红色
	"silent":  lipgloss.Color("#9B59B6"), // 紫色
}

// LogsPageState 日志页面状态
type LogsPageState struct {
	Logs          []model.LogEntry // 日志列表
	LogLevel      int              // 当前级别索引
	FilterText    string           // 搜索关键词
	FilterMode    bool             // 是否处于过滤输入模式
	SelectedLog   int              // 选中的日志索引
	ScrollTop     int              // 滚动偏移
	HScrollOffset int              // 水平滚动偏移
	Width         int              // 页面宽度
	Height        int              // 页面高度
}

// RenderLogsPage 渲染日志页面
func RenderLogsPage(state LogsPageState) string {
	var sections []string

	// 渲染日志级别标签栏
	levelBar := renderLevelBar(state.LogLevel)
	sections = append(sections, levelBar)
	sections = append(sections, "")

	// 渲染搜索框
	searchBox := renderLogSearchBox(state.FilterText, state.FilterMode)
	sections = append(sections, searchBox)
	sections = append(sections, "")

	// 过滤日志
	filteredLogs := filterLogs(state.Logs, logLevels[state.LogLevel], state.FilterText)

	// 渲染统计信息
	statsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
	stats := fmt.Sprintf("共 %d 条日志 (级别: %s)", len(filteredLogs), logLevels[state.LogLevel])
	sections = append(sections, statsStyle.Render(stats))
	sections = append(sections, "")

	// 计算可显示的日志行数
	availableHeight := state.Height - 12
	if availableHeight < 5 {
		availableHeight = 5
	}

	// 渲染日志列表
	logList := renderLogList(filteredLogs, state.SelectedLog, state.ScrollTop, availableHeight, state.Width, state.HScrollOffset)
	sections = append(sections, logList)

	return strings.Join(sections, "\n")
}

// renderLevelBar 渲染日志级别标签栏
func renderLevelBar(selectedLevel int) string {
	var tabs []string

	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#333333")).
		Padding(0, 1)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Padding(0, 1)

	for i, level := range logLevels {
		color := logLevelColors[level]
		indicator := lipgloss.NewStyle().Foreground(color).Render("●")

		if i == selectedLevel {
			tabs = append(tabs, activeStyle.Render(indicator+" "+level))
		} else {
			tabs = append(tabs, inactiveStyle.Render(indicator+" "+level))
		}
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, tabs...)
}

// renderLogSearchBox 渲染搜索框
func renderLogSearchBox(filterText string, filterMode bool) string {
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

	return label + input
}

// filterLogs 过滤日志（显示选定级别及更高级别的日志）
func filterLogs(logs []model.LogEntry, level, filter string) []model.LogEntry {
	var filtered []model.LogEntry
	levelIndex := getLevelIndex(level)

	for _, log := range logs {
		// 只显示选定级别及更高级别的日志
		logLevelIndex := getLevelIndex(log.Type)
		if logLevelIndex < levelIndex {
			continue
		}

		// 关键词过滤
		if filter != "" && !strings.Contains(strings.ToLower(log.Payload), strings.ToLower(filter)) {
			continue
		}

		filtered = append(filtered, log)
	}

	return filtered
}

// getLevelIndex 获取日志级别索引
func getLevelIndex(level string) int {
	for i, l := range logLevels {
		if l == level {
			return i
		}
	}
	return 1 // 默认info
}

// renderLogList 渲染日志列表
func renderLogList(logs []model.LogEntry, selectedIdx, scrollTop, maxLines, width, hOffset int) string {
	if len(logs) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))
		return emptyStyle.Render("暂无日志")
	}

	var lines []string
	maxWidth := width - 20 // 预留边距

	// 调整滚动位置确保选中项可见
	if selectedIdx < scrollTop {
		scrollTop = selectedIdx
	}
	if selectedIdx >= scrollTop+maxLines {
		scrollTop = selectedIdx - maxLines + 1
	}

	endIdx := scrollTop + maxLines
	if endIdx > len(logs) {
		endIdx = len(logs)
	}

	for i := scrollTop; i < endIdx; i++ {
		log := logs[i]
		line := renderLogEntry(log, i == selectedIdx, maxWidth, hOffset)
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderLogEntry 渲染单条日志
func renderLogEntry(log model.LogEntry, selected bool, maxWidth int, hOffset int) string {
	// 获取日志级别颜色
	color := logLevelColors[log.Type]
	if color == "" {
		color = lipgloss.Color("#CCCCCC")
	}

	// 级别标签
	levelStyle := lipgloss.NewStyle().
		Foreground(color).
		Width(8)

	// 时间
	timeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#666666"))
	timeStr := log.Timestamp.Format("15:04:05")

	// 内容
	contentStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#CCCCCC"))

	// 对日志内容进行 URL 解码
	content := log.Payload
	if decoded, err := url.QueryUnescape(content); err == nil {
		content = decoded
	}
	if hOffset > 0 && len(content) > hOffset {
		content = content[hOffset:]
	} else if hOffset > 0 {
		content = ""
	}
	// 截取显示宽度
	displayWidth := maxWidth - 20
	if displayWidth > 0 && len(content) > displayWidth {
		content = content[:displayWidth]
	}

	// 构建行
	line := fmt.Sprintf("%s %s %s",
		timeStyle.Render(timeStr),
		levelStyle.Render(strings.ToUpper(log.Type)),
		contentStyle.Render(content),
	)

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
