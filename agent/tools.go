package agent

import (
	"context"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
)

// --- Tools Implementations ---

// TimeTool implements MCPTool for getting current time
type TimeTool struct{}

func (t *TimeTool) Spec() mcp.Tool {
	return mcp.NewTool("get_current_time",
		mcp.WithDescription("Get the current server time"),
	)
}

func (t *TimeTool) Execute(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText(time.Now().Format(time.RFC3339)), nil
}

// EchoTool implements MCPTool for echoing messages
type EchoTool struct{}

func (t *EchoTool) Spec() mcp.Tool {
	return mcp.NewTool("echo_message",
		mcp.WithDescription("Echo back the input message"),
		mcp.WithString("message", mcp.Required(), mcp.Description("The message to echo")),
	)
}

func (t *EchoTool) Execute(ctx context.Context, args map[string]interface{}) (*mcp.CallToolResult, error) {
	msg, ok := args["message"].(string)
	if !ok {
		return mcp.NewToolResultError("message argument is required and must be a string"), nil
	}
	return mcp.NewToolResultText(fmt.Sprintf("Echo: %s", msg)), nil
}
