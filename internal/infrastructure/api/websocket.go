package api

import (
	"encoding/json"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WSMessage WebSocket消息类型
type WSMessage struct {
	Type string
	Data []byte
}

// MemoryData 内存数据
type MemoryData struct {
	Inuse   int64 `json:"inuse"`
	OSLimit int64 `json:"oslimit"`
}

// TrafficData 流量数据
type TrafficData struct {
	Up   int64 `json:"up"`
	Down int64 `json:"down"`
}

// ConnectionsData 连接数据（与ConnectionsResponse结构相同）
type ConnectionsData struct {
	DownloadTotal int64            `json:"downloadTotal"`
	UploadTotal   int64            `json:"uploadTotal"`
	Connections   []ConnectionData `json:"connections"`
}

// ConnectionData 单个连接数据
type ConnectionData struct {
	ID            string         `json:"id"`
	Metadata      ConnectionMeta `json:"metadata"`
	Upload        int64          `json:"upload"`
	Download      int64          `json:"download"`
	Start         string         `json:"start"`
	Chains        []string       `json:"chains"`
	Rule          string         `json:"rule"`
	RulePayload   string         `json:"rulePayload"`
	DownloadSpeed int64          `json:"downloadSpeed"`
	UploadSpeed   int64          `json:"uploadSpeed"`
}

// ConnectionMeta 连接元数据
type ConnectionMeta struct {
	Network         string `json:"network"`
	Type            string `json:"type"`
	SourceIP        string `json:"sourceIP"`
	DestinationIP   string `json:"destinationIP"`
	SourcePort      string `json:"sourcePort"`
	DestinationPort string `json:"destinationPort"`
	Host            string `json:"host"`
	Process         string `json:"process"`
	ProcessPath     string `json:"processPath"`
}

// LogData 日志数据
type LogData struct {
	Type    string `json:"type"`    // debug, info, warning, error, silent
	Payload string `json:"payload"` // 日志内容
}

// WSClient WebSocket客户端
type WSClient struct {
	baseURL         string
	secret          string
	memoryConn      *websocket.Conn
	trafficConn     *websocket.Conn
	connectionsConn *websocket.Conn
	logsConn        *websocket.Conn

	memoryMu      sync.Mutex
	trafficMu     sync.Mutex
	connectionsMu sync.Mutex
	logsMu        sync.Mutex

	memoryHandler      func(MemoryData)
	trafficHandler     func(TrafficData)
	connectionsHandler func(ConnectionsData)
	logsHandler        func(LogData)

	logLevel  string // 日志级别过滤
	stopChan  chan struct{}
	isRunning bool
	runningMu sync.Mutex
}

// NewWSClient 创建WebSocket客户端
func NewWSClient(baseURL, secret string) *WSClient {
	return &WSClient{
		baseURL:  baseURL,
		secret:   secret,
		stopChan: make(chan struct{}),
	}
}

// SetMemoryHandler 设置内存数据处理器
func (c *WSClient) SetMemoryHandler(handler func(MemoryData)) {
	c.memoryHandler = handler
}

// SetTrafficHandler 设置流量数据处理器
func (c *WSClient) SetTrafficHandler(handler func(TrafficData)) {
	c.trafficHandler = handler
}

// SetConnectionsHandler 设置连接数据处理器
func (c *WSClient) SetConnectionsHandler(handler func(ConnectionsData)) {
	c.connectionsHandler = handler
}

// SetLogsHandler 设置日志数据处理器
func (c *WSClient) SetLogsHandler(handler func(LogData)) {
	c.logsHandler = handler
}

// SetLogLevel 设置日志级别过滤
func (c *WSClient) SetLogLevel(level string) {
	c.logLevel = level
}

// buildWSURL 构建WebSocket URL
func (c *WSClient) buildWSURL(endpoint string) string {
	// 将http/https转换为ws/wss
	wsURL := strings.Replace(c.baseURL, "https://", "wss://", 1)
	wsURL = strings.Replace(wsURL, "http://", "ws://", 1)

	// 构建完整URL
	u, err := url.Parse(wsURL + "/" + endpoint)
	if err != nil {
		return ""
	}

	// 添加token参数
	if c.secret != "" {
		q := u.Query()
		q.Set("token", c.secret)
		u.RawQuery = q.Encode()
	}

	return u.String()
}

// Start 启动WebSocket连接
func (c *WSClient) Start() error {
	c.runningMu.Lock()
	if c.isRunning {
		c.runningMu.Unlock()
		return nil
	}
	c.isRunning = true
	c.stopChan = make(chan struct{})
	c.runningMu.Unlock()

	// 启动memory连接
	go c.connectMemory()
	// 启动traffic连接
	go c.connectTraffic()
	// 启动connections连接
	go c.connectConnections()
	// 启动logs连接
	go c.connectLogs()

	return nil
}

// Stop 停止WebSocket连接
func (c *WSClient) Stop() {
	c.runningMu.Lock()
	if !c.isRunning {
		c.runningMu.Unlock()
		return
	}
	c.isRunning = false
	close(c.stopChan)
	c.runningMu.Unlock()

	c.memoryMu.Lock()
	if c.memoryConn != nil {
		c.memoryConn.Close()
		c.memoryConn = nil
	}
	c.memoryMu.Unlock()

	c.trafficMu.Lock()
	if c.trafficConn != nil {
		c.trafficConn.Close()
		c.trafficConn = nil
	}
	c.trafficMu.Unlock()

	c.connectionsMu.Lock()
	if c.connectionsConn != nil {
		c.connectionsConn.Close()
		c.connectionsConn = nil
	}
	c.connectionsMu.Unlock()

	c.logsMu.Lock()
	if c.logsConn != nil {
		c.logsConn.Close()
		c.logsConn = nil
	}
	c.logsMu.Unlock()
}

// IsRunning 检查是否正在运行
func (c *WSClient) IsRunning() bool {
	c.runningMu.Lock()
	defer c.runningMu.Unlock()
	return c.isRunning
}

// connectMemory 连接memory WebSocket
func (c *WSClient) connectMemory() {
	wsURL := c.buildWSURL("memory")
	if wsURL == "" {
		return
	}

	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			// 连接失败，等待后重试
			time.Sleep(2 * time.Second)
			continue
		}

		c.memoryMu.Lock()
		c.memoryConn = conn
		c.memoryMu.Unlock()

		// 读取消息
		c.readMemoryMessages(conn)

		// 连接断开，等待后重连
		c.memoryMu.Lock()
		c.memoryConn = nil
		c.memoryMu.Unlock()

		select {
		case <-c.stopChan:
			return
		case <-time.After(1 * time.Second):
			// 重连
		}
	}
}

