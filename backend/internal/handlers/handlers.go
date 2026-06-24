package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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
	futuClient  *futu.Client
}

func NewHandler(store *store.MemoryStore, cfg *config.Config, agentEngine *agent.Engine, futuClient *futu.Client) *Handler {
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

func (h *Handler) GetDecisions(w http.ResponseWriter, r *http.Request) {
	limitStr := r.URL.Query().Get("limit")
	limit := 50
	if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
		limit = l
	}

	decisions := h.store.GetDecisions(limit)
	if decisions == nil {
		decisions = []store.TradeDecision{}
	}

	respondJSON(w, http.StatusOK, decisions)
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
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
