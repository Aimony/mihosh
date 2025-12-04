package ui

import (
	"fmt"
	"strings"

	"github.com/aimony/mihomo-cli/api"
	"github.com/aimony/mihomo-cli/config"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PageType 页面类型
type PageType int

const (
	PageNodes PageType = iota
	PageConnections
	PageSettings
	PageHelp
)

// Model TUI 主模型
type Model struct {
	client          *api.Client
	config          *config.Config
	groups          map[string]api.Group
	proxies         map[string]api.Proxy
	connections     *api.ConnectionsResponse
	groupNames      []string
	selectedGroup   int
	selectedProxy   int
	selectedSetting int
	width           int
	height          int
	err             error
	testURL         string
	timeout         int
	testing         bool
	currentProxies  []string
	currentPage     PageType
	editMode        bool
	editValue       string
}

// 消息类型
type (
	groupsMsg       map[string]api.Group
	proxiesMsg      map[string]api.Proxy
	connectionsMsg  *api.ConnectionsResponse
	errMsg          error
	testDoneMsg     struct {
		name  string
		delay int
	}
	configSavedMsg struct{}
)

// 快捷键定义
type keyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Enter       key.Binding
	Test        key.Binding
	TestAll     key.Binding
	Quit        key.Binding
	Refresh     key.Binding
	NextPage    key.Binding
	PrevPage    key.Binding
	Page1       key.Binding
	Page2       key.Binding
	Page3       key.Binding
	Page4       key.Binding
	Escape      key.Binding
	Save        key.Binding
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "上"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "下"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "左"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "右"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "确认"),
	),
	Test: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "测速"),
	),
	TestAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "全测"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "刷新"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "退出"),
	),
	NextPage: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "下一页"),
	),
	PrevPage: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "上一页"),
	),
	Page1: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "节点"),
	),
	Page2: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "连接"),
	),
	Page3: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "设置"),
	),
	Page4: key.NewBinding(
		key.WithKeys("4"),
		key.WithHelp("4", "帮助"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "取消"),
	),
	Save: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "保存"),
	),
}

// NewModel 创建新的 TUI 模型
func NewModel(client *api.Client, testURL string, timeout int) Model {
	cfg, _ := config.Load()
	return Model{
		client:      client,
		config:      cfg,
		testURL:     testURL,
		timeout:     timeout,
		currentPage: PageNodes,
	}
}

// Init 初始化
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		fetchGroups(m.client),
		fetchProxies(m.client),
	)
}

// Update 更新
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// 编辑模式特殊处理
		if m.editMode {
			return m.handleEditMode(msg)
		}

		// 全局快捷键
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.NextPage):
			m.currentPage = (m.currentPage + 1) % 4
			return m, m.onPageChange()

		case key.Matches(msg, keys.PrevPage):
			m.currentPage = (m.currentPage + 3) % 4
			return m, m.onPageChange()

		case key.Matches(msg, keys.Page1):
			m.currentPage = PageNodes
			return m, m.onPageChange()

		case key.Matches(msg, keys.Page2):
			m.currentPage = PageConnections
			return m, fetchConnections(m.client)

		case key.Matches(msg, keys.Page3):
			m.currentPage = PageSettings
			return m, nil

		case key.Matches(msg, keys.Page4):
			m.currentPage = PageHelp
			return m, nil

		case key.Matches(msg, keys.Refresh):
			return m, m.refreshCurrentPage()
		}

		// 页面特定快捷键
		switch m.currentPage {
		case PageNodes:
			return m.updateNodesPage(msg)
		case PageConnections:
			return m.updateConnectionsPage(msg)
		case PageSettings:
			return m.updateSettingsPage(msg)
		case PageHelp:
			return m.updateHelpPage(msg)
		}

	case groupsMsg:
		m.groups = msg
		m.groupNames = make([]string, 0, len(msg))
		for name := range msg {
			m.groupNames = append(m.groupNames, name)
		}
		m.updateCurrentProxies()

	case proxiesMsg:
		m.proxies = msg

	case connectionsMsg:
		m.connections = msg

	case testDoneMsg:
		m.testing = false
		return m, fetchProxies(m.client)

	case configSavedMsg:
		m.editMode = false
		m.err = nil

	case errMsg:
		m.err = msg
		m.testing = false
	}

	return m, nil
}

// onPageChange 页面切换时的处理
func (m Model) onPageChange() tea.Cmd {
	m.err = nil
	switch m.currentPage {
	case PageConnections:
		return fetchConnections(m.client)
	}
	return nil
}

// refreshCurrentPage 刷新当前页面
func (m Model) refreshCurrentPage() tea.Cmd {
	switch m.currentPage {
	case PageNodes:
		return tea.Batch(fetchGroups(m.client), fetchProxies(m.client))
	case PageConnections:
		return fetchConnections(m.client)
	case PageSettings:
		cfg, _ := config.Load()
		m.config = cfg
		return nil
	}
	return nil
}

