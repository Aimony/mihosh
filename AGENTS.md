# AGENTS.md

本文档为参与 Mihosh 项目开发的 AI 智能体及代码助手提供完整的项目上下文、开发规范和决策指南。

---

## 1. 项目基本信息

### 1.1 项目定位
- **项目名称**：Mihosh
- **定义**：一款现代化的 Mihomo (Clash Meta) 终端控制台客户端
- **核心能力**：
  - 🔗 实时连接监控（流量统计、抓包级过滤排查）
  - 🚀 节点管理与性能验证（支持批量测速、单节点切换）
  - 📊 控制台日志抓取与搜索（支持多级别过滤）
  - 📋 规则组可视化调试

### 1.2 技术栈总览

| 层级 | 技术选型 | 版本 |
|------|---------|------|
| **编程语言** | Go | 1.24+ (toolchain go1.24.11) |
| **CLI 框架** | spf13/cobra + spf13/viper | cobra v1.10.1 |
| **TUI 框架** | charmbracelet/bubbletea | v1.3.10 |
| **UI 组件** | charmbracelet/bubbles, lipgloss | v0.21.0, v1.1.0 |
| **通信协议** | gorilla/websocket | v1.5.3 |
| **配置格式** | YAML | viper 管理 |
| **测试框架** | testify | v1.11.1 |
| **并发原语** | x/sync/semaphore | Go 标准库 |

---

## 2. 架构总览与目录结构

### 2.1 分层架构模型

Mihosh 采用**分层 + DDD 弱化**的组织方式：

```
┌─────────────────────────────────┐
│     CLI 入口层                   │  (cmd/, internal/cli/)
│  cobra 命令解析与执行           │
└────────────┬────────────────────┘
             │
┌────────────▼────────────────────┐
│     TUI 渲染层                   │  (internal/ui/tui/)
│  Bubbletea Model/Update/View    │  ✓ 五页面状态隔离
│  管理交互、状态、消息路由        │  ✓ Ring Buffer 优化
└────────────┬────────────────────┘
             │
┌────────────▼────────────────────┐
│     业务服务层                   │  (internal/app/service/)
│  ProxyService, ConfigService     │  ✓ Semaphore 并发控制
│  ConnectionService              │  ✓ 业务编排
└────────────┬────────────────────┘
             │
┌────────────▼────────────────────┐
│     基础设施层                   │  (internal/infrastructure/)
│  HTTP 客户端、WebSocket 适配器   │  ✓ 泛型重连生命周期
│  配置读写、API 封装              │
└────────────┬────────────────────┘
             │
┌────────────▼────────────────────┐
│     数据模型层                   │  (internal/domain/model/)
│  Proxy, Connection, Log, Rule    │
└─────────────────────────────────┘
```

### 2.2 目录结构详解

```
d:\Code_Project\Personal\mihosh/
├── main.go                       # 应用启动点
├── go.mod                        # 模块定义
├── go.sum                        # 依赖锁文件
│
├── internal/
│   ├── cli/                      # CLI 命令解析层
│   │   ├── root.go               # TUI 主命令启动
│   │   ├── list.go               # 列表查询命令
│   │   ├── ...
│   │
│   ├── domain/model/             # 业务数据模型
│   │   ├── proxy.go              # Proxy, Group 定义
│   │   ├── connection.go         # Connection 连接数据
│   │   └── ...
│   │
│   ├── infrastructure/           # 基础设施适配器 (API, Config)
│   │   ├── api/
│   │   │   ├── client.go         # HTTP 客户端
│   │   │   └── websocket.go      # WebSocket 适配 ⭐
│   │
│   ├── app/service/              # 业务服务编排
│   │   ├── proxy.go              # 代理管理服务 (Semaphore) ⭐
│   │   └── ...
│   │
│   └── ui/tui/                   # TUI 界面主层 (FSD 架构) ⭐⭐⭐
│       ├── model.go              # 主 Model (精简全局状态)
│       ├── update.go             # 消息路由与页面分发
│       ├── view.go               # 顶级布局渲染
│       ├── page_renders.go       # 页面委托器 (Bridge to Features)
│       │
│       └── features/             # 业务特性模块 (Feature-Sliced Design)
│           ├── nodes/            # 节点管理特性
│           │   ├── state.go      # NodesState 定义与逻辑
│           │   ├── view.go       # 节点页面渲染
│           │   └── ...
│           ├── connections/      # 连接监控特性
│           ├── logs/             # 日志查看特性
│           ├── rules/            # 规则管理特性
│           └── settings/         # 设置编辑特性
│
├── pkg/utils/                    # 工具函数库
├── docs/                         # 文档
├── build/                        # 构建脚本
└── ...
```

