package research

import (
	"policy-backend/agent"
)

// Config 研究工作流的完整配置
type Config struct {
	// Planner 使用高能力模型进行搜索决策
	PlannerLLM agent.LLMConfig

	// Worker 使用轻量模型进行内容摘要
	WorkerLLM agent.LLMConfig

	// 搜索工具配置
	SearchConfig agent.ToolConfig

	// 并发配置
	MaxWorkers int // 默认 5
	MaxURLs    int // 默认 10
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		MaxWorkers: 5,
		MaxURLs:    10,
	}
}

// WithPlannerLLM 设置 Planner 的 LLM 配置
func (c *Config) WithPlannerLLM(cfg agent.LLMConfig) *Config {
	c.PlannerLLM = cfg
	return c
}

// WithWorkerLLM 设置 Worker 的 LLM 配置
func (c *Config) WithWorkerLLM(cfg agent.LLMConfig) *Config {
	c.WorkerLLM = cfg
	return c
}

// WithSearchConfig 设置搜索配置
func (c *Config) WithSearchConfig(cfg agent.ToolConfig) *Config {
	c.SearchConfig = cfg
	return c
}

// WithMaxWorkers 设置最大并发数
func (c *Config) WithMaxWorkers(n int) *Config {
	c.MaxWorkers = n
	return c
}

// WithMaxURLs 设置最大 URL 数量
func (c *Config) WithMaxURLs(n int) *Config {
	c.MaxURLs = n
	return c
}
