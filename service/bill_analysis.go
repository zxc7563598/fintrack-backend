package service

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"strings"

	deepseek "github.com/cohesion-org/deepseek-go"
	"github.com/zxc7563598/fintrack-backend/service/ai"
)

type BillAnalysis struct {
	ai *ai.AIClient
}

func NewBillAnalysis(aiClient *ai.AIClient) *BillAnalysis {
	return &BillAnalysis{ai: aiClient}
}

// AnalysisFromBytes 使用 CSV 字节流生成损友式吐槽报告
func (b *BillAnalysis) AnalysisFromBytes(ctx context.Context, csvBytes []byte) (string, error) {
	reader := csv.NewReader(bytes.NewReader(csvBytes))
	// 读取表头
	_, err := reader.Read()
	if err != nil {
		return "", fmt.Errorf("读取表头失败: %w", err)
	}
	const chunkSize = 1000
	lineCount := 0
	part := 1
	var chunkBuilder strings.Builder
	var partialReports []string
	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return "", fmt.Errorf("读取 CSV 出错: %w", err)
		}
		chunkBuilder.WriteString(strings.Join(record, ","))
		chunkBuilder.WriteString("\n")
		lineCount++
		if lineCount >= chunkSize {
			report, err := b.callAI(ctx, chunkBuilder.String(), part)
			if err != nil {
				return "", err
			}
			partialReports = append(partialReports, report)

			chunkBuilder.Reset()
			lineCount = 0
			part++
		}
	}
	if lineCount > 0 {
		report, err := b.callAI(ctx, chunkBuilder.String(), part)
		if err != nil {
			return "", err
		}
		partialReports = append(partialReports, report)
	}
	// 最终汇总
	finalPrompt := `
以下是多个分块的损友式吐槽总结，请把它们合并成一份完整的吐槽报告，保持幽默生动风格。

要求：
- 输出内容必须直接开始吐槽，不能出现任何标题、标注、提示语、总结、报告名称、【】符号，但该有的换行可以有。
- 只输出吐槽正文，不允许出现额外的说明或格式。

以下是多份吐槽总结：
` + strings.Join(partialReports, "\n\n")
	finalMsgs := []deepseek.ChatCompletionMessage{
		{Role: deepseek.ChatMessageRoleUser, Content: finalPrompt},
	}
	return b.ai.Chat(ctx, finalMsgs, deepseek.DeepSeekChat)
}

// callAI 与之前一样
func (b *BillAnalysis) callAI(ctx context.Context, csvChunk string, part int) (string, error) {
	prompt := fmt.Sprintf(`
我要给你一份CSV数据（第 %d 段），里面记录了我的收支明细（平台,收支类型,交易类型,商品名称,对方,支付方式,金额,交易时间,备注）。

请直接生成一份“损友式”的财务吐槽报告：
- 语气像朋友聊天一样，随意吐槽我把钱花哪儿了、赚的钱从哪里来。
- 可以夸张、幽默、调侃、吐槽，带些生活化口语。
- 可以点出最大的一笔支出、最常消费的对象、最高收入或者最离谱的交易，但不要写成“最大/最奇葩”等硬性标签，直接自然描述。
- 收支类型有“收入”“支出”“不计收支”“未知”，其中“不计收支”和“未知”不能算进收入或支出。
- 所有金额必须精确，不能使用“约”“大概”等模糊词。

请严格遵守以下规则：
- 输出内容必须直接开始吐槽，不能出现任何标题、标注、提示语、总结、报告名称、【】符号，但该有的换行可以有。
- 只输出吐槽正文，不允许出现额外的说明或格式。
- 所有金额必须精确，不能用“约”“大概”等模糊词

CSV 数据如下：
%s
`, part, csvChunk)
	msgs := []deepseek.ChatCompletionMessage{
		{Role: deepseek.ChatMessageRoleUser, Content: prompt},
	}
	return b.ai.Chat(ctx, msgs, deepseek.DeepSeekChat)
}