// View 渲染视图
func (m Model) View() string {
	if m.width == 0 {
		return "初始化中..."
	}

	// 渲染标签栏
	tabs := m.renderTabs()

	// 渲染内容区域
	var content string
	switch m.currentPage {
	case PageNodes:
		content = m.renderNodesPage()
	case PageConnections:
		content = m.renderConnectionsPage()
	case PageSettings:
		content = m.renderSettingsPage()
	case PageHelp:
		content = m.renderHelpPage()
	}

	// 渲染状态栏
	statusBar := m.renderStatusBar()

	// 组合布局
	return lipgloss.JoinVertical(
		lipgloss.Left,
		tabs,
		"",
		content,
		"",
		statusBar,
	)
}

// renderTabs 渲染标签栏
func (m Model) renderTabs() string {
	activeTabStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00BFFF")).
		Background(lipgloss.Color("#333")).
		Padding(0, 2)

	inactiveTabStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Padding(0, 2)

	tabs := []string{
		"[1] 节点管理",
		"[2] 连接监控",
		"[3] 设置",
		"[4] 帮助",
	}

	var rendered []string
	for i, tab := range tabs {
		if PageType(i) == m.currentPage {
			rendered = append(rendered, activeTabStyle.Render("● "+tab))
		} else {
			rendered = append(rendered, inactiveTabStyle.Render("  "+tab))
		}
	}

	tabBar := lipgloss.JoinHorizontal(lipgloss.Top, rendered...)
	
	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666")).
		Render(strings.Repeat("─", m.width))

	return lipgloss.JoinVertical(lipgloss.Left, tabBar, divider)
}

// renderStatusBar 渲染状态栏
func (m Model) renderStatusBar() string {
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888"))

	errorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF0000"))

	var status string
	if m.err != nil {
		status = errorStyle.Render(fmt.Sprintf("错误: %v", m.err))
	} else if m.testing {
		status = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Render("⏳ 测速中...")
	} else {
		// 显示连接状态
		status = statusStyle.Render("●连接正常")
	}

	// 按 ? 显示帮助
	helpHint := statusStyle.Render(" | 按Tab切换页面 | 按数字键快速跳转 | 按q退出")

	divider := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666")).
		Render(strings.Repeat("─", m.width))

	statusBar := status + helpHint

	return lipgloss.JoinVertical(lipgloss.Left, divider, statusBar)
}

// updateCurrentProxies 更新当前显示的代理列表
func (m *Model) updateCurrentProxies() {
	if len(m.groupNames) > 0 && m.selectedGroup < len(m.groupNames) {
		groupName := m.groupNames[m.selectedGroup]
		if group, ok := m.groups[groupName]; ok {
			m.currentProxies = group.All
		}
	}
}

// getDelayColor 根据延迟获取颜色
func (m Model) getDelayColor(delay int) lipgloss.Color {
	if delay < 200 {
		return lipgloss.Color("#00FF00") // 绿色
	} else if delay < 500 {
		return lipgloss.Color("#FFFF00") // 黄色
	}
	return lipgloss.Color("#FF0000") // 红色
}

// 命令函数
func fetchGroups(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		groups, err := client.GetGroups()
		if err != nil {
			return errMsg(err)
		}
		return groupsMsg(groups)
	}
}

func fetchProxies(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		proxies, err := client.GetProxies()
		if err != nil {
			return errMsg(err)
		}
		return proxiesMsg(proxies)
	}
}

func fetchConnections(client *api.Client) tea.Cmd {
	return func() tea.Msg {
		conns, err := client.GetConnections()
		if err != nil {
			return errMsg(err)
		}
		return connectionsMsg(conns)
	}
}

func selectProxy(client *api.Client, group, proxy string) tea.Cmd {
	return func() tea.Msg {
		if err := client.SelectProxy(group, proxy); err != nil {
			return errMsg(err)
		}
		return fetchProxies(client)()
	}
}

func testProxy(client *api.Client, name, testURL string, timeout int) tea.Cmd {
	return func() tea.Msg {
		delay, err := client.TestProxyDelay(name, testURL, timeout)
		if err != nil {
			return errMsg(err)
		}
		return testDoneMsg{name: name, delay: delay}
	}
}

func testGroup(client *api.Client, group, testURL string, timeout int) tea.Cmd {
	return func() tea.Msg {
		if err := client.TestGroupDelay(group, testURL, timeout); err != nil {
			return errMsg(err)
		}
		return testDoneMsg{}
	}
}

func saveConfig(cfg *config.Config) tea.Cmd {
	return func() tea.Msg {
		if err := config.Save(cfg); err != nil {
			return errMsg(err)
		}
		return configSavedMsg{}
	}
}
