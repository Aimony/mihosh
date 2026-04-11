package logs

import (
	"regexp"
)

// ParsedLog 解析后的日志结构
type ParsedLog struct {
	Protocol   string // TCP, UDP
	SourceIP   string
	SourcePort string
	DestHost   string // 域名或IP
	DestPort   string
	MatchRule  string
	ProxyChain string
	Raw        string
}

// logPattern 匹配 Mihomo 日志格式:
// [TCP] 172.18.0.6:39412 --> accounts.google.com:443 match DomainSuffix(google.com) using 节点选择 [新加坡 1]
var logPattern = regexp.MustCompile(`\[(\w+)\]\s+([\d.]+):(\d+)\s+-->\s+(.+?):(\d+)\s+match\s+(.+?)\s+using\s+(.+)`)

// ParseLogPayload 尽力解析日志 Payload（best-effort）
func ParseLogPayload(payload string) *ParsedLog {
	matches := logPattern.FindStringSubmatch(payload)
	if matches == nil {
		return &ParsedLog{Raw: payload}
	}
	return &ParsedLog{
		Protocol:   matches[1],
		SourceIP:   matches[2],
		SourcePort: matches[3],
		DestHost:   matches[4],
		DestPort:   matches[5],
		MatchRule:  matches[6],
		ProxyChain: matches[7],
		Raw:        payload,
	}
}
