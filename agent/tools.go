package agent

import (
	"context"
	"fmt"
	"time"
)

// --- 内置工具实现 ---
// 每个工具都实现 ToolImplementation 接口。
// 要添加新工具，只需创建一个新结构体并实现 Spec() 和 Execute() 方法。

// TimeTool 返回当前服务器时间。
type TimeTool struct{}

func (t *TimeTool) Spec() Tool {
	return NewTool("get_current_time").
		WithDescription("获取 RFC3339 格式的当前服务器时间").
		Build()
}

func (t *TimeTool) Execute(ctx context.Context, args map[string]any) (*ToolResult, error) {
	return NewToolResultText(time.Now().Format(time.RFC3339)), nil
}

// 确保 TimeTool 实现了 ToolImplementation 接口
var _ ToolImplementation = (*TimeTool)(nil)

// EchoTool 原样返回输入的消息。
type EchoTool struct{}

func (t *EchoTool) Spec() Tool {
	return NewTool("echo_message").
		WithDescription("回显输入的消息").
		WithStringParam("message", "要回显的消息", true).
		Build()
}

func (t *EchoTool) Execute(ctx context.Context, args map[string]any) (*ToolResult, error) {
	msg, ok := args["message"].(string)
	if !ok {
		return NewToolResultError("message 参数是必需的且必须是字符串"), nil
	}
	return NewToolResultText(fmt.Sprintf("Echo: %s", msg)), nil
}

// 确保 EchoTool 实现了 ToolImplementation 接口
var _ ToolImplementation = (*EchoTool)(nil)
