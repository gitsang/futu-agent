package futu

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	futuapi "github.com/shing1211/futuapi4go/pkg/futuapi"
	"github.com/shing1211/futuapi4go/client"
	"github.com/shing1211/futuapi4go/pkg/constant"
)

var marketMap = map[string]constant.TrdMarket{
	"HK": constant.TrdMarket_HK,
	"US": constant.TrdMarket_US,
	"CN": constant.TrdMarket_CN,
}

var sideMap = map[string]constant.TrdSide{
	"BUY":  constant.TrdSide_Buy,
	"SELL": constant.TrdSide_Sell,
}

type Client struct {
	sdkClient *client.Client
	accID     uint64
	accMap    map[string]uint64
	connected bool
	mu        sync.RWMutex
}

func NewClient(host string, port int) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Connecting to Futu OpenD at %s", addr)

	sdkClient, err := futuapi.NewClient(addr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Futu OpenD: %w", err)
	}

	c := &Client{
		sdkClient: sdkClient,
		accMap:    make(map[string]uint64),
		connected: true,
	}

	if err := c.discoverAccounts(); err != nil {
		sdkClient.Close()
		return nil, fmt.Errorf("failed to discover accounts: %w", err)
	}

	log.Printf("Connected to Futu OpenD, account ID: %d", c.accID)
	return c, nil
}

func (c *Client) discoverAccounts() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	accounts, err := client.GetAccountList(ctx, c.sdkClient)
	if err != nil {
		return fmt.Errorf("GetAccountList failed: %w", err)
	}

	if len(accounts) == 0 {
		return fmt.Errorf("no trading accounts found")
	}

	c.accID = accounts[0].AccID
	log.Printf("Found %d accounts, using accID: %d", len(accounts), c.accID)

	for _, acc := range accounts {
		log.Printf("  Account %d: markets=%v, env=%d", acc.AccID, acc.TrdMarketAuthList, acc.TrdEnv)
		
		for _, market := range acc.TrdMarketAuthList {
			switch constant.TrdMarket(market) {
			case constant.TrdMarket_HK:
				c.accMap["HK"] = acc.AccID
			case constant.TrdMarket_US:
				c.accMap["US"] = acc.AccID
			case constant.TrdMarket_CN:
				c.accMap["CN"] = acc.AccID
			}
		}
	}

	log.Printf("Account mapping: %v", c.accMap)
	return nil
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.sdkClient != nil {
		c.sdkClient.Close()
		c.connected = false
	}
}

func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

type AccountFunds struct {
	Market      string  `json:"market"`
	Currency    string  `json:"currency"`
	TotalAssets float64 `json:"total_assets"`
	Cash        float64 `json:"cash"`
	MarketValue float64 `json:"market_value"`
}

type Position struct {
	Code          string  `json:"code"`
	Market        string  `json:"market"`
	Name          string  `json:"name"`
	Quantity      int     `json:"quantity"`
	AvgCost       float64 `json:"avg_cost"`
	CurrentPrice  float64 `json:"current_price"`
	UnrealizedPnL float64 `json:"unrealized_pnl"`
}

func (c *Client) GetAccountFunds(ctx context.Context, market string) (*AccountFunds, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Futu OpenD")
	}

	accID := c.accID
	if id, ok := c.accMap[market]; ok {
		accID = id
	}

	funds, err := client.GetFunds(ctx, c.sdkClient, accID)
	if err != nil {
		return nil, fmt.Errorf("GetFunds failed: %w", err)
	}

	currency := "CNY"
	switch market {
	case "HK":
		currency = "HKD"
	case "US":
		currency = "USD"
	}

	return &AccountFunds{
		Market:      market,
		Currency:    currency,
		TotalAssets: funds.TotalAssets,
		Cash:        funds.Cash,
		MarketValue: funds.MarketVal,
	}, nil
}

