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

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	Total      int         `json:"total"`
	TotalPages int         `json:"total_pages"`
}

const maxDecisions = 10000

type MemoryStore struct {
	mu         sync.RWMutex
	decisions  []TradeDecision
	idIndex    map[int64]int
	nextID     int64
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		decisions: make([]TradeDecision, 0),
		idIndex:   make(map[int64]int),
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
	s.idIndex[decision.ID] = len(s.decisions) - 1
	
	if len(s.decisions) > maxDecisions {
		s.rebuildIndex()
	}
	
	return decision
}

func (s *MemoryStore) rebuildIndex() {
	s.decisions = s.decisions[len(s.decisions)-maxDecisions:]
	s.idIndex = make(map[int64]int, len(s.decisions))
	for i, d := range s.decisions {
		s.idIndex[d.ID] = i
	}
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

func (s *MemoryStore) GetDecisionsPaginated(page, pageSize int) PaginatedResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	total := len(s.decisions)
	totalPages := (total + pageSize - 1) / pageSize
	if totalPages == 0 {
		totalPages = 1
	}

	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	result := make([]TradeDecision, 0, end-start)
	for i := total - 1 - start; i >= total-end; i-- {
		if i >= 0 && i < total {
			result = append(result, s.decisions[i])
		}
	}

	return PaginatedResponse{
		Data:       result,
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

func (s *MemoryStore) GetDecision(id int64) *TradeDecision {
	s.mu.RLock()
	defer s.mu.RUnlock()

	idx, exists := s.idIndex[id]
	if !exists {
		return nil
	}

	if idx < 0 || idx >= len(s.decisions) {
		return nil
	}

	return &s.decisions[idx]
}

func (s *MemoryStore) MarkExecuted(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	idx, exists := s.idIndex[id]
	if !exists {
		return false
	}

	if idx < 0 || idx >= len(s.decisions) {
		return false
	}

	s.decisions[idx].Executed = true
	return true
}
