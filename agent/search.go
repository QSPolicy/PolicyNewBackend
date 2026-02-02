package agent

import (
	"context"
	"fmt"
	"strings"
)

// SearchResultItem 代表单个搜索结果，抽象自百度/Google等搜索引擎的响应
// 尽量涵盖常见的搜索结果字段，以便大模型能获取充分的上下文
type SearchResultItem struct {
	ID        string  `json:"id,omitempty"`        // 引用编号或唯一标识
	Title     string  `json:"title"`               // 网页标题
	URL       string  `json:"url"`                 // 网页地址
	Content   string  `json:"content"`             // 网页内容摘要或片段 (2000字以内)
	Website   string  `json:"website"`             // 站点名称 (e.g. "知乎")
	Icon      string  `json:"icon,omitempty"`      // 网站图标地址
	Date      string  `json:"date,omitempty"`      // 网页发布日期
	Type      string  `json:"type"`                // 资源类型: web, video, image, pdf 等
	Score     float64 `json:"score,omitempty"`     // 相关性评分 [0,1]，对应 rerank_score
	Authority float64 `json:"authority,omitempty"` // 权威性评分 [0,1]，对应 authority_score
	Anchor    string  `json:"anchor,omitempty"`    // 锚文本
}

// SearchOptions 定义搜索时的过滤参数
type SearchOptions struct {
	Query         string   // 搜索关键词
	Limit         int      // 返回结果数量限制 (对应百度API的 top_k)
	TimeRange     string   // 时间限制: week, month, semiyear, year (对应百度的 search_recency_filter)
	ResourceTypes []string // 资源类型: web, video, image (对应百度的 resource_type_filter)
	Sites         []string // 指定搜索站点 (对应百度的 search_filter.match.site)
	BlockSites    []string // 屏蔽站点 (对应百度的 block_websites)
}

// SearchEngine 定义了搜索引擎的通用接口
// 未来接入真实 API 时，只需实现此接口并进行适配
type SearchEngine interface {
	Search(ctx context.Context, opts SearchOptions) ([]SearchResultItem, error)
}

// SearchTool 是包装了搜索引擎的 Tool 实现
type SearchTool struct {
	Engine SearchEngine
}

func NewSearchTool(engine SearchEngine) *SearchTool {
	return &SearchTool{Engine: engine}
}

func (t *SearchTool) Spec() Tool {
	return NewTool("search_internet").
		WithDescription("使用搜索引擎查询实时信息。支持指定时间范围、资源类型和特定站点。").
		WithStringParam("query", "搜索关键词", true).
		WithNumberParam("limit", "返回结果数量限制 (默认 10，最大 50)", false).
		WithEnumParam("time_range", "时间范围限制: week(7天), month(30天), semiyear(180天), year(365天)", []string{"week", "month", "semiyear", "year"}, false).
		WithStringParam("resource_types", "资源类型，多个用逗号分隔 (可选: web, video, image，默认 web)", false).
		WithStringParam("sites", "指定搜索站点域名，多个用逗号分隔，最多20个 (例如: 'zhihu.com,github.com')", false).
		Build()
}

func (t *SearchTool) Execute(ctx context.Context, args map[string]any) (*ToolResult, error) {
	query, ok := args["query"].(string)
	if !ok {
		return NewToolResultError("query 参数是必需的且必须是字符串"), nil
	}

	limit := 10
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
		if limit > 50 {
			limit = 50 // 百度API限制最大50
		}
	}

	opts := SearchOptions{
		Query:         query,
		Limit:         limit,
		ResourceTypes: []string{"web"}, // 默认只搜索网页
	}

	// 解析可选参数
	if tr, ok := args["time_range"].(string); ok && tr != "" {
		opts.TimeRange = tr
	}

	if rt, ok := args["resource_types"].(string); ok && rt != "" {
		opts.ResourceTypes = splitAndTrim(rt)
	}

	if s, ok := args["sites"].(string); ok && s != "" {
		opts.Sites = splitAndTrim(s)
	}

	results, err := t.Engine.Search(ctx, opts)
	if err != nil {
		return NewToolResultError(fmt.Sprintf("搜索执行失败: %v", err)), nil
	}

	return NewToolResult(results), nil
}

// 辅助函数：分割并修剪字符串
func splitAndTrim(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		trimmed := strings.TrimSpace(p)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// 确保 SearchTool 实现了 ToolImplementation 接口
var _ ToolImplementation = (*SearchTool)(nil)

// --- Mock Search Engine Implementation ---

type MockSearchEngine struct{}

func (e *MockSearchEngine) Search(ctx context.Context, opts SearchOptions) ([]SearchResultItem, error) {
	results := make([]SearchResultItem, 0, opts.Limit)

	// 构造模拟数据的后缀，用于展示参数是否生效
	suffix := ""
	if len(opts.ResourceTypes) > 0 {
		suffix += fmt.Sprintf(" [Types: %v]", opts.ResourceTypes)
	}
	if len(opts.Sites) > 0 {
		suffix += fmt.Sprintf(" [Sites: %v]", opts.Sites)
	}
	if opts.TimeRange != "" {
		suffix += fmt.Sprintf(" [Time: %v]", opts.TimeRange)
	}

	for i := 1; i <= opts.Limit; i++ {
		results = append(results, SearchResultItem{
			ID:        fmt.Sprintf("%d", i),
			Title:     fmt.Sprintf("模拟搜索结果 %d: %s%s", i, opts.Query, suffix),
			URL:       fmt.Sprintf("https://example.com/result/%d", i),
			Content:   fmt.Sprintf("这是关于 '%s' 的第 %d 条详细内容的模拟摘要。在真实场景中，这里会有约200字的上下文信息...", opts.Query, i),
			Website:   "Example Search Source",
			Icon:      "https://example.com/favicon.ico",
			Date:      "2023-10-27",
			Type:      "web",
			Score:     0.95 - (float64(i) * 0.05),
			Authority: 0.8,
			Anchor:    opts.Query,
		})
	}
	return results, nil
}
