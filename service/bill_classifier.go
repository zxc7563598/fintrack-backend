package service

import (
	"context"
	"fmt"

	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/zxc7563598/fintrack-backend/service/ai"
)

type BillClassifier struct {
	ai *ai.AIClient
}

func NewBillClassifier(aiClient *ai.AIClient) *BillClassifier {
	return &BillClassifier{ai: aiClient}
}

func (b *BillClassifier) Classify(ctx context.Context, accounts string) (string, error) {
	// 封装提示词
	prompt := "你是一个账户分类助手。\n" +
		"我会给你一组账单中的“账户名称”，其中很多账户是同一个账户叠加了优惠或红包信息。\n" +
		"你的任务是：\n" +
		"1. 提取出核心账户（如“兴业银行信用卡(0104)”、“余额宝”、“花呗”、“账户余额”等）。\n" +
		"2. 把所有带有附加描述的账户（例如“兴业银行信用卡(0104)&红包”、“余额宝&红包”）统一映射到它们所属的核心账户。\n" +
		"3. 输出一个 JSON 对象，key 为原始账户名称，value 为归类后的核心账户名称。\n" +
		"4. 只输出纯 JSON，不要任何多余的文字，不要 ```json ```，不要 ```，不要多余解释。\n" +
		"输入账户：" + accounts

	msgs := []deepseek.ChatCompletionMessage{
		{Role: deepseek.ChatMessageRoleUser, Content: prompt},
	}
	fmt.Println("发送哪哦那个:", prompt)
	return b.ai.Chat(ctx, msgs, deepseek.DeepSeekChat)
}