---

## 3. 核心开发原则

### 3.1 三大编程哲学

任何代码生成或架构决策**必须**严格遵守以下原则：

1. **🎯 KISS (Keep It Simple, Stupid)**
   - 代码以最简单的方式解决问题
   - **不追求通用性**，只做当前需求
   - **坚决避免**过度设计、防御性代码、过多参数化
   - 三行重复代码 > 一个不必要的工具函数

2. **🔬 First Principles (第一性原理分析)**
   - 遇到性能瓶颈、报错或设计困境时，从根源拆解
   - **必读相关代码**后再下结论（"事实为本"）
   - 不套用模板方案，不盲目跟风
   - 例：内存泄漏 → 分析 goroutine 泄漏 → 分析 channel 阻塞 → 寻找 deadlock

3. **📌 Fact-Based (基于事实)**
   - 架构迁移前**必须读代码**，不猜测
   - 性能优化前**必须测量和分析**
   - 复杂功能前**必须理解现有代码**和设计

---

## 4. 快速开发命令

### 4.1 环境设置

```bash
# 验证 Go 版本
go version                           # 需要 1.24+

# 安装依赖
go mod download
go mod tidy

# 查看项目配置示例
cat config.example.yaml
```

### 4.2 编译与运行

```bash
# 在项目根目录运行

# 完整编译 (生成 Windows 可执行文件)
go build -o mihosh.exe .

# 快速运行 TUI (需要本地 Mihomo 代理服务)
go run .

# 编译为 Linux 可执行文件
GOOS=linux GOARCH=amd64 go build -o mihosh .

# 快速迭代 (安装 air，支持热重载)
# https://github.com/cosmtrek/air
air
```

### 4.3 测试命令

```bash
# 运行所有单元测试
go test ./...

# 运行特定包测试（带详细输出）
go test -v ./internal/app/service

# 运行测试并生成覆盖率报告
go test -cover ./...

# 运行单个测试函数
go test -run TestFunctionName ./package

# 性能测试 (如果有 Benchmark)
go test -bench=. ./package
```

### 4.4 代码质量检查

```bash
# 语法与样式检查 (需要安装 golangci-lint)
golangci-lint run ./...

# 格式化代码
go fmt ./...
gofmt -w .

# 查找潜在问题
go vet ./...

# 模块依赖审计
go mod verify
```

### 4.5 调试与分析

```bash
# 生成 CPU 和内存分析数据 (支持 pprof)
go run -cpuprofile=cpu.prof -memprofile=mem.prof .
go tool pprof cpu.prof

# 查看 goroutine 泄漏检查 (运行后 Ctrl+C)
# 观察 "goroutine" 输出

# WebSocket 连接测试 (用 curl 或专门工具)
curl -i -N -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  http://localhost:7890/api/traffic
```

---

## 5. TUI 架构详细规范

### 5.1 Model 与分层设计（★ 优先避免 ★）

**当前现状**：
- 主 Model 包含 **约 20 个字段** (model.go: ~100 行)
- 重构后状态通过 `features/` 下的子状态管理，主 Model 仅持有引用。
- 状态完全按页面隔离，符合 Feature-Sliced Design (FSD) 最佳实践。

**必须遵守的规则**：
```
MUST: 新增状态时，优先添加到对应的 *State 结构体中
      (如新增节点页面字段 → 添加到 NodesState)

NEVER: 直接在 Model 中添加游离的新字段
       (如 isLoading, selectedIndex 等通用字段)

ALWAYS: 通过 tea.Msg 进行页面间的消息传递
        (避免直接修改其他页面的 State)
```

**关键数据结构**（内存占用）：
| 页面 | State 大小 | 备注 |
|------|-----------|------|
| NodesState | ~2KB | 含两列表滚动位置、缓存数据 |
| ConnsState | ~4KB | Ring Buffer (1000 条，每条 ~100B) |
| LogsState | ~3KB | Ring Buffer (1000 条) |
| RulesState | ~1KB | 单表，过滤索引缓存 |
| SettingsState | ~0.5KB | 简单编辑模式字段 |

