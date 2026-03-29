package common

import "github.com/charmbracelet/lipgloss"

// 基础调色板 (主要从现有代码中提取)
const (
	ColorPrimary   = "#1E90FF" // 蓝色 (DodgerBlue)
	ColorSecondary = "#00BFFF" // 深天蓝 (DeepSkyBlue)
	ColorSuccess   = "#2ECC71" // 绿色 (Emerald)
	ColorWarning   = "#FFD700" // 黄色 (Gold)
	ColorDanger    = "#E74C3C" // 红色 (Alizarin)
	ColorInfo      = "#00CED1" // 青色 (DarkTurquoise)
	ColorMuted     = "#888888" // 灰色 (中灰)
	ColorDim       = "#555555" // 灰色 (深灰)
	ColorHighlight = "#333333" // 背景高亮色
	ColorWhite     = "#FFFFFF"
	ColorActive    = "#00FF00" // 激活态 (绿色)
	ColorOrange    = "#E67E22" // 橙色
	ColorPurple    = "#9B59B6" // 紫色
	ColorGray      = "#95A5A6" // 石棉灰
	ColorBorder    = "#666666" // 边框灰
)

// LipGloss 颜色对象
var (
	CPrimary   = lipgloss.Color(ColorPrimary)
	CSecondary = lipgloss.Color(ColorSecondary)
	CSuccess   = lipgloss.Color(ColorSuccess)
	CWarning   = lipgloss.Color(ColorWarning)
	CDanger    = lipgloss.Color(ColorDanger)
	CInfo      = lipgloss.Color(ColorInfo)
	CMuted     = lipgloss.Color(ColorMuted)
	CDim       = lipgloss.Color(ColorDim)
	CHighlight = lipgloss.Color(ColorHighlight)
	CWhite     = lipgloss.Color(ColorWhite)
	CActive    = lipgloss.Color(ColorActive)
	COrange    = lipgloss.Color(ColorOrange)
	CPurple    = lipgloss.Color(ColorPurple)
	CGray      = lipgloss.Color(ColorGray)
	CBorder    = lipgloss.Color(ColorBorder)
)
