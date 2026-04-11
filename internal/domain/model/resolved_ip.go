package model

// ResolvedIP 内网IP解析结果
type ResolvedIP struct {
	IP          string // 原始IP
	IsPrivate   bool   // 是否内网IP
	NetworkType string // "docker", "tailscale", "local", "lan", "unknown"
	AppName     string // 容器名 / 机器名 / 降级提示
	AppDetail   string // 容器镜像 / OS / 额外信息
}
