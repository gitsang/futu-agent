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
	"github.com/gitsang/futu-agent/backend/internal/services/universe"
	"github.com/gitsang/futu-agent/backend/internal/store"
)

type Engine struct {
	store          *store.MemoryStore
	futuClient     *futu.CachedClient
	llmClient      *llm.Client
	config         *config.Config
	universeService *universe.Service
	agents         map[string]*AgentWorker
	mu             sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	tradingEnabled bool
	recentOrders   map[string]time.Time
	recentOrdersMu sync.RWMutex
}

type AgentWorker struct {
	AgentID string
	Config  config.ResolvedAgentConfig
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
	mu      sync.Mutex
}

func NewEngine(store *store.MemoryStore, futuClient *futu.CachedClient, llmClient *llm.Client, cfg *config.Config, universeService *universe.Service) *Engine {
	return &Engine{
		store:          store,
		futuClient:     futuClient,
		llmClient:      llmClient,
		config:         cfg,
		universeService: universeService,
		agents:         make(map[string]*AgentWorker),
		tradingEnabled: cfg.TradingEnabled,
		recentOrders:   make(map[string]time.Time),
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

	for _, agentCfg := range e.config.GetResolvedAgents() {
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

	e.cancelStaleOrders()

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

func (e *Engine) cancelStaleOrders() {
	markets := []string{"HK", "US", "CN"}
	for _, market := range markets {
		orders, err := e.futuClient.GetOrders(e.ctx, market)
		if err != nil {
			log.Printf("Failed to get orders for market %s: %v", market, err)
			continue
		}

		log.Printf("Checking %d orders in market %s for stale orders", len(orders), market)

		for _, order := range orders {
			if order.Status != "SUBMITTED" {
				continue
			}

			createTime, err := time.ParseInLocation("2006-01-02 15:04:05", order.CreateTime, time.Local)
			if err != nil {
				log.Printf("Failed to parse create time for order %s: %v", order.OrderID, err)
				continue
			}

			age := time.Since(createTime)
			log.Printf("Order %s age: %v (threshold: 30m)", order.OrderID, age)

			if age > 30*time.Minute {
				orderID, err := parseOrderID(order.OrderID)
				if err != nil {
					log.Printf("Failed to parse order ID %s: %v", order.OrderID, err)
					continue
				}

				log.Printf("Cancelling stale order: %s %s %s %d @ %.2f (created: %s, age: %v)",
					order.Side, market, order.Code, int(order.Qty), order.Price, order.CreateTime, age)

				if err := e.futuClient.CancelOrder(e.ctx, market, orderID); err != nil {
					log.Printf("Failed to cancel order %s: %v", order.OrderID, err)
				}
			}
		}
	}
}

func parseOrderID(orderIDStr string) (uint64, error) {
	var orderID uint64
	_, err := fmt.Sscanf(orderIDStr, "%d", &orderID)
	return orderID, err
}

func (e *Engine) executeAgent(worker *AgentWorker) {
	ctx, cancel := context.WithTimeout(e.ctx, 5*time.Minute)
	defer cancel()

	market := worker.Config.Market
	marketStatus := futu.GetMarketStatus(market)
	log.Printf("Executing agent %s (市场: %s, 状态: %s)", worker.AgentID, market, marketStatus)

	if !futu.IsMarketOpen(market) {
		log.Printf("市场 %s 休中，跳过交易决策", market)
		return
	}

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

	stockUniverse := worker.Config.StockUniverse
	candidates := []futu.StockScreener{}
	if stockUniverse.Source == "screen" {
		if e.universeService != nil {
			candidates = e.universeService.GetCandidates(stockUniverse.ID)
		}
		if len(candidates) == 0 {
			log.Printf("股票池 %s 无候选股票，使用默认股票池", stockUniverse.ID)
			candidates = e.getDefaultCandidates(market)
		}

		if stockUniverse.ScreenConfig.Limit > 0 && len(candidates) > stockUniverse.ScreenConfig.Limit {
			candidates = candidates[:stockUniverse.ScreenConfig.Limit]
		}
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
	if len(candidates) > 0 {
		marketDataLines = append(marketDataLines, "")
		marketDataLines = append(marketDataLines, "【候选股票（从全市场筛选）】")
		for _, candidate := range candidates {
			marketDataLines = append(marketDataLines, fmt.Sprintf("- %s (%s): 现价%.2f, 涨跌幅%.2f%%, PE=%.1f, PB=%.1f, 市值%.0f亿",
				candidate.Name, candidate.Code, candidate.Price, candidate.ChangePct, candidate.PE, candidate.PB, candidate.MarketValue))
		}
	}

	marketData := strings.Join(marketDataLines, "\n")

	positionsJSON, _ := json.Marshal(positions)
	accountJSON, _ := json.Marshal(accountFunds)
	candidatesJSON, _ := json.Marshal(candidates)

	tradingStrategy := worker.Config.TradingStrategy
	if tradingStrategy == "" {
		tradingStrategy = "基于技术分析的通用交易策略"
	}

	candidatesText := string(candidatesJSON)
	decision, err := e.llmClient.AnalyzeAndDecide(ctx, marketData, string(positionsJSON), string(accountJSON), candidatesText, tradingStrategy, worker.Config.Rules)
	if err != nil {
		log.Printf("Agent %s failed to analyze: %v", worker.AgentID, err)
		return
	}

	log.Printf("Agent %s decision: %s %s %d @ %.2f - %s", worker.AgentID, decision.Action, decision.Code, decision.Quantity, decision.Price, decision.Reason)

	if decision.Action == "HOLD" {
		return
	}

	if e.tradingEnabled {
		if e.isDuplicateOrder(decision.Code, decision.Action) {
			log.Printf("Agent %s 重复下单检测: %s %s 已在最近5分钟内下过单，跳过", worker.AgentID, decision.Action, decision.Code)
			return
		}

		decision.Price = e.optimizeOrderPrice(ctx, decision, market)
		log.Printf("Agent %s 优化后价格: %.2f", worker.AgentID, decision.Price)

		if err := e.validateOrder(decision, positions, accountFunds); err != nil {
			log.Printf("Agent %s 订单校验失败: %v", worker.AgentID, err)
			return
		}

		orderID, err := e.futuClient.PlaceOrder(ctx, decision.Market, decision.Code, decision.Action, decision.Price, decision.Quantity)
		if err != nil {
			log.Printf("Failed to execute trade: %v", err)
			return
		}
		log.Printf("Trade executed successfully, order ID: %s", orderID)
		e.recordOrder(decision.Code, decision.Action)
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

func (e *Engine) validateOrder(decision *llm.TradeDecision, positions []futu.Position, funds *futu.AccountFunds) error {
	if decision.Action == "SELL" {
		for _, pos := range positions {
			if pos.Code == decision.Code {
				if pos.Quantity < decision.Quantity {
					return fmt.Errorf("持仓不足: 持有%d股, 卖出%d股", pos.Quantity, decision.Quantity)
				}
				return nil
			}
		}
		return fmt.Errorf("没有持仓: 未持有%s", decision.Code)
	}

	if decision.Action == "BUY" {
		requiredFunds := decision.Price * float64(decision.Quantity)
		if funds.Cash < requiredFunds {
			return fmt.Errorf("资金不足: 需要%.2f, 可用%.2f", requiredFunds, funds.Cash)
		}
	}

	return nil
}

func (e *Engine) optimizeOrderPrice(ctx context.Context, decision *llm.TradeDecision, market string) float64 {
	quote, err := e.futuClient.GetQuote(ctx, market, decision.Code)
	if err != nil {
		log.Printf("无法获取行情，使用原始价格: %v", err)
		return decision.Price
	}

	currentPrice := quote.Price
	if currentPrice <= 0 {
		log.Printf("当前价格无效 (%.2f)，使用原始价格", currentPrice)
		return decision.Price
	}

	var optimizedPrice float64
	if decision.Action == "BUY" {
		optimizedPrice = currentPrice * 1.005
	} else {
		optimizedPrice = currentPrice * 0.995
	}

	optimizedPrice = float64(int(optimizedPrice*100)) / 100

	log.Printf("价格优化: LLM价格=%.2f, 当前价=%.2f, 优化后=%.2f", decision.Price, currentPrice, optimizedPrice)
	return optimizedPrice
}

func (e *Engine) isDuplicateOrder(code, action string) bool {
	key := fmt.Sprintf("%s:%s", code, action)

	e.recentOrdersMu.RLock()
	lastOrderTime, exists := e.recentOrders[key]
	e.recentOrdersMu.RUnlock()

	if !exists {
		return false
	}

	return time.Since(lastOrderTime) < 5*time.Minute
}

func (e *Engine) recordOrder(code, action string) {
	key := fmt.Sprintf("%s:%s", code, action)

	e.recentOrdersMu.Lock()
	defer e.recentOrdersMu.Unlock()

	e.recentOrders[key] = time.Now()

	for k, t := range e.recentOrders {
		if time.Since(t) > 10*time.Minute {
			delete(e.recentOrders, k)
		}
	}
}

func (e *Engine) getDefaultCandidates(market string) []futu.StockScreener {
	defaultStocks := map[string][]futu.StockScreener{
		"US": {
			{Code: "AAPL", Name: "Apple"},
			{Code: "MSFT", Name: "Microsoft"},
			{Code: "GOOGL", Name: "Alphabet"},
			{Code: "AMZN", Name: "Amazon"},
			{Code: "NVDA", Name: "NVIDIA"},
			{Code: "META", Name: "Meta"},
			{Code: "TSLA", Name: "Tesla"},
			{Code: "JPM", Name: "JPMorgan"},
			{Code: "V", Name: "Visa"},
			{Code: "JNJ", Name: "Johnson & Johnson"},
		},
		"HK": {
			{Code: "00700", Name: "腾讯控股"},
			{Code: "09988", Name: "阿里巴巴-SW"},
			{Code: "03690", Name: "美团-W"},
			{Code: "09999", Name: "网易-S"},
			{Code: "01810", Name: "小米集团-W"},
			{Code: "00005", Name: "汇丰控股"},
			{Code: "00941", Name: "中国移动"},
			{Code: "02318", Name: "中国平安"},
			{Code: "00388", Name: "香港交易所"},
			{Code: "01299", Name: "友邦保险"},
		},
		"CN": {
			{Code: "600519", Name: "贵州茅台"},
			{Code: "601318", Name: "中国平安"},
			{Code: "600036", Name: "招商银行"},
			{Code: "000858", Name: "五粮液"},
			{Code: "000333", Name: "美的集团"},
			{Code: "601166", Name: "兴业银行"},
			{Code: "600276", Name: "恒瑞医药"},
			{Code: "000002", Name: "万科A"},
			{Code: "600030", Name: "中信证券"},
			{Code: "002415", Name: "海康威视"},
		},
	}

	if stocks, ok := defaultStocks[market]; ok {
		log.Printf("使用默认股票池: %d 只股票", len(stocks))
		return stocks
	}

	log.Printf("未找到市场 %s 的默认股票池", market)
	return []futu.StockScreener{}
}

func (a *AgentWorker) Stop() {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.cancel != nil {
		a.cancel()
	}
	a.running = false
}
