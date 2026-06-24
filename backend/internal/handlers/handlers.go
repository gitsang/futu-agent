package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/gitsang/futu-agent/backend/internal/models"
	"github.com/gitsang/futu-agent/backend/internal/services/agent"
	"github.com/gitsang/futu-agent/backend/internal/services/futu"
)

type Handler struct {
	db          *sql.DB
	agentEngine *agent.Engine
	futuClient  *futu.Client
}

func NewHandler(db *sql.DB, agentEngine *agent.Engine, futuClient *futu.Client) *Handler {
	return &Handler{
		db:          db,
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

	rows, err := h.db.Query(`
		SELECT id, agent_id, stock_code, market, action, quantity, price, reason, executed, created_at
		FROM trade_decisions
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch decisions")
		return
	}
	defer rows.Close()

	var decisions []models.DecisionResponse
	for rows.Next() {
		var d models.TradeDecision
		if err := rows.Scan(&d.ID, &d.AgentID, &d.StockCode, &d.Market, &d.Action, &d.Quantity, &d.Price, &d.Reason, &d.Executed, &d.CreatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to scan decision")
			return
		}
		decisions = append(decisions, models.DecisionResponse{
			ID:        d.ID,
			AgentID:   d.AgentID,
			StockCode: d.StockCode,
			Market:    d.Market,
			Action:    d.Action,
			Quantity:  d.Quantity,
			Price:     d.Price,
			Reason:    d.Reason,
			Executed:  d.Executed,
			CreatedAt: d.CreatedAt,
		})
	}

	if decisions == nil {
		decisions = []models.DecisionResponse{}
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

	var d models.TradeDecision
	err = h.db.QueryRow(`
		SELECT id, agent_id, stock_code, market, action, quantity, price, reason, llm_response, executed, executed_at, created_at
		FROM trade_decisions
		WHERE id = $1
	`, id).Scan(&d.ID, &d.AgentID, &d.StockCode, &d.Market, &d.Action, &d.Quantity, &d.Price, &d.Reason, &d.LLMResponse, &d.Executed, &d.ExecutedAt, &d.CreatedAt)

	if err == sql.ErrNoRows {
		respondError(w, http.StatusNotFound, "Decision not found")
		return
	}
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch decision")
		return
	}

	respondJSON(w, http.StatusOK, d)
}

func (h *Handler) GetAgents(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT id, agent_id, name, description, llm_model, llm_endpoint, trading_strategy, enabled, created_at, updated_at
		FROM agent_configs
		ORDER BY created_at DESC
	`)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch agents")
		return
	}
	defer rows.Close()

	var agents []models.AgentConfigResponse
	for rows.Next() {
		var a models.AgentConfig
		if err := rows.Scan(&a.ID, &a.AgentID, &a.Name, &a.Description, &a.LLMModel, &a.LLMEndpoint, &a.TradingStrategy, &a.Enabled, &a.CreatedAt, &a.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to scan agent")
			return
		}
		agents = append(agents, models.AgentConfigResponse{
			ID:              a.ID,
			AgentID:         a.AgentID,
			Name:            a.Name,
			Description:     a.Description,
			LLMModel:        a.LLMModel,
			LLMEndpoint:     a.LLMEndpoint,
			TradingStrategy: a.TradingStrategy,
			Enabled:         a.Enabled,
		})
	}

	if agents == nil {
		agents = []models.AgentConfigResponse{}
	}

	respondJSON(w, http.StatusOK, agents)
}

func (h *Handler) CreateAgent(w http.ResponseWriter, r *http.Request) {
	var req models.CreateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.AgentID == "" {
		respondError(w, http.StatusBadRequest, "Agent ID is required")
		return
	}

	var id int64
	err := h.db.QueryRow(`
		INSERT INTO agent_configs (agent_id, name, description, llm_model, llm_endpoint, trading_strategy, risk_parameters)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, req.AgentID, req.Name, req.Description, req.LLMModel, req.LLMEndpoint, req.TradingStrategy, req.RiskParameters).Scan(&id)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to create agent")
		return
	}

	respondJSON(w, http.StatusCreated, map[string]interface{}{
		"id":      id,
		"agent_id": req.AgentID,
	})
}

func (h *Handler) UpdateAgent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid agent ID")
		return
	}

	var req models.UpdateAgentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	query := `
		UPDATE agent_configs
		SET name = COALESCE(NULLIF($2, ''), name),
		    description = COALESCE(NULLIF($3, ''), description),
		    llm_model = COALESCE(NULLIF($4, ''), llm_model),
		    llm_endpoint = COALESCE(NULLIF($5, ''), llm_endpoint),
		    trading_strategy = COALESCE(NULLIF($6, ''), trading_strategy),
		    risk_parameters = COALESCE($7, risk_parameters),
		    enabled = COALESCE($8, enabled),
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $1
	`

	_, err = h.db.Exec(query, id, req.Name, req.Description, req.LLMModel, req.LLMEndpoint, req.TradingStrategy, req.RiskParameters, req.Enabled)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update agent")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Agent updated successfully"})
}

func (h *Handler) DeleteAgent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid agent ID")
		return
	}

	result, err := h.db.Exec("DELETE FROM agent_configs WHERE id = $1", id)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to delete agent")
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondError(w, http.StatusNotFound, "Agent not found")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Agent deleted successfully"})
}

func (h *Handler) StartAgent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	agentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid agent ID")
		return
	}

	if err := h.agentEngine.StartAgent(agentID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to start agent")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Agent started successfully"})
}

func (h *Handler) StopAgent(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	agentID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid agent ID")
		return
	}

	if err := h.agentEngine.StopAgent(agentID); err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to stop agent")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Agent stopped successfully"})
}

func (h *Handler) GetConfig(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query("SELECT key, value, description, updated_at FROM system_configs")
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to fetch config")
		return
	}
	defer rows.Close()

	configs := make(map[string]models.SystemConfig)
	for rows.Next() {
		var c models.SystemConfig
		if err := rows.Scan(&c.Key, &c.Value, &c.Description, &c.UpdatedAt); err != nil {
			respondError(w, http.StatusInternalServerError, "Failed to scan config")
			return
		}
		configs[c.Key] = c
	}

	respondJSON(w, http.StatusOK, configs)
}

func (h *Handler) UpdateConfig(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Query().Get("key")
	if key == "" {
		respondError(w, http.StatusBadRequest, "Config key is required")
		return
	}

	var req models.UpdateConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	_, err := h.db.Exec(`
		INSERT INTO system_configs (key, value, updated_at)
		VALUES ($1, $2, CURRENT_TIMESTAMP)
		ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = CURRENT_TIMESTAMP
	`, key, req.Value)

	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to update config")
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"message": "Config updated successfully"})
}

func (h *Handler) GetStatus(w http.ResponseWriter, r *http.Request) {
	dbStatus := "connected"
	if err := h.db.Ping(); err != nil {
		dbStatus = "disconnected"
	}

	var activeAgents int
	h.db.QueryRow("SELECT COUNT(*) FROM agent_configs WHERE enabled = TRUE").Scan(&activeAgents)

	respondJSON(w, http.StatusOK, models.StatusResponse{
		ServerStatus:    "running",
		DatabaseStatus:  dbStatus,
		FutuOpendStatus: h.agentEngine.GetFutuOpendStatus(),
		TradingEnabled:  h.agentEngine.IsTradingEnabled(),
		ActiveAgents:    activeAgents,
	})
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