**示例：正确的新状态添加方式**
```go
// ✓ 正确：添加到 ConnsState
type ConnsState struct {
    connections *model.ConnectionsResponse
    // ... 既有字段
    newFilterMode bool              // 新增：连接过滤模式
    newFilterExpiry time.Time        // 新增：过滤过期时间
}

// ✗ 错误：添加到 Model
type Model struct {
    // ... 既有字段
    connFilterMode bool             // 不要这样做！
    connFilterExpiry time.Time
}
```

---

### 5.2 tea.Msg 消息体系（★ 严格约束 ★）

**当前消息类型总数**：**16 个**

**消息分类**：

| 分类 | 定义位置 | 示例 | 用途 |
|------|---------|------|------|
| **数据消息** | model.go | `proxiesMsg`, `connectionsMsg`, `errMsg` | API 响应反应到 UI |
| **用户操作** | update.go | `KeyMsg`, `MouseMsg` | 键盘/鼠标输入 |
| **后台任务** | commands.go | `testDoneMsg`, `ipInfoMsg` | 异步任务完成回调 |
| **定时器** | commands.go | `connTickMsg`, `logsTickMsg` | 定时 1s 刷新信号 |
| **系统事件** | bubbletea | `WindowSizeMsg`, `tea.Quit()` | 窗口大小、退出 |

**新增消息的规范**：
```go
// ✓ 规范命名：使用 XxxMsg 后缀
type siteTestMsg struct {
    name  string
    delay int
    err   error
}

// ✓ 消息定义位置：
// - 后台任务消息 → commands.go
// - 数据消息 → model.go
// - 事件消息 → 靠近发起方

// ✗ 避免：过大的消息体
type megaMsg struct {
    // 包含整个数据库查询结果
    results []model.Connection   // 不要这样
}
```

**消息路由流程**：
```
Model.Update(msg)
├─ 全局消息处理 (WindowSizeMsg, KeyMsg, MouseMsg)
├─ 页面消息分发 (dispatchKeyToPage)
│  ├─ PageNodes → NodesState.Update()
│  ├─ PageConnections → ConnsState.Update()
│  ├─ PageLogs → LogsState.Update()
│  ├─ PageRules → RulesState.Update()
│  └─ PageSettings → SettingsState.Update()
└─ 全局数据消息 (proxiesMsg, connectionsMsg, errMsg)
   └─ 应用到对应 State
```

---

### 5.3 并发控制规范（★ 严格执行 ★）

#### 5.3.1 Goroutine 约束（§5.2）

**规则**：批量网络操作**必须**使用 Worker Pool / Semaphore 限制并发数

**当前实施状态** ✓：
```go
// File: internal/app/service/proxy.go:60-67
const testAllConcurrency = 20

func (s *ProxyService) TestAllProxies(proxies []string) map[string]int {
    sem := make(chan struct{}, testAllConcurrency)
    
    for _, proxy := range proxies {
        sem <- struct{}{}                      // 获取许可证
        go func(p string) {
            defer func() { <-sem }()            // 释放许可证
            // 执行网络请求
        }(proxy)
    }
}
```

**执行清单**：
- ✓ 节点测速：Semaphore (20 并发)
- ✓ 连接查询：无需限制 (API 本身限制)
- ✓ WebSocket 读取：4 条独立连接 (无竞争)
- ✓ 日志查询：无需限制 (本地缓冲查询)

**新增场景的规范**：
```go
// ✗ 避免：无限制并发
for _, item := range items {
    go func(i interface{}) {
        // 执行网络操作
    }(item)
}

// ✓ 规范：受控并发 (≤ 20)
const maxWorkers = 20
sem := make(chan struct{}, maxWorkers)
for _, item := range items {
    sem <- struct{}{}
    go func(i interface{}) {
        defer func() { <-sem }()
        // 执行网络操作
    }(item)
}
```

---

#### 5.3.2 内存优化：Ring Buffer（§5.3）

**规则**：固定长度的日志/历史记录**必须**使用 Ring Buffer，禁止切片截断

