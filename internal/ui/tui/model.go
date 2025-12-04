package tui

import (
	"github.com/aimony/mihomo-cli/internal/app/service"
	"github.com/aimony/mihomo-cli/internal/domain/model"
	"github.com/aimony/mihomo-cli/internal/infrastructure/api"
	"github.com/aimony/mihomo-cli/internal/infrastructure/config"
	"github.com/aimony/mihomo-cli/internal/ui/tui/components"
	"github.com/charmbracelet/bubbles/key"
)

// Model TUI 主模型
type Model struct {
	client          *api.Client
	config          *config.Config
	proxySvc        *service.ProxyService
	configSvc       *service.ConfigService
	connSvc         *service.ConnectionService
	groups          map[string]model.Group
	proxies         map[string]model.Proxy
	connections     *model.ConnectionsResponse
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
	currentPage     components.PageType
	editMode        bool
	editValue       string
	testFailures    []string // 记录测速失败的节点
}

// 消息类型
type (
	groupsMsg       map[string]model.Group
	proxiesMsg      map[string]model.Proxy
	connectionsMsg  *model.ConnectionsResponse
	errMsg          error
	testDoneMsg     struct {
		name  string
		delay int
		err   error
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
	
	// 创建服务实例
	proxySvc := service.NewProxyService(client, testURL, timeout)
	configSvc := service.NewConfigService()
	connSvc := service.NewConnectionService(client)

	return Model{
		client:      client,
		config:      cfg,
		proxySvc:    proxySvc,
		configSvc:   configSvc,
		connSvc:     connSvc,
		testURL:     testURL,
		timeout:     timeout,
		currentPage: components.PageNodes,
	}
}
