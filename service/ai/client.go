package ai

import (
	"context"

	deepseek "github.com/cohesion-org/deepseek-go"
)

// 封装 DeepSeek 调用
type AIClient struct {
	Client *deepseek.Client
}

// Chat 封装一次聊天请求
func (c *AIClient) Chat(ctx context.Context, messages []deepseek.ChatCompletionMessage, model string) (string, error) {
	req := &deepseek.ChatCompletionRequest{
		Model:    model,
		Messages: messages,
	}
	resp, err := c.Client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", nil
	}
	return resp.Choices[0].Message.Content, nil
}
