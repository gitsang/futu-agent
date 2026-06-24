package models

import (
	"database/sql"
	"encoding/json"
	"time"
)

type TradeDecision struct {
	ID          int64           `json:"id"`
	AgentID     string          `json:"agent_id"`
	StockCode   string          `json:"stock_code"`
	Market      string          `json:"market"`
	Action      string          `json:"action"`
	Quantity    int             `json:"quantity"`
	Price       float64         `json:"price"`
	Reason      string          `json:"reason"`
	LLMResponse json.RawMessage `json:"llm_response"`
	Executed    bool            `json:"executed"`
	ExecutedAt  *time.Time      `json:"executed_at"`
	CreatedAt   time.Time       `json:"created_at"`
}

type Position struct {
	ID            int64     `json:"id"`
	StockCode     string    `json:"stock_code"`
	Market        string    `json:"market"`
	Quantity      int       `json:"quantity"`
	AvgCost       float64   `json:"avg_cost"`
	CurrentPrice  float64   `json:"current_price"`
	UnrealizedPnL float64   `json:"unrealized_pnl"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type AccountFunds struct {
	ID           int64     `json:"id"`
	Market       string    `json:"market"`
	Currency     string    `json:"currency"`
	TotalAssets  float64   `json:"total_assets"`
	Cash         float64   `json:"cash"`
	MarketValue  float64   `json:"market_value"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AgentConfig struct {
	ID              int64           `json:"id"`
	AgentID         string          `json:"agent_id"`
	Market          string          `json:"market"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	LLMModel        string          `json:"llm_model"`
	LLMEndpoint     string          `json:"llm_endpoint"`
	TradingStrategy string          `json:"trading_strategy"`
	RiskParameters  json.RawMessage `json:"risk_parameters"`
	Enabled         bool            `json:"enabled"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type SystemConfig struct {
	Key         string    `json:"key"`
	Value       string    `json:"value"`
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AccountFundsResponse struct {
	TotalAssets float64 `json:"total_assets"`
	Cash        float64 `json:"cash"`
	MarketValue float64 `json:"market_value"`
}

type PositionResponse struct {
	StockCode    string  `json:"stock_code"`
	Market       string  `json:"market"`
	Quantity     int     `json:"quantity"`
	AvgCost      float64 `json:"avg_cost"`
	CurrentPrice float64 `json:"current_price"`
	UnrealizedPnL float64 `json:"unrealized_pnl"`
}

type DecisionResponse struct {
	ID        int64           `json:"id"`
	AgentID   string          `json:"agent_id"`
	StockCode string          `json:"stock_code"`
	Market    string          `json:"market"`
	Action    string          `json:"action"`
	Quantity  int             `json:"quantity"`
	Price     float64         `json:"price"`
	Reason    string          `json:"reason"`
	Executed  bool            `json:"executed"`
	CreatedAt time.Time       `json:"created_at"`
}

type AgentConfigResponse struct {
	ID              int64  `json:"id"`
	AgentID         string `json:"agent_id"`
	Market          string `json:"market"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	LLMModel        string `json:"llm_model"`
	LLMEndpoint     string `json:"llm_endpoint"`
	TradingStrategy string `json:"trading_strategy"`
	Enabled         bool   `json:"enabled"`
}

type CreateAgentRequest struct {
	AgentID         string          `json:"agent_id"`
	Market          string          `json:"market"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	LLMModel        string          `json:"llm_model"`
	LLMEndpoint     string          `json:"llm_endpoint"`
	TradingStrategy string          `json:"trading_strategy"`
	RiskParameters  json.RawMessage `json:"risk_parameters"`
}

type UpdateAgentRequest struct {
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	LLMModel        string          `json:"llm_model"`
	LLMEndpoint     string          `json:"llm_endpoint"`
	TradingStrategy string          `json:"trading_strategy"`
	RiskParameters  json.RawMessage `json:"risk_parameters"`
	Enabled         *bool           `json:"enabled"`
}

type UpdateConfigRequest struct {
	Value string `json:"value"`
}

type StatusResponse struct {
	ServerStatus    string `json:"server_status"`
	DatabaseStatus  string `json:"database_status"`
	FutuOpendStatus string `json:"futu_opend_status"`
	TradingEnabled  bool   `json:"trading_enabled"`
	ActiveAgents    int    `json:"active_agents"`
}

func NullString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func NullFloat64(nf sql.NullFloat64) float64 {
	if nf.Valid {
		return nf.Float64
	}
	return 0
}

func NullInt64(ni sql.NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}
