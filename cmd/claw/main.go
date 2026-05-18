// cmd/claw/main.go
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/lusir01/go-tiny-claw/internal/engine"
	"github.com/lusir01/go-tiny-claw/internal/feishu"
	"github.com/lusir01/go-tiny-claw/internal/provider"
	"github.com/lusir01/go-tiny-claw/internal/tools"
)

func main() {
	workDir, _ := os.Getwd()
	workDir += "/workspace"

	llmProvider := provider.NewZhipuClaudeProvider("glm-5.1")

	registry := tools.NewRegistry()
	registry.Register(tools.NewReadFileTool(workDir))
	registry.Register(tools.NewWriteFileTool(workDir))
	registry.Register(tools.NewBashTool(workDir))
	registry.Register(tools.NewEditFileTool(workDir))

	eng := engine.NewAgentEngine(llmProvider, registry, workDir, true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// 飞书模式：有环境变量时后台启动
	if os.Getenv("FEISHU_APP_ID") != "" && os.Getenv("FEISHU_APP_SECRET") != "" {
		bot := feishu.NewFeishuBot(eng)
		go func() {
			log.Println("🚀 飞书 WebSocket 长连接模式启动...")
			if err := bot.StartWebSocket(ctx); err != nil {
				log.Printf("❌ WebSocket 连接失败: %v\n", err)
			}
		}()
	}

	// 终端交互模式：始终启动
	fmt.Println("🖥️  Go Tiny Claw 终端模式 (输入 exit 或 quit 退出)")
	fmt.Println("─────────────────────────────────────────────────")

	reporter := engine.NewTerminalReporter()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\n> ")

		inputCh := make(chan string, 1)
		go func() {
			if scanner.Scan() {
				inputCh <- scanner.Text()
			} else {
				inputCh <- ""
			}
		}()

		select {
		case <-sigChan:
			fmt.Println("\n📴 再见！")
			cancel()
			return
		case input := <-inputCh:
			input = strings.TrimSpace(input)
			if input == "" {
				continue
			}
			if input == "exit" || input == "quit" {
				fmt.Println("📴 再见！")
				cancel()
				return
			}

			runCtx, runCancel := context.WithCancel(ctx)
			done := make(chan struct{})

			go func() {
				defer close(done)
				if err := eng.Run(runCtx, input, reporter); err != nil && runCtx.Err() == nil {
					log.Printf("❌ Agent 运行失败: %v\n", err)
				}
			}()

			select {
			case <-done:
				runCancel()
			case <-sigChan:
				runCancel()
				<-done
				fmt.Println("\n📴 再见！")
				cancel()
				return
			}
		}
	}
}
