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

func (c *Client) AnalyzeAndDecide(ctx context.Context, marketData, positions, accountInfo string) (*TradeDecision, error) {
	systemPrompt := `You are an aggressive AI trading agent for simulated trading. Your goal is to maximize returns through active trading.

Output a JSON object with the following structure:
{
  "action": "BUY" | "SELL" | "HOLD",
  "code": "stock code",
  "market": "HK" | "US" | "CN",
  "quantity": number,
  "price": number,
  "reason": "explanation for the decision"
}

Trading Strategy:
1. Be AGGRESSIVE - this is simulated trading, take risks!
2. Look for opportunities to buy on dips (when price drops > 1%)
3. Take profits when gains exceed 3%
4. Cut losses when losses exceed 5%
5. Use 30-50% of available cash per trade
6. Trade frequently - don't just hold!
7. Use technical indicators: volume spikes, price momentum, support/resistance
8. If you have > 70% cash, you MUST find something to buy
9. If a position is down > 3%, consider averaging down or cutting loss

Rules:
1. Only output valid JSON, no other text
2. Never output HOLD if cash > 70% of total assets - find something to buy!
3. Be decisive - make a trade decision every cycle
4. For simulation, err on the side of action, not inaction`

	userPrompt := fmt.Sprintf(`Current Market Data:
%s

Current Positions:
%s

Account Information:
%s

Analyze and provide a trading decision in JSON format. Remember: be aggressive, trade actively!`, marketData, positions, accountInfo)

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
