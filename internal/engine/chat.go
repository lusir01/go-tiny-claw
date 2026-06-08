package engine

import (
	"context"
	"fmt"
	"sync"

	"github.com/lusir01/go-tiny-claw/internal/schema"
)

// Chat 执行一轮对话，接收完整历史并返回追加了本轮内容的新历史。
// 调用方负责在轮次间保存和传递 history，从而实现多轮上下文。
// 首次调用时传入空 slice，方法会自动注入 system prompt。
func (e *AgentEngine) Chat(ctx context.Context, userPrompt string, history []schema.Message, reporter Reporter) ([]schema.Message, error) {
	if len(history) == 0 {
		history = append(history, e.composer.Build())
	}
	history = append(history, schema.Message{Role: schema.RoleUser, Content: userPrompt})

	for {
		availableTools := e.registry.GetAvailableTools()

		if e.EnableThinking {
			if reporter != nil {
				reporter.OnThinking(ctx)
			}
			thinkResp, err := e.provider.Generate(ctx, history, nil)
			if err != nil {
				return history, fmt.Errorf("Thinking 生成失败: %w", err)
			}
			if thinkResp.Content != "" {
				history = append(history, *thinkResp)
			}
		}

		actionResp, err := e.provider.Generate(ctx, history, availableTools)
		if err != nil {
			return history, fmt.Errorf("Action 生成失败: %w", err)
		}
		history = append(history, *actionResp)

		if actionResp.Content != "" && reporter != nil {
			reporter.OnMessage(ctx, actionResp.Content)
		}

		if len(actionResp.ToolCalls) == 0 {
			break
		}

		observationMsgs := make([]schema.Message, len(actionResp.ToolCalls))
		var wg sync.WaitGroup

		for i, toolCall := range actionResp.ToolCalls {
			wg.Add(1)
			go func(idx int, call schema.ToolCall) {
				defer wg.Done()
				if reporter != nil {
					reporter.OnToolCall(ctx, call.Name, string(call.Arguments))
				}
				result := e.registry.Execute(ctx, call)
				if reporter != nil {
					display := result.Output
					if len(display) > 200 {
						display = display[:200] + "... (已截断)"
					}
					reporter.OnToolResult(ctx, call.Name, display, result.IsError)
				}
				observationMsgs[idx] = schema.Message{
					Role:       schema.RoleUser,
					Content:    result.Output,
					ToolCallID: call.ID,
				}
			}(i, toolCall)
		}

		wg.Wait()
		for _, obs := range observationMsgs {
			history = append(history, obs)
		}
	}

	return history, nil
}
