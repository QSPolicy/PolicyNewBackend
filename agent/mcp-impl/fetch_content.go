package mcpimpl

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"policy-backend/agent/def"
	"policy-backend/utils"
	"time"
)

// === Tool: Fetch Content ===

// FetchContentTool 是一个能够抓取网页内容的工具
type FetchContentTool struct {
	client *http.Client
}

func NewFetchContentTool() *FetchContentTool {
	return &FetchContentTool{
		client: &http.Client{
			// 设置合理的超时，防止长时间挂起
			Timeout: 15 * time.Second,
		},
	}
}

func (t *FetchContentTool) Spec() def.Tool {
	return def.NewTool("fetch_content").
		WithDescription("获取指定URL的网页文本内容。用于读取外部网页以获取更详细的信息。").
		WithStringParam("url", "要读取的网页URL", true).
		Build()
}

func (t *FetchContentTool) Execute(ctx context.Context, args map[string]any) (*def.ToolResult, error) {
	url, ok := args["url"].(string)
	if !ok {
		return def.NewToolResultError("url parameter is required"), nil
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return def.NewToolResultError(fmt.Sprintf("Failed to create request: %v", err)), nil
	}

	// 模拟浏览器 User-Agent，避免部分网站拒绝
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// 执行请求
	resp, err := t.client.Do(req)
	if err != nil {
		utils.Log.Warn(fmt.Sprintf("Fetch failed for %s: %v", url, err))
		return def.NewToolResultError(fmt.Sprintf("Network error fetching URL: %v", err)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return def.NewToolResultError(fmt.Sprintf("HTTP Error %d", resp.StatusCode)), nil
	}

	// 限制读取大小（例如 50KB），防止大文件消耗过多资源
	const maxBytes = 50 * 1024
	bodyBytes, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes))
	if err != nil {
		return def.NewToolResultError(fmt.Sprintf("Error reading body: %v", err)), nil
	}

	// 简单的 HTML 清理逻辑
	// 注意：在实际生产环境中，这应该替换为 goquery 或 html2text 等专业库
	rawText := string(bodyBytes)
	cleanText := simplifyHTML(rawText)

	return def.NewToolResultText(cleanText), nil
}

// simplifyHTML 是一个占位符，用于未来的 HTML 清理
func simplifyHTML(html string) string {
	// 这里可以添加逻辑去除 script, style 标签等
	if len(html) > 5000 {
		return html[:5000] + "...(truncated)"
	}
	return html
}