**当前实施** ✓：
```go
// File: internal/ui/tui/conns_state.go
const closedConnCap = 1000
type ConnsState struct {
    closedConns [closedConnCap]model.Connection  // ← 环形缓冲
    closedHead  int
    closedCount int
}

// 写入操作 (O(1) 恒定时间)
func (s *ConnsState) AddClosedConn(conn model.Connection) {
    s.closedConns[s.closedHead] = conn
    s.closedHead = (s.closedHead + 1) % closedConnCap
    if s.closedCount < closedConnCap {
        s.closedCount++
    }
}

// ✗ 禁止: 这是 O(n) GC 炸弹
// s.connections = append([]model.Connection{newConn}, s.connections...)

// ✗ 禁止: 切片截断导致内存泄漏
// s.logs = s.logs[1:]  // 持续堆积内存
```

**应用场景**：
- ✓ 历史连接 (1000 条，按关闭时间) → conns_state.go
- ✓ 日志记录 (1000 条，按时间) → logs_state.go
- ✓ 图表数据 (60 个数据点) → model.ChartData

#### 5.3.3 横向滚动边界约束（§5.3）

**规则**：日志页面水平滚动必须设置上限，禁止无限滚动导致 TUI 布局溢出

**问题**：用户持续按住右方向键时，`logHScrollOffset` 无上限递增，导致渲染内容超出 TUI 可视区域

**解决方案**：
```go
// 在 State 中增加 maxHScrollOffset 字段
type State struct {
    logHScrollOffset   int
    maxHScrollOffset  int  // 新增：动态计算的最大滚动上限
}

// 在 UpdateMaxHScrollOffset 中根据页面宽度动态计算上限
func (s State) UpdateMaxHScrollOffset(width, height int) State {
    sidebarRenderedWidth := 19
    pageWidth := width - sidebarRenderedWidth - 2
    fixedOverhead := 8 + 1 + 20
    maxOffset := pageWidth - fixedOverhead - 20
    if maxOffset < 0 {
        maxOffset = 0
    }
    s.maxHScrollOffset = maxOffset
    if s.logHScrollOffset > s.maxHScrollOffset {
        s.logHScrollOffset = s.maxHScrollOffset
    }
    return s
}

// Right 键处理器增加边界检查
case key.Matches(msg, common.Keys.Right):
    if s.logHScrollOffset < s.maxHScrollOffset {
        s.logHScrollOffset += 10
        if s.logHScrollOffset > s.maxHScrollOffset {
            s.logHScrollOffset = s.maxHScrollOffset
        }
    }
```

**触发时机**：
- 窗口大小变化时（`tea.WindowSizeMsg`）
- 初始页面加载时

**最大列宽计算公式**：`maxOffset = pageWidth - 29`，其中 29 是日志前缀（时间戳 + 级别 + 边距）的固定宽度

**新增固定记录的规范**：
```go
// ✓ 规范：定义容量常量 + 环形数组
const siteTestHistoryCap = 100
type SiteTestHistory struct {
    items [siteTestHistoryCap]model.SiteTest
    head  int
    count int
}

// ✗ 避免：无限增长的切片
type SiteTestHistory struct {
    items []model.SiteTest  // 永远增长，直到 OOM
}
```

---

#### 5.3.3 WebSocket 重连生命周期（§5.4）

**规则**：所有 WebSocket 通道**必须**使用统一的泛型重连逻辑

**当前实施** ✓：
```go
// File: internal/infrastructure/api/websocket.go:224-280
// 泛型函数处理所有四路连接的重连
func connectStream[T any](c *WSClient, stream string, 
    handler func(T), url string) {
    
    for {
        // 1. 建立连接
        ws, err := websocket.Dial(url, ...)
        if err != nil {
            time.Sleep(2 * time.Second)  // 重连延迟
            continue
        }
        
        // 2. 读取消息循环
        for {
            var data T
            err = ws.ReadJSON(&data)
            if err != nil {
                break  // 连接断开，重新连接
            }
            handler(data)  // 处理数据
        }
        
        ws.Close()
        
        // 3. 响应停止信号
        select {
        case <-c.stopChan:
            return
        default:
            continue
        }
    }
}

// 四路流的启动
go connectStream(c, "memory", c.memoryHandler, memoryURL)
go connectStream(c, "traffic", c.trafficHandler, trafficURL)
go connectStream(c, "connections", c.connectionsHandler, connURL)
go connectStream(c, "logs", c.logsHandler, logsURL)
```

