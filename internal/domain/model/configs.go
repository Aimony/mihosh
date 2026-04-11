package model

// ConfigsResponse 配置信息响应
type ConfigsResponse struct {
	Mode string `json:"mode"`
}

// UpdateConfigRequest 更新配置请求
type UpdateConfigRequest struct {
	Mode string `json:"mode,omitempty"`
}
