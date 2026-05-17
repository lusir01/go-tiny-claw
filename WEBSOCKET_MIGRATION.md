# 飞书 Bot WebSocket 长连接迁移指南

## 改造说明

已将飞书 Bot 从 **HTTP 回调方式** 改为 **WebSocket 长连接方式**。

## 主要变更

### 1. 代码变更

#### `internal/feishu/bot.go`
- ✅ 新增 `StartWebSocket()` 方法：启动 WebSocket 长连接
- ✅ 保留 `GetEventDispatcher()` 方法：兼容 HTTP 回调方式（可选）
- ✅ 重构 `createEventDispatcher()` 方法：HTTP 和 WebSocket 共用事件处理逻辑

#### `cmd/claw/main.go`
- ✅ 移除 HTTP 服务器相关代码（`http.ListenAndServe`）
- ✅ 改用 `bot.StartWebSocket()` 启动长连接
- ✅ 添加优雅关闭机制（监听 Ctrl+C 信号）

### 2. 依赖变更

新增依赖（已自动安装）：
```go
github.com/gorilla/websocket v1.5.0
github.com/gogo/protobuf v1.3.2
```

## 使用方式

### 环境变量配置

**WebSocket 方式仅需 2 个环境变量：**
```bash
export FEISHU_APP_ID="cli_xxxxxxxxxxxxx"
export FEISHU_APP_SECRET="xxxxxxxxxxxxxxxxxxxxxxxx"
```

**不再需要（HTTP 回调方式才需要）：**
```bash
# export FEISHU_ENCRYPT_KEY="xxxxxx"  # WebSocket 不需要
# export FEISHU_VERIFY_TOKEN="xxxxxx"  # WebSocket 不需要
```

### 启动服务

```bash
# 编译
go build -o bin/claw ./cmd/claw

# 运行
./bin/claw
```

### 停止服务

按 `Ctrl+C` 优雅关闭，程序会自动断开 WebSocket 连接。

## 核心优势

| 对比维度 | HTTP 回调方式 | WebSocket 长连接方式 ✅ |
|---------|--------------|-------------------|
| **部署要求** | 需要公网 IP + 域名 + SSL 证书 | **无需公网 IP，内网可运行** |
| **配置复杂度** | 需要配置回调 URL、Nginx 反向代理 | **仅需 AppID 和 AppSecret** |
| **开发体验** | 需要内网穿透工具（ngrok/frp） | **本地直接运行** |
| **重连机制** | 无需重连（HTTP 无状态） | **自动重连（SDK 内置）** |
| **环境变量** | 4 个（AppID、AppSecret、EncryptKey、VerifyToken） | **2 个（AppID、AppSecret）** |

## 飞书开放平台配置

### 1. 开启长连接模式

进入飞书开放平台 → 应用详情 → 事件订阅：
- ✅ 选择 **"长连接模式"**
- ❌ 无需填写 **"请求地址 URL"**（HTTP 回调方式才需要）

### 2. 订阅事件

确保已订阅以下事件：
- `im.message.receive_v1`（接收消息）
- `im.message.message_read_v1`（消息已读，可选）

### 3. 权限配置

确保应用已开通以下权限：
- 获取与发送单聊、群组消息
- 以应用身份读取通讯录

## 技术细节

### WebSocket 连接流程

```
1. 程序启动
   ↓
2. 创建 ws.Client（传入 AppID、AppSecret）
   ↓
3. 注册事件处理器（EventDispatcher）
   ↓
4. 调用 wsClient.Start(ctx) 建立长连接
   ↓
5. 持续接收飞书推送的事件消息
   ↓
6. 收到 Ctrl+C 信号 → 取消 context → 断开连接
```

### 自动重连机制

SDK 内置自动重连功能（`ws.WithAutoReconnect(true)`）：
- 网络断开时自动重连
- 重连间隔：指数退避策略
- 无需手动处理重连逻辑

### 并发处理

收到消息后，自动开启 goroutine 处理：
```go
go b.handleAgentRun(chatId, contentStr)
```

## 回退到 HTTP 方式（可选）

如果需要回退到 HTTP 回调方式，修改 `cmd/claw/main.go`：

```go
// 使用 HTTP 回调方式
handler := httpserverext.NewEventHandlerFunc(bot.GetEventDispatcher())
http.HandleFunc("/webhook/event", handler)
http.ListenAndServe(":48080", nil)
```

并配置完整的 4 个环境变量（包括 EncryptKey 和 VerifyToken）。

## 常见问题

### Q1: 启动后没有收到消息？
**A:** 检查以下几点：
1. 飞书开放平台是否选择了 **"长连接模式"**
2. 环境变量 `FEISHU_APP_ID` 和 `FEISHU_APP_SECRET` 是否正确
3. 应用是否已发布并添加到群聊/私聊中
4. 查看日志是否有 WebSocket 连接成功的提示

### Q2: 如何查看 WebSocket 连接状态？
**A:** 启动时会输出日志：
```
🔌 正在启动 WebSocket 长连接模式...
✅ WebSocket 客户端已创建，正在连接飞书服务器...
```

### Q3: 生产环境推荐哪种方式？
**A:** 
- **小型项目/内网部署**：推荐 WebSocket（简单、无需公网 IP）
- **大型项目/需要负载均衡**：推荐 HTTP 回调（可水平扩展）

## 相关文档

- [飞书开放平台 - 长连接模式](https://open.feishu.cn/document/server-docs/event-subscription-guide/event-subscription-configure-/request-url-configuration-case#d286cc88)
- [oapi-sdk-go WebSocket 文档](https://github.com/larksuite/oapi-sdk-go)

---

**改造完成时间**: 2026-05-17  
**改造人**: Claude Sonnet 4.6