func (c *Client) GetAllAccountFunds(ctx context.Context) ([]AccountFunds, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Futu OpenD")
	}

	var result []AccountFunds
	for _, market := range []string{"CN", "HK", "US"} {
		accID := c.accID
		if id, ok := c.accMap[market]; ok {
			accID = id
		}

		funds, err := client.GetFunds(ctx, c.sdkClient, accID)
		if err != nil {
			log.Printf("Failed to get funds for %s (accID: %d): %v", market, accID, err)
			continue
		}

		currency := "CNY"
		switch market {
		case "HK":
			currency = "HKD"
		case "US":
			currency = "USD"
		}

		result = append(result, AccountFunds{
			Market:      market,
			Currency:    currency,
			TotalAssets: funds.TotalAssets,
			Cash:        funds.Cash,
			MarketValue: funds.MarketVal,
		})
	}

	return result, nil
}

func (c *Client) GetPositions(ctx context.Context, market string) ([]Position, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Futu OpenD")
	}

	var result []Position
	
	if market == "" || market == "ALL" {
		for _, accID := range c.accMap {
			positions, err := client.GetPositionList(ctx, c.sdkClient, accID)
			if err != nil {
				log.Printf("Failed to get positions for accID %d: %v", accID, err)
				continue
			}

			for _, pos := range positions {
				marketStr := marketFromCode(pos.Code)
				result = append(result, Position{
					Code:          pos.Code,
					Market:        marketStr,
					Name:          pos.Name,
					Quantity:      int(pos.Quantity),
					AvgCost:       pos.CostPrice,
					CurrentPrice:  pos.CurPrice,
					UnrealizedPnL: pos.UnrealizedPL,
				})
			}
		}
	} else {
		accID := c.accID
		if id, ok := c.accMap[market]; ok {
			accID = id
		}

		positions, err := client.GetPositionList(ctx, c.sdkClient, accID)
		if err != nil {
			return nil, fmt.Errorf("GetPositionList failed: %w", err)
		}

		for _, pos := range positions {
			marketStr := marketFromCode(pos.Code)
			result = append(result, Position{
				Code:          pos.Code,
				Market:        marketStr,
				Name:          pos.Name,
				Quantity:      int(pos.Quantity),
				AvgCost:       pos.CostPrice,
				CurrentPrice:  pos.CurPrice,
				UnrealizedPnL: pos.UnrealizedPL,
			})
		}
	}

	return result, nil
}

func (c *Client) PlaceOrder(ctx context.Context, market, code, side string, price float64, quantity int) (string, error) {
	if !c.IsConnected() {
		return "", fmt.Errorf("not connected to Futu OpenD")
	}

	accID := c.accID
	if id, ok := c.accMap[market]; ok {
		accID = id
	}

	trdMarket, ok := marketMap[market]
	if !ok {
		return "", fmt.Errorf("unsupported market: %s", market)
	}

	trdSide, ok := sideMap[side]
	if !ok {
		return "", fmt.Errorf("unsupported side: %s", side)
	}

	result, err := client.PlaceOrder(
		ctx,
		c.sdkClient,
		accID,
		trdMarket,
		code,
		trdSide,
		constant.OrderType_Normal,
		price,
		float64(quantity),
		0,
	)
	if err != nil {
		return "", fmt.Errorf("PlaceOrder failed: %w", err)
	}

	orderID := fmt.Sprintf("%d", result.OrderID)
	log.Printf("Order placed: %s %s %s %d @ %.2f (OrderID: %s)", side, market, code, quantity, price, orderID)
	return orderID, nil
}

// marketFromCode determines market from stock code pattern:
// CN: 6 digits starting with 6/0/3
// HK: 5 digits starting with 0
// US: 1-5 uppercase letters
func marketFromCode(code string) string {
	if len(code) == 0 {
		return "UNKNOWN"
	}
	
	if len(code) == 6 && (code[0] == '6' || code[0] == '0' || code[0] == '3') {
		return "CN"
	}
	
	if len(code) == 5 && code[0] == '0' {
		return "HK"
	}
	
	if len(code) >= 1 && len(code) <= 5 {
		allUpper := true
		for _, ch := range code {
			if ch < 'A' || ch > 'Z' {
				allUpper = false
				break
			}
		}
		if allUpper {
			return "US"
		}
	}
	
	return "UNKNOWN"
}

type Quote struct {
	Code       string
	Market     string
	Price      float64
	High       float64
	Low        float64
	Open       float64
	Close      float64
	Volume     int64
	Turnover   float64
	UpdateTime time.Time
}

type OrderResult struct {
	OrderID   string
	OrderIDEx string
}
