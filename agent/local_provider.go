package agent

import (
	"context"
	"fmt"
)

// LocalToolProvider 是一个简单的进程内工具注册表。
// 它在内存中存储工具及其处理程序，以便直接调用。
type LocalToolProvider struct {
	tools    []Tool
	handlers map[string]ToolImplementation
}

// NewLocalToolProvider 创建一个新的本地工具提供者，并注册默认工具。
func NewLocalToolProvider() *LocalToolProvider {
	p := &LocalToolProvider{
		tools:    make([]Tool, 0),
		handlers: make(map[string]ToolImplementation),
	}

	// 注册默认工具
	p.Register(&TimeTool{})
	p.Register(&EchoTool{})

	return p
}

// Register 向提供者添加新的工具实现。
// 这是扩展点 —— 可以在不修改现有代码的情况下添加新工具。
func (p *LocalToolProvider) Register(impl ToolImplementation) {
	spec := impl.Spec()
	p.tools = append(p.tools, spec)
	p.handlers[spec.Name] = impl
}

// ListTools 返回所有注册的工具。
func (p *LocalToolProvider) ListTools() []Tool {
	return p.tools
}

// ExecuteTool 按名称执行工具。
func (p *LocalToolProvider) ExecuteTool(ctx context.Context, name string, args map[string]any) (*ToolResult, error) {
	impl, ok := p.handlers[name]
	if !ok {
		return nil, fmt.Errorf("未找到工具: %s", name)
	}
	return impl.Execute(ctx, args)
}

// 确保 LocalToolProvider 实现了 ToolProvider 接口
var _ ToolProvider = (*LocalToolProvider)(nil)
