package research

// PageSummary 存储对于单个URL的摘要结果
type PageSummary struct {
	URL       string  `json:"url"`
	Title     string  `json:"title"`
	Summary   string  `json:"summary"`
	Relevance float64 `json:"relevance"` // 0.0 - 1.0
	Error     string  `json:"error,omitempty"`
}

// Report 研究工作流的最终输出
type Report struct {
	Query     string        `json:"query"`
	Summaries []PageSummary `json:"summaries"`
	Overview  string        `json:"overview"` // 最终的综合报告
}

// Input 研究工作流的输入
type Input struct {
	Query      string `json:"query"`
	MaxURLs    int    `json:"max_urls,omitempty"`    // 最多收集的 URL 数量
	MaxWorkers int    `json:"max_workers,omitempty"` // 最大并发 worker 数
}

// StoreKeys 定义 Store 中使用的键常量
const (
	KeyQuery         = "query"
	KeyCollectedURLs = "collected_urls"
	KeySummaries     = "summaries"
)
