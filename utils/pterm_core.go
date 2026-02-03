package utils

import (
	"fmt"
	"time"

	"github.com/pterm/pterm"
	"go.uber.org/zap/zapcore"
)

// PtermCore 是一个自定义的 zapcore.Core，使用 pterm 进行漂亮的终端输出
type PtermCore struct {
	zapcore.LevelEnabler
	fields []zapcore.Field
}

// NewPtermCore 创建一个新的 PtermCore
func NewPtermCore(level zapcore.LevelEnabler) *PtermCore {
	return &PtermCore{
		LevelEnabler: level,
		fields:       make([]zapcore.Field, 0),
	}
}

// With 添加字段上下文到新的 Core
func (c *PtermCore) With(fields []zapcore.Field) zapcore.Core {
	// 复制现有字段并添加新字段
	newFields := make([]zapcore.Field, len(c.fields)+len(fields))
	copy(newFields, c.fields)
	copy(newFields[len(c.fields):], fields)

	return &PtermCore{
		LevelEnabler: c.LevelEnabler,
		fields:       newFields,
	}
}

// Check 检查日志级别是否启用
func (c *PtermCore) Check(ent zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(ent.Level) {
		return ce.AddCore(ent, c)
	}
	return ce
}

// Write 写入日志条目
func (c *PtermCore) Write(ent zapcore.Entry, fields []zapcore.Field) error {
	// 合并上下文中的字段和当前条目的字段
	allFields := make([]zapcore.Field, 0, len(c.fields)+len(fields))
	allFields = append(allFields, c.fields...)
	allFields = append(allFields, fields...)

	// 将字段转换为 map 以便更容易查找特定键（如 HTTP 字段）
	fieldMap := make(map[string]interface{})
	enc := zapcore.NewMapObjectEncoder()
	for _, f := range allFields {
		f.AddTo(enc)
	}
	fieldMap = enc.Fields

	// 生成日志前缀
	levelStyle := pterm.NewStyle(pterm.FgWhite, pterm.BgBlack)
	levelText := ent.Level.CapitalString()
	
	switch ent.Level {
	case zapcore.DebugLevel:
		levelStyle = pterm.NewStyle(pterm.FgGray)
	case zapcore.InfoLevel:
		levelStyle = pterm.NewStyle(pterm.FgLightCyan)
	case zapcore.WarnLevel:
		levelStyle = pterm.NewStyle(pterm.FgYellow)
	case zapcore.ErrorLevel:
		levelStyle = pterm.NewStyle(pterm.FgRed, pterm.Bold)
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		levelStyle = pterm.NewStyle(pterm.FgWhite, pterm.BgRed, pterm.Bold)
	}

	// 格式化时间
	timeStr := ent.Time.Format("15:04:05.000")
	prefix := pterm.Gray(timeStr) + " " + levelStyle.Sprint(fmt.Sprintf("%-5s", levelText)) + " "

	// 检查是否是 HTTP 日志（包含特定的 HTTP 字段）
	if isHTTPLog(fieldMap) {
		c.writeHTTPLog(prefix, ent.Message, fieldMap)
	} else {
		c.writeStandardLog(prefix, ent.Message, fieldMap)
	}

	return nil
}

// Sync 同步日志（对于 stdout 通常不需要操作，但为了接口兼容）
func (c *PtermCore) Sync() error {
	return nil
}

// isHTTPLog 判断是否为 HTTP 请求日志
func isHTTPLog(fields map[string]interface{}) bool {
	_, hasStatus := fields["status"]
	_, hasMethod := fields["method"]
	_, hasURI := fields["uri"]
	return hasStatus && hasMethod && hasURI
}

// writeHTTPLog 格式化 HTTP 日志
func (c *PtermCore) writeHTTPLog(prefix string, msg string, fields map[string]interface{}) {
	// 提取关键信息
	var status int64
	switch v := fields["status"].(type) {
	case int:
		status = int64(v)
	case int64:
		status = v
	case float64:
		status = int64(v)
	}
	
	method, _ := fields["method"].(string)
	uri, _ := fields["uri"].(string)
	
	var latency int64
	switch v := fields["latency_ms"].(type) {
	case int:
		latency = int64(v)
	case int64:
		latency = v
	case float64:
		latency = int64(v)
	case time.Duration:
		latency = int64(v / time.Millisecond)
	}

	ip, _ := fields["ip"].(string)
	
	// 状态码样式
	var statusStyle *pterm.Style
	switch {
	case status >= 200 && status < 300:
		statusStyle = pterm.NewStyle(pterm.FgGreen)
	case status >= 300 && status < 400:
		statusStyle = pterm.NewStyle(pterm.FgBlue)
	case status >= 400 && status < 500:
		statusStyle = pterm.NewStyle(pterm.FgYellow)
	case status >= 500:
		statusStyle = pterm.NewStyle(pterm.FgRed)
	default:
		statusStyle = pterm.NewStyle(pterm.FgWhite)
	}

	// 拼接输出
	// 格式: [TIME] [LEVEL] [STATUS] METHOD URI (LATENCY) - MSG
	
	output := fmt.Sprintf("%s%s %s %s (%dms) - %s",
		prefix,
		statusStyle.Sprintf("[%d]", status),
		pterm.Magenta(method),
		pterm.Cyan(uri),
		latency,
		msg,
	)

	// 如果有错误信息，单独追加
	if errVal, ok := fields["error"]; ok {
		output += fmt.Sprintf("\n    └─ %s: %v", pterm.Red("Error"), errVal)
	}
	
	// 如果是详细 debug，显示其他字段
	// ...

	pterm.Println(output)
	if ip != "" {
		// 可选：在下一行显示 IP 等辅助信息，或者保持简洁
	}
}

// writeStandardLog 格式化普通日志
func (c *PtermCore) writeStandardLog(prefix string, msg string, fields map[string]interface{}) {
	pterm.Println(prefix + msg)
	
	// 打印所有字段
	if len(fields) > 0 {
		var nodes []pterm.TreeNode
		for k, v := range fields {
			// 过滤调用栈等不需要显示的字段
			if k == "stacktrace" {
				continue
			}

			// 构建节点
			node := pterm.TreeNode{Text: pterm.Cyan(k)}

			// 检查是否为嵌套 map (复杂结构)
			if nestedMap, ok := v.(map[string]interface{}); ok {
				node.Children = c.buildTreeNodes(nestedMap)
			} else {
				// 简单值直接显示
				node.Text = pterm.Sprintf("%s: %v", pterm.Gray(k), v)
			}
			nodes = append(nodes, node)
		}

		if len(nodes) > 0 {
			// 使用 Tree 渲染，不设 Root Text 使其从第一层开始展示
			_ = pterm.DefaultTree.WithRoot(pterm.TreeNode{Children: nodes}).Render()
		}
	}
}

// buildTreeNodes 递归构建 pterm 树节点
func (c *PtermCore) buildTreeNodes(m map[string]interface{}) []pterm.TreeNode {
	var nodes []pterm.TreeNode
	for k, v := range m {
		node := pterm.TreeNode{Text: pterm.Cyan(k)}
		
		if nestedMap, ok := v.(map[string]interface{}); ok {
			// 如果是嵌套 map，递归构建子节点
			node.Children = c.buildTreeNodes(nestedMap)
		} else {
			// 如果是叶子节点，显示值
			node.Text = pterm.Sprintf("%s: %v", pterm.Cyan(k), pterm.Green(v))
		}
		nodes = append(nodes, node)
	}
	return nodes
}
