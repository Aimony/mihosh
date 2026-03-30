package common

import "github.com/charmbracelet/lipgloss"

// 基础 UI 样式定义
var (
	// Tab 渲染
	TabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Background(CSecondary).
			Foreground(CWhite).
			Padding(0, 1)

	TabInactiveStyle = lipgloss.NewStyle().
				Background(CHighlight).
				Foreground(CMuted).
				Padding(0, 1)

	// 文本样式
	BoldStyle      = lipgloss.NewStyle().Bold(true)
	ActiveStyle    = lipgloss.NewStyle().Foreground(CActive).Bold(true)
	InactiveStyle  = lipgloss.NewStyle().Foreground(CPrimary)
	DimStyle       = lipgloss.NewStyle().Foreground(CDim)
	MutedStyle     = lipgloss.NewStyle().Foreground(CMuted)
	HighlightStyle = lipgloss.NewStyle().Background(CHighlight).Foreground(CWhite)
	ErrorStyle     = lipgloss.NewStyle().Foreground(CDanger)
	SuccessStyle   = lipgloss.NewStyle().Foreground(CSuccess)
	WarningStyle   = lipgloss.NewStyle().Foreground(CWarning)

	// 表格/列表样式
	TableHeaderStyle = lipgloss.NewStyle().Foreground(CMuted).Bold(true)
	TableBorderStyle = lipgloss.NewStyle().Foreground(CBorder)
	SelectedStyle    = lipgloss.NewStyle().Foreground(CActive).Bold(true)

	// 页面标题
	PageHeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(CWarning).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(CBorder)
)
