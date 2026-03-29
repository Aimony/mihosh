package tui

import (
	"strings"

	"github.com/aimony/mihosh/internal/domain/model"
	"github.com/aimony/mihosh/internal/infrastructure/api"
	"github.com/aimony/mihosh/internal/ui/tui/pages"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

const closedConnCap = 1000

// ConnsState 连接页面完整状态
type ConnsState struct {
	connections        *model.ConnectionsResponse
	prevConnIDs        map[string]model.Connection
	// Ring Buffer for closed connections
	closedConns [closedConnCap]model.Connection
	closedHead  int // 写入位置（下一条写入的索引）
	closedCount int // 已写入的总条数（上限 closedConnCap）

	selectedConn       int
	connScrollTop      int
	connFilterMode     bool
	connFilter         string
	connDetailMode     bool
	connDetailSnapshot *model.Connection
	connIPInfo         *model.IPInfo
	connDetailScroll   int
	connViewMode       int // 0=活跃, 1=历史

	siteTests        []model.SiteTest
	selectedSiteTest int
	proxyAddr        string
}

// NewConnsState 初始化连接状态
func NewConnsState(proxyAddr string, siteTests []model.SiteTest) ConnsState {
	return ConnsState{
		proxyAddr: proxyAddr,
		siteTests: siteTests,
	}
}

// ClosedConnections 返回历史连接（最新在前）
func (s ConnsState) ClosedConnections() []model.Connection {
	if s.closedCount == 0 {
		return nil
	}
	result := make([]model.Connection, s.closedCount)
	// Ring Buffer: 从最新写入位置向前读
	for i := 0; i < s.closedCount; i++ {
		idx := (s.closedHead - 1 - i + closedConnCap) % closedConnCap
		result[i] = s.closedConns[idx]
	}
	return result
}

// appendClosed 向 Ring Buffer 追加一条历史连接
func (s *ConnsState) appendClosed(conn model.Connection) {
	s.closedConns[s.closedHead] = conn
	s.closedHead = (s.closedHead + 1) % closedConnCap
	if s.closedCount < closedConnCap {
		s.closedCount++
	}
}

// ToPageState 转换为渲染层所需的 ConnectionsPageState
func (s ConnsState) ToPageState(chartData *model.ChartData, width, height int) pages.ConnectionsPageState {
	return pages.ConnectionsPageState{
		Connections:        s.connections,
		Width:              width,
		Height:             height,
		SelectedIndex:      s.selectedConn,
		ScrollTop:          s.connScrollTop,
		FilterText:         s.connFilter,
		FilterMode:         s.connFilterMode,
		DetailMode:         s.connDetailMode,
		SelectedConnection: s.connDetailSnapshot,
		IPInfo:             s.connIPInfo,
		DetailScroll:       s.connDetailScroll,
		ChartData:          chartData,
		ViewMode:           s.connViewMode,
		ClosedConnections:  s.ClosedConnections(),
		SiteTests:          s.siteTests,
		SelectedSiteTest:   s.selectedSiteTest,
	}
}

// Update 处理连接页面按键
func (s ConnsState) Update(msg tea.KeyMsg, client *api.Client, timeout int) (ConnsState, tea.Cmd) {
	// 详情模式
	if s.connDetailMode {
		switch {
		case key.Matches(msg, keys.Escape), key.Matches(msg, keys.Enter):
			s.connDetailMode = false
			s.connDetailSnapshot = nil
			s.connIPInfo = nil
			s.connDetailScroll = 0
		case key.Matches(msg, keys.Up):
			if s.connDetailScroll > 0 {
				s.connDetailScroll--
			}
		case key.Matches(msg, keys.Down):
			s.connDetailScroll++
		}
		return s, nil
	}

	// 过滤输入模式
	if s.connFilterMode {
		return s.handleConnFilterMode(msg)
	}

	switch {
	case key.Matches(msg, keys.Up):
		if s.selectedConn > 0 {
			s.selectedConn--
			if s.selectedConn < s.connScrollTop {
				s.connScrollTop = s.selectedConn
			}
		}

	case key.Matches(msg, keys.Down):
		connCount := s.filteredConnCount()
		if s.selectedConn < connCount-1 {
			s.selectedConn++
		}

	case key.Matches(msg, keys.Enter):
		conn := s.selectedConnection()
		if conn != nil {
			snapshot := *conn
			s.connDetailSnapshot = &snapshot
			s.connDetailMode = true
			s.connIPInfo = nil
			return s, fetchIPInfo(conn.Metadata.DestinationIP)
		}

	case msg.String() == "x":
		conn := s.selectedConnection()
		if conn != nil {
			return s, tea.Batch(
				closeConnection(client, conn.ID),
				fetchConnections(client),
			)
		}

	case msg.String() == "X":
		return s, tea.Batch(
			closeAllConnections(client),
			fetchConnections(client),
		)

	case msg.String() == "/":
		s.connFilterMode = true

	case msg.String() == "h":
		s.connViewMode = (s.connViewMode + 1) % 2
		s.selectedConn = 0
		s.connScrollTop = 0

	case msg.String() == "s":
		if len(s.siteTests) > 0 && s.selectedSiteTest < len(s.siteTests) {
			site := s.siteTests[s.selectedSiteTest]
			s.siteTests[s.selectedSiteTest].Testing = true
			return s, testSiteDelay(s.proxyAddr, site.Name, site.URL, timeout)
		}

	case msg.String() == "S":
		if len(s.siteTests) > 0 {
			for i := range s.siteTests {
				s.siteTests[i].Testing = true
			}
			return s, testAllSites(s.proxyAddr, s.siteTests, timeout)
		}

	case key.Matches(msg, keys.Left):
		if s.selectedSiteTest > 0 {
			s.selectedSiteTest--
		}

	case key.Matches(msg, keys.Right):
		if s.selectedSiteTest < len(s.siteTests)-1 {
			s.selectedSiteTest++
		}

	case key.Matches(msg, keys.Escape):
		if s.connFilter != "" {
			s.connFilter = ""
			s.selectedConn = 0
			s.connScrollTop = 0
		}
	}

	return s, nil
}

// HandleMouseScroll 鼠标滚轮处理
func (s ConnsState) HandleMouseScroll(up bool) ConnsState {
	if s.connDetailMode {
		if up {
			if s.connDetailScroll > 0 {
				s.connDetailScroll--
			}
		} else {
			s.connDetailScroll++
		}
		return s
	}
	count := s.filteredConnCount()
	if up {
		if s.selectedConn > 0 {
			s.selectedConn--
			if s.selectedConn < s.connScrollTop {
				s.connScrollTop = s.selectedConn
			}
		}
	} else {
		if s.selectedConn < count-1 {
			s.selectedConn++
		}
	}
	return s
}

// ApplyWSConnections 处理 WebSocket 连接推送（含历史记录检测）
func (s ConnsState) ApplyWSConnections(data api.ConnectionsData) ConnsState {
	currentIDs := make(map[string]model.Connection, len(data.Connections))
	for _, conn := range data.Connections {
		currentIDs[conn.ID] = model.Connection{
			ID:            conn.ID,
			Upload:        conn.Upload,
			Download:      conn.Download,
			Start:         conn.Start,
			Chains:        conn.Chains,
			Rule:          conn.Rule,
			RulePayload:   conn.RulePayload,
			DownloadSpeed: conn.DownloadSpeed,
			UploadSpeed:   conn.UploadSpeed,
			Metadata: model.Metadata{
				Network:         conn.Metadata.Network,
				Type:            conn.Metadata.Type,
				SourceIP:        conn.Metadata.SourceIP,
				DestinationIP:   conn.Metadata.DestinationIP,
				SourcePort:      conn.Metadata.SourcePort,
				DestinationPort: conn.Metadata.DestinationPort,
				Host:            conn.Metadata.Host,
				Process:         conn.Metadata.Process,
				ProcessPath:     conn.Metadata.ProcessPath,
			},
		}
	}

	// 检测已关闭的连接写入 Ring Buffer
	if s.prevConnIDs != nil {
		for id, conn := range s.prevConnIDs {
			if _, exists := currentIDs[id]; !exists {
				s.appendClosed(conn)
			}
		}
	}
	s.prevConnIDs = currentIDs
	s.connections = convertToConnectionsResponse(data)
	return s
}

// ApplyConnections 应用 REST API 返回的连接数据
func (s ConnsState) ApplyConnections(resp *model.ConnectionsResponse) ConnsState {
	s.connections = resp
	return s
}

// ApplySiteTestResult 应用网站测速结果
func (s ConnsState) ApplySiteTestResult(name string, delay int, err error) ConnsState {
	for i := range s.siteTests {
		if s.siteTests[i].Name == name {
			s.siteTests[i].Testing = false
			if err != nil {
				s.siteTests[i].Delay = 0
				s.siteTests[i].Error = "timeout"
			} else {
				s.siteTests[i].Delay = delay
				s.siteTests[i].Error = ""
			}
			break
		}
	}
	return s
}

// ApplyConnectionClosed 处理单连接关闭后的索引调整
func (s ConnsState) ApplyConnectionClosed() ConnsState {
	if s.selectedConn > 0 {
		s.selectedConn--
	}
	return s
}

// ApplyAllConnectionsClosed 所有连接关闭后重置状态
func (s ConnsState) ApplyAllConnectionsClosed() ConnsState {
	s.selectedConn = 0
	s.connScrollTop = 0
	return s
}

// ApplyIPInfo 更新 IP 地理信息
func (s ConnsState) ApplyIPInfo(info *model.IPInfo) ConnsState {
	s.connIPInfo = info
	return s
}

// ResetPrevConnIDs 重置连接快照（切换到连接页时调用）
func (s ConnsState) ResetPrevConnIDs() ConnsState {
	s.prevConnIDs = nil
	return s
}

// UpdateProxyAddr 更新代理地址（设置页保存后调用）
func (s ConnsState) UpdateProxyAddr(addr string) ConnsState {
	s.proxyAddr = addr
	return s
}

// handleConnFilterMode 连接过滤输入模式
func (s ConnsState) handleConnFilterMode(msg tea.KeyMsg) (ConnsState, tea.Cmd) {
	switch {
	case key.Matches(msg, keys.Escape):
		s.connFilterMode = false
	case key.Matches(msg, keys.Enter):
		s.connFilterMode = false
		s.selectedConn = 0
		s.connScrollTop = 0
	case key.Matches(msg, keys.Backspace):
		if len(s.connFilter) > 0 {
			s.connFilter = s.connFilter[:len(s.connFilter)-1]
		}
	default:
		input := msg.String()
		if len(input) == 1 && input[0] >= 32 && input[0] < 127 {
			s.connFilter += input
		}
	}
	return s, nil
}

// filteredConnCount 过滤后的连接数量
func (s ConnsState) filteredConnCount() int {
	var conns []model.Connection
	if s.connViewMode == 0 {
		if s.connections == nil {
			return 0
		}
		conns = s.connections.Connections
	} else {
		conns = s.ClosedConnections()
	}
	if s.connFilter == "" {
		return len(conns)
	}
	count := 0
	filter := strings.ToLower(s.connFilter)
	for _, conn := range conns {
		if connMatchesFilter(conn, filter) {
			count++
		}
	}
	return count
}

// selectedConnection 获取当前选中的连接
func (s ConnsState) selectedConnection() *model.Connection {
	var conns []model.Connection
	if s.connViewMode == 0 {
		if s.connections == nil || len(s.connections.Connections) == 0 {
			return nil
		}
		conns = s.connections.Connections
	} else {
		closed := s.ClosedConnections()
		if len(closed) == 0 {
			return nil
		}
		conns = closed
	}

	if s.connFilter == "" {
		if s.selectedConn >= 0 && s.selectedConn < len(conns) {
			return &conns[s.selectedConn]
		}
		return nil
	}

	filter := strings.ToLower(s.connFilter)
	idx := 0
	for i := range conns {
		if connMatchesFilter(conns[i], filter) {
			if idx == s.selectedConn {
				return &conns[i]
			}
			idx++
		}
	}
	return nil
}

// connMatchesFilter 检查连接是否匹配过滤词
func connMatchesFilter(conn model.Connection, filter string) bool {
	if filter == "" {
		return true
	}
	if strings.Contains(strings.ToLower(conn.Metadata.Host), filter) {
		return true
	}
	if strings.Contains(strings.ToLower(conn.Rule), filter) {
		return true
	}
	if strings.Contains(strings.ToLower(conn.Metadata.DestinationIP), filter) {
		return true
	}
	for _, chain := range conn.Chains {
		if strings.Contains(strings.ToLower(chain), filter) {
			return true
		}
	}
	return false
}
