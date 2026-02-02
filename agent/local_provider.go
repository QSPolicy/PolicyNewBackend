package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"policy-backend/utils"
	"time"

	"go.uber.org/zap"
)

// LocalToolProvider 是一个简单的进程内工具注册表。
// 它在内存中存储工具及其处理程序，以便直接调用。
type LocalToolProvider struct {
	tools    []Tool
	handlers map[string]ToolImplementation
}

// NewLocalToolProvider 创建一个新的本地工具提供者，并注册所有的工具。
// 它接受可选的配置参数，用于初始化工具。
func NewLocalToolProvider(cfg *ToolConfig) *LocalToolProvider {
	p := &LocalToolProvider{
		tools:    make([]Tool, 0),
		handlers: make(map[string]ToolImplementation),
	}

	// 注册搜索工具
	// 优先使用 Config 中的 Key，如果不存在则使用 Mock
	var searchEngine SearchEngine
	if cfg != nil && cfg.BaiduAPIKey != "" {
		searchEngine = NewBaiduSearchEngine(cfg.BaiduAPIKey)
	} else {
		searchEngine = &MockSearchEngine{}
	}
	p.Register(NewSearchTool(searchEngine))

	// 在这里继续注册其他所有工具...
	// p.Register(...)

	return p
}

// ToolConfig 定义工具提供者所需的配置
type ToolConfig struct {
	BaiduAPIKey string
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
		utils.Log.Warn("[MCP] 工具未找到", zap.String("tool", name))
		return nil, fmt.Errorf("未找到工具: %s", name)
	}

	// 记录工具调用开始
	argsJSON, _ := json.Marshal(args)
	utils.Log.Info("[MCP] 工具调用开始",
		zap.String("tool", name),
		zap.String("args", string(argsJSON)),
	)

	startTime := time.Now()
	result, err := impl.Execute(ctx, args)
	duration := time.Since(startTime)

	if err != nil {
		utils.Log.Error("[MCP] 工具执行错误",
			zap.String("tool", name),
			zap.Duration("duration", duration),
			zap.Error(err),
		)
		return nil, err
	}

	// 记录工具调用结果
	resultJSON, _ := json.Marshal(result.Content)
	resultPreview := string(resultJSON)
	if len(resultPreview) > 500 {
		resultPreview = resultPreview[:500] + "...(truncated)"
	}

	utils.Log.Info("[MCP] 工具调用完成",
		zap.String("tool", name),
		zap.Duration("duration", duration),
		zap.Bool("isError", result.IsError),
		zap.String("result_preview", resultPreview),
	)

	return result, nil
}

// 确保 LocalToolProvider 实现了 ToolProvider 接口
var _ ToolProvider = (*LocalToolProvider)(nil)
