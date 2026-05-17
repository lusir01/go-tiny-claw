// cmd/claw/main.go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/lusir01/go-tiny-claw/internal/engine"
	"github.com/lusir01/go-tiny-claw/internal/feishu"
	"github.com/lusir01/go-tiny-claw/internal/provider"
	"github.com/lusir01/go-tiny-claw/internal/tools"
)

func main() {
	workDir, _ := os.Getwd()

	llmProvider := provider.NewZhipuClaudeProvider("glm-5.1")

	registry := tools.NewRegistry()
	registry.Register(tools.NewReadFileTool(workDir))
	registry.Register(tools.NewWriteFileTool(workDir))
	registry.Register(tools.NewBashTool(workDir))
	registry.Register(tools.NewEditFileTool(workDir))

	// 开启慢思考
	eng := engine.NewAgentEngine(llmProvider, registry, workDir, false)

	// 初始化飞书 Bot
	bot := feishu.NewFeishuBot(eng)

	// 创建可取消的 context，用于优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听系统信号（Ctrl+C 等）
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 在 goroutine 中启动 WebSocket 长连接
	go func() {
		log.Println("🚀 go-tiny-claw 正在启动 WebSocket 长连接模式...")
		if err := bot.StartWebSocket(ctx); err != nil {
			log.Fatalf("❌ WebSocket 连接失败: %v", err)
		}
	}()

	// 等待退出信号
	<-sigChan
	log.Println("\n📴 收到退出信号，正在优雅关闭...")
	cancel() // 取消 context，触发 WebSocket 断开
	log.Println("✅ 服务已停止")
}
