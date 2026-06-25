package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/gitsang/futu-agent/backend/internal/config"
	"github.com/gitsang/futu-agent/backend/internal/services/agent"
	"github.com/gitsang/futu-agent/backend/internal/services/futu"
	"github.com/gitsang/futu-agent/backend/internal/store"
)

type Handler struct {
	store       *store.MemoryStore
	cfg         *config.Config
	agentEngine *agent.Engine
	futuClient  *futu.CachedClient
}

func NewHandler(store *store.MemoryStore, cfg *config.Config, agentEngine *agent.Engine, futuClient *futu.CachedClient) *Handler {
	return &Handler{
		store:       store,
		cfg:         cfg,
		agentEngine: agentEngine,
		futuClient:  futuClient,
	}
}

func (h *Handler) GetAccountFunds(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")
	if market == "" {
		market = "ALL"
	}

	if !validateMarket(market) {
		respondError(w, http.StatusBadRequest, "Invalid market parameter")
		return
	}

	funds, err := h.futuClient.GetAccountFunds(r.Context(), market)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch account funds")
		return
	}

	respondJSON(w, http.StatusOK, funds)
}

func (h *Handler) GetAllAccountFunds(w http.ResponseWriter, r *http.Request) {
	funds, err := h.futuClient.GetAllAccountFunds(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch account funds")
		return
	}

	respondJSON(w, http.StatusOK, funds)
}

func (h *Handler) GetPositions(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")

	positions, err := h.futuClient.GetPositions(r.Context(), market)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch positions")
		return
	}

	if positions == nil {
		positions = []futu.Position{}
	}

	respondJSON(w, http.StatusOK, positions)
}

func (h *Handler) GetOrders(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")

	orders, err := h.futuClient.GetOrders(r.Context(), market)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch orders")
		return
	}

	if orders == nil {
		orders = []futu.Order{}
	}

	respondJSON(w, http.StatusOK, orders)
}

func (h *Handler) GetDecisions(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")
	
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	pageSize := 20

	if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
		page = p
	}
	if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
		pageSize = ps
	}

	limitStr := r.URL.Query().Get("limit")
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		pageSize = l
		page = 1
	}

	result := h.store.GetDecisionsPaginated(page, pageSize, market)
	respondJSON(w, http.StatusOK, result)
}

func (h *Handler) GetDecision(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid decision ID")
		return
	}

	decision := h.store.GetDecision(id)
	if decision == nil {
		respondError(w, http.StatusNotFound, "Decision not found")
		return
	}

	respondJSON(w, http.StatusOK, decision)
}

func (h *Handler) GetAgents(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")

	agents := h.cfg.GetAgentsByMarket(market)
	if agents == nil {
		agents = []config.AgentConfig{}
	}

	respondJSON(w, http.StatusOK, agents)
}

func (h *Handler) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")

	var req struct {
		Enabled *bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Enabled == nil {
		respondError(w, http.StatusBadRequest, "Enabled field is required")
		return
	}

	if !h.cfg.UpdateAgent(idStr, *req.Enabled) {
		respondError(w, http.StatusNotFound, "Agent not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Agent updated successfully"})
}

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"trading_enabled": h.cfg.TradingEnabled,
		"llm_model":       h.cfg.LLMModel,
		"futu_opend_host": h.cfg.FutuOpendHost,
	})
}

func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	respondError(w, http.StatusNotImplemented, "Config update not implemented")
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"server_status":    "running",
		"futu_opend_status": h.agentEngine.GetFutuOpendStatus(),
		"trading_enabled":  h.cfg.TradingEnabled,
		"active_agents":    h.countEnabledAgents(),
	})
}

func (h *Handler) GetTradingStats(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")
	if market == "" {
		market = "ALL"
	}

	if !validateMarket(market) {
		respondError(w, http.StatusBadRequest, "Invalid market parameter")
		return
	}

	stats, err := h.futuClient.GetTradingStats(r.Context(), market)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch trading stats")
		return
	}

	respondJSON(w, http.StatusOK, stats)
}

func (h *Handler) GetMarketOverview(w http.ResponseWriter, r *http.Request) {
	overview, err := h.futuClient.GetMarketOverview(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch market overview")
		return
	}

	respondJSON(w, http.StatusOK, overview)
}

func (h *Handler) GetTradeHistory(w http.ResponseWriter, r *http.Request) {
	market := r.URL.Query().Get("market")
	daysStr := r.URL.Query().Get("days")

	days := 30
	if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
		days = d
	}

	history, err := h.futuClient.GetTradeHistory(r.Context(), market, days)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch trade history")
		return
	}

	respondJSON(w, http.StatusOK, history)
}

func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"services": map[string]string{
			"futu_opend": h.agentEngine.GetFutuOpendStatus(),
			"server":     "running",
		},
	}
	respondJSON(w, http.StatusOK, health)
}

func (h *Handler) GetAccountInfo(w http.ResponseWriter, r *http.Request) {
	funds, err := h.futuClient.GetAllAccountFunds(r.Context())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch account info")
		return
	}

	positions, err := h.futuClient.GetPositions(r.Context(), "ALL")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch positions")
		return
	}

	totalAssets := 0.0
	totalCash := 0.0
	totalMarketValue := 0.0
	for _, f := range funds {
		totalAssets += f.TotalAssets
		totalCash += f.Cash
		totalMarketValue += f.MarketValue
	}

	info := map[string]interface{}{
		"total_assets":    totalAssets,
		"total_cash":      totalCash,
		"total_market_value": totalMarketValue,
		"position_count":  len(positions),
		"account_count":   len(funds),
		"markets":         []string{"CN", "HK", "US"},
	}
	respondJSON(w, http.StatusOK, info)
}

func (h *Handler) countEnabledAgents() int {
	count := 0
	for _, agent := range h.cfg.Agents {
		if agent.Enabled {
			count++
		}
	}
	return count
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func validateMarket(market string) bool {
	validMarkets := map[string]bool{"HK": true, "US": true, "CN": true, "ALL": true, "": true}
	return validMarkets[market]
}
