package tui

import (
	"github.com/aimony/mihosh/internal/app/service"
	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/infrastructure/api"
	"github.com/aimony/mihosh/internal/infrastructure/config"
	"github.com/aimony/mihosh/internal/ui/tui/components"
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
	editCursor      int      // 编辑时光标位置
	testFailures    []string // 记录测速失败的节点
	// 连接页面状态
	selectedConn       int               // 选中的连接索引
	connScrollTop      int               // 连接列表滚动偏移
	connFilterMode     bool              // 是否处于过滤输入模式
	connFilter         string            // 连接过滤关键词
	connDetailMode     bool              // 是否处于详情查看模式
	connDetailSnapshot *model.Connection // 详情模式下的连接快照
	connIPInfo         *model.IPInfo     // 目标IP的地理信息
	connDetailScroll   int               // 详情页面滚动偏移量
	// 图表数据
	chartData    *model.ChartData // 图表历史数据
	lastUpload   int64            // 上次上传总量（用于计算速度）
	lastDownload int64            // 上次下载总量
	// WebSocket客户端
	wsClient  *api.WSClient    // WebSocket流客户端
	wsMsgChan chan interface{} // WebSocket消息通道
	// 历史连接
	closedConnections []model.Connection          // 已关闭的连接历史（最多1000条）
	connViewMode      int                         // 0=活跃连接, 1=历史连接
	prevConnIDs       map[string]model.Connection // 上次推送的连接ID映射（用于检测关闭）
	// 日志页面状态
	logs             []model.LogEntry // 日志列表（最多保留1000条）
	logLevel         int              // 当前级别索引（0=debug, 1=info, 2=warning, 3=error, 4=silent）
	logFilter        string           // 搜索关键词
	logFilterMode    bool             // 是否处于过滤输入模式
	selectedLog      int              // 选中的日志索引
	logScrollTop     int              // 日志列表滚动偏移
	logHScrollOffset int              // 日志水平滚动偏移
	// 规则页面状态
	rules          []model.Rule // 规则列表
	ruleFilter     string       // 搜索关键词
	ruleFilterMode bool         // 是否处于过滤输入模式
	selectedRule   int          // 选中的规则索引
	ruleScrollTop  int          // 规则列表滚动偏移
}

// 消息类型
type (
	groupsMsg      map[string]model.Group
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
	Page6     key.Binding
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
		key.WithHelp("5", "帮助"),
	),
	Page6: key.NewBinding(
		key.WithKeys("6"),
		key.WithHelp("6", "设置"),
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
		// 配置加载失败时使用默认配置
		cfg = &config.DefaultConfig
	}

	// 创建服务实例
	proxySvc := service.NewProxyService(client, testURL, timeout)
	configSvc := service.NewConfigService()
	connSvc := service.NewConnectionService(client)

	// 创建WebSocket客户端
	wsClient := api.NewWSClient(cfg.APIAddress, cfg.Secret)

	return Model{
		client:      client,
		config:      cfg,
		proxySvc:    proxySvc,
		configSvc:   configSvc,
		connSvc:     connSvc,
		testURL:     testURL,
		timeout:     timeout,
		currentPage: components.PageNodes,
		chartData:   model.NewChartData(60),
		wsClient:    wsClient,
		wsMsgChan:   make(chan interface{}, 100), // 带缓冲的消息通道
		logLevel:    1,                           // 默认info级别
	}
}
