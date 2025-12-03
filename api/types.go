package api

// Proxy 代理信息
type Proxy struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	UDP     bool     `json:"udp"`
	History []Delay  `json:"history"`
	All     []string `json:"all,omitempty"`
	Now     string   `json:"now,omitempty"`
}

// Delay 延迟信息
type Delay struct {
	Time  string `json:"time"`
	Delay int    `json:"delay"`
}

// ProxiesResponse 代理列表响应
type ProxiesResponse struct {
	Proxies map[string]Proxy `json:"proxies"`
}

// Group 策略组
type Group struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	Now     string   `json:"now"`
	All     []string `json:"all"`
	History []Delay  `json:"history"`
}

// GroupsResponse 策略组列表响应
type GroupsResponse struct {
	Proxies map[string]Group `json:"proxies"`
}

// DelayTestResponse 延迟测试响应
type DelayTestResponse struct {
	Delay int `json:"delay"`
}

// SelectProxyRequest 选择代理请求
type SelectProxyRequest struct {
	Name string `json:"name"`
}

// Connection 连接信息
type Connection struct {
	ID          string   `json:"id"`
	Metadata    Metadata `json:"metadata"`
	Upload      int64    `json:"upload"`
	Download    int64    `json:"download"`
	Start       string   `json:"start"`
	Chains      []string `json:"chains"`
	Rule        string   `json:"rule"`
	RulePayload string   `json:"rulePayload"`
}

// Metadata 连接元数据
type Metadata struct {
	Network         string `json:"network"`
	Type            string `json:"type"`
	SourceIP        string `json:"sourceIP"`
	DestinationIP   string `json:"destinationIP"`
	SourcePort      string `json:"sourcePort"`
	DestinationPort string `json:"destinationPort"`
	Host            string `json:"host"`
}

// ConnectionsResponse 连接列表响应
type ConnectionsResponse struct {
	DownloadTotal int64        `json:"downloadTotal"`
	UploadTotal   int64        `json:"uploadTotal"`
	Connections   []Connection `json:"connections"`
}

// Traffic 流量信息
type Traffic struct {
	Up   int64 `json:"up"`
	Down int64 `json:"down"`
}
