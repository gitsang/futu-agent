package universe

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gitsang/futu-agent/backend/internal/config"
	"github.com/gitsang/futu-agent/backend/internal/services/futu"
)

type Service struct {
	futuClient *futu.CachedClient
	universes  map[string]*UniverseState
	mu         sync.RWMutex
	ctx        context.Context
	cancel     context.CancelFunc
}

type UniverseState struct {
	Config     config.StockUniverseTemplate
	Candidates []futu.StockScreener
	LastUpdate time.Time
	mu         sync.RWMutex
}

func NewService(futuClient *futu.CachedClient) *Service {
	return &Service{
		futuClient: futuClient,
		universes:  make(map[string]*UniverseState),
	}
}

func (s *Service) Start(ctx context.Context, universes []config.StockUniverseTemplate) error {
	s.mu.Lock()
	if s.cancel != nil {
		s.cancel()
	}
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.universes = make(map[string]*UniverseState, len(universes))
	for _, universe := range universes {
		if universe.ID == "" {
			s.mu.Unlock()
			return fmt.Errorf("stock universe ID cannot be empty")
		}
		s.universes[universe.ID] = &UniverseState{Config: universe}
	}
	s.mu.Unlock()

	for _, universe := range universes {
		go s.runUniverse(universe.ID)
	}

	log.Printf("Universe service started with %d universes", len(universes))
	return nil
}

func (s *Service) Stop() {
	s.mu.RLock()
	cancel := s.cancel
	s.mu.RUnlock()

	if cancel != nil {
		cancel()
	}
	log.Println("Universe service stopped")
}

func (s *Service) GetCandidates(universeID string) []futu.StockScreener {
	state := s.getUniverse(universeID)
	if state == nil {
		return nil
	}

	state.mu.RLock()
	defer state.mu.RUnlock()

	candidates := make([]futu.StockScreener, len(state.Candidates))
	copy(candidates, state.Candidates)
	return candidates
}

func (s *Service) RefreshNow(universeID string) error {
	state := s.getUniverse(universeID)
	if state == nil {
		return fmt.Errorf("unknown stock universe %s", universeID)
	}
	return s.refresh(state)
}

func (s *Service) runUniverse(universeID string) {
	state := s.getUniverse(universeID)
	if state == nil {
		return
	}

	interval := scheduleDuration(state.Config.Schedule)
	s.refresh(state)
	log.Printf("Universe %s refreshed: %d candidates (next refresh in %v)", universeID, len(s.GetCandidates(universeID)), interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			if err := s.refresh(state); err != nil {
				log.Printf("Universe %s refresh failed: %v", universeID, err)
			}
			log.Printf("Universe %s refreshed: %d candidates (next refresh in %v)", universeID, len(s.GetCandidates(universeID)), interval)
		}
	}
}

func (s *Service) refresh(state *UniverseState) error {
	state.mu.RLock()
	universe := state.Config
	state.mu.RUnlock()

	if universe.Source != "screen" {
		return nil
	}

	market := universe.ScreenConfig.Market
	if market == "" {
		market = universe.Market
	}

	minPrice, maxPrice, minVolume := screenFilters(universe.ScreenConfig.Filters)
	candidates, err := s.futuClient.ScreenStocks(s.ctx, market, minPrice, maxPrice, minVolume)
	if err != nil {
		return err
	}
	if universe.ScreenConfig.Limit > 0 && len(candidates) > universe.ScreenConfig.Limit {
		candidates = candidates[:universe.ScreenConfig.Limit]
	}

	state.mu.Lock()
	state.Candidates = append([]futu.StockScreener(nil), candidates...)
	state.LastUpdate = time.Now()
	state.mu.Unlock()

	return nil
}

func (s *Service) getUniverse(universeID string) *UniverseState {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.universes[universeID]
}

func scheduleDuration(schedule string) time.Duration {
	if schedule == "" {
		return 24 * time.Hour
	}
	interval, err := time.ParseDuration(schedule)
	if err != nil || interval <= 0 {
		return 24 * time.Hour
	}
	return interval
}

func screenFilters(filters []config.StockUniverseFilterConfig) (float64, float64, int64) {
	var minPrice, maxPrice float64
	var minVolume int64
	for _, filter := range filters {
		switch filter.Field {
		case "price":
			switch filter.Operator {
			case ">", ">=":
				minPrice = filter.Value
			case "<", "<=":
				maxPrice = filter.Value
			}
		case "volume":
			switch filter.Operator {
			case ">", ">=":
				minVolume = int64(filter.Value)
			}
		}
	}
	return minPrice, maxPrice, minVolume
}
