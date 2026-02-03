package def

// Tool 表示一个可调用的工具定义。
// 这是我们自己的轻量级模型，独立于任何外部库。
type Tool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema ToolInputSchema `json:"inputSchema"`
}

// ToolInputSchema 定义了工具参数的 JSON Schema。
// 遵循 JSON Schema 规范以兼容 OpenAI 的函数调用。
type ToolInputSchema struct {
	Type       string                    `json:"type"` // 始终为 "object"
	Properties map[string]PropertySchema `json:"properties,omitempty"`
	Required   []string                  `json:"required,omitempty"`
}

// PropertySchema 定义了单个参数的 schema。
type PropertySchema struct {
	Type        string          `json:"type"`
	Description string          `json:"description,omitempty"`
	Enum        []string        `json:"enum,omitempty"`
	Default     any             `json:"default,omitempty"`
	Items       *PropertySchema `json:"items,omitempty"`
}

// ToolResult 表示工具执行的结果。
type ToolResult struct {
	Content any  `json:"content"`
	IsError bool `json:"isError,omitempty"`
}

// NewToolResult 创建一个成功的工具结果。
func NewToolResult(content any) *ToolResult {
	return &ToolResult{Content: content, IsError: false}
}

// NewToolResultText 创建一个成功的文本结果。
func NewToolResultText(text string) *ToolResult {
	return &ToolResult{Content: text, IsError: false}
}

// NewToolResultError 创建一个错误结果。
func NewToolResultError(errMsg string) *ToolResult {
	return &ToolResult{Content: errMsg, IsError: true}
}

// ToolBuilder 提供了一个用于构建工具定义的流式 API。
// 遵循建造者模式以提高可扩展性（开闭原则）。
type ToolBuilder struct {
	tool Tool
}

// NewTool 开始使用给定的名称构建一个新工具。
func NewTool(name string) *ToolBuilder {
	return &ToolBuilder{
		tool: Tool{
			Name: name,
			InputSchema: ToolInputSchema{
				Type:       "object",
				Properties: make(map[string]PropertySchema),
			},
		},
	}
}

// WithDescription 设置工具的描述。
func (b *ToolBuilder) WithDescription(desc string) *ToolBuilder {
	b.tool.Description = desc
	return b
}

// WithStringParam 为工具添加一个字符串参数。
func (b *ToolBuilder) WithStringParam(name, description string, required bool) *ToolBuilder {
	b.tool.InputSchema.Properties[name] = PropertySchema{
		Type:        "string",
		Description: description,
	}
	if required {
		b.tool.InputSchema.Required = append(b.tool.InputSchema.Required, name)
	}
	return b
}

// WithNumberParam 为工具添加一个数字参数。
func (b *ToolBuilder) WithNumberParam(name, description string, required bool) *ToolBuilder {
	b.tool.InputSchema.Properties[name] = PropertySchema{
		Type:        "number",
		Description: description,
	}
	if required {
		b.tool.InputSchema.Required = append(b.tool.InputSchema.Required, name)
	}
	return b
}

// WithBoolParam 为工具添加一个布尔参数。
func (b *ToolBuilder) WithBoolParam(name, description string, required bool) *ToolBuilder {
	b.tool.InputSchema.Properties[name] = PropertySchema{
		Type:        "boolean",
		Description: description,
	}
	if required {
		b.tool.InputSchema.Required = append(b.tool.InputSchema.Required, name)
	}
	return b
}

// WithEnumParam 添加一个枚举（具有允许值的字符串）参数。
func (b *ToolBuilder) WithEnumParam(name, description string, values []string, required bool) *ToolBuilder {
	b.tool.InputSchema.Properties[name] = PropertySchema{
		Type:        "string",
		Description: description,
		Enum:        values,
	}
	if required {
		b.tool.InputSchema.Required = append(b.tool.InputSchema.Required, name)
	}
	return b
}

// WithStringArrayParam 为工具添加一个字符串数组参数。
func (b *ToolBuilder) WithStringArrayParam(name, description string, required bool) *ToolBuilder {
	b.tool.InputSchema.Properties[name] = PropertySchema{
		Type:        "array",
		Description: description,
		Items: &PropertySchema{
			Type: "string",
		},
	}
	if required {
		b.tool.InputSchema.Required = append(b.tool.InputSchema.Required, name)
	}
	return b
}

// Build 返回构建好的 Tool。
func (b *ToolBuilder) Build() Tool {
	return b.tool
}
