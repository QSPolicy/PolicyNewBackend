package workflow

import (
	"context"

	"policy-backend/agent"
)

// Phase 定义工作流中的一个阶段
type Phase interface {
	Name() string
	Run(ctx context.Context, store *Store) error
}

// Engine 是工作流引擎的通用接口
type Engine interface {
	// Run 执行工作流，返回结果
	Run(ctx context.Context, input any) (any, error)
	// GetPhases 返回工作流的所有阶段（用于调试/监控）
	GetPhases() []Phase
}

// AgentConfig 定义一个 Agent 的配置
type AgentConfig struct {
	LLM   agent.LLMConfig
	Tools []string // 工具名称列表
}

// PipelineConfig 定义流水线式工作流的配置
type PipelineConfig struct {
	Name        string
	Description string
	Phases      []PhaseConfig
}

// PhaseConfig 定义单个阶段的配置
type PhaseConfig struct {
	Name        string
	AgentConfig AgentConfig
	Parallel    bool // 是否并行执行（用于 fan-out）
	MaxWorkers  int  // 并行时的最大 worker 数
}

// BaseEngine 提供引擎的基础实现
type BaseEngine struct {
	phases []Phase
}

func NewBaseEngine() *BaseEngine {
	return &BaseEngine{
		phases: make([]Phase, 0),
	}
}

func (e *BaseEngine) AddPhase(p Phase) {
	e.phases = append(e.phases, p)
}

func (e *BaseEngine) GetPhases() []Phase {
	return e.phases
}

// RunSequential 按顺序执行所有阶段
func (e *BaseEngine) RunSequential(ctx context.Context, store *Store) error {
	for _, phase := range e.phases {
		if err := phase.Run(ctx, store); err != nil {
			return err
		}
	}
	return nil
}
