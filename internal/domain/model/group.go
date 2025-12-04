package model

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
