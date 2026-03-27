package tui

import (
	"github.com/aimony/mihosh/internal/ui/tui/pages"
)

// renderNodesPage 渲染节点管理页面
func (m Model) renderNodesPage() string {
	state := pages.NodesPageState{
		Groups:            m.groups,
		Proxies:           m.proxies,
		GroupNames:        m.groupNames,
		SelectedGroup:     m.selectedGroup,
		SelectedProxy:     m.selectedProxy,
		CurrentProxies:    m.currentProxies,
		Testing:           m.testing,
		TestFailures:      m.testFailures,
		ShowFailureDetail: m.showFailureDetail,
		Width:             m.width,
		Height:            m.height,
		GroupScrollTop:    m.groupScrollTop,
		ProxyScrollTop:    m.proxyScrollTop,
	}
	pageHeight := m.height - 8 // 减去状态栏、边框、标题等占用
	state.Height = pageHeight
	return pages.RenderNodesPage(state)
}

// renderConnectionsPage 渲染连接监控页面
func (m Model) renderConnectionsPage() string {
	state := pages.ConnectionsPageState{
		Connections:        m.connections,
		Width:              m.width,
		Height:             m.height,
		SelectedIndex:      m.selectedConn,
		ScrollTop:          m.connScrollTop,
		FilterText:         m.connFilter,
		FilterMode:         m.connFilterMode,
		DetailMode:         m.connDetailMode,
		SelectedConnection: m.connDetailSnapshot, // 使用快照
		IPInfo:             m.connIPInfo,         // IP地理信息
		DetailScroll:       m.connDetailScroll,   // 详情滚动偏移
		ChartData:          m.chartData,          // 图表数据
		ViewMode:           m.connViewMode,       // 0=活跃, 1=历史
		ClosedConnections:  m.closedConnections,  // 历史连接
		SiteTests:          m.siteTests,          // 网站测速数据
		SelectedSiteTest:   m.selectedSiteTest,   // 选中的网站索引
	}
	pageHeight := m.height - 8 // 减去状态栏、边框、标题等占用
	state.Height = pageHeight
	return pages.RenderConnectionsPage(state)
}

// renderSettingsPage 渲染设置页面
func (m Model) renderSettingsPage() string {
	state := pages.SettingsPageState{
		Config:          m.config,
		SelectedSetting: m.selectedSetting,
		EditMode:        m.editMode,
		EditValue:       m.editValue,
		EditCursor:      m.editCursor,
	}
	pageHeight := m.height - 8 // 减去状态栏、边框、标题等占用
	return pages.RenderSettingsPage(state, m.width, pageHeight)
}

// renderHelpPage 渲染帮助页面
func (m Model) renderHelpPage() string {
	// 帮助页面作为一个弹窗，可以使用较大的宽度，但高度仍需限制在终端范围内
	return pages.RenderHelpPage(m.width, m.height)
}

// renderLogsPage 渲染日志页面
func (m Model) renderLogsPage() string {
	state := pages.LogsPageState{
		Logs:               m.logs,
		FilteredLogIndices: m.filteredLogIndices,
		LogLevel:           m.logLevel,
		FilterText:    m.logFilter,
		FilterMode:    m.logFilterMode,
		SelectedLog:   m.selectedLog,
		ScrollTop:     m.logScrollTop,
		HScrollOffset: m.logHScrollOffset,
		Width:         m.width,
		Height:        m.height - 8, // 减去状态栏、边框、标题等占用
	}
	return pages.RenderLogsPage(state)
}

// renderRulesPage 渲染规则页面
func (m Model) renderRulesPage() string {
	state := pages.RulesPageState{
		Rules:               m.rules,
		FilteredRuleIndices: m.filteredRuleIndices,
		FilterText:          m.ruleFilter,
		FilterMode:   m.ruleFilterMode,
		SelectedRule: m.selectedRule,
		ScrollTop:    m.ruleScrollTop,
		Width:        m.width,
		Height:       m.height - 8, // 减去状态栏、边框、标题等占用
	}
	return pages.RenderRulesPage(state)
}
