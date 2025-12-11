package model

import "time"

// LogEntry 日志条目
type LogEntry struct {
	Type      string    // debug, info, warning, error, silent
	Payload   string    // 日志内容
	Timestamp time.Time // 接收时间
}
