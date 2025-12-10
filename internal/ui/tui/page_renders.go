package tui

import (
	"github.com/aimony/mihosh/internal/ui/tui/pages"
)

// renderNodesPage 渲染节点管理页面
func (m Model) renderNodesPage() string {
	state := pages.NodesPageState{
		Groups:         m.groups,
		Proxies:        m.proxies,
		GroupNames:     m.groupNames,
		SelectedGroup:  m.selectedGroup,
		SelectedProxy:  m.selectedProxy,
		CurrentProxies: m.currentProxies,
		Testing:        m.testing,
		TestFailures:   m.testFailures,
		Width:          m.width,
	}
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
		SelectedConnection: m.getSelectedConnection(),
	}
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
	return pages.RenderSettingsPage(state)
}

// renderHelpPage 渲染帮助页面
func (m Model) renderHelpPage() string {
	return pages.RenderHelpPage()
}
