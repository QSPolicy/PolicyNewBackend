package agent

import (
"context"

"github.com/mark3labs/mcp-go/mcp"
)

// MCPServer defines the interface for an MCP provider
type MCPServer interface {
	ListTools() []mcp.Tool
	ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (*mcp.CallToolResult, error)
}

// MCPTool defines the interface for a specific tool implementation
type MCPTool interface {
	Spec() mcp.Tool
	Execute(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error)
}
