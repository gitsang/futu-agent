package store

import (
	"sync"
	"time"
)

type TradeDecision struct {
	ID        int64     `json:"id"`
	AgentID   string    `json:"agent_id"`
	StockCode string    `json:"stock_code"`
	Market    string    `json:"market"`
	Action    string    `json:"action"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	Reason    string    `json:"reason"`
	Executed  bool      `json:"executed"`
	CreatedAt time.Time `json:"created_at"`
}

type MemoryStore struct {
	mu         sync.RWMutex
	decisions  []TradeDecision
	nextID     int64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		decisions: make([]TradeDecision, 0),
		nextID:    1,
	}
}

func (s *MemoryStore) SaveDecision(decision TradeDecision) TradeDecision {
	s.mu.Lock()
	defer s.mu.Unlock()

	decision.ID = s.nextID
	decision.CreatedAt = time.Now()
	s.nextID++

	s.decisions = append(s.decisions, decision)
	return decision
}

func (s *MemoryStore) GetDecisions(limit int) []TradeDecision {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.decisions) {
		limit = len(s.decisions)
	}

	start := len(s.decisions) - limit
	if start < 0 {
		start = 0
	}

	result := make([]TradeDecision, limit)
	copy(result, s.decisions[start:])

	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}

	return result
}

func (s *MemoryStore) GetDecision(id int64) *TradeDecision {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for i := range s.decisions {
		if s.decisions[i].ID == id {
			return &s.decisions[i]
		}
	}
	return nil
}

func (s *MemoryStore) MarkExecuted(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.decisions {
		if s.decisions[i].ID == id {
			s.decisions[i].Executed = true
			return true
		}
	}
	return false
}
