package connections

import (
	"fmt"
	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/pkg/utils"
	"github.com/charmbracelet/lipgloss"
)

// 表格列宽配置
const (
	ColWidthClose = 4
	ColWidthHost  = 22
	ColWidthType  = 12
	ColWidthRule  = 16
	ColWidthChain = 16
	ColWidthDL    = 10
	ColWidthUL    = 10
	ColWidthTime  = 8
)

// RenderTableHeader 渲染表头
func RenderTableHeader(style lipgloss.Style) string {
	header := fmt.Sprintf(
		"%-*s %-*s %-*s %-*s %-*s %-*s %-*s %-*s",
		ColWidthClose, "",
		ColWidthHost, "主机",
		ColWidthType, "类型",
		ColWidthRule, "规则",
		ColWidthChain, "代理链",
		ColWidthDL, "↓下载",
		ColWidthUL, "↑上传",
		ColWidthTime, "时长",
	)
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

	// 代理链
	chain := "DIRECT"
	if len(conn.Chains) > 0 {
		chain = conn.Chains[len(conn.Chains)-1]
	}
	chain = utils.TruncateString(chain, ColWidthChain)

	// 流量
	download := utils.FormatBytes(conn.Download)
	upload := utils.FormatBytes(conn.Upload)

	// 连接时长
	duration := utils.FormatDuration(conn.Start)

	row := fmt.Sprintf(
		"%s%s %s %s %s %s %s %s %s",
		prefix,
		utils.PadString("×", ColWidthClose-2),
		utils.PadString(host, ColWidthHost),
		utils.PadString(connType, ColWidthType),
		utils.PadString(rule, ColWidthRule),
		utils.PadString(chain, ColWidthChain),
		utils.PadString(download, ColWidthDL),
		utils.PadString(upload, ColWidthUL),
		utils.PadString(duration, ColWidthTime),
	)

	return style.Render(row)
}

// TruncateString 根据显示宽度截断字符串（支持中文）
// 暴露给外部使用，虽然 utils 中已有，但在 connections 包内部可能也需要
func TruncateString(s string, maxWidth int) string {
	return utils.TruncateString(s, maxWidth)
}
