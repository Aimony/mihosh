# Mihosh

一个功能完整的 mihomo 终端管理工具，支持在终端直接操作 mihomo 的核心功能，包括节点切换、测速等，避免频繁切换到 Web UI。

## 功能特性

- 🎯 **交互式 TUI 界面** - 使用 bubbletea 框架构建的美观终端界面
- 🚀 **节点切换** - 快速切换代理节点
- ⚡ **节点测速** - 测试单个节点或整个策略组的延迟
- 📊 **实时信息** - 显示策略组、节点列表和延迟信息
- 🔧 **命令行模式** - 支持脚本化操作
- 📡 **连接管理** - 查看当前连接信息

## 安装

### 从源码编译

```bash
# 克隆仓库
git clone https://github.com/aimony/mihosh.git
cd mihosh

# 编译
go build -o mihomo.exe

# 或者直接运行
go run main.go
```

### 二进制文件

下载对应平台的可执行文件到系统 PATH 目录。

## 快速开始

### 1. 初始化配置

首次使用需要初始化配置：

```bash
mihomo config init
```

按提示输入：
- Mihomo API 地址（默认：`http://127.0.0.1:9090`）
- API 密钥（如果 mihomo 配置了 secret）
- 测速 URL（默认：`http://www.gstatic.com/generate_204`）

配置文件保存在 `~/.mihosh/config.yaml`

### 2. 启动交互界面

```bash
mihomo
```

## 使用指南

### 交互式界面

启动后进入 TUI 界面，支持以下快捷键：

| 快捷键 | 功能 |
|--------|------|
| `↑/↓` 或 `k/j` | 上下选择节点 |
| `←/→` 或 `h/l` | 切换策略组 |
| `Enter` | 切换到选中的节点 |
| `t` | 测速当前选中的节点 |
| `a` | 测速当前策略组的所有节点 |
| `r` | 刷新数据 |
| `q` | 退出 |

界面显示：
- **策略组列表** - 显示所有策略组及当前选中的节点
- **节点列表** - 显示当前策略组的所有节点
- **延迟信息** - 用颜色标识（绿色 <200ms，黄色 <500ms，红色 ≥500ms）
- **当前节点** - 用 ✓ 标记

### 命令行模式

#### 列出所有策略组和节点

```bash
mihomo list
```

#### 切换节点

```bash
mihomo select <策略组名> <节点名>
```

示例：
```bash
mihomo select PROXY 香港节点1
```

#### 测速节点

测速单个节点：
```bash
mihomo test <节点名>
```

测速整个策略组：
```bash
mihomo test-group <策略组名>
```

#### 查看连接信息

```bash
mihomo connections
```

#### 配置管理

查看当前配置：
```bash
mihomo config show
```

重新初始化配置：
```bash
mihomo config init
```

## 配置文件

配置文件位于 `~/.mihosh/config.yaml`：

```yaml
api_address: http://127.0.0.1:9090
secret: your-secret-here
test_url: http://www.gstatic.com/generate_204
timeout: 5000
```

配置说明：
- `api_address` - Mihomo API 地址
- `secret` - API 密钥（与 mihomo 配置文件中的 secret 一致）
- `test_url` - 测速使用的 URL
- `timeout` - 请求超时时间（毫秒）

## 常见问题

### 1. 连接失败

检查：
- Mihomo 是否正在运行
- API 地址是否正确
- API 密钥是否匹配

### 2. 找不到策略组或节点

确保 mihomo 配置文件中有对应的策略组和节点配置。

### 3. 测速超时

可以在配置文件中增加 `timeout` 值，或更改 `test_url` 为响应更快的 URL。

## 技术栈

- **Go** - 主要编程语言
- **Bubbletea** - TUI 框架
- **Lipgloss** - 终端样式库
- **Cobra** - 命令行参数解析
- **Viper** - 配置文件管理

## 开发

```bash
# 安装依赖
go mod tidy

# 运行测试
go test ./...

# 编译
go build -o mihomo.exe
```

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
