package agent

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// LocalMCPServer wraps the MCP server instance and allows direct internal access
type LocalMCPServer struct {
	server   *server.MCPServer
	tools    []mcp.Tool
	handlers map[string]server.ToolHandlerFunc
}

// NewLocalMCPServer creates a new local MCP server with predefined tools
func NewLocalMCPServer() *LocalMCPServer {
	s := server.NewMCPServer(
		"local-optimization-utils",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	l := &LocalMCPServer{
		server:   s,
		tools:    make([]mcp.Tool, 0),
		handlers: make(map[string]server.ToolHandlerFunc),
	}

	// Register default tools
	l.RegisterTool(&TimeTool{})
	l.RegisterTool(&EchoTool{})

	return l
}

// RegisterTool adds a new tool to the server
func (l *LocalMCPServer) RegisterTool(t MCPTool) {
	spec := t.Spec()
	// Create adapter handler
	handler := func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			// Handle empty args case or invalid type
			if request.Params.Arguments == nil {
				args = make(map[string]interface{})
			} else {
				return mcp.NewToolResultError("arguments seem invalid"), nil
			}
		}
		return t.Execute(ctx, args)
	}

	l.server.AddTool(spec, handler)
	l.tools = append(l.tools, spec)
	l.handlers[spec.Name] = handler
}

// ListTools returns the list of available tools
func (l *LocalMCPServer) ListTools() []mcp.Tool {
	return l.tools
}

// ExecuteTool executes a tool directly
func (l *LocalMCPServer) ExecuteTool(ctx context.Context, name string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	handler, ok := l.handlers[name]
	if !ok {
		return nil, fmt.Errorf("tool not found: %s", name)
	}

	req := mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "tools/call",
		},
		Params: mcp.CallToolParams{
			Name:      name,
			Arguments: args,
		},
	}

	return handler(ctx, req)
}
