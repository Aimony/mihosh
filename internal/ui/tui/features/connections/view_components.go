package connections

import (
	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/ui/tui/features/connections/components"
	"github.com/charmbracelet/lipgloss"
)

const (
	connectionsBaseUsedLines     = 8
	connectionsMinDisplayRows    = 5
	connectionsSiteCardsTopLine  = 2
	connectionsSiteCardHeight    = 5
	connectionsSiteCardMinWidth  = 12
	connectionsSiteCardMaxWidth  = 20
	connectionsSiteLayoutColumns = 4
	connectionsSiteCardOuterPad  = 3
)

// MouseTarget 表示 connections 页面鼠标命中的组件
type MouseTarget int

const (
	ConnectionsMouseTargetNone MouseTarget = iota
	MouseTargetConnection
	MouseTargetSiteTest
)

// MouseHit 是 connections 页面鼠标命中结果
type MouseHit struct {
	Target MouseTarget
	Index  int
}

type connectionsListWindow struct {
	ScrollTop   int
	VisibleRows int
	ShowTopHint bool
}

// ResolveMouseHit 根据 pageContent 内的坐标定位命中的连接行/网站卡片。
func ResolveMouseHit(state PageState, pageX, pageY int) MouseHit {
	if pageX < 0 || pageY < 0 {
		return MouseHit{Target: ConnectionsMouseTargetNone, Index: -1}
	}
	if state.ViewMode == 0 && state.Connections == nil {
		return MouseHit{Target: ConnectionsMouseTargetNone, Index: -1}
	}

	line := 2 // 页面标题 + 空行

	if state.ViewMode == 0 {
		if state.ChartData != nil {
			chartsSection := components.RenderChartsSection(state.ChartData, state.Width)
			if chartsSection != "" {
				line += lipgloss.Height(chartsSection) + 1 // 图表区域 + 空行
			}
		}
		if len(state.SiteTests) > 0 {
			siteStart := line
			if idx := resolveSiteTestMouseHit(state, pageX, pageY-siteStart); idx >= 0 {
				return MouseHit{
					Target: MouseTargetSiteTest,
					Index:  idx,
				}
			}
			siteSection := components.RenderSiteTestSection(state.SiteTests, state.SelectedSiteTest, state.Width)
			line += lipgloss.Height(siteSection) + 1 // 网站测速区域 + 空行
		}
	}

	line++ // 统计行
	if state.FilterMode || state.FilterText != "" {
		line++ // 过滤行
	}
	line += 2 // 表头 + 分隔线

	filteredConns := filterConnections(connectionsByViewMode(state), state.FilterText)
	if len(filteredConns) == 0 {
		return MouseHit{Target: ConnectionsMouseTargetNone, Index: -1}
	}

	window := resolveConnectionsListWindow(state, len(filteredConns))
	dataStart := line
	if window.ShowTopHint {
		dataStart++
	}

	if pageY >= dataStart && pageY < dataStart+window.VisibleRows {
		return MouseHit{
			Target: MouseTargetConnection,
			Index:  window.ScrollTop + (pageY - dataStart),
		}
	}

	return MouseHit{Target: ConnectionsMouseTargetNone, Index: -1}
}

func resolveSiteTestMouseHit(state PageState, pageX int, siteSectionY int) int {
	if siteSectionY < connectionsSiteCardsTopLine || siteSectionY >= connectionsSiteCardsTopLine+connectionsSiteCardHeight {
		return -1
	}

	cardOuterWidth := calcSiteCardOuterWidth(state.Width)
	if cardOuterWidth <= 0 {
		return -1
	}

	idx := pageX / cardOuterWidth
	if idx < 0 || idx >= len(state.SiteTests) {
		return -1
	}
	return idx
}

func calcSiteCardOuterWidth(pageWidth int) int {
	cardWidth := (pageWidth - 10) / connectionsSiteLayoutColumns
	if cardWidth < connectionsSiteCardMinWidth {
		cardWidth = connectionsSiteCardMinWidth
	}
	if cardWidth > connectionsSiteCardMaxWidth {
		cardWidth = connectionsSiteCardMaxWidth
	}
	return cardWidth + connectionsSiteCardOuterPad
}

func resolveConnectionsListWindow(state PageState, total int) connectionsListWindow {
	if total <= 0 {
		return connectionsListWindow{}
	}

	selected := state.SelectedIndex
	if selected < 0 {
		selected = 0
	}
	if selected >= total {
		selected = total - 1
	}

	maxDisplay := calcConnectionsMaxDisplay(state)
	scrollTop := state.ScrollTop
	if scrollTop < 0 {
		scrollTop = 0
	}
	if scrollTop >= total {
		scrollTop = total - 1
	}

	if selected >= scrollTop+maxDisplay {
		scrollTop = selected - maxDisplay + 1
	}
	if selected < scrollTop {
		scrollTop = selected
	}

	endIdx := scrollTop + maxDisplay
	if endIdx > total {
		endIdx = total
	}

	return connectionsListWindow{
		ScrollTop:   scrollTop,
		VisibleRows: endIdx - scrollTop,
		ShowTopHint: scrollTop > 0,
	}
}

func calcConnectionsMaxDisplay(state PageState) int {
	usedLines := connectionsBaseUsedLines
	if state.ViewMode == 0 {
		if state.ChartData != nil {
			usedLines += 6
		}
		if len(state.SiteTests) > 0 {
			usedLines += 8
		}
	}
	if state.FilterMode || state.FilterText != "" {
		usedLines++
	}

	maxDisplay := state.Height - usedLines
	if maxDisplay < connectionsMinDisplayRows {
		maxDisplay = connectionsMinDisplayRows
	}
	return maxDisplay
}

func connectionsByViewMode(state PageState) []model.Connection {
	if state.ViewMode == 0 {
		if state.Connections == nil {
			return nil
		}
		return state.Connections.Connections
	}
	return state.ClosedConnections
}
