package tui

import (
	"strings"
	"time"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/ui/tui/pages"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const logCap = 1000

// LogsState 日志页面完整状态（使用 Ring Buffer 存储日志）
type LogsState struct {
	// Ring Buffer
	logBuf  [logCap]model.LogEntry
	logHead int // 写入位置
	logCount int // 已写入总数（上限 logCap）

	filteredLogIndices []int
	logLevel           int    // 0=debug,1=info,2=warning,3=error,4=silent
	logFilter          string
	logFilterMode      bool
	selectedLog        int
	logScrollTop       int
	logHScrollOffset   int
}

// NewLogsState 初始化日志状态
func NewLogsState() LogsState {
	return LogsState{logLevel: 1} // 默认 info
}

// logs 返回日志列表（最新在前，用于渲染）
func (s LogsState) logs() []model.LogEntry {
	if s.logCount == 0 {
		return nil
	}
	result := make([]model.LogEntry, s.logCount)
	for i := 0; i < s.logCount; i++ {
		idx := (s.logHead - 1 - i + logCap) % logCap
		result[i] = s.logBuf[idx]
	}
	return result
}

// AppendLog 追加一条日志并更新过滤缓存
func (s LogsState) AppendLog(logType, payload string) LogsState {
	entry := model.LogEntry{
		Type:      logType,
		Payload:   payload,
		Timestamp: time.Now(),
	}
	s.logBuf[s.logHead] = entry
	s.logHead = (s.logHead + 1) % logCap
	if s.logCount < logCap {
		s.logCount++
	}
	s.updateFilteredLogs()
	return s
}

// ClearLogs 清空所有日志
func (s LogsState) ClearLogs() LogsState {
	s.logHead = 0
	s.logCount = 0
	s.filteredLogIndices = nil
	s.selectedLog = 0
	s.logScrollTop = 0
	return s
}

// ToPageState 转换为渲染层所需的 LogsPageState
func (s LogsState) ToPageState(width, height int) pages.LogsPageState {
	return pages.LogsPageState{
		Logs:               s.logs(),
		FilteredLogIndices: s.filteredLogIndices,
		LogLevel:           s.logLevel,
		FilterText:         s.logFilter,
		FilterMode:         s.logFilterMode,
		SelectedLog:        s.selectedLog,
		ScrollTop:          s.logScrollTop,
		HScrollOffset:      s.logHScrollOffset,
		Width:              width,
		Height:             height,
	}
}

// Update 处理日志页面按键
func (s LogsState) Update(msg tea.KeyMsg) (LogsState, tea.Cmd) {
	if s.logFilterMode {
		return s.handleLogFilterMode(msg)
	}

	switch {
	case key.Matches(msg, keys.Up):
		if s.selectedLog > 0 {
			s.selectedLog--
			if s.selectedLog < s.logScrollTop {
				s.logScrollTop = s.selectedLog
			}
		}

	case key.Matches(msg, keys.Down):
		if s.selectedLog < len(s.filteredLogIndices)-1 {
			s.selectedLog++
		}

	case key.Matches(msg, keys.Left):
		if s.logHScrollOffset > 0 {
			s.logHScrollOffset -= 10
			if s.logHScrollOffset < 0 {
				s.logHScrollOffset = 0
			}
		}

	case key.Matches(msg, keys.Right):
		s.logHScrollOffset += 10

	case msg.String() == "<" || msg.String() == ",":
		if s.logLevel > 0 {
			s.logLevel--
			s.updateFilteredLogs()
		}
		s.selectedLog = 0
		s.logScrollTop = 0

	case msg.String() == ">" || msg.String() == ".":
		if s.logLevel < 4 {
			s.logLevel++
			s.updateFilteredLogs()
		}
		s.selectedLog = 0
		s.logScrollTop = 0

	case msg.String() == "/":
		s.logFilterMode = true

	case key.Matches(msg, keys.Clear):
		s = s.ClearLogs()

	case key.Matches(msg, keys.Escape):
		if s.logFilter != "" {
			s.logFilter = ""
			s.selectedLog = 0
			s.logScrollTop = 0
			s.updateFilteredLogs()
		}
	}

	return s, nil
}

// HandleMouseScroll 鼠标滚轮处理
func (s LogsState) HandleMouseScroll(up bool) LogsState {
	count := len(s.filteredLogIndices)
	if up {
		if s.selectedLog > 0 {
			s.selectedLog--
			if s.selectedLog < s.logScrollTop {
				s.logScrollTop = s.selectedLog
			}
		}
	} else {
		if s.selectedLog < count-1 {
			s.selectedLog++
		}
	}
	return s
}

// handleLogFilterMode 日志过滤输入模式
func (s LogsState) handleLogFilterMode(msg tea.KeyMsg) (LogsState, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		s.logFilterMode = false
	case key.Matches(msg, keys.Enter):
		s.logFilterMode = false
		s.selectedLog = 0
		s.logScrollTop = 0
	case key.Matches(msg, keys.Backspace):
		if len(s.logFilter) > 0 {
			s.logFilter = s.logFilter[:len(s.logFilter)-1]
			s.updateFilteredLogs()
		}
	default:
		input := msg.String()
		if len(input) == 1 && input[0] >= 32 && input[0] < 127 {
			s.logFilter += input
			s.updateFilteredLogs()
		}
	}
	return s, nil
}

// updateFilteredLogs 重建过滤索引缓存（索引对应 logs() 的下标）
func (s *LogsState) updateFilteredLogs() {
	logList := s.logs()
	var indices []int
	for i, log := range logList {
		logLevelIndex := getLogLevelIndex(log.Type)
		if logLevelIndex < s.logLevel {
			continue
		}
		if s.logFilter != "" && !strings.Contains(strings.ToLower(log.Payload), strings.ToLower(s.logFilter)) {
			continue
		}
		indices = append(indices, i)
	}
	s.filteredLogIndices = indices
}

// getLogLevelIndex 获取日志级别索引（0=debug,1=info,2=warning,3=error,4=silent）
func getLogLevelIndex(level string) int {
	levels := []string{"debug", "info", "warning", "error", "silent"}
	for i, l := range levels {
		if l == level {
			return i
		}
	}
	return 1
}
