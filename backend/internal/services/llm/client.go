package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/gitsang/futu-agent/backend/internal/config"
)

type Client struct {
	baseURL    string
	model      string
	apiKey     string
	httpClient *http.Client
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatResponse struct {
	Choices []Choice `json:"choices"`
}

type Choice struct {
	Message Message `json:"message"`
}

type TradeDecision struct {
	Action   string  `json:"action"`
	Code     string  `json:"code"`
	Market   string  `json:"market"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Reason   string  `json:"reason"`
}

func NewClient(baseURL, model, apiKey, proxy string) *Client {
	transport := &http.Transport{}

	if proxy != "" {
		proxyURL, err := url.Parse(proxy)
		if err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		}
	}

	return &Client{
		baseURL: baseURL,
		model:   model,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Transport: transport,
			Timeout:   120 * time.Second,
		},
	}
}

func (c *Client) ChatCompletion(ctx context.Context, messages []Message) (string, error) {
	reqBody := ChatRequest{
		Model:    c.model,
		Messages: messages,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	reqURL := fmt.Sprintf("%s/chat/completions", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", reqURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response from LLM")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func (c *Client) AnalyzeAndDecide(ctx context.Context, marketData, positions, accountInfo, tradingStrategy string, rules config.AgentRules) (*TradeDecision, error) {
	var aggressionDesc, goalDesc, actionBias string
	switch rules.AggressionLevel {
	case "aggressive":
		aggressionDesc = "激进"
		goalDesc = "你的目标是通过频繁交易最大化收益"
		actionBias = "积极寻找交易机会"
	case "moderate":
		aggressionDesc = "稳健"
		goalDesc = "你的目标是在风险可控的前提下追求收益"
		actionBias = "在机会明确时果断行动"
	default: // conservative
		aggressionDesc = "保守"
		goalDesc = "你的目标是保护本金，谨慎参与"
		actionBias = "只在高确定性机会出现时行动"
	}

	systemPrompt := fmt.Sprintf(`你是一个%s的AI交易代理，用于模拟交易。%s。

输出以下JSON格式的决策：
{
  "action": "BUY" | "SELL" | "HOLD",
  "code": "股票代码",
  "market": "HK" | "US" | "CN",
  "quantity": 数量,
  "price": 价格,
  "reason": "决策原因"
}

交易策略：
%s

通用规则：
1. 跌幅超过%.1f%%时寻找抄底机会
2. 涨幅超过%.1f%%时考虑止盈
3. 亏损超过%.1f%%时考虑止损
4. 每次交易使用%d-%d%%的可用资金
5. %s
6. 使用技术指标：成交量、价格动量、支撑阻力位
7. 如果现金超过总资产的%d%%，必须找到买入机会
8. 如果持仓亏损超过%.1f%%，考虑加仓或止损

手数规则（必须遵守）：
- %s

规则：
1. 只输出有效的JSON，不要输出其他文本
2. 如果现金超过总资产的%d%%，绝不能输出HOLD - 必须找到买入机会！
3. %s
4. 模拟交易时，偏向行动而非不行动`,
		aggressionDesc,
		goalDesc,
		tradingStrategy,
		rules.BuyOnDipThreshold,
		rules.TakeProfitThreshold,
		rules.StopLossThreshold,
		rules.CashUsageMin,
		rules.CashUsageMax,
		actionBias,
		rules.MaxCashRatio,
		rules.PositionLossThreshold,
		rules.LotSizeRule,
		rules.MaxCashRatio,
		actionBias)

	userPrompt := fmt.Sprintf(`当前市场数据：
%s

当前持仓：
%s

账户信息：
%s

请分析并提供JSON格式的交易决策。`, marketData, positions, accountInfo)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: userPrompt},
	}

	response, err := c.ChatCompletion(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	var decision TradeDecision
	if err := json.Unmarshal([]byte(response), &decision); err != nil {
		return nil, fmt.Errorf("failed to parse LLM response as JSON: %w (response: %s)", err, response)
	}

	return &decision, nil
}
