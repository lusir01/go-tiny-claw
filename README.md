# Go Tiny Claw

一个基于 Go 语言实现的轻量级 AI Agent 框架，支持飞书机器人集成和多种 AI 模型提供商。

## 功能特性

- 🤖 支持飞书机器人 WebSocket 长连接模式
- 🔧 可扩展的工具系统（文件读写、命令执行等）
- 🎯 统一的 Provider 接口，支持多种 AI 模型
- 📝 完整的消息上下文管理
- ⚡ 异步任务处理，不阻塞消息回调

## 快速开始

### 1. 环境要求

- Go 1.21 或更高版本
- 飞书开放平台账号（如需使用飞书机器人）
- 智谱 AI API Key（或其他兼容 OpenAI 的 API）

### 2. 配置环境变量

在你的 shell 配置文件（`~/.zshrc` 或 `~/.bashrc`）中添加以下环境变量：

```bash
# 飞书 Bot 配置
export FEISHU_APP_ID="你的飞书应用ID"
export FEISHU_APP_SECRET="你的飞书应用密钥"

# 智谱 AI 配置
export ZHIPU_API_KEY="你的智谱API密钥"
export ZHIPU_BASE_URL="https://open.bigmodel.cn/api/paas/v4"  # 可选，默认使用官方地址
```

配置完成后，重新加载配置文件：

```bash
source ~/.zshrc  # 或 source ~/.bashrc
```

### 3. 获取配置信息

#### 飞书配置

1. 访问 [飞书开放平台](https://open.feishu.cn/app)
2. 创建企业自建应用
3. 获取 `App ID` 和 `App Secret`
4. 在"事件订阅"中选择 **"长连接模式"**（无需配置回调 URL）
5. 订阅事件：`im.message.receive_v1`（接收消息）
6. 在"权限管理"中添加"获取与发送单聊、群组消息"权限
7. 发布应用并添加到群聊或单聊中

#### 智谱 AI 配置

1. 访问 [智谱开放平台](https://open.bigmodel.cn/)
2. 注册并创建 API Key
3. 复制 API Key 到配置文件

### 4. 编译和运行

```bash
# 安装依赖
go mod download

# 编译项目
go build -o bin/claw ./cmd/claw

# 运行飞书机器人（WebSocket 长连接模式）
./bin/claw

# 停止服务：按 Ctrl+C 优雅关闭
```

## 项目结构

```
.
├── cmd/
│   └── claw/          # 主程序入口（WebSocket 长连接模式）
├── internal/
│   ├── engine/        # Agent 核心引擎
│   ├── feishu/        # 飞书机器人集成（WebSocket）
│   ├── provider/      # AI 模型提供商（智谱 AI）
│   ├── schema/        # 数据结构定义
│   └── tools/         # 工具系统（文件读写、命令执行等）
├── .env.example       # 环境变量配置示例（仅占位符）
├── setup_env.sh.example  # 环境变量脚本示例（仅占位符）
└── .gitignore         # Git 忽略规则（包含敏感文件）
```

## 支持的 AI 模型

- 智谱 AI (GLM-4)
- 任何兼容 OpenAI API 的服务

## 安全说明

⚠️ **重要提示**：

1. **永远不要**将包含真实密钥的配置文件提交到 Git
2. 环境变量应配置在 `~/.zshrc` 或 `~/.bashrc` 中，不要提交到代码仓库
3. 不要在代码中硬编码 API Key 或密钥
4. 如果不小心泄露了密钥，请立即在对应平台重置
5. `.env.example` 和 `setup_env.sh.example` 仅包含占位符，可以安全提交

## 开发指南

### 添加新工具

在 `internal/tools/` 目录下创建新的工具文件，实现 `Tool` 接口：

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]interface{}
    Execute(ctx context.Context, args map[string]interface{}) (string, error)
}
```

### 添加新的 Provider

在 `internal/provider/` 目录下实现 `Provider` 接口：

```go
type Provider interface {
    Generate(ctx context.Context, msgs []schema.Message, tools []schema.ToolDefinition) (*schema.Message, error)
}
```

## 常见问题

### Q: 如何切换不同的 AI 模型？

A: 修改代码中的 `NewZhipuClaudeProvider` 或 `NewZhipuOpenAIProvider` 调用，传入不同的模型名称。

### Q: 飞书机器人收不到消息？

A: 检查以下几点：
1. 确认环境变量配置正确（`env | grep -E "FEISHU|ZHIPU"` 验证）
2. 确认飞书应用已选择 **"长连接模式"**（不是 HTTP 回调）
3. 确认已订阅 `im.message.receive_v1` 事件
4. 查看控制台日志，确认 WebSocket 连接成功（应显示 `connected to wss://...`）
5. 确认机器人已被添加到对应的群聊或单聊
6. 确认应用已发布并开启相关权限

### Q: API 调用失败？

A: 检查：
1. API Key 是否正确
2. Base URL 是否正确（默认使用智谱官方地址）
3. 网络连接是否正常
4. API 配额是否充足

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
