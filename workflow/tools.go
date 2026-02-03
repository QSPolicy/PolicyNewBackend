package workflow

import (
	"context"
	"fmt"
	"policy-backend/agent/def"
)

// === Generic Context Tool: Set Collector ===

// SetCollectorTool 是一个通用的工作流工具，允许 Agent 将数据收集到 Context Store 的集合中（去重）。
// 例如：收集搜索到的 URLs。
type SetCollectorTool struct {
	store       *Store
	name        string
	description string
	storeKey    string // 存入 Store 的哪个 Key
	paramName   string // 工具参数名，例如 "url"
}

// NewSetCollectorTool 创建一个新的集合收集工具
func NewSetCollectorTool(store *Store, name, description, storeKey, paramName string) *SetCollectorTool {
	return &SetCollectorTool{
		store:       store,
		name:        name,
		description: description,
		storeKey:    storeKey,
		paramName:   paramName,
	}
}

func (t *SetCollectorTool) Spec() def.Tool {
	return def.NewTool(t.name).
		WithDescription(t.description).
		WithStringParam(t.paramName, fmt.Sprintf("The value to collect for %s", t.paramName), true).
		Build()
}

func (t *SetCollectorTool) Execute(ctx context.Context, args map[string]any) (*def.ToolResult, error) {
	val, ok := args[t.paramName].(string)
	if !ok {
		return def.NewToolResultError(fmt.Sprintf("Parameter '%s' is required and must be a string", t.paramName)), nil
	}

	t.store.AddToSet(t.storeKey, val)
	return def.NewToolResultText(fmt.Sprintf("Successfully collected: %s", val)), nil
}

// === Generic Context Tool: Key Value Setter ===

// KVSetterTool 允许 Agent 直接设置 Context Store 中的某个 Key 的值。
// 例如：设置 "final_answer" 或 "search_summary"。
type KVSetterTool struct {
	store       *Store
	name        string
	description string
	storeKey    string
	paramName   string
}

// NewKVSetterTool 创建一个新的 KV 设置工具
func NewKVSetterTool(store *Store, name, description, storeKey, paramName string) *KVSetterTool {
	return &KVSetterTool{
		store:       store,
		name:        name,
		description: description,
		storeKey:    storeKey,
		paramName:   paramName,
	}
}

func (t *KVSetterTool) Spec() def.Tool {
	return def.NewTool(t.name).
		WithDescription(t.description).
		WithStringParam(t.paramName, fmt.Sprintf("The value to set for %s", t.paramName), true).
		Build()
}

func (t *KVSetterTool) Execute(ctx context.Context, args map[string]any) (*def.ToolResult, error) {
	val, ok := args[t.paramName].(string)
	if !ok {
		return def.NewToolResultError(fmt.Sprintf("Parameter '%s' is required", t.paramName)), nil
	}

	t.store.Set(t.storeKey, val)
	return def.NewToolResultText(fmt.Sprintf("Successfully set value for %s", t.storeKey)), nil
}
