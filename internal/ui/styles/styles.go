package styles

import "github.com/charmbracelet/lipgloss"

// 颜色定义
var (
	ColorPrimary   = lipgloss.Color("#00BFFF")
	ColorSecondary = lipgloss.Color("#888")
	ColorSuccess   = lipgloss.Color("#00FF00")
	ColorWarning   = lipgloss.Color("#FFFF00")
	ColorDanger    = lipgloss.Color("#FF0000")
	ColorOrange    = lipgloss.Color("#FFA500")
	ColorGray      = lipgloss.Color("#666")
	ColorDarkBg    = lipgloss.Color("#333")
)

// Tab 样式
var (
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Background(ColorDarkBg).
			Padding(0, 2)

	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				Padding(0, 2)
)

// 状态栏样式
var (
	StatusStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorDanger)

	TestingStyle = lipgloss.NewStyle().
			Foreground(ColorOrange)
)

// 分隔线样式
var DividerStyle = lipgloss.NewStyle().
	Foreground(ColorGray)

// 标题样式
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)
)

// 列表样式
var (
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF"))

	DisabledItemStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary)
)

// 表格样式
var (
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary)

	TableRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFF"))

	TableAltRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#DDD"))
)

// 输入框样式
var (
	InputStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	InputLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary)
)
