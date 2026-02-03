package workflow

import (
	"context"
	"sync"

	"policy-backend/utils"

	"go.uber.org/zap"
)

// Task 定义一个可执行的任务
type Task interface {
	ID() string
	Execute(ctx context.Context) (any, error)
}

// TaskResult 包装任务执行结果
type TaskResult struct {
	TaskID string
	Result any
	Error  error
}

// WorkerPool 管理并发任务执行
type WorkerPool struct {
	maxWorkers int
	sem        chan struct{}
}

// NewWorkerPool 创建一个新的 Worker 池
func NewWorkerPool(maxWorkers int) *WorkerPool {
	if maxWorkers <= 0 {
		maxWorkers = 5
	}
	return &WorkerPool{
		maxWorkers: maxWorkers,
		sem:        make(chan struct{}, maxWorkers),
	}
}

// Execute 并行执行所有任务并返回结果
func (p *WorkerPool) Execute(ctx context.Context, tasks []Task) []TaskResult {
	results := make([]TaskResult, len(tasks))
	var wg sync.WaitGroup
	var mu sync.Mutex
	resultIndex := 0

	for _, task := range tasks {
		wg.Add(1)
		go func(t Task) {
			defer wg.Done()

			// 获取信号量
			select {
			case p.sem <- struct{}{}:
				defer func() { <-p.sem }()
			case <-ctx.Done():
				mu.Lock()
				results[resultIndex] = TaskResult{
					TaskID: t.ID(),
					Error:  ctx.Err(),
				}
				resultIndex++
				mu.Unlock()
				return
			}

			// 执行任务
			result, err := t.Execute(ctx)
			if err != nil {
				utils.Log.Warn("Task failed",
					zap.String("task_id", t.ID()),
					zap.Error(err),
				)
			}

			mu.Lock()
			results[resultIndex] = TaskResult{
				TaskID: t.ID(),
				Result: result,
				Error:  err,
			}
			resultIndex++
			mu.Unlock()
		}(task)
	}

	wg.Wait()
	return results
}

// ExecuteFunc 简化版本，直接接受函数切片
func (p *WorkerPool) ExecuteFunc(ctx context.Context, fns []func(context.Context) (any, error)) []TaskResult {
	tasks := make([]Task, len(fns))
	for i, fn := range fns {
		tasks[i] = &funcTask{
			id: string(rune('A' + i)),
			fn: fn,
		}
	}
	return p.Execute(ctx, tasks)
}

// funcTask 将函数包装为 Task
type funcTask struct {
	id string
	fn func(context.Context) (any, error)
}

func (t *funcTask) ID() string {
	return t.id
}

func (t *funcTask) Execute(ctx context.Context) (any, error) {
	return t.fn(ctx)
}
