package schema

import "encoding/json"

// Role 定义消息的角色，这是与大模型沟通的基石
type Role string

const (
	RoleSystem    Role = "system"    // 系统提示词：确立 Agent 的性格与红线
	RoleUser      Role = "user"      // 用户输入 / 工具执行的返回结果 (Observation)
	RoleAssistant Role = "assistant" // 模型的输出：包含推理(Reasoning)或工具调用(ToolCall)
)

// Message 代表上下文中传递的单条消息
type Message struct {
	Role    Role   `json:"role"`
	Content string `json:"content"` // 存放纯文本内容

	// 如果模型决定调用工具，此字段将被填充 (支持并行调用多个工具)
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`

	// 如果这是对某个工具调用的响应，此字段必须填写，以告知模型上下文的关联性
	ToolCallID string `json:"tool_call_id,omitempty"`
}

// ToolCall 代表模型请求调用某个具体的工具
type ToolCall struct {
	ID   string `json:"id"`   // 工具调用的唯一 ID
	Name string `json:"name"` // 想要调用的工具名称 (例如 "bash")
	// Arguments 存放 JSON 参数。使用 RawMessage 是为了延迟解析，将解析责任交给具体的工具
	Arguments json.RawMessage `json:"arguments"`
}

// ToolResult 代表工具在本地执行完毕后返回的物理结果
type ToolResult struct {
	ToolCallID string `json:"tool_call_id"`
	Output     string `json:"output"`   // 工具执行的控制台输出或报错堆栈
	IsError    bool   `json:"is_error"` // 标记是否失败，供后续的驾驭工程进行错误自愈
}

// ToolDefinition 描述了一个大模型可以调用的工具元信息 (供模型理解工具有什么用)
type ToolDefinition struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"input_schema"` // 对应 JSON Schema
}
