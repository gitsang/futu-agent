package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gitsang/futu-agent/backend/internal/config"
	"github.com/gitsang/futu-agent/backend/internal/services/futu"
	"github.com/gitsang/futu-agent/backend/internal/services/llm"
	"github.com/gitsang/futu-agent/backend/internal/store"
)

type Engine struct {
	store          *store.MemoryStore
	futuClient     *futu.Client
	llmClient      *llm.Client
	config         *config.Config
	agents         map[string]*AgentWorker
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	tradingEnabled bool
}

type AgentWorker struct {
	AgentID string
	Config  config.AgentConfig
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
	mu      sync.Mutex
}

func NewEngine(store *store.MemoryStore, futuClient *futu.Client, llmClient *llm.Client, cfg *config.Config) *Engine {
	return &Engine{
		store:          store,
		futuClient:     futuClient,
		llmClient:      llmClient,
		config:         cfg,
		agents:         make(map[string]*AgentWorker),
		tradingEnabled: cfg.TradingEnabled,
	}
}

func (e *Engine) Start(ctx context.Context) error {
	e.ctx, e.cancel = context.WithCancel(ctx)

	e.loadAgents()

	go e.runLoop()

	log.Println("Agent engine started")
	return nil
}

func (e *Engine) Stop() {
	e.cancel()

	e.mu.Lock()
	defer e.mu.Unlock()

	for _, agent := range e.agents {
		agent.Stop()
	}

	log.Println("Agent engine stopped")
}

func (e *Engine) loadAgents() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, agentCfg := range e.config.Agents {
		if agentCfg.Enabled {
			worker := &AgentWorker{
				AgentID: agentCfg.ID,
				Config:  agentCfg,
				running: true,
			}
			e.agents[agentCfg.ID] = worker
			log.Printf("Loaded agent: %s (%s)", agentCfg.ID, agentCfg.Name)
		}
	}
}

func (e *Engine) runLoop() {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-e.ctx.Done():
			return
		case <-ticker.C:
			e.executeCycle()
		}
	}
}

func (e *Engine) executeCycle() {
	if !e.tradingEnabled {
		return
	}

	if !e.futuClient.IsConnected() {
		log.Println("Futu client not connected, skipping cycle")
		return
	}

	e.mu.RLock()
	agents := make([]*AgentWorker, 0, len(e.agents))
	for _, agent := range e.agents {
		if agent.Config.Enabled {
			agents = append(agents, agent)
		}
	}
	e.mu.RUnlock()

	for _, agent := range agents {
		go e.executeAgent(agent)
	}
}

func (e *Engine) executeAgent(worker *AgentWorker) {
	ctx, cancel := context.WithTimeout(e.ctx, 5*time.Minute)
	defer cancel()

	log.Printf("Executing agent %s", worker.AgentID)

	market := worker.Config.Market

	accountFunds, err := e.futuClient.GetAccountFunds(ctx, market)
	if err != nil {
		log.Printf("Failed to get account funds: %v", err)
		return
	}

	positions, err := e.futuClient.GetPositions(ctx, market)
	if err != nil {
		log.Printf("Failed to get positions: %v", err)
		return
	}

	var marketDataLines []string
	marketDataLines = append(marketDataLines, fmt.Sprintf("=== %s市场数据 ===", market))
	marketDataLines = append(marketDataLines, "")
	marketDataLines = append(marketDataLines, "【账户概况】")
	marketDataLines = append(marketDataLines, fmt.Sprintf("总资产: %.2f", accountFunds.TotalAssets))
	marketDataLines = append(marketDataLines, fmt.Sprintf("可用资金: %.2f", accountFunds.Cash))
	marketDataLines = append(marketDataLines, fmt.Sprintf("持仓市值: %.2f", accountFunds.MarketValue))
	if accountFunds.TotalAssets > 0 {
		marketDataLines = append(marketDataLines, fmt.Sprintf("仓位比例: %.1f%%", accountFunds.MarketValue/accountFunds.TotalAssets*100))
	}
	marketDataLines = append(marketDataLines, "")
	marketDataLines = append(marketDataLines, "【当前持仓及行情】")
	if len(positions) == 0 {
		marketDataLines = append(marketDataLines, "无持仓")
	} else {
		for _, pos := range positions {
			pnlPct := 0.0
			if pos.AvgCost > 0 {
				pnlPct = (pos.CurrentPrice - pos.AvgCost) / pos.AvgCost * 100
			}
			marketDataLines = append(marketDataLines, fmt.Sprintf("- %s (%s): 持有%d股, 成本价%.2f, 现价%.2f, 盈亏%.2f%%", 
				pos.Name, pos.Code, pos.Quantity, pos.AvgCost, pos.CurrentPrice, pnlPct))

			quote, err := e.futuClient.GetQuote(ctx, market, pos.Code)
			if err == nil {
				marketDataLines = append(marketDataLines, fmt.Sprintf("  行情: 今开%.2f 最高%.2f 最低%.2f 昨收%.2f 涨跌幅%.2f%% 振幅%.2f%% 换手率%.2f%% 成交量%d 成交额%.2f", 
					quote.Open, quote.High, quote.Low, quote.LastClose, quote.ChangePct, quote.Amplitude, quote.TurnoverRate, quote.Volume, quote.Turnover))
			}
		}
	}

	marketData := strings.Join(marketDataLines, "\n")

	positionsJSON, _ := json.Marshal(positions)
	accountJSON, _ := json.Marshal(accountFunds)

	decision, err := e.llmClient.AnalyzeAndDecide(ctx, marketData, string(positionsJSON), string(accountJSON))
	if err != nil {
		log.Printf("Agent %s failed to analyze: %v", worker.AgentID, err)
		return
	}

	log.Printf("Agent %s decision: %s %s %d @ %.2f - %s", worker.AgentID, decision.Action, decision.Code, decision.Quantity, decision.Price, decision.Reason)

	if decision.Action == "HOLD" {
		return
	}

	if e.tradingEnabled {
		orderID, err := e.futuClient.PlaceOrder(ctx, decision.Market, decision.Code, decision.Action, decision.Price, decision.Quantity)
		if err != nil {
			log.Printf("Failed to execute trade: %v", err)
			return
		}
		log.Printf("Trade executed successfully, order ID: %s", orderID)
	}

	e.store.SaveDecision(store.TradeDecision{
		AgentID:   worker.AgentID,
		StockCode: decision.Code,
		Market:    decision.Market,
		Action:    decision.Action,
		Quantity:  decision.Quantity,
		Price:     decision.Price,
		Reason:    decision.Reason,
		Executed:  true,
	})
}

func (e *Engine) GetFutuOpendStatus() string {
	if e.futuClient.IsConnected() {
		return "connected"
	}
	return "disconnected"
}

func (e *Engine) IsTradingEnabled() bool {
	return e.tradingEnabled
}

func (a *AgentWorker) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}
	a.running = false
}
