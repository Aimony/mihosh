# Mihosh

[English](README.md) | 简体中文

一个功能完整的 mihomo 终端管理工具（TUI），让你在终端优雅地管理代理节点、监控连接、查看日志，无需频繁切换到 Web UI。

## 技术栈

![Go](https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white)
![Bubbletea](https://img.shields.io/badge/Bubbletea-FF69B4?style=flat-square&logo=go&logoColor=white)
![Lipgloss](https://img.shields.io/badge/Lipgloss-9B59B6?style=flat-square&logo=go&logoColor=white)
![Cobra](https://img.shields.io/badge/Cobra-2ECC71?style=flat-square&logo=go&logoColor=white)
![Viper](https://img.shields.io/badge/Viper-E74C3C?style=flat-square&logo=go&logoColor=white)
![WebSocket](https://img.shields.io/badge/WebSocket-010101?style=flat-square&logo=socket.io&logoColor=white)

## 功能预览

| 页面 | 功能 |
|------|------|
| 🎯 **节点管理** | 快速切换代理节点，支持单节点/批量测速 |
| 📊 **连接监控** | 实时查看活跃连接，流量/内存图表，一键关闭连接 |
| 📝 **日志查看** | 实时日志流，支持级别过滤和关键词搜索 |
| 📋 **规则管理** | 查看代理规则，支持多关键词搜索 |
| ⚙️ **设置** | 在界面中直接修改配置 |
| ❓ **帮助** | 内置快捷键说明 |

## 安装



```bash
# 一键安装 (Linux/macOS)
curl -fsSL https://raw.githubusercontent.com/aimony/mihosh/main/install.sh | bash
```


## 快速开始

### 1. 初始化配置

```bash
mihosh config init
```

按提示输入 Mihomo API 地址和密钥，配置保存在 `~/.mihosh/config.yaml`

### 2. 启动

```bash
mihosh
```

启动后进入交互式 TUI 界面，按 `5` 或 `Tab` 切换到帮助页查看完整快捷键。

## 配置文件

配置文件位于 `~/.mihosh/config.yaml`：

```yaml
api_address: http://127.0.0.1:9090
secret: your-secret-here
test_url: http://www.gstatic.com/generate_204
timeout: 5000
```

## 命令行模式（可选）

除了 TUI 界面，也支持命令行直接操作：

```bash
mihosh list                          # 列出策略组
mihosh select <组名> <节点名>          # 切换节点
mihosh test <节点名>                  # 测速节点
mihosh connections                   # 查看连接
```

## 常见问题

| 问题 | 解决方案 |
|------|---------|
| 连接失败 | 检查 Mihomo 是否运行、API 地址和密钥是否正确 |
| 找不到节点 | 确保 mihomo 配置文件中有对应的策略组配置 |
| 测速超时 | 增加 `timeout` 值或更换 `test_url` |

## 开发

```bash
go mod tidy      # 安装依赖
go test ./...    # 运行测试
go build         # 编译
```

## 许可证

MIT License
