package ui

import (
	"fmt"
	"strings"

	"github.com/aimony/mihomo-cli/api"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model TUI 主模型
type Model struct {
	client          *api.Client
	groups          map[string]api.Group
	proxies         map[string]api.Proxy
	groupNames      []string
	selectedGroup   int
	selectedProxy   int
	width           int
	height          int
	err             error
	testURL         string
	timeout         int
	testing         bool
	currentProxies  []string
}

// 消息类型
type (
	groupsMsg   map[string]api.Group
	proxiesMsg  map[string]api.Proxy
	errMsg      error
	testDoneMsg struct {
		name  string
		delay int
	}
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
		key.WithHelp("←/h", "切换到策略组"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "切换到节点"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "选择节点"),
	),
	Test: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "测速当前节点"),
	),
	TestAll: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "测速所有节点"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "刷新"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "退出"),
	),
}

// NewModel 创建新的 TUI 模型
func NewModel(client *api.Client, testURL string, timeout int) Model {
	return Model{
		client:  client,
		testURL: testURL,
		timeout: timeout,
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
		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

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
			if len(m.groupNames) > 0 {
				groupName := m.groupNames[m.selectedGroup]
				m.testing = true
				return m, testGroup(m.client, groupName, m.testURL, m.timeout)
			}

		case key.Matches(msg, keys.Refresh):
			return m, tea.Batch(
				fetchGroups(m.client),
				fetchProxies(m.client),
			)
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

	case testDoneMsg:
		m.testing = false
		// 刷新代理信息
		return m, fetchProxies(m.client)

	case errMsg:
		m.err = msg
		m.testing = false
	}

	return m, nil
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

// View 渲染视图
func (m Model) View() string {
	if m.width == 0 {
		return "初始化中..."
	}

	// 样式定义
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#00BFFF")).
		Padding(0, 1)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FFD700")).
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(lipgloss.Color("#666"))

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)

	activeStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#1E90FF"))

	// 顶部标题
	title := titleStyle.Render("🚀 Mihomo CLI")
	
	// 策略组列表
	groupList := m.renderGroupList(selectedStyle, activeStyle)
	
	// 节点列表
	proxyList := m.renderProxyList(selectedStyle, activeStyle)

	// 帮助信息
	help := m.renderHelp()

	// 错误信息
	errMsg := ""
	if m.err != nil {
		errMsg = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Render(fmt.Sprintf("\n错误: %v", m.err))
	}

	// 测试状态
	testingMsg := ""
	if m.testing {
		testingMsg = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA500")).
			Render("\n⏳ 测速中...")
	}

	// 布局
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		"",
		headerStyle.Width(m.width-4).Render("策略组"),
		groupList,
		"",
		headerStyle.Width(m.width-4).Render("节点列表"),
		proxyList,
		errMsg,
		testingMsg,
		"",
		help,
	)

	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(content)
}

// renderGroupList 渲染策略组列表
func (m Model) renderGroupList(selectedStyle, activeStyle lipgloss.Style) string {
	if len(m.groupNames) == 0 {
		return "  正在加载..."
	}

	var lines []string
	for i, name := range m.groupNames {
		group := m.groups[name]
		prefix := "  "
		if i == m.selectedGroup {
			prefix = "► "
		}

		line := fmt.Sprintf("%s%s (%s) → %s", prefix, name, group.Type, group.Now)
		
		if i == m.selectedGroup {
			line = selectedStyle.Render(line)
		} else if group.Now != "" {
			line = activeStyle.Render(line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
}

// renderProxyList 渲染节点列表
func (m Model) renderProxyList(selectedStyle, activeStyle lipgloss.Style) string {
	if len(m.currentProxies) == 0 {
		return "  无可用节点"
	}

	var currentNode string
	if len(m.groupNames) > 0 && m.selectedGroup < len(m.groupNames) {
		groupName := m.groupNames[m.selectedGroup]
		if group, ok := m.groups[groupName]; ok {
			currentNode = group.Now
		}
	}

	var lines []string
	for i, name := range m.currentProxies {
		proxy, exists := m.proxies[name]
		
		prefix := "  "
		suffix := ""
		
		if i == m.selectedProxy {
			prefix = "► "
		}
		
		if name == currentNode {
			suffix = " ✓"
		}

		// 获取延迟信息
		delay := ""
		if exists && len(proxy.History) > 0 {
			lastDelay := proxy.History[len(proxy.History)-1].Delay
			if lastDelay > 0 {
				delayColor := m.getDelayColor(lastDelay)
				delay = lipgloss.NewStyle().
					Foreground(delayColor).
					Render(fmt.Sprintf(" (%dms)", lastDelay))
			}
		}

		line := fmt.Sprintf("%s%s%s%s", prefix, name, delay, suffix)
		
		if i == m.selectedProxy {
			line = selectedStyle.Render(line)
		} else if name == currentNode {
			line = activeStyle.Render(line)
		}

		lines = append(lines, line)
	}

	return strings.Join(lines, "\n")
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

// renderHelp 渲染帮助信息
func (m Model) renderHelp() string {
	help := []string{
		"[↑/↓] 选择",
		"[←/→] 切换组",
		"[Enter] 切换节点",
		"[t] 测速",
		"[a] 测速全部",
		"[r] 刷新",
		"[q] 退出",
	}

	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888")).
		Render(strings.Join(help, " | "))
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