// connectTraffic 连接traffic WebSocket
func (c *WSClient) connectTraffic() {
	wsURL := c.buildWSURL("traffic")
	if wsURL == "" {
		return
	}

	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			// 连接失败，等待后重试
			time.Sleep(2 * time.Second)
			continue
		}

		c.trafficMu.Lock()
		c.trafficConn = conn
		c.trafficMu.Unlock()

		// 读取消息
		c.readTrafficMessages(conn)

		// 连接断开，等待后重连
		c.trafficMu.Lock()
		c.trafficConn = nil
		c.trafficMu.Unlock()

		select {
		case <-c.stopChan:
			return
		case <-time.After(1 * time.Second):
			// 重连
		}
	}
}

// readMemoryMessages 读取memory消息
func (c *WSClient) readMemoryMessages(conn *websocket.Conn) {
	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var data MemoryData
		if err := json.Unmarshal(message, &data); err != nil {
			continue
		}

		if c.memoryHandler != nil {
			c.memoryHandler(data)
		}
	}
}

// readTrafficMessages 读取traffic消息
func (c *WSClient) readTrafficMessages(conn *websocket.Conn) {
	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var data TrafficData
		if err := json.Unmarshal(message, &data); err != nil {
			continue
		}

		if c.trafficHandler != nil {
			c.trafficHandler(data)
		}
	}
}

// connectConnections 连接connections WebSocket
func (c *WSClient) connectConnections() {
	wsURL := c.buildWSURL("connections")
	if wsURL == "" {
		return
	}

	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			// 连接失败，等待后重试
			time.Sleep(2 * time.Second)
			continue
		}

		c.connectionsMu.Lock()
		c.connectionsConn = conn
		c.connectionsMu.Unlock()

		// 读取消息
		c.readConnectionsMessages(conn)

		// 连接断开，等待后重连
		c.connectionsMu.Lock()
		c.connectionsConn = nil
		c.connectionsMu.Unlock()

		select {
		case <-c.stopChan:
			return
		case <-time.After(1 * time.Second):
			// 重连
		}
	}
}

// readConnectionsMessages 读取connections消息
func (c *WSClient) readConnectionsMessages(conn *websocket.Conn) {
	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var data ConnectionsData
		if err := json.Unmarshal(message, &data); err != nil {
			continue
		}

		if c.connectionsHandler != nil {
			c.connectionsHandler(data)
		}
	}
}

// connectLogs 连接logs WebSocket
func (c *WSClient) connectLogs() {
	// 构建logs URL，带level参数
	level := c.logLevel
	if level == "" {
		level = "debug" // 使用debug级别获取所有日志，由客户端过滤
	}
	wsURL := c.buildWSURL("logs?level=" + level)
	if wsURL == "" {
		return
	}

	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			// 连接失败，等待后重试
			time.Sleep(2 * time.Second)
			continue
		}

		c.logsMu.Lock()
		c.logsConn = conn
		c.logsMu.Unlock()

		// 读取消息
		c.readLogsMessages(conn)

		// 连接断开，等待后重连
		c.logsMu.Lock()
		c.logsConn = nil
		c.logsMu.Unlock()

		select {
		case <-c.stopChan:
			return
		case <-time.After(1 * time.Second):
			// 重连
		}
	}
}

// readLogsMessages 读取logs消息
func (c *WSClient) readLogsMessages(conn *websocket.Conn) {
	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		_, message, err := conn.ReadMessage()
		if err != nil {
			return
		}

		var data LogData
		if err := json.Unmarshal(message, &data); err != nil {
			continue
		}

		if c.logsHandler != nil {
			c.logsHandler(data)
		}
	}
}
