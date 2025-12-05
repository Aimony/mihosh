package pages

import (
	"fmt"
	"strings"

	"github.com/aimony/mihosh/internal/infrastructure/config"
	"github.com/aimony/mihosh/pkg/utils"
	"github.com/charmbracelet/lipgloss"
)

var SettingKeys = []string{"api-address", "secret", "test-url", "timeout"}
var SettingLabels = []string{"API 地址", "密钥", "测速URL", "超时(ms)"}

// SettingsPageState 设置页面状态
type SettingsPageState struct {
	Config          *config.Config
	SelectedSetting int
	EditMode        bool
	EditValue       string
}

// GetSettingValue 获取配置值
func GetSettingValue(cfg *config.Config, index int) string {
	if cfg == nil {
		return ""
	}

	switch index {
	case 0:
		return cfg.APIAddress
	case 1:
		return cfg.Secret
	case 2:
		return cfg.TestURL
	case 3:
		return fmt.Sprintf("%d", cfg.Timeout)
	}
	return ""
}

// RenderSettingsPage 渲染设置页面
func RenderSettingsPage(state SettingsPageState) string {
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
	for i, label := range SettingLabels {
		prefix := "  "
		if i == state.SelectedSetting {
			prefix = "► "
		}

		value := GetSettingValue(state.Config, i)

		// 密钥特殊处理
		if i == 1 && value != "" {
			value = utils.MaskSecret(value)
		}

		// 如果正在编辑此项
		if state.EditMode && i == state.SelectedSetting {
			value = editStyle.Render(state.EditValue + "▋")
		}

		line := fmt.Sprintf("%s%s: %s", prefix, label, value)

		if i == state.SelectedSetting {
			line = selectedStyle.Render(line)
		} else {
			line = normalStyle.Render(line)
		}

		lines = append(lines, line)
	}

	// 操作提示
	var helpText string
	if state.EditMode {
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
