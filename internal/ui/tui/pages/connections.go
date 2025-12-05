package pages

import (
	"fmt"
	"strings"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/pkg/utils"
	"github.com/charmbracelet/lipgloss"
)

// ConnectionsPageState 连接页面状态
type ConnectionsPageState struct {
	Connections *model.ConnectionsResponse
	Width       int
}

// RenderConnectionsPage 渲染连接监控页面
func RenderConnectionsPage(state ConnectionsPageState) string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700"))

	infoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00BFFF"))

	if state.Connections == nil {
		return "正在加载连接信息..."
	}

	// 统计information
	stats := fmt.Sprintf(
		"总连接: %s | 上传: %s | 下载: %s",
		infoStyle.Render(fmt.Sprintf("%d", len(state.Connections.Connections))),
		infoStyle.Render(utils.FormatBytes(state.Connections.UploadTotal)),
		infoStyle.Render(utils.FormatBytes(state.Connections.DownloadTotal)),
	)

	// 连接列表
	var connLines []string
	if len(state.Connections.Connections) == 0 {
		connLines = append(connLines, "  无活跃连接")
	} else {
		maxDisplay := 15
		for i, conn := range state.Connections.Connections {
			if i >= maxDisplay {
				remaining := len(state.Connections.Connections) - maxDisplay
				connLines = append(connLines, fmt.Sprintf("  ... 还有 %d 个连接", remaining))
				break
			}

			proxy := "DIRECT"
			if len(conn.Chains) > 0 {
				proxy = conn.Chains[len(conn.Chains)-1]
			}

			line := fmt.Sprintf(
				"  %s:%s → %s:%s via %s",
				conn.Metadata.SourceIP,
				conn.Metadata.SourcePort,
				conn.Metadata.Host,
				conn.Metadata.DestinationPort,
				infoStyle.Render(proxy),
			)
			connLines = append(connLines, line)
		}
	}

	helpText := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Render("[r]刷新")

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("连接监控"),
		"",
		stats,
		"",
		strings.Join(connLines, "\n"),
		"",
		helpText,
	)
}