**执行清单**：
- ✓ 内存流 (memory WebSocket)
- ✓ 流量流 (traffic WebSocket)
- ✓ 连接流 (connections WebSocket)
- ✓ 日志流 (logs WebSocket)

**新增实时数据源的规范**：
```go
// ✓ 规范：使用泛型 connectStream
go connectStream(c, "new-metric", 
    func(data model.NewMetric) {
        c.newMetricHandler(data)
    }, 
    newMetricURL)

// ✗ 避免：复制 Copy-Paste 重连代码
go func() {  // 冗余代码！
    for {
        ws, _ := websocket.Dial(...)
        for {
            var data model.NewMetric
            ws.ReadJSON(&data)
            // 处理...
        }
    }
}()
```

---

## 6. 页面系统架构（五大页面）

### 6.1 页面分类与职责

| 页面 | 状态定义 | 渲染逻辑 | 核心职责 |
|------|---------|---------|---------|
| **Nodes (节点)** | features/nodes/state.go | features/nodes/view.go | 策略组/节点管理、并发测速 |
| **Connections (连接)** | features/connections/state.go | features/connections/view.go | 活跃连接监控、Ring Buffer 历史 |
| **Logs (日志)** | features/logs/state.go | features/logs/view.go | 实时日志流、Ring Buffer 缓存、水平滚动边界保护 |
| **Rules (规则)** | features/rules/state.go | features/rules/view.go | 规则查询与过滤 |
| **Settings (设置)** | features/settings/state.go | features/settings/view.go | UI 交互式配置编辑 |

### 6.2 页面生命周期流程

**以 Nodes 页面为例**：

```
1. 页面初始化 (Model.Init / onPageChange)
   └─ fetchGroups() 获取策略组列表

2. 获取数据后更新状态
   └─ groupsMsg → NodesState.groups 和 groupNames

3. 用户交互 (KeyMsg)
   └─ ↑/↓ 导航 → selectedGroup / selectedProxy 更新
   └─ Enter 确认 → selectProxy() 发起 HTTP 请求
   └─ T 测速 → testSingleProxy() 异步任务

4. 异步任务完成 (testDoneMsg)
   └─ testPending-- 
   └─ 触发重新渲染

5. 页面渲染 (Model.View)
   └─ pages.RenderNodesPage(m.nodesState)
   └─ 使用 lipgloss 拼装 UI 布局

6. 切换页面时清理
   └─ onPageChange() 可选清空临时状态
```

### 6.3 页面状态初始化

**规范**：所有新页面**必须**在 model.go 中定义其 State 结构体和初始化函数

```go
// File: internal/ui/tui/model.go

// 1. 在 Model struct 中添加
type Model struct {
    // ...
    myPageState MyPageState
}

// 2. 定义 State 结构体
type MyPageState struct {
    data          []interface{}
    selectedIndex int
    filter        string
    isLoading     bool
}

// 3. 提供初始化函数
func NewMyPageState() MyPageState {
    return MyPageState{
        selectedIndex: 0,
    }
}

// 4. 在 NewModel() 中初始化
func NewModel(...) Model {
    return Model{
        myPageState: NewMyPageState(),
        // ...
    }
}

// 5. 在 update.go 中添加消息处理
case MyPageState state := msg:
    m.myPageState = state.Update(msg, m.client)
```

---

## 7. 服务层规范

### 7.1 ProxyService：并发控制示范

**文件**：`internal/app/service/proxy.go`

**关键方法**：

```go
func (s *ProxyService) TestAllProxies(proxies []string) map[string]int {
    const testAllConcurrency = 20
    sem := make(chan struct{}, testAllConcurrency)
    
    results := make(map[string]int)
    var mu sync.Mutex
    var wg sync.WaitGroup
    
    for _, proxy := range proxies {
        wg.Add(1)
        sem <- struct{}{}
        
        go func(p string) {
            defer wg.Done()
            defer func() { <-sem }()
            
            delay, err := s.client.TestProxyDelay(p, s.testURL, s.timeout)
            
            mu.Lock()
            if err == nil {
                results[p] = delay
            }
            mu.Unlock()
        }(proxy)
    }
    
    wg.Wait()
    return results
}
```

**设计要点**：
- ✓ Semaphore 限制并发为 20
- ✓ WaitGroup 等待所有任务完成
- ✓ Mutex 保护共享 results map
- ✓ 延迟释放 Semaphore (defer 在最前)

---

### 7.2 ConfigService 与 ConnectionService

