package logs

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/ui/tui/components/common"
	"github.com/charmbracelet/lipgloss"
)

const (
	logsFixedLines     = 8
	logsMinHeight      = 5
	logsDefaultPadding = 20
	logsLevelWidth     = 8
)

// 日志级别列表
var logLevels = []string{"debug", "info", "warning", "error", "silent"}

// 日志级别颜色
var logLevelColors = map[string]lipgloss.Color{
	"debug":   common.CMuted,
	"info":    common.CSecondary,
	"warning": common.CWarning,
	"error":   common.CDanger,
	"silent":  common.CPurple,
}

// PageState 日志页面状态
type PageState struct {
	Logs               []model.LogEntry // 日志列表
	FilteredLogIndices []int            // 过滤后的日志索引
	LogLevel           int              // 当前级别索引
	FilterText         string           // 搜索关键词
	FilterMode         bool             // 是否处于过滤输入模式
	SelectedLog        int              // 选中的日志索引
	ScrollTop          int              // 滚动偏移
	HScrollOffset      int              // 水平滚动偏移
	Width              int              // 页面宽度
	Height             int              // 页面高度
}

// RenderLogsPage 渲染日志页面
func RenderLogsPage(state PageState) string {
	var sections []string

	// 渲染日志级别标签栏
	levelBar := renderLevelBar(state.LogLevel)
	sections = append(sections, levelBar)
	sections = append(sections, "")

	// 渲染搜索框
	searchBox := renderLogSearchBox(state.FilterText, state.FilterMode)
	sections = append(sections, searchBox)
	sections = append(sections, "")

	// 过滤日志 (使用缓存的索引)
	var filteredLogs []model.LogEntry
	for _, idx := range state.FilteredLogIndices {
		if idx >= 0 && idx < len(state.Logs) {
			filteredLogs = append(filteredLogs, state.Logs[idx])
		}
	}

	// 渲染统计信息
	stats := fmt.Sprintf("共 %d 条日志 (级别: %s)", len(filteredLogs), logLevels[state.LogLevel])
	sections = append(sections, common.MutedStyle.Render(stats))
	sections = append(sections, "")

	// 计算可显示的日志行数 (级别栏 + 搜索框 + 统计 + 间隔)
	availableHeight := state.Height - logsFixedLines
	if availableHeight < logsMinHeight {
		availableHeight = logsMinHeight
	}

	// 渲染日志列表
	logList := renderLogList(filteredLogs, state.SelectedLog, state.ScrollTop, availableHeight, state.Width, state.HScrollOffset)
	sections = append(sections, logList)

	// 统一底部的提示信息
	helpText := "[↑/↓]选择 [{/}]级别 [/]搜索 [c]清空 [Esc]清除搜索 [r]刷新"
	mainContent := strings.Join(sections, "\n")
	contentLines := strings.Count(mainContent, "\n") + 1

	footer := common.RenderFooter(state.Width, state.Height, contentLines, helpText)
	return mainContent + footer
}

// renderLevelBar 渲染日志级别标签栏
func renderLevelBar(selectedLevel int) string {
	var tabs []string

	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(common.CWhite).
		Background(common.CHighlight).
		Padding(0, 1)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(common.CMuted).
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
	if filterMode {
		inputStyle := lipgloss.NewStyle().Foreground(common.CWhite).Background(common.CHighlight)
		label := common.MutedStyle.Render("搜索: ")
		input := inputStyle.Render(filterText + "█")
		return label + input
	}

	label := common.MutedStyle.Render("搜索: ")
	input := lipgloss.NewStyle().Foreground(common.CWhite).Render(filterText)
	return label + input
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
		return common.MutedStyle.Render("暂无日志")
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
	color := logLevelColors[log.Type]
	if color == "" {
		color = common.CMuted
	}

	levelStyle := lipgloss.NewStyle().
		Foreground(color).
		Width(logsLevelWidth)

	timeStr := log.Timestamp.Format("15:04:05")
	timePart := common.DimStyle.Render(timeStr)

	contentStyle := lipgloss.NewStyle().Foreground(common.CMuted)

	content := log.Payload
	if decoded, err := url.QueryUnescape(content); err == nil {
		content = decoded
	}

	usableWidth := maxWidth - logsDefaultPadding
	if usableWidth < 1 {
		usableWidth = 1
	}

	if hOffset > 0 {
		contentWidth := 0
		for i, r := range content {
			if r > 127 {
				contentWidth += 2
			} else {
				contentWidth++
			}
			if contentWidth > hOffset {
				content = content[i:]
				break
			}
		}
		if contentWidth <= hOffset {
			content = ""
		}
	}

	displayWidth := usableWidth
	if len(content) > displayWidth {
		content = content[:displayWidth]
	}

	line := fmt.Sprintf("%s %s %s",
		timePart,
		levelStyle.Render(strings.ToUpper(log.Type)),
		contentStyle.Render(content),
	)

	if selected {
		line = lipgloss.NewStyle().
			Background(common.CHighlight).
			Render(common.SymbolSelectActive + line)
	} else {
		line = common.SymbolSelectInactive + line
	}

	return line
}
