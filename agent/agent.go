package agent

import (
	"context"
	"fmt"

	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/shared"
	"go.uber.org/zap"

	"policy-backend/config"
	"policy-backend/utils"
)

type Agent struct {
	client *openai.Client
	model  shared.ChatModel
	tools  ToolHandler
}

func NewAgent(llmCfg config.LLMConfig, toolHandler ToolHandler) *Agent {
	opts := []option.RequestOption{
		option.WithAPIKey(llmCfg.APIKey),
	}
	if llmCfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(llmCfg.BaseURL))
	}

	c := openai.NewClient(opts...)
	return &Agent{
		client: &c,
		model:  shared.ChatModel(llmCfg.Model),
		tools:  toolHandler,
	}
}

func (a *Agent) Chat(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion) (*openai.ChatCompletion, error) {
	tools := a.tools.GetTools()
	currentMessages := messages

	for {
		params := openai.ChatCompletionNewParams{
			Messages: currentMessages,
			Model:    a.model,
			Tools:    tools,
		}

		chatCompletion, err := a.client.Chat.Completions.New(ctx, params)
		if err != nil {
			return nil, err
		}

		choice := chatCompletion.Choices[0]

		if len(choice.Message.ToolCalls) == 0 {
			return chatCompletion, nil
		}

		// Build assistant message with tool calls
		toolCallsParam := make([]openai.ChatCompletionMessageToolCallParam, 0, len(choice.Message.ToolCalls))
		for _, tc := range choice.Message.ToolCalls {
			toolCallsParam = append(toolCallsParam, openai.ChatCompletionMessageToolCallParam{
				ID:   tc.ID,
				Type: "function",
				Function: openai.ChatCompletionMessageToolCallFunctionParam{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}

		assistantMsg := openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				ToolCalls: toolCallsParam,
			},
		}

		if choice.Message.Content != "" {
			assistantMsg.OfAssistant.Content = openai.ChatCompletionAssistantMessageParamContentUnion{
				OfString: openai.String(choice.Message.Content),
			}
		}

		currentMessages = append(currentMessages, assistantMsg)

		// Execute tools and add results
		for _, toolCall := range choice.Message.ToolCalls {
			utils.Log.Info("Agent calling tool", zap.String("tool", toolCall.Function.Name))
			resultStr, err := a.tools.ExecuteTool(ctx, toolCall)
			if err != nil {
				utils.Log.Error("Error executing tool", zap.String("tool", toolCall.Function.Name), zap.Error(err))
				resultStr = fmt.Sprintf("Error: %v", err)
			}

			currentMessages = append(currentMessages, openai.ChatCompletionMessageParamUnion{
				OfTool: &openai.ChatCompletionToolMessageParam{
					ToolCallID: toolCall.ID,
					Content: openai.ChatCompletionToolMessageParamContentUnion{
						OfString: openai.String(resultStr),
					},
				},
			})
		}
	}
}

// ChatStream executes a streaming chat completion.
// If onChunk is provided, it receives content chunks as they arrive.
func (a *Agent) ChatStream(ctx context.Context, messages []openai.ChatCompletionMessageParamUnion, onChunk func(string)) (*openai.ChatCompletion, error) {
	tools := a.tools.GetTools()
	currentMessages := messages

	type toolAccumulator struct {
		ID        string
		Name      string
		Arguments string
	}

	for {
		params := openai.ChatCompletionNewParams{
			Messages: currentMessages,
			Model:    a.model,
			Tools:    tools,
		}

		stream := a.client.Chat.Completions.NewStreaming(ctx, params)

		var accumulatedContent string
		toolCallsMap := make(map[int64]*toolAccumulator)

		for stream.Next() {
			chunk := stream.Current()
			if len(chunk.Choices) == 0 {
				continue
			}

			delta := chunk.Choices[0].Delta

			// Handle Content
			if delta.Content != "" {
				accumulatedContent += delta.Content
				if onChunk != nil {
					onChunk(delta.Content)
				}
			}

			// Handle ToolCalls
			for _, tc := range delta.ToolCalls {
				idx := tc.Index
				if _, ok := toolCallsMap[idx]; !ok {
					toolCallsMap[idx] = &toolAccumulator{}
				}
				acc := toolCallsMap[idx]

				if tc.ID != "" {
					acc.ID = tc.ID
				}
				if tc.Function.Name != "" {
					acc.Name += tc.Function.Name
				}
				if tc.Function.Arguments != "" {
					acc.Arguments += tc.Function.Arguments
				}
			}
		}

		if err := stream.Err(); err != nil {
			return nil, err
		}
		stream.Close()

		// Reconstruct complete tool calls
		var toolCalls []openai.ChatCompletionMessageToolCall
		if len(toolCallsMap) > 0 {
			// Find max index
			var maxIdx int64 = -1
			for idx := range toolCallsMap {
				if idx > maxIdx {
					maxIdx = idx
				}
			}

			for i := int64(0); i <= maxIdx; i++ {
				if acc, ok := toolCallsMap[i]; ok {
					toolCalls = append(toolCalls, openai.ChatCompletionMessageToolCall{
						ID:   acc.ID,
						Type: "function",
						Function: openai.ChatCompletionMessageToolCallFunction{
							Name:      acc.Name,
							Arguments: acc.Arguments,
						},
					})
				}
			}
		}

		// If no tool calls, returns the accumulated message.
		if len(toolCalls) == 0 {
			msg := openai.ChatCompletionMessage{
				Role:    "assistant",
				Content: accumulatedContent,
			}

			return &openai.ChatCompletion{
				Choices: []openai.ChatCompletionChoice{
					{
						Message: msg,
					},
				},
			}, nil
		}

		// If there are tool calls, we must execute them and continue the loop.

		toolCallsParam := make([]openai.ChatCompletionMessageToolCallParam, 0, len(toolCalls))
		for _, tc := range toolCalls {
			toolCallsParam = append(toolCallsParam, openai.ChatCompletionMessageToolCallParam{
				ID:   tc.ID,
				Type: "function",
				Function: openai.ChatCompletionMessageToolCallFunctionParam{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}

		assistantMsg := openai.ChatCompletionMessageParamUnion{
			OfAssistant: &openai.ChatCompletionAssistantMessageParam{
				ToolCalls: toolCallsParam,
			},
		}
		if accumulatedContent != "" {
			assistantMsg.OfAssistant.Content = openai.ChatCompletionAssistantMessageParamContentUnion{
				OfString: openai.String(accumulatedContent),
			}
		}
		currentMessages = append(currentMessages, assistantMsg)

		// Execute tools
		for _, tc := range toolCalls {
			utils.Log.Info("Agent calling tool (streaming mode)", zap.String("tool", tc.Function.Name))
			resultStr, err := a.tools.ExecuteTool(ctx, tc)
			if err != nil {
				utils.Log.Error("Error executing tool", zap.String("tool", tc.Function.Name), zap.Error(err))
				resultStr = fmt.Sprintf("Error: %v", err)
			}

			currentMessages = append(currentMessages, openai.ChatCompletionMessageParamUnion{
				OfTool: &openai.ChatCompletionToolMessageParam{
					ToolCallID: tc.ID,
					Content: openai.ChatCompletionToolMessageParamContentUnion{
						OfString: openai.String(resultStr),
					},
				},
			})
		}
		// Loop continues to send tool results back to LLM...
	}
}