**ConfigService**：配置加载/保存/修改
- `Load(path)` - 从文件加载配置
- `Save(path)` - 保存配置到文件
- `Update(key, value)` - 更新配置项

**ConnectionService**：连接与地理位置查询
- `QueryConnections()` - 获取活跃连接
- `QueryIPInfo(ip)` - 查询 IP 地理位置 (ip-api.com API)

---

## 8. 项目特定的约束与陷阱

### 8.1 必须遵守的规范

| 序号 | 规范 | 违反后果 | 优先级 |
|------|------|---------|--------|
| ★★★ | 新增状态添加到 *State，不添加到 Model | Model 字段爆炸 | **CRITICAL** |
| ★★★ | 所有 WebSocket 使用泛型 connectStream | 代码重复、难维护 | **CRITICAL** |
| ★★★ | 批量网络操作使用 Semaphore (≤20) | 文件句柄耗尽、系统崩溃 | **CRITICAL** |
| ★★★ | 固定长历史用 Ring Buffer | 内存泄漏、GC 压力 | **CRITICAL** |
| ★★ | 页面间通信使用 tea.Msg，不直接修改 | 状态不可控、Bug 难追踪 | **HIGH** |
| ★★ | 新增消息使用 XxxMsg 命名 + 文档注释 | 消息体系混乱 | **HIGH** |
| ★ | 定期 `go test ./...` 验证编译 | 破坏性修改未被发现 | **MEDIUM** |

### 8.2 常见陷阱与调试

**问题 1：窗口大小变化导致文字错位**
- 原因：缓存宽度未同步更新
- 解决：在 `Model.Update()` 中监听 `WindowSizeMsg`，更新所有 State 的缓存宽度

**问题 2：WebSocket 连接反复断开**
- 原因：网络不稳定 / Token 过期 / 代理地址错误
- 调试：查看 `internal/infrastructure/api/websocket.go` 的错误日志；确认代理服务地址正确

**问题 3：节点测速卡顿**
- 原因：Semaphore 未生效，直接发起无限 goroutine
- 调试：检查 `ProxyService.TestAllProxies()` 是否使用了 Semaphore

**问题 4：日志/连接列表内存占用持续增长**
- 原因：未使用 Ring Buffer，使用了切片截断
- 调试：检查 `features/connections/state.go` 和 `features/logs/state.go` 是否使用 Ring Buffer

**问题 5：文件句柄耗尽 ("too many open files")**
- 原因：WebSocket 连接泄漏 / 并发请求过多
- 调试：检查 WSClient 是否正确调用 `Close()`；检查 Semaphore 是否生效

---

## 9. 开发与审查工作流

### 9.1 复杂需求的处理流程

**适用场景**：架构重构、新增页面、性能优化、并发控制改进

```
1. 需求理解阶段
   ├─ 理解用户需求与目标
   ├─ 识别涉及的文件范围
   └─ 确认是否存在相关约束 (如并发数限制)

2. 代码探索阶段 (First Principles)
   ├─ 读取相关文件，理解现有实现
   ├─ 分析数据流向与消息传递
   ├─ 识别可能的性能瓶颈或设计缺陷
   └─ 查看 AGENTS.md 的相关规范

3. 计划制定阶段
   ├─ 设计实现方案 (多个选项对比)
   ├─ 评估对现有代码的影响范围
   ├─ 列出具体的实现步骤 (包括测试)
   └─ 生成 Implementation Plan 文档

4. 用户审核阶段
   └─ 展示方案给人类开发者审核
   └─ 根据反馈调整设计

5. 实现阶段
   ├─ 按步骤编写代码
   ├─ 边实现边跑测试 (go test ./...)
   ├─ 及时提交 git commit
   └─ 必要时生成性能数据对比

6. 验收阶段
   ├─ 运行完整的 go build + go test
   ├─ 手动验证新功能的用户交互
   ├─ 检查是否符合 AGENTS.md 规范
   └─ 准备 Release Notes
```

### 9.2 提交 Commit 的规范

**Commit 消息格式**（参考 Conventional Commits）：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 分类**：
- `feat` - 新增功能
- `fix` - 修复 Bug
- `refactor` - 代码重构 (不改功能)
- `perf` - 性能优化
- `test` - 测试代码
- `docs` - 文档更新
- `chore` - 构建、依赖、工具链

