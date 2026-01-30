package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
)

// ToolHandler defines the interface required by the Agent to interact with tools.
// This decouples the Agent from the specific tool implementation (e.g., MCP).
type ToolHandler interface {
	GetTools() []openai.ChatCompletionToolParam
	ExecuteTool(ctx context.Context, toolCall openai.ChatCompletionMessageToolCall) (string, error)
}

// MCPAdapter adapts an MCPServer to the ToolHandler interface.
// It translates OpenAI tool calls to MCP tool executions and vice versa.
type MCPAdapter struct {
	server MCPServer
}

// NewMCPAdapter creates a new adapter for the given MCP server.
func NewMCPAdapter(server MCPServer) *MCPAdapter {
	return &MCPAdapter{
		server: server,
	}
}

// GetTools lists available tools from the MCP server and converts them to OpenAI format.
func (a *MCPAdapter) GetTools() []openai.ChatCompletionToolParam {
	mcpTools := a.server.ListTools()
	var openaiTools []openai.ChatCompletionToolParam

	for _, t := range mcpTools {
		// Convert MCP ToolInputSchema to map[string]any for FunctionParameters
		schemaMap := map[string]any{
			"type": t.InputSchema.Type,
		}
		if t.InputSchema.Properties != nil {
			schemaMap["properties"] = t.InputSchema.Properties
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

// ExecuteTool calls the corresponding tool on the MCP server.
func (a *MCPAdapter) ExecuteTool(ctx context.Context, toolCall openai.ChatCompletionMessageToolCall) (string, error) {
	args := make(map[string]interface{})
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
		return "", fmt.Errorf("failed to unmarshal arguments: %w", err)
	}

	result, err := a.server.ExecuteTool(ctx, toolCall.Function.Name, args)
	if err != nil {
		return "", err
	}
	if result.IsError {
		return "", fmt.Errorf("tool execution failed")
	}
	bytes, err := json.Marshal(result.Content)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tool result: %w", err)
	}
	return string(bytes), nil
}
