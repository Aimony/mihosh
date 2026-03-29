package tui

import (
	"github.com/aimony/mihosh/internal/app/service"
	"github.com/aimony/mihosh/internal/infrastructure/config"
	"github.com/aimony/mihosh/internal/ui/tui/pages"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

// SettingsState 设置页面完整状态
type SettingsState struct {
	selectedSetting int
	editMode        bool
	editValue       string
	editCursor      int
}

// ToPageState 转换为渲染层所需的 SettingsPageState
func (s SettingsState) ToPageState(cfg *config.Config) pages.SettingsPageState {
	return pages.SettingsPageState{
		Config:          cfg,
		SelectedSetting: s.selectedSetting,
		EditMode:        s.editMode,
		EditValue:       s.editValue,
		EditCursor:      s.editCursor,
	}
}

// Update 处理设置页面按键，返回：(新状态, 更新后的cfg, 更新后的proxyAddr, cmd)
// proxyAddr 为空字符串时表示无变化
func (s SettingsState) Update(msg tea.KeyMsg, cfg *config.Config, configSvc *service.ConfigService) (SettingsState, *config.Config, string, tea.Cmd) {
	if s.editMode {
		return s.handleEditMode(msg, cfg, configSvc)
	}

	switch {
	case key.Matches(msg, keys.Up):
		if s.selectedSetting > 0 {
			s.selectedSetting--
		}
	case key.Matches(msg, keys.Down):
		if s.selectedSetting < len(pages.SettingKeys)-1 {
			s.selectedSetting++
		}
	case key.Matches(msg, keys.Enter):
		s.editMode = true
		s.editValue = pages.GetSettingValue(cfg, s.selectedSetting)
		s.editCursor = len(s.editValue)
	}

	return s, cfg, "", nil
}

// HandleMouseScroll 鼠标滚轮处理
func (s SettingsState) HandleMouseScroll(up bool) SettingsState {
	if up {
		if s.selectedSetting > 0 {
			s.selectedSetting--
		}
	} else {
		if s.selectedSetting < len(pages.SettingKeys)-1 {
			s.selectedSetting++
		}
	}
	return s
}

// handleEditMode 处理编辑模式按键，返回更新后的 cfg 和 proxyAddr（空表示无变化）
func (s SettingsState) handleEditMode(msg tea.KeyMsg, cfg *config.Config, configSvc *service.ConfigService) (SettingsState, *config.Config, string, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		s.editMode = false
		s.editValue = ""
		s.editCursor = 0

	case key.Matches(msg, keys.Enter):
		settingKey := pages.SettingKeys[s.selectedSetting]
		if err := configSvc.SetConfigValue(settingKey, s.editValue); err != nil {
			// 保存失败：保持编辑模式，但不更新 cfg
			return s, cfg, "", nil
		}
		newCfg, _ := configSvc.LoadConfig()
		s.editMode = false
		s.editValue = ""
		s.editCursor = 0
		return s, newCfg, newCfg.ProxyAddress, nil

	case msg.String() == "left":
		if s.editCursor > 0 {
			s.editCursor--
		}

	case msg.String() == "right":
		if s.editCursor < len(s.editValue) {
			s.editCursor++
		}

	case key.Matches(msg, keys.Home):
		s.editCursor = 0

	case key.Matches(msg, keys.End):
		s.editCursor = len(s.editValue)

	case key.Matches(msg, keys.Backspace):
		if s.editCursor > 0 {
			s.editValue = s.editValue[:s.editCursor-1] + s.editValue[s.editCursor:]
			s.editCursor--
		}

	case key.Matches(msg, keys.Delete):
		if s.editCursor < len(s.editValue) {
			s.editValue = s.editValue[:s.editCursor] + s.editValue[s.editCursor+1:]
		}

	default:
		input := msg.String()
		if len(input) > 0 && (len(input) > 1 || (input[0] >= 32 && input[0] < 127)) {
			s.editValue = s.editValue[:s.editCursor] + input + s.editValue[s.editCursor:]
			s.editCursor += len(input)
		}
	}

	return s, cfg, "", nil
}
