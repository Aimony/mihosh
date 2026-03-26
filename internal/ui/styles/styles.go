package styles

import "github.com/charmbracelet/lipgloss"

// ============================================================
//  Tokyo Night Deep — 全局调色板
// ============================================================

// 表面 / 背景色
var (
	ColorBackground = lipgloss.Color("#1A1B26") // 深蓝灰背景
	ColorSurface    = lipgloss.Color("#24283b") // 稍亮的面板背景
	ColorOverlay    = lipgloss.Color("#414868") // 弹出层 / 悬浮层
)

// 主强调色
var (
	ColorPrimary   = lipgloss.Color("#7AA2F7") // 现代蓝
	ColorSecondary = lipgloss.Color("#BB9AF7") // 紫色 — 选中态
)

// 语义状态色
var (
	ColorSuccess = lipgloss.Color("#9ECE6A") // 延迟 <100ms
	ColorWarning = lipgloss.Color("#E0AF68") // 延迟 100-300ms
	ColorDanger  = lipgloss.Color("#F7768E") // 超时 / 离线
)

// 中性色
var (
	ColorBorder = lipgloss.Color("#414868") // 边框
	ColorGray   = lipgloss.Color("#565f89") // 次要文字
	ColorDim    = lipgloss.Color("#3b4261") // 最暗文字 / 禁用
	ColorText   = lipgloss.Color("#c0caf5") // 正文
	ColorBright = lipgloss.Color("#ffffff") // 高亮文字
)

// ============================================================
//  公共样式预设
// ============================================================

// 状态栏样式
var (
	StatusStyle = lipgloss.NewStyle().
			Foreground(ColorGray)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorDanger)

	TestingStyle = lipgloss.NewStyle().
			Foreground(ColorWarning)
)

// 分隔线
var DividerStyle = lipgloss.NewStyle().
	Foreground(ColorBorder)

// 标题
var (
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Padding(0, 1)

	SubtitleStyle = lipgloss.NewStyle().
			Foreground(ColorGray)
)

// 列表
var (
	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				Bold(true)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	DisabledItemStyle = lipgloss.NewStyle().
				Foreground(ColorGray)
)

// 表格
var (
	TableHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(ColorPrimary)

	TableRowStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	TableAltRowStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#a9b1d6"))
)

// 输入框
var (
	InputStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1)

	InputLabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorGray)
)
