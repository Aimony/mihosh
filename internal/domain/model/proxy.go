package model

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

// DelayTestResponse 延迟测试响应
type DelayTestResponse struct {
	Delay int `json:"delay"`
}

// SelectProxyRequest 选择代理请求
type SelectProxyRequest struct {
	Name string `json:"name"`
}
