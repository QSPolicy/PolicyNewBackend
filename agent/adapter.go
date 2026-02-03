package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"policy-backend/agent/def"

	"github.com/openai/openai-go"
)

// ToolHandler 定义了 Agent 与工具交互所需的接口。
// 这将 Agent 与具体的工具实现（如 MCP）解耦。
type ToolHandler interface {
	GetTools() []openai.ChatCompletionToolParam
	ExecuteTool(ctx context.Context, toolCall openai.ChatCompletionMessageToolCall) (string, error)
}

// OpenAIToolAdapter 将 ToolProvider 适配为 ToolHandler 接口。
// 它在我们的领域模型和 OpenAI 的 API 格式之间进行转换。
// 这是适配器模式的应用 —— 将一个接口转换为另一个接口。
type OpenAIToolAdapter struct {
	provider def.ToolProvider
}

// NewOpenAIToolAdapter 为给定的工具提供者创建一个新的适配器。
func NewOpenAIToolAdapter(provider def.ToolProvider) *OpenAIToolAdapter {
	return &OpenAIToolAdapter{
		provider: provider,
	}
}

// GetTools 将我们的 Tool 定义转换为 OpenAI 的函数调用格式。
func (a *OpenAIToolAdapter) GetTools() []openai.ChatCompletionToolParam {
	tools := a.provider.ListTools()
	var openaiTools []openai.ChatCompletionToolParam

	for _, t := range tools {
		// 构建符合 JSON Schema 的参数
		schemaMap := map[string]any{
			"type": t.InputSchema.Type,
		}
		if len(t.InputSchema.Properties) > 0 {
			properties := make(map[string]any)
			for k, v := range t.InputSchema.Properties {
				properties[k] = v
			}
			schemaMap["properties"] = properties
		}
		if len(t.InputSchema.Required) > 0 {
			schemaMap["required"] = t.InputSchema.Required
		}

		openaiTools = append(openaiTools, openai.ChatCompletionToolParam{
			Type: "function",
			Function: openai.FunctionDefinitionParam{
				Name:        t.Name,
				Description: openai.String(t.Description),
				Parameters:  openai.FunctionParameters(schemaMap),
			},
		})
	}
	return openaiTools
}

// ExecuteTool 解析 OpenAI 的工具调用并委派给提供者执行。
func (a *OpenAIToolAdapter) ExecuteTool(ctx context.Context, toolCall openai.ChatCompletionMessageToolCall) (string, error) {
	// 从 JSON 解析参数
	args := make(map[string]any)
	if toolCall.Function.Arguments != "" {
		if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
			return "", fmt.Errorf("无法解析参数: %w", err)
		}
	}

	// 通过提供者执行
	result, err := a.provider.ExecuteTool(ctx, toolCall.Function.Name, args)
	if err != nil {
		return "", err
	}

	// 处理错误结果
	if result.IsError {
		if errMsg, ok := result.Content.(string); ok {
			return "", fmt.Errorf("工具错误: %s", errMsg)
		}
		return "", fmt.Errorf("工具执行失败")
	}

	// 序列化结果内容
	switch v := result.Content.(type) {
	case string:
		return v, nil
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return "", fmt.Errorf("无法序列化工具结果: %w", err)
		}
		return string(bytes), nil
	}
}

// 确保 OpenAIToolAdapter 实现了 ToolHandler
var _ ToolHandler = (*OpenAIToolAdapter)(nil)
