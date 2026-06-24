package agent

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/gitsang/futu-agent/backend/internal/config"
	"github.com/gitsang/futu-agent/backend/internal/services/futu"
	"github.com/gitsang/futu-agent/backend/internal/services/llm"
)

type Engine struct {
	db          *sql.DB
	futuClient  *futu.Client
	llmClient   *llm.Client
	config      *config.Config
	agents      map[int64]*AgentWorker
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
	tradingEnabled bool
}

type AgentWorker struct {
	ID        int64
	AgentID   string
	Config    AgentConfig
	ctx       context.Context
	cancel    context.CancelFunc
	running   bool
	mu        sync.Mutex
}

type AgentConfig struct {
	Name            string
	Description     string
	TradingStrategy string
	Enabled         bool
}

func NewEngine(db *sql.DB, futuClient *futu.Client, llmClient *llm.Client, cfg *config.Config) *Engine {
	return &Engine{
		db:             db,
		futuClient:     futuClient,
		llmClient:      llmClient,
		config:         cfg,
		agents:         make(map[int64]*AgentWorker),
		tradingEnabled: cfg.TradingEnabled,
	}
}

func (e *Engine) Start(ctx context.Context) error {
	e.ctx, e.cancel = context.WithCancel(ctx)

	if err := e.loadAgents(); err != nil {
		return fmt.Errorf("failed to load agents: %w", err)
	}

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

func (e *Engine) loadAgents() error {
	rows, err := e.db.Query(`
		SELECT id, agent_id, name, description, trading_strategy, enabled
		FROM agent_configs
		WHERE enabled = TRUE
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var id int64
		var agentID, name, description, strategy string
		var enabled bool

		if err := rows.Scan(&id, &agentID, &name, &description, &strategy, &enabled); err != nil {
			log.Printf("Failed to scan agent config: %v", err)
			continue
		}

		worker := &AgentWorker{
			ID:      id,
			AgentID: agentID,
			Config: AgentConfig{
				Name:            name,
				Description:     description,
				TradingStrategy: strategy,
				Enabled:         enabled,
			},
		}

		e.agents[id] = worker
	}

	return nil
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

	accountFunds, err := e.futuClient.GetAccountFunds(ctx, "CN")
	if err != nil {
		log.Printf("Failed to get account funds: %v", err)
		return
	}

	positions, err := e.futuClient.GetPositions(ctx, "CN")
	if err != nil {
		log.Printf("Failed to get positions: %v", err)
		return
	}

	var marketDataLines []string
	marketDataLines = append(marketDataLines, "=== A股市场数据 ===")
	marketDataLines = append(marketDataLines, "")
	marketDataLines = append(marketDataLines, "【账户概况】")
	marketDataLines = append(marketDataLines, fmt.Sprintf("总资产: %.2f", accountFunds.TotalAssets))
	marketDataLines = append(marketDataLines, fmt.Sprintf("可用资金: %.2f", accountFunds.Cash))
	marketDataLines = append(marketDataLines, fmt.Sprintf("持仓市值: %.2f", accountFunds.MarketValue))
	marketDataLines = append(marketDataLines, fmt.Sprintf("仓位比例: %.1f%%", accountFunds.MarketValue/accountFunds.TotalAssets*100))
	marketDataLines = append(marketDataLines, "")
	marketDataLines = append(marketDataLines, "【当前持仓】")
	if len(positions) == 0 {
		marketDataLines = append(marketDataLines, "无持仓")
	} else {
		for _, pos := range positions {
			pnlPct := (pos.CurrentPrice - pos.AvgCost) / pos.AvgCost * 100
			marketDataLines = append(marketDataLines, fmt.Sprintf("- %s.%s: 持有%d股, 成本价%.2f, 现价%.2f, 盈亏%.2f (%.2f%%)", 
				pos.Market, pos.Code, pos.Quantity, pos.AvgCost, pos.CurrentPrice, pos.UnrealizedPnL, pnlPct))
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

	e.saveDecision(worker.AgentID, decision)

	if e.tradingEnabled {
		e.executeTrade(ctx, decision)
	}
}

func (e *Engine) saveDecision(agentID string, decision *llm.TradeDecision) {
	decisionJSON, _ := json.Marshal(decision)

	_, err := e.db.Exec(`
		INSERT INTO trade_decisions (agent_id, stock_code, market, action, quantity, price, reason, llm_response, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, agentID, decision.Code, decision.Market, decision.Action, decision.Quantity, decision.Price, decision.Reason, decisionJSON, time.Now())

	if err != nil {
		log.Printf("Failed to save decision: %v", err)
	}
}

func (e *Engine) executeTrade(ctx context.Context, decision *llm.TradeDecision) {
	log.Printf("Executing trade: %s %s %d @ %.2f", decision.Action, decision.Code, decision.Quantity, decision.Price)

	orderID, err := e.futuClient.PlaceOrder(ctx, decision.Market, decision.Code, decision.Action, decision.Price, decision.Quantity)
	if err != nil {
		log.Printf("Failed to execute trade: %v", err)
		return
	}

	log.Printf("Trade executed successfully, order ID: %s", orderID)

	_, err = e.db.Exec(`
		UPDATE trade_decisions 
		SET executed = TRUE, executed_at = $1 
		WHERE stock_code = $2 AND market = $3 AND action = $4 AND created_at = (
			SELECT MAX(created_at) FROM trade_decisions WHERE stock_code = $2 AND market = $3 AND action = $4
		)
	`, time.Now(), decision.Code, decision.Market, decision.Action)
	if err != nil {
		log.Printf("Failed to update decision status: %v", err)
	}
}

func (e *Engine) StartAgent(id int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	worker, exists := e.agents[id]
	if !exists {
		var agentID, name, description, strategy string
		var enabled bool
		err := e.db.QueryRow(`
			SELECT agent_id, name, description, trading_strategy, enabled
			FROM agent_configs WHERE id = $1
		`, id).Scan(&agentID, &name, &description, &strategy, &enabled)
		if err != nil {
			return fmt.Errorf("agent %d not found: %w", id, err)
		}

		worker = &AgentWorker{
			ID:      id,
			AgentID: agentID,
			Config: AgentConfig{
				Name:            name,
				Description:     description,
				TradingStrategy: strategy,
				Enabled:         true,
			},
		}
		e.agents[id] = worker
	}

	worker.mu.Lock()
	defer worker.mu.Unlock()

	if worker.running {
		return fmt.Errorf("agent %d already running", id)
	}

	_, err := e.db.Exec("UPDATE agent_configs SET enabled = TRUE, updated_at = CURRENT_TIMESTAMP WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to enable agent: %w", err)
	}

	worker.Config.Enabled = true
	worker.running = true

	log.Printf("Agent %s started", worker.AgentID)
	return nil
}

func (e *Engine) StopAgent(id int64) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	worker, exists := e.agents[id]
	if !exists {
		return fmt.Errorf("agent %d not found", id)
	}

	worker.mu.Lock()
	defer worker.mu.Unlock()

	if !worker.running {
		return fmt.Errorf("agent %d not running", id)
	}

	_, err := e.db.Exec("UPDATE agent_configs SET enabled = FALSE, updated_at = CURRENT_TIMESTAMP WHERE id = $1", id)
	if err != nil {
		return fmt.Errorf("failed to disable agent: %w", err)
	}

	worker.Config.Enabled = false
	worker.running = false

	log.Printf("Agent %s stopped", worker.AgentID)
	return nil
}

func (a *AgentWorker) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}
	a.running = false
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
