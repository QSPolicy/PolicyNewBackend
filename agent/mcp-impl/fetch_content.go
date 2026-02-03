package mcpimpl

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"policy-backend/agent/def"
	"policy-backend/utils"
	"strings"
	"time"

	"codeberg.org/readeck/go-readability/v2"
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
			Timeout: 30 * time.Second,
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
	rawURL, ok := args["url"].(string)
	if !ok {
		return def.NewToolResultError("url parameter is required"), nil
	}

	// 验证 URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil || parsedURL.Scheme == "" || parsedURL.Host == "" {
		return def.NewToolResultError(fmt.Sprintf("Invalid URL: %s", rawURL)), nil
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, "GET", rawURL, nil)
	if err != nil {
		return def.NewToolResultError(fmt.Sprintf("Failed to create request: %v", err)), nil
	}

	// 模拟浏览器 User-Agent，避免部分网站拒绝
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	// 执行请求
	resp, err := t.client.Do(req)
	if err != nil {
		utils.Log.Warn(fmt.Sprintf("Fetch failed for %s: %v", rawURL, err))
		return def.NewToolResultError(fmt.Sprintf("Network error fetching URL: %v", err)), nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return def.NewToolResultError(fmt.Sprintf("HTTP Error %d", resp.StatusCode)), nil
	}

	// 使用 go-readability 解析内容
	// 注意：go-readability 会尝试提取主要内容，去除广告和导航栏等
	article, err := readability.FromReader(resp.Body, parsedURL)
	if err != nil {
		utils.Log.Warn(fmt.Sprintf("Readability failed for %s: %v", rawURL, err))
		return def.NewToolResultError(fmt.Sprintf("Failed to extract content: %v", err)), nil
	}

	// 构建返回信息，包含标题和正文
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Title: %s\n", article.Title()))
	if article.Byline() != "" {
		sb.WriteString(fmt.Sprintf("Byline: %s\n", article.Byline()))
	}
	if article.Excerpt() != "" {
		sb.WriteString(fmt.Sprintf("Excerpt: %s\n", article.Excerpt()))
	}

	sb.WriteString("\n=== Content ===\n\n")

	// 获取纯文本内容
	var textBuilder strings.Builder
	err = article.RenderText(&textBuilder)
	if err != nil {
		// 降级处理：如果不成功，至少返回标题
		utils.Log.Warn(fmt.Sprintf("Failed to render text for %s: %v", rawURL, err))
		sb.WriteString("Error decoding content text.")
	} else {
		content := textBuilder.String()
		// 简单的截断过长的内容，防止上下文溢出
		if len(content) > 100000 {
			content = content[:100000] + "\n...(truncated)"
		}
		sb.WriteString(content)
	}

	return def.NewToolResultText(sb.String()), nil
}
