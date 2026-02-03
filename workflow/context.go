package workflow

import (
	"context"
	"sync"
)

// Status 定义工作流/任务的状态
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
)

// Context 是工作流执行的通用上下文接口
// 具体的业务实现应该嵌入此接口或组合使用
type Context interface {
	context.Context
	GetStatus() Status
	SetStatus(Status)
	GetError() error
	SetError(error)
}

// BaseContext 提供 Context 接口的基础实现
// 业务代码可以嵌入此结构体来获得基本功能
type BaseContext struct {
	context.Context
	mu     sync.RWMutex
	status Status
	err    error
}

func NewBaseContext(ctx context.Context) *BaseContext {
	return &BaseContext{
		Context: ctx,
		status:  StatusPending,
	}
}

func (c *BaseContext) GetStatus() Status {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.status
}

func (c *BaseContext) SetStatus(s Status) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.status = s
}

func (c *BaseContext) GetError() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.err
}

func (c *BaseContext) SetError(err error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.err = err
	if err != nil {
		c.status = StatusFailed
	}
}

// Store 提供线程安全的键值存储，用于在阶段间传递数据
type Store struct {
	mu   sync.RWMutex
	data map[string]any
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]any),
	}
}

func (s *Store) Set(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

func (s *Store) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

func (s *Store) GetString(key string) string {
	v, ok := s.Get(key)
	if !ok {
		return ""
	}
	str, _ := v.(string)
	return str
}

func (s *Store) GetSlice(key string) []any {
	v, ok := s.Get(key)
	if !ok {
		return nil
	}
	slice, _ := v.([]any)
	return slice
}

// AppendToSlice 线程安全地向切片追加元素
func (s *Store) AppendToSlice(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.data[key]
	if !ok {
		s.data[key] = []any{value}
		return
	}
	slice, ok := existing.([]any)
	if !ok {
		s.data[key] = []any{value}
		return
	}
	s.data[key] = append(slice, value)
}

// AddToSet 线程安全地向集合添加元素（去重）
func (s *Store) AddToSet(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	existing, ok := s.data[key]
	if !ok {
		s.data[key] = map[string]bool{value: true}
		return
	}
	set, ok := existing.(map[string]bool)
	if !ok {
		s.data[key] = map[string]bool{value: true}
		return
	}
	set[value] = true
}

// GetSet 获取集合的所有键
func (s *Store) GetSet(key string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	existing, ok := s.data[key]
	if !ok {
		return nil
	}
	set, ok := existing.(map[string]bool)
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(set))
	for k := range set {
		keys = append(keys, k)
	}
	return keys
}
