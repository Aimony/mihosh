package components

import (
	"fmt"
	"strings"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/pkg/utils"
	"github.com/charmbracelet/lipgloss"
)

// 表格列宽配置
const (
	ColWidthClose = 4
	ColWidthHost  = 30
	ColWidthType  = 12
	ColWidthRule  = 18
	ColWidthChain = 20
	ColWidthDL    = 10
	ColWidthUL    = 10
	ColWidthTime  = 8
)

// RenderTableHeader 渲染表头
func RenderTableHeader(style lipgloss.Style) string {
	header := strings.Join([]string{
		alignCenter("", ColWidthClose),
		alignLeft("主机", ColWidthHost),
		alignLeft("类型", ColWidthType),
		alignLeft("规则", ColWidthRule),
		alignLeft("代理链", ColWidthChain),
		alignRight("↓下载", ColWidthDL),
		alignRight("↑上传", ColWidthUL),
		alignRight("时长", ColWidthTime),
	}, " ")
	return style.Render(header)
}

// RenderConnectionRow 渲染单行连接
func RenderConnectionRow(conn model.Connection, style lipgloss.Style, prefix string) string {
	// 主机
	host := conn.Metadata.Host
	if host == "" {
		host = conn.Metadata.DestinationIP
	}
	host = utils.TruncateString(host, ColWidthHost)

	// 类型
	connType := fmt.Sprintf("%s/%s", conn.Metadata.Network, conn.Metadata.Type)
	connType = utils.TruncateString(connType, ColWidthType)

	// 规则
	rule := conn.Rule
	if conn.RulePayload != "" && utils.DisplayWidth(rule)+utils.DisplayWidth(conn.RulePayload)+1 <= ColWidthRule {
		rule = fmt.Sprintf("%s:%s", rule, conn.RulePayload)
	}
	rule = utils.TruncateString(rule, ColWidthRule)

	// 代理链（chains 为倒序，反转后用 → 拼接显示完整链路）
	chain := "DIRECT"
	if len(conn.Chains) > 0 {
		parts := make([]string, len(conn.Chains))
		for i, c := range conn.Chains {
			parts[len(conn.Chains)-1-i] = c
		}
		chain = strings.Join(parts, " → ")
	}
	chain = utils.TruncateString(chain, ColWidthChain)

	// 流量
	download := utils.FormatBytes(conn.Download)
	upload := utils.FormatBytes(conn.Upload)

	// 连接时长
	duration := utils.FormatDuration(conn.Start)

	row := prefix + strings.Join([]string{
		alignCenter("×", ColWidthClose),
		alignLeft(host, ColWidthHost),
		alignLeft(connType, ColWidthType),
		alignLeft(rule, ColWidthRule),
		alignLeft(chain, ColWidthChain),
		alignRight(download, ColWidthDL),
		alignRight(upload, ColWidthUL),
		alignRight(duration, ColWidthTime),
	}, " ")

	return style.Render(row)
}

// TruncateString 根据显示宽度截断字符串（支持中文）
// 暴露给外部使用，虽然 utils 中已有，但在 connections 包内部可能也需要
func TruncateString(s string, maxWidth int) string {
	return utils.TruncateString(s, maxWidth)
}

func alignLeft(s string, width int) string {
	return utils.PadString(utils.TruncateString(s, width), width)
}

func alignRight(s string, width int) string {
	s = utils.TruncateString(s, width)
	padding := width - utils.DisplayWidth(s)
	if padding <= 0 {
		return s
	}
	return strings.Repeat(" ", padding) + s
}

func alignCenter(s string, width int) string {
	s = utils.TruncateString(s, width)
	padding := width - utils.DisplayWidth(s)
	if padding <= 0 {
		return s
	}

	left := padding / 2
	right := padding - left
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", right)
}
