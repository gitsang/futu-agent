package futu

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

type CacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

type QuoteCache struct {
	entries map[string]CacheEntry
	mu      sync.RWMutex
	ttl     time.Duration
}

func NewQuoteCache(ttl time.Duration) *QuoteCache {
	cache := &QuoteCache{
		entries: make(map[string]CacheEntry),
		ttl:     ttl,
	}

	// Start cleanup goroutine
	go cache.cleanup()

	return cache
}

func (c *QuoteCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(entry.expiresAt) {
		return nil, false
	}

	return entry.data, true
}

func (c *QuoteCache) Set(key string, data interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = CacheEntry{
		data:      data,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *QuoteCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.entries, key)
}

func (c *QuoteCache) cleanup() {
	ticker := time.NewTicker(c.ttl / 2)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if now.After(entry.expiresAt) {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

func (c *QuoteCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

func (c *QuoteCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]CacheEntry)
}

type CachedClient struct {
	*Client
	quoteCache    *QuoteCache
	fundsCache    *QuoteCache
	positionCache *QuoteCache
}

func NewCachedClient(host string, port int) (*CachedClient, error) {
	client, err := NewClient(host, port)
	if err != nil {
		return nil, err
	}

	return &CachedClient{
		Client:        client,
		quoteCache:    NewQuoteCache(30 * time.Second), // Quotes cached for 30s
		fundsCache:    NewQuoteCache(60 * time.Second), // Funds cached for 60s
		positionCache: NewQuoteCache(30 * time.Second), // Positions cached for 30s
	}, nil
}

func (c *CachedClient) GetQuoteWithCache(ctx context.Context, market, code string) (*Quote, error) {
	cacheKey := fmt.Sprintf("quote:%s:%s", market, code)

	// Check cache first
	if cached, ok := c.quoteCache.Get(cacheKey); ok {
		log.Printf("Cache hit for %s", cacheKey)
		return cached.(*Quote), nil
	}

	// Fetch from API
	quote, err := c.GetQuote(ctx, market, code)
	if err != nil {
		return nil, err
	}

	// Store in cache
	c.quoteCache.Set(cacheKey, quote)
	log.Printf("Cached quote for %s", cacheKey)

	return quote, nil
}

func (c *CachedClient) GetAccountFundsWithCache(ctx context.Context, market string) (*AccountFunds, error) {
	cacheKey := fmt.Sprintf("funds:%s", market)

	if cached, ok := c.fundsCache.Get(cacheKey); ok {
		log.Printf("Cache hit for %s", cacheKey)
		return cached.(*AccountFunds), nil
	}

	funds, err := c.GetAccountFunds(ctx, market)
	if err != nil {
		return nil, err
	}

	c.fundsCache.Set(cacheKey, funds)
	return funds, nil
}

func (c *CachedClient) GetPositionsWithCache(ctx context.Context, market string) ([]Position, error) {
	cacheKey := fmt.Sprintf("positions:%s", market)

	if cached, ok := c.positionCache.Get(cacheKey); ok {
		log.Printf("Cache hit for %s", cacheKey)
		return cached.([]Position), nil
	}

	positions, err := c.GetPositions(ctx, market)
	if err != nil {
		return nil, err
	}

	c.positionCache.Set(cacheKey, positions)
	return positions, nil
}

func (c *CachedClient) InvalidateQuotesCache() {
	c.quoteCache.Clear()
}

func (c *CachedClient) InvalidateFundsCache() {
	c.fundsCache.Clear()
}

func (c *CachedClient) InvalidatePositionsCache() {
	c.positionCache.Clear()
}

func (c *CachedClient) InvalidateAllCache() {
	c.quoteCache.Clear()
	c.fundsCache.Clear()
	c.positionCache.Clear()
}

func (c *CachedClient) GetCacheStats() map[string]int {
	return map[string]int{
		"quotes":    c.quoteCache.Size(),
		"funds":     c.fundsCache.Size(),
		"positions": c.positionCache.Size(),
	}
}
