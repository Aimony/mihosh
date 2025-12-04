package ui

import (
	"fmt"
	"strings"

	"github.com/aimony/mihomo-cli/config"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ==================== 设置页面 ====================

var settingKeys = []string{"api-address", "secret", "test-url", "timeout"}
var settingLabels = []string{"API 地址", "密钥", "测速URL", "超时(ms)"}

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
		if m.selectedSetting < len(settingKeys)-1 {
			m.selectedSetting++
		}

	case key.Matches(msg, keys.Enter):
		// 进入编辑模式
		m.editMode = true
		m.editValue = m.getSettingValue(m.selectedSetting)
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
		return m, nil

	case key.Matches(msg, keys.Enter):
		// 保存配置
		key := settingKeys[m.selectedSetting]
		if err := config.Set(key, m.editValue); err != nil {
			m.err = err
		} else {
			// 重新加载配置
			cfg, _ := config.Load()
			m.config = cfg
			m.editMode = false
			m.editValue = ""
		}
		return m, nil

	case msg.String() == "backspace":
		if len(m.editValue) > 0 {
			m.editValue = m.editValue[:len(m.editValue)-1]
		}

	default:
		// 添加字符到编辑值
		if len(msg.String()) == 1 {
			m.editValue += msg.String()
		}
	}

	return m, nil
}

// getSettingValue 获取配置值
func (m Model) getSettingValue(index int) string {
	if m.config == nil {
		return ""
	}

	switch index {
	case 0:
		return m.config.APIAddress
	case 1:
		return m.config.Secret
	case 2:
		return m.config.TestURL
	case 3:
		return fmt.Sprintf("%d", m.config.Timeout)
	}
	return ""
}

// renderSettingsPage 渲染设置页面
func (m Model) renderSettingsPage() string {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	normalStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF"))

	editStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFF00")).
		Background(lipgloss.Color("#333")).
		Padding(0, 1)

	// 配置项列表
	var lines []string
	for i, label := range settingLabels {
		prefix := "  "
		if i == m.selectedSetting {
			prefix = "► "
		}

		value := m.getSettingValue(i)
		
		// 密钥特殊处理
		if i == 1 && value != "" {
			if len(value) <= 4 {
				value = "****"
			} else {
				value = value[:2] + "****" + value[len(value)-2:]
			}
		}

		// 如果正在编辑此项
		if m.editMode && i == m.selectedSetting {
			value = editStyle.Render(m.editValue + "▋")
		}

		line := fmt.Sprintf("%s%s: %s", prefix, label, value)

		if i == m.selectedSetting {
			line = selectedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}

		lines = append(lines, line)
	}

	// 操作提示
	var helpText string
	if m.editMode {
		helpText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888")).
			Render("[Enter]保存 [Esc]取消")
	} else {
		helpText = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888")).
			Render("[↑/↓]选择 [Enter]编辑")
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		headerStyle.Render("设置"),
		"",
		strings.Join(lines, "\n"),
		"",
		"",
		helpText,
	)
}

// ==================== 帮助页面 ====================

// updateHelpPage 更新帮助页面
func (m Model) updateHelpPage(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// 帮助页面没有特殊交互，只响应全局快捷键
	return m, nil
}

// renderHelpPage 渲染帮助页面
func (m Model) renderHelpPage() string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700"))

	sectionStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00BFFF"))

	keyStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00"))

	descStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC"))

	helpContent := []string{
		titleStyle.Render("Mihomo CLI 使用帮助"),
		"",
		"",
		sectionStyle.Render("全局快捷键:"),
		"  " + keyStyle.Render("Tab      ") + " - " + descStyle.Render("下一页"),
		"  " + keyStyle.Render("Shift+Tab") + " - " + descStyle.Render("上一页"),
		"  " + keyStyle.Render("1-4      ") + " - " + descStyle.Render("快速跳转页面"),
		"  " + keyStyle.Render("r        ") + " - " + descStyle.Render("刷新当前页面"),
		"  " + keyStyle.Render("q        ") + " - " + descStyle.Render("退出程序"),
		"",
		"",
		sectionStyle.Render("节点管理页面 [1]:"),
		"  " + keyStyle.Render("↑/↓ 或 k/j") + " - " + descStyle.Render("选择节点"),
		"  " + keyStyle.Render("←/→ 或 h/l") + " - " + descStyle.Render("切换策略组"),
		"  " + keyStyle.Render("Enter     ") + " - " + descStyle.Render("切换到选中节点"),
		"  " + keyStyle.Render("t         ") + " - " + descStyle.Render("测速当前节点"),
		"  " + keyStyle.Render("a         ") + " - " + descStyle.Render("测速当前组所有节点"),
		"",
		"",
		sectionStyle.Render("连接监控页面 [2]:"),
		"  " + keyStyle.Render("r         ") + " - " + descStyle.Render("刷新连接列表"),
		"",
		"",
		sectionStyle.Render("设置页面 [3]:"),
		"  " + keyStyle.Render("↑/↓      ") + " - " + descStyle.Render("选择配置项"),
		"  " + keyStyle.Render("Enter     ") + " - " + descStyle.Render("编辑配置项"),
		"  " + keyStyle.Render("Esc       ") + " - " + descStyle.Render("取消编辑"),
		"",
		"",
		sectionStyle.Render("延迟颜色说明:"),
		"  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).Render("●") + " " + descStyle.Render("绿色 - 小于200ms"),
		"  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFF00")).Render("●") + " " + descStyle.Render("黄色 - 200-500ms"),
		"  " + lipgloss.NewStyle().Foreground(lipgloss.Color("#FF0000")).Render("●") + " " + descStyle.Render("红色 - 大于500ms"),
		"",
		"",
		descStyle.Render("💡 提示: 所有命令行功能都可以在这个TUI界面中完成！"),
	}

	return strings.Join(helpContent, "\n")
}
