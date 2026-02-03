package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"policy-backend/agent"
	"policy-backend/config"
	"policy-backend/research"
	"policy-backend/utils"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Initialize logger
	utils.InitLogger(&cfg.Log)

	// Check for LLM configs
	if len(cfg.Agent.LLMConfigs) == 0 {
		fmt.Println("错误: 未配置 LLM，请在 config.yaml 中添加 llm_configs")
		os.Exit(1)
	}

	reader := bufio.NewReader(os.Stdin)

	// Select LLM Config
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
			idx-- // Convert to 0-based
		}

		if idx < 0 || idx >= len(cfg.Agent.LLMConfigs) {
			fmt.Println("无效选择，使用默认配置 #1")
			idx = 0
		}
		llmCfg = cfg.Agent.LLMConfigs[idx]
		fmt.Printf("已选择: %s\n", llmCfg.Name)
	}

	// Check Baidu Search API Key
	if cfg.Agent.BaiduSearchAPIKey == "" {
		fmt.Println("警告: 未配置 Baidu Search API Key，研究功能可能受限")
	}

	// Configure Research Engine
	researchCfg := research.DefaultConfig().
		WithPlannerLLM(llmCfg).
		WithWorkerLLM(llmCfg).
		WithSearchConfig(agent.ToolConfig{
			BaiduAPIKey: cfg.Agent.BaiduSearchAPIKey,
		}).
		WithMaxWorkers(5).
		WithMaxURLs(5)

	engine := research.NewEngine(researchCfg)

	// Start Interaction Loop
	fmt.Println("\n=== Research Engine 测试 ===")
	fmt.Println("输入 'quit' 或 'exit' 退出")
	fmt.Println()

	for {
		fmt.Print("请输入研究主题: ")
		input, err := reader.ReadString('\n')
		if err != nil {
			// 如果是 EOF (通常是 Ctrl+D)，则优雅退出
			if err.Error() == "EOF" {
				fmt.Println("\n再见!")
				break
			}
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

		fmt.Printf("\n正在研究 '%s' ... (请稍候)\n", input)

		ctx := context.Background()
		report, err := engine.Run(ctx, research.Input{
			Query: input,
		})

		if err != nil {
			fmt.Printf("错误: %v\n", err)
			continue
		}

		fmt.Println("\n=== 研究报告 ===")
		fmt.Printf("查询主题: %s\n", report.Query)
		fmt.Printf("执行结果: %s\n", report.Overview)

		if len(report.Summaries) > 0 {
			fmt.Println("\n=== 详细分析结果 ===")
			for i, s := range report.Summaries {
				fmt.Printf("\n[资源 #%d]\n", i+1)
				if s.Error != "" {
					fmt.Printf("URL: %s\n", s.URL)
					fmt.Printf("状态: 获取/处理失败\n")
					fmt.Printf("错误: %s\n", s.Error)
					continue
				}

				fmt.Printf("标题: %s\n", s.Title)
				fmt.Printf("URL: %s\n", s.URL)
				fmt.Printf("相关度: %.2f\n", s.Relevance)
				fmt.Println("摘要内容:")

				// 简单的缩进处理，让阅读更舒服
				lines := strings.Split(s.Summary, "\n")
				for _, line := range lines {
					if strings.TrimSpace(line) != "" {
						fmt.Printf("  %s\n", line)
					}
				}
			}
		} else {
			fmt.Println("\n(未收集到有效的详细资料)")
		}

		fmt.Println("\n==================")
		fmt.Println()
	}
}