**Scope 范围**：
- `tui` - TUI 界面层
- `service` - 业务服务层
- `api` - HTTP/WebSocket 客户端
- `cli` - CLI 命令层
- `model` - 数据模型

**示例**：
```
feat(tui): 新增连接过滤功能

- 在 ConnsState 中添加 connFilter 和 connFilterMode 字段
- 实现连接列表的关键词过滤逻辑
- 支持正则表达式过滤 (可选)

Closes #123
```

---

## 10. 常见问题与答疑 (FAQ)

### Q: 如何添加一个新的实时数据源 (如 CPU 使用率)?

**A**: 
1. 定义数据模型 (internal/domain/model/cpu.go)
2. 在 WSClient 中添加对应的 handler
3. 使用 `connectStream[T]` 泛型函数启动连接
4. 在 TUI 中创建新 State (可选) 或添加到现有页面

```go
// 1. 定义模型
type CPUUsage struct {
    Percent float64 `json:"percent"`
}

// 2. 在 WSClient 中添加 handler
type WSClient struct {
    cpuHandler func(CPUUsage)
}

// 3. 启动连接
go connectStream(c, "cpu", c.cpuHandler, cpuURL)

// 4. 在页面中使用
// ConnsState 中添加 cpuHistory [100]float64
// 更新时同步更新 ChartData
```

---

### Q: 如何添加一个新的 CLI 命令?

**A**:
1. 在 `internal/cli/` 中创建新文件 (如 `newcmd.go`)
2. 定义 `var newCmd = &cobra.Command{...}`
3. 在 `internal/cli/root.go` 中注册: `rootCmd.AddCommand(newCmd)`
4. 测试: `go run ./cmd/app newcmd`

```go
// File: internal/cli/newcmd.go
var newCmd = &cobra.Command{
    Use:   "newcmd",
    Short: "新命令描述",
    RunE: func(cmd *cobra.Command, args []string) error {
        // 命令逻辑
        return nil
    },
}

// File: internal/cli/root.go
func init() {
    rootCmd.AddCommand(newCmd)
}
```

---

### Q: 如何调试 WebSocket 连接问题?

**A**:
1. 查看 WSClient 的日志输出
2. 使用 `curl` 或 `wscat` 测试 WebSocket 连接:
```bash
wscat -c ws://localhost:7890/api/traffic
```
3. 检查 `internal/infrastructure/api/websocket.go` 的 `connectStream` 函数中的错误处理
4. 确认代理服务地址、Token 是否正确

---

### Q: 如何优化节点测速的速度?

**A**:
1. **不应该**直接增加 Semaphore 的并发数 (默认 20 是安全上限)
2. **可以做**的优化:
   - 减少单次测试的 `timeout` 参数
   - 并行发起多个 `TestAllProxies()` 调用 (但需小心)
   - 实现测速超时及时中止 (已有实现)

```go
// ✓ 推荐：减少超时时间 (但注意网络延迟)
testTimeout := 3000  // ms, 而非默认 5000

// ✗ 避免：增加并发数
const testAllConcurrency = 50  // 风险！
```

---

## 11. 术语与概念速查表

| 术语 | 定义 | 相关文件 |
|------|------|---------|
| **Model** | Bubbletea 中的主状态结构体 | model.go |
| **tea.Msg** | Bubbletea 消息接口的实现 | model.go, update.go |
| **Update()** | 处理消息并更新状态的函数 | update.go, *_state.go |
| **View()** | 返回 UI 字符串的函数 | view.go, pages/*.go |
| **Cmd/Batch** | 返回异步任务的函数 | commands.go |
| **State 隔离** | 每个页面有独立的 State 结构体 | *_state.go (5 个) |
| **Ring Buffer** | 环形缓冲区 (固定大小) | conns_state.go, logs_state.go |
| **Semaphore** | 信号量，限制并发数 | proxy.go, WSClient |
| **Handler** | 异步数据处理回调函数 | websocket.go, commands.go |
| **tea.Tick()** | 定时器消息 | commands.go |

---

## 12. 资源与参考

### 12.1 官方文档

- **Bubbletea 文档**: https://github.com/charmbracelet/bubbletea
- **Cobra 文档**: https://cobra.dev/
- **Go 官方**: https://golang.org/
- **WebSocket in Go**: https://pkg.go.dev/github.com/gorilla/websocket


