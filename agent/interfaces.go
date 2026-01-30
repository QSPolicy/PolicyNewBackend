package agent

import "context"

// ToolProvider 定义了工具注册表的接口。
// 任何实现（本地、远程等）都必须满足此接口。
// 这遵循了接口隔离原则 —— 客户端只依赖于它们需要的东西。
type ToolProvider interface {
	// ListTools 返回所有可用的工具。
	ListTools() []Tool
	// ExecuteTool 根据名称和给定的参数执行工具。
	ExecuteTool(ctx context.Context, name string, args map[string]any) (*ToolResult, error)
}

// ToolImplementation 定义了单个工具实现的接口。
// 添加新工具只需实现此接口（开闭原则）。
type ToolImplementation interface {
	// Spec 返回工具的定义（名称、描述、参数）。
	Spec() Tool
	// Execute 使用给定的参数运行工具。
	Execute(ctx context.Context, args map[string]any) (*ToolResult, error)
}
