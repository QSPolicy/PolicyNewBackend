package mcpimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"policy-backend/utils"
	"time"

	"go.uber.org/zap"
)

// BaiduSearchEngine 实现了基于百度搜索API的 SearchEngine 接口
type BaiduSearchEngine struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

// NewBaiduSearchEngine 创建一个新的百度搜索引擎实例
func NewBaiduSearchEngine(apiKey string) *BaiduSearchEngine {
	return &BaiduSearchEngine{
		APIKey:  apiKey,
		BaseURL: "https://qianfan.baidubce.com",
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// --- 百度 API 请求结构 ---

type baiduSearchRequest struct {
	Messages            []baiduMessage        `json:"messages"`
	SearchSource        string                `json:"search_source,omitempty"`
	ResourceTypeFilter  []baiduSearchResource `json:"resource_type_filter,omitempty"`
	SearchFilter        *baiduSearchFilter    `json:"search_filter,omitempty"`
	BlockWebsites       []string              `json:"block_websites,omitempty"`
	SearchRecencyFilter string                `json:"search_recency_filter,omitempty"`
}

type baiduMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type baiduSearchResource struct {
	Type string `json:"type"`
	TopK int    `json:"top_k"`
}

type baiduSearchFilter struct {
	Match *baiduFilterMatch `json:"match,omitempty"`
}

type baiduFilterMatch struct {
	Site []string `json:"site,omitempty"`
}

// --- 百度 API 响应结构 ---

type baiduSearchResponse struct {
	RequestID  string           `json:"request_id"`
	Code       string           `json:"code,omitempty"`
	Message    string           `json:"message,omitempty"`
	References []baiduReference `json:"references"`
}

type baiduReference struct {
	ID             int     `json:"id"`
	Title          string  `json:"title"`
	URL            string  `json:"url"`
	Content        string  `json:"content"`
	Website        string  `json:"website"`
	Icon           string  `json:"icon"`
	Date           string  `json:"date"`
	Type           string  `json:"type"`
	WebAnchor      string  `json:"web_anchor"`
	RerankScore    float64 `json:"rerank_score"`
	AuthorityScore float64 `json:"authority_score"`
}

// Search 执行百度搜索
func (e *BaiduSearchEngine) Search(ctx context.Context, opts SearchOptions) ([]SearchResultItem, error) {
	// 构建请求体
	req := e.buildRequest(opts)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		utils.Log.Error("[BaiduSearch] 序列化请求失败", zap.Error(err))
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	utils.Log.Debug("[BaiduSearch] 发送请求",
		zap.String("query", opts.Query),
		zap.Int("limit", opts.Limit),
		zap.String("time_range", opts.TimeRange),
		zap.Strings("resource_types", opts.ResourceTypes),
		zap.Strings("sites", opts.Sites),
		zap.String("request_body", string(jsonBody)),
	)

	// 创建 HTTP 请求
	httpReq, err := http.NewRequestWithContext(ctx, "POST", e.BaseURL+"/v2/ai_search/web_search", bytes.NewReader(jsonBody))
	if err != nil {
		utils.Log.Error("[BaiduSearch] 创建请求失败", zap.Error(err))
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Appbuilder-Authorization", "Bearer "+e.APIKey)

	// 发送请求
	startTime := time.Now()
	resp, err := e.HTTPClient.Do(httpReq)
	if err != nil {
		utils.Log.Error("[BaiduSearch] HTTP请求失败",
			zap.Error(err),
			zap.Duration("duration", time.Since(startTime)),
		)
		return nil, fmt.Errorf("请求百度搜索API失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		utils.Log.Error("[BaiduSearch] 读取响应失败", zap.Error(err))
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	utils.Log.Debug("[BaiduSearch] 收到响应",
		zap.Int("status_code", resp.StatusCode),
		zap.Duration("duration", time.Since(startTime)),
		zap.Int("response_size", len(body)),
	)

	// 解析响应
	var baiduResp baiduSearchResponse
	if err := json.Unmarshal(body, &baiduResp); err != nil {
		utils.Log.Error("[BaiduSearch] 解析响应失败",
			zap.Error(err),
			zap.String("body", string(body)),
		)
		return nil, fmt.Errorf("解析响应失败: %w (body: %s)", err, string(body))
	}

	// 检查错误
	if baiduResp.Code != "" {
		utils.Log.Error("[BaiduSearch] API返回错误",
			zap.String("code", baiduResp.Code),
			zap.String("message", baiduResp.Message),
			zap.String("request_id", baiduResp.RequestID),
		)
		return nil, fmt.Errorf("百度搜索API错误 [%s]: %s", baiduResp.Code, baiduResp.Message)
	}

	// 转换为通用结果格式
	results := e.convertResults(baiduResp.References)

	utils.Log.Info("[BaiduSearch] 搜索完成",
		zap.String("query", opts.Query),
		zap.String("request_id", baiduResp.RequestID),
		zap.Int("result_count", len(results)),
		zap.Duration("total_duration", time.Since(startTime)),
	)

	return results, nil
}

// buildRequest 根据 SearchOptions 构建百度API请求
func (e *BaiduSearchEngine) buildRequest(opts SearchOptions) baiduSearchRequest {
	req := baiduSearchRequest{
		Messages: []baiduMessage{
			{Role: "user", Content: opts.Query},
		},
		SearchSource: "baidu_search_v2",
	}

	// 设置资源类型过滤
	if len(opts.ResourceTypes) > 0 {
		for _, rt := range opts.ResourceTypes {
			topK := opts.Limit
			// 根据百度API限制调整 top_k
			switch rt {
			case "web":
				if topK > 50 {
					topK = 50
				}
			case "video":
				if topK > 10 {
					topK = 10
				}
			case "image":
				if topK > 30 {
					topK = 30
				}
			}
			req.ResourceTypeFilter = append(req.ResourceTypeFilter, baiduSearchResource{
				Type: rt,
				TopK: topK,
			})
		}
	} else {
		// 默认只搜索网页
		req.ResourceTypeFilter = []baiduSearchResource{
			{Type: "web", TopK: opts.Limit},
		}
	}

	// 设置站点过滤
	if len(opts.Sites) > 0 {
		req.SearchFilter = &baiduSearchFilter{
			Match: &baiduFilterMatch{
				Site: opts.Sites,
			},
		}
	}

	// 设置屏蔽站点
	if len(opts.BlockSites) > 0 {
		req.BlockWebsites = opts.BlockSites
	}

	// 设置时间过滤
	if opts.TimeRange != "" {
		req.SearchRecencyFilter = opts.TimeRange
	}

	return req
}

// convertResults 将百度API响应转换为通用格式
func (e *BaiduSearchEngine) convertResults(refs []baiduReference) []SearchResultItem {
	results := make([]SearchResultItem, 0, len(refs))
	for _, ref := range refs {
		results = append(results, SearchResultItem{
			ID:        fmt.Sprintf("%d", ref.ID),
			Title:     ref.Title,
			URL:       ref.URL,
			Content:   ref.Content,
			Website:   ref.Website,
			Icon:      ref.Icon,
			Date:      ref.Date,
			Type:      ref.Type,
			Score:     ref.RerankScore,
			Authority: ref.AuthorityScore,
			Anchor:    ref.WebAnchor,
		})
	}
	return results
}

// 确保 BaiduSearchEngine 实现了 SearchEngine 接口
var _ SearchEngine = (*BaiduSearchEngine)(nil)
