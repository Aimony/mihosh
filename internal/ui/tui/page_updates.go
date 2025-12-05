package tui

import (
	"github.com/aimony/mihosh/internal/ui/tui/pages"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// updateNodesPage 更新节点页面
func (m Model) updateNodesPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Up):
		if m.selectedProxy > 0 {
			m.selectedProxy--
		}

	case key.Matches(msg, keys.Down):
		if m.selectedProxy < len(m.currentProxies)-1 {
			m.selectedProxy++
		}

	case key.Matches(msg, keys.Left):
		if m.selectedGroup > 0 {
			m.selectedGroup--
			m.updateCurrentProxies()
			m.selectedProxy = 0
		}

	case key.Matches(msg, keys.Right):
		if m.selectedGroup < len(m.groupNames)-1 {
			m.selectedGroup++
			m.updateCurrentProxies()
			m.selectedProxy = 0
		}

	case key.Matches(msg, keys.Enter):
		if len(m.currentProxies) > 0 && m.selectedProxy < len(m.currentProxies) {
			groupName := m.groupNames[m.selectedGroup]
			proxyName := m.currentProxies[m.selectedProxy]
			return m, selectProxy(m.client, groupName, proxyName)
		}

	case key.Matches(msg, keys.Test):
		if len(m.currentProxies) > 0 && m.selectedProxy < len(m.currentProxies) {
			proxyName := m.currentProxies[m.selectedProxy]
			m.testing = true
			return m, testProxy(m.client, proxyName, m.testURL, m.timeout)
		}

	case key.Matches(msg, keys.TestAll):
		if len(m.groupNames) > 0 && len(m.currentProxies) > 0 {
			m.testing = true
			m.testFailures = []string{} // 清空之前的失败记录
			return m, testAllProxies(m.client, m.currentProxies, m.testURL, m.timeout)
		}
	}

	return m, nil
}

// updateConnectionsPage 更新连接页面
func (m Model) updateConnectionsPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Refresh):
		return m, fetchConnections(m.client)
	}
	return m, nil
}

// updateSettingsPage 更新设置页面
func (m Model) updateSettingsPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 如果在编辑模式,交给编辑处理器
	if m.editMode {
		return m.handleEditMode(msg)
	}

	switch {
	case key.Matches(msg, keys.Up):
		if m.selectedSetting > 0 {
			m.selectedSetting--
		}

	case key.Matches(msg, keys.Down):
		if m.selectedSetting < len(pages.SettingKeys)-1 {
			m.selectedSetting++
		}

	case key.Matches(msg, keys.Enter):
		// 进入编辑模式
		m.editMode = true
		m.editValue = pages.GetSettingValue(m.config, m.selectedSetting)
		m.editCursor = len(m.editValue) // 光标初始化到末尾
		return m, nil
	}

	return m, nil
}

// handleEditMode 处理编辑模式的按键
func (m Model) handleEditMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		m.editMode = false
		m.editValue = ""
		m.editCursor = 0
		return m, nil

	case key.Matches(msg, keys.Enter):
		// 保存配置
		key := pages.SettingKeys[m.selectedSetting]
		if err := m.configSvc.SetConfigValue(key, m.editValue); err != nil {
			m.err = err
		} else {
			// 重新加载配置
			cfg, _ := m.configSvc.LoadConfig()
			m.config = cfg
			m.editMode = false
			m.editValue = ""
			m.editCursor = 0
		}
		return m, nil

	case key.Matches(msg, keys.Left):
		// 光标左移
		if m.editCursor > 0 {
			m.editCursor--
		}

	case key.Matches(msg, keys.Right):
		// 光标右移
		if m.editCursor < len(m.editValue) {
			m.editCursor++
		}

	case key.Matches(msg, keys.Home):
		// 光标移到开头
		m.editCursor = 0

	case key.Matches(msg, keys.End):
		// 光标移到末尾
		m.editCursor = len(m.editValue)

	case key.Matches(msg, keys.Backspace):
		// 删除光标前一个字符
		if m.editCursor > 0 {
			m.editValue = m.editValue[:m.editCursor-1] + m.editValue[m.editCursor:]
			m.editCursor--
		}

	case key.Matches(msg, keys.Delete):
		// 删除光标后一个字符
		if m.editCursor < len(m.editValue) {
			m.editValue = m.editValue[:m.editCursor] + m.editValue[m.editCursor+1:]
		}

	default:
		// 处理字符输入和粘贴
		input := msg.String()
		// 过滤掉控制字符，但允许多字符粘贴
		if len(input) > 0 && (len(input) > 1 || (input[0] >= 32 && input[0] < 127)) {
			// 在光标位置插入
			m.editValue = m.editValue[:m.editCursor] + input + m.editValue[m.editCursor:]
			m.editCursor += len(input)
		}
	}

	return m, nil
}

// updateHelpPage 更新帮助页面
func (m Model) updateHelpPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 帮助页面没有特殊交互，只响应全局快捷键
	return m, nil
}
