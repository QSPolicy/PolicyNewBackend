package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/openai/openai-go"

	"policy-backend/agent"
	"policy-backend/config"
	"policy-backend/utils"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 初始化日志
	utils.InitLogger(&cfg.Log)

	// 检查是否有 LLM 配置
	if len(cfg.Agent.LLMConfigs) == 0 {
		fmt.Println("错误: 没有配置 LLM，请在 config.yaml 中添加 llm_configs")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	// 选择配置
	var llmCfg agent.LLMConfig
	if len(cfg.Agent.LLMConfigs) == 1 {
		llmCfg = cfg.Agent.LLMConfigs[0]
		fmt.Printf("自动选择唯一配置: %s (%s)\n", llmCfg.Name, llmCfg.Model)
	} else {
		fmt.Println("可用模型配置:")
		for i, c := range cfg.Agent.LLMConfigs {
			fmt.Printf("  [%d] %s (Model: %s)\n", i+1, c.Name, c.Model)
		}

		fmt.Print("请选择序号 (默认 1): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		var idx int
		if input == "" {
			idx = 0
		} else {
			fmt.Sscanf(input, "%d", &idx)
			idx-- // 转为 0-based
		}

		if idx < 0 || idx >= len(cfg.Agent.LLMConfigs) {
			fmt.Println("无效选择，使用默认配置 #1")
			idx = 0
		}
		llmCfg = cfg.Agent.LLMConfigs[idx]
		fmt.Printf("已选择: %s\n", llmCfg.Name)
	}

	// 创建 Tool Provider
	toolProvider := agent.NewLocalToolProvider(&agent.ToolConfig{
		BaiduAPIKey: cfg.Agent.BaiduSearchAPIKey,
	})
	fmt.Println("已注册的工具:")
	for _, t := range toolProvider.ListTools() {
		fmt.Printf("  - %s: %s\n", t.Name, t.Description)
	}

	// 创建 Adapter
	adapter := agent.NewOpenAIToolAdapter(toolProvider)

	// 创建 Agent
	ag := agent.NewAgent(llmCfg, adapter)

	// 开始对话
	fmt.Println("\n=== Agent 对话测试 ===")
	fmt.Println("输入 'quit' 或 'exit' 退出")
	fmt.Println("输入 'stream' 切换流式/非流式模式")
	fmt.Println()

	messages := []openai.ChatCompletionMessageParamUnion{}
	streamMode := false

	for {
		fmt.Print("You: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("读取输入错误: %v\n", err)
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		if input == "quit" || input == "exit" {
			fmt.Println("再见!")
			break
		}

		if input == "stream" {
			streamMode = !streamMode
			if streamMode {
				fmt.Println("已切换到流式模式")
			} else {
				fmt.Println("已切换到非流式模式")
			}
			continue
		}

		if input == "clear" {
			messages = []openai.ChatCompletionMessageParamUnion{}
			fmt.Println("对话历史已清空")
			continue
		}

		// 添加用户消息
		messages = append(messages, openai.ChatCompletionMessageParamUnion{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(input),
				},
			},
		})

		ctx := context.Background()
		fmt.Print("Assistant: ")

		var resp *openai.ChatCompletion
		if streamMode {
			// 流式输出
			resp, err = ag.ChatStream(ctx, messages, func(chunk string) {
				fmt.Print(chunk)
			})
			fmt.Println() // 换行
		} else {
			// 非流式输出
			resp, err = ag.Chat(ctx, messages)
			if err == nil && len(resp.Choices) > 0 {
				fmt.Println(resp.Choices[0].Message.Content)
			}
		}

		if err != nil {
			fmt.Printf("错误: %v\n", err)
			// 移除失败的用户消息
			messages = messages[:len(messages)-1]
			continue
		}

		// 添加助手回复到历史
		if len(resp.Choices) > 0 {
			messages = append(messages, openai.ChatCompletionMessageParamUnion{
				OfAssistant: &openai.ChatCompletionAssistantMessageParam{
					Content: openai.ChatCompletionAssistantMessageParamContentUnion{
						OfString: openai.String(resp.Choices[0].Message.Content),
					},
				},
			})
		}
		fmt.Println()
	}
}
