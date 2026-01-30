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
