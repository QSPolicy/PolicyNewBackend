package research

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openai/openai-go"
	"go.uber.org/zap"

	"policy-backend/agent"
	mcpimpl "policy-backend/agent/mcp-impl"
	"policy-backend/utils"
	"policy-backend/workflow"
)

// Engine 实现研究工作流
type Engine struct {
	config *Config
}

// NewEngine 创建研究引擎
func NewEngine(cfg *Config) *Engine {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Engine{config: cfg}
}

// Run 执行研究工作流
func (e *Engine) Run(ctx context.Context, input Input) (*Report, error) {
	store := workflow.NewStore()
	store.Set(KeyQuery, input.Query)

	utils.Log.Info("Research workflow started", zap.String("query", input.Query))

	// Phase 1: Planner
	if err := e.runPlannerPhase(ctx, store); err != nil {
		return nil, fmt.Errorf("planner phase failed: %w", err)
	}

	urls := store.GetSet(KeyCollectedURLs)
	utils.Log.Info("Planner completed", zap.Int("urls_collected", len(urls)))

	if len(urls) == 0 {
		return &Report{
			Query:    input.Query,
			Overview: "No relevant URLs found during search.",
		}, nil
	}

	// Phase 2: Workers
	summaries := e.runWorkerPhase(ctx, store, urls)

	// Phase 3: Aggregate
	report := &Report{
		Query:     input.Query,
		Summaries: summaries,
		Overview:  fmt.Sprintf("Analyzed %d URLs for query: %s", len(summaries), input.Query),
	}

	utils.Log.Info("Research workflow completed", zap.Int("summaries", len(summaries)))
	return report, nil
}

func (e *Engine) runPlannerPhase(ctx context.Context, store *workflow.Store) error {
	// 创建工具提供者
	provider := agent.NewLocalToolProvider(&e.config.SearchConfig)

	// 注册 collect_url 工具
	// 使用 workflow 通用集合收集工具
	collectTool := workflow.NewSetCollectorTool(
		store,
		"collect_url",
		"记录一个看起来包含相关信息的URL，供后续详细分析。当你发现搜索结果中有一个链接可能包含用户问题的答案时，请使用此工具。",
		KeyCollectedURLs,
		"url",
	)
	provider.Register(collectTool)

	// 创建 Agent
	toolHandler := agent.NewOpenAIToolAdapter(provider)
	plannerAgent := agent.NewAgent(e.config.PlannerLLM, toolHandler)

	// 构建消息
	query := store.GetString(KeyQuery)
	messages := []openai.ChatCompletionMessageParamUnion{
		{
			OfSystem: &openai.ChatCompletionSystemMessageParam{
				Content: openai.ChatCompletionSystemMessageParamContentUnion{
					OfString: openai.String(GetPlannerPrompt(query)),
				},
			},
		},
		{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(fmt.Sprintf("Please research: %s", query)),
				},
			},
		},
	}

	// 执行
	_, err := plannerAgent.Chat(ctx, messages)
	return err
}

func (e *Engine) runWorkerPhase(ctx context.Context, store *workflow.Store, urls []string) []PageSummary {
	query := store.GetString(KeyQuery)
	pool := workflow.NewWorkerPool(e.config.MaxWorkers)

	// 构建任务
	tasks := make([]workflow.Task, len(urls))
	for i, url := range urls {
		tasks[i] = &summarizeTask{
			url:    url,
			query:  query,
			config: e.config,
		}
	}

	// 并行执行
	results := pool.Execute(ctx, tasks)

	// 收集结果
	summaries := make([]PageSummary, 0, len(results))
	for _, r := range results {
		if r.Error != nil {
			summaries = append(summaries, PageSummary{
				URL:   r.TaskID,
				Error: r.Error.Error(),
			})
			continue
		}
		if summary, ok := r.Result.(*PageSummary); ok && summary != nil {
			summaries = append(summaries, *summary)
		}
	}

	return summaries
}

// summarizeTask 实现 workflow.Task
type summarizeTask struct {
	url    string
	query  string
	config *Config
}

func (t *summarizeTask) ID() string {
	return t.url
}

func (t *summarizeTask) Execute(ctx context.Context) (any, error) {
	// 创建工具提供者
	provider := agent.NewLocalToolProvider(nil)
	fetchTool := mcpimpl.NewFetchContentTool()
	provider.Register(fetchTool)

	// 创建 Agent
	toolHandler := agent.NewOpenAIToolAdapter(provider)
	workerAgent := agent.NewAgent(t.config.WorkerLLM, toolHandler)

	// 构建消息
	messages := []openai.ChatCompletionMessageParamUnion{
		{
			OfSystem: &openai.ChatCompletionSystemMessageParam{
				Content: openai.ChatCompletionSystemMessageParamContentUnion{
					OfString: openai.String(WorkerSystemPrompt),
				},
			},
		},
		{
			OfUser: &openai.ChatCompletionUserMessageParam{
				Content: openai.ChatCompletionUserMessageParamContentUnion{
					OfString: openai.String(fmt.Sprintf("URL: %s\nQuery: %s", t.url, t.query)),
				},
			},
		},
	}

	// 执行
	resp, err := workerAgent.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}

	// 解析结果
	content := resp.Choices[0].Message.Content
	var summary PageSummary
	if err := json.Unmarshal([]byte(content), &summary); err != nil {
		// 如果解析失败，返回原始内容
		return &PageSummary{
			URL:       t.url,
			Summary:   content,
			Relevance: 0.5,
		}, nil
	}

	summary.URL = t.url
	return &summary, nil
}
