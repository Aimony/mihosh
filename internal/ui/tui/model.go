package tui

import (
	"context"

	"github.com/aimony/mihosh/internal/app/service"
	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/infrastructure/api"
	"github.com/aimony/mihosh/internal/infrastructure/config"
	"github.com/aimony/mihosh/internal/ui/tui/components"
	"github.com/aimony/mihosh/internal/ui/tui/components/common"
	"github.com/charmbracelet/bubbles/key"
)

// Model TUI 主模型（仅保留全局共享状态）
type Model struct {
	// 基础设施
	client    *api.Client
	config    *config.Config
	proxySvc  *service.ProxyService
	configSvc *service.ConfigService
	connSvc   *service.ConnectionService

	// 路由与布局
	currentPage components.PageType
	width       int
	height      int
	showHelp    bool

	// 测速参数（供 NodesState 使用）
	testURL string
	timeout int

	// 共享图表数据（Connections 页面和 StatusBar 共用）
	chartData *model.ChartData

	// 全局错误（状态栏显示）
	err error

	// WebSocket
	wsClient  *api.WSClient
	wsMsgChan chan interface{}
	wsCtx     context.Context
	wsCancel  context.CancelFunc

	// 五个页面子状态
	nodesState    NodesState
	connsState    ConnsState
	logsState     LogsState
	rulesState    RulesState
	settingsState SettingsState
}

// 消息类型
type (
	groupsMsg struct {
		groups       map[string]model.Group
		orderedNames []string
	}
	proxiesMsg     map[string]model.Proxy
	connectionsMsg *model.ConnectionsResponse
	errMsg         error
	testDoneMsg    struct {
		name  string
		delay int
		err   error
	}
	configSavedMsg          struct{}
	connectionClosedMsg     struct{ id string }
	allConnectionsClosedMsg struct{}
	testAllDoneMsg          struct {
		results map[string]int
	}
)

// 快捷键定义
type keyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Enter     key.Binding
	Test      key.Binding
	TestAll   key.Binding
	Quit      key.Binding
	Refresh   key.Binding
	NextPage  key.Binding
	PrevPage  key.Binding
	Page1     key.Binding
	Page2     key.Binding
	Page3     key.Binding
	Page4     key.Binding
	Page5     key.Binding
	Escape    key.Binding
	Save      key.Binding
	Backspace key.Binding
	Delete    key.Binding
	Home      key.Binding
	End       key.Binding
	Clear     key.Binding
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
		key.WithHelp("3", "日志"),
	),
	Page4: key.NewBinding(
		key.WithKeys("4"),
		key.WithHelp("4", "规则"),
	),
	Page5: key.NewBinding(
		key.WithKeys("5"),
		key.WithHelp("5", "设置"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "取消"),
	),
	Save: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "保存"),
	),
	Backspace: key.NewBinding(
		key.WithKeys("backspace", "ctrl+h"),
		key.WithHelp("backspace", "删除"),
	),
	Delete: key.NewBinding(
		key.WithKeys("delete"),
		key.WithHelp("delete", "删除后字符"),
	),
	Home: key.NewBinding(
		key.WithKeys("home", "ctrl+a"),
		key.WithHelp("home", "行首"),
	),
	End: key.NewBinding(
		key.WithKeys("end", "ctrl+e"),
		key.WithHelp("end", "行尾"),
	),
	Clear: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "清空"),
	),
}

// NewModel 创建新的 TUI 模型
func NewModel(client *api.Client, testURL string, timeout int) Model {
	cfg, err := config.Load()
	if err != nil || cfg == nil {
		cfg = &config.DefaultConfig
	}

	proxySvc := service.NewProxyService(client, testURL, timeout)
	configSvc := service.NewConfigService()
	connSvc := service.NewConnectionService(client)

	wsClient := api.NewWSClient(cfg.APIAddress, cfg.Secret)
	wsCtx, wsCancel := context.WithCancel(context.Background())

	return Model{
		client:        client,
		config:        cfg,
		proxySvc:      proxySvc,
		configSvc:     configSvc,
		connSvc:       connSvc,
		testURL:       testURL,
		timeout:       timeout,
		currentPage:   components.PageNodes,
		chartData:     model.NewChartData(common.ChartPoints),
		wsClient:      wsClient,
		wsMsgChan:     make(chan interface{}, common.WSMsgChanCap),
		wsCtx:         wsCtx,
		wsCancel:      wsCancel,
		nodesState:    NodesState{},
		connsState:    NewConnsState(cfg.ProxyAddress, model.DefaultSiteTests()),
		logsState:     NewLogsState(),
		rulesState:    RulesState{},
		settingsState: SettingsState{},
	}
}
