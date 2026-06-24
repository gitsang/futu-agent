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
	accMap    map[string][]uint64
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
		accMap:    make(map[string][]uint64),
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
				c.accMap["HK"] = append(c.accMap["HK"], acc.AccID)
			case constant.TrdMarket_US:
				c.accMap["US"] = append(c.accMap["US"], acc.AccID)
			case constant.TrdMarket_CN:
				c.accMap["CN"] = append(c.accMap["CN"], acc.AccID)
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

	accIDs := c.accMap[market]
	if len(accIDs) == 0 {
		accIDs = []uint64{c.accID}
	}

	var totalAssets, totalCash, totalMarketValue float64
	for _, accID := range accIDs {
		funds, err := client.GetFunds(ctx, c.sdkClient, accID)
		if err != nil {
			log.Printf("Failed to get funds for accID %d: %v", accID, err)
			continue
		}
		totalAssets += funds.TotalAssets
		totalCash += funds.Cash
		totalMarketValue += funds.MarketVal
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
		TotalAssets: totalAssets,
		Cash:        totalCash,
		MarketValue: totalMarketValue,
	}, nil
}

func (c *Client) GetAllAccountFunds(ctx context.Context) ([]AccountFunds, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Futu OpenD")
	}

	var result []AccountFunds
	for _, market := range []string{"CN", "HK", "US"} {
		accIDs := c.accMap[market]
		if len(accIDs) == 0 {
			continue
		}

		var totalAssets, totalCash, totalMarketValue float64
		for _, accID := range accIDs {
			funds, err := client.GetFunds(ctx, c.sdkClient, accID)
			if err != nil {
				log.Printf("Failed to get funds for %s (accID: %d): %v", market, accID, err)
				continue
			}
			totalAssets += funds.TotalAssets
			totalCash += funds.Cash
			totalMarketValue += funds.MarketVal
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
			TotalAssets: totalAssets,
			Cash:        totalCash,
			MarketValue: totalMarketValue,
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
		seen := make(map[uint64]bool)
		for _, accIDs := range c.accMap {
			for _, accID := range accIDs {
				if seen[accID] {
					continue
				}
				seen[accID] = true
				
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
		}
	} else {
		accIDs := c.accMap[market]
		if len(accIDs) == 0 {
			accIDs = []uint64{c.accID}
		}

		for _, accID := range accIDs {
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
	}

	return result, nil
}

func (c *Client) PlaceOrder(ctx context.Context, market, code, side string, price float64, quantity int) (string, error) {
	if !c.IsConnected() {
		return "", fmt.Errorf("not connected to Futu OpenD")
	}

	accIDs := c.accMap[market]
	if len(accIDs) == 0 {
		accIDs = []uint64{c.accID}
	}
	accID := accIDs[0]

	trdMarket, ok := marketMap[market]
	if !ok {
		return "", fmt.Errorf("unsupported market: %s", market)
	}

	trdSide, ok := sideMap[side]
	if !ok {
		return "", fmt.Errorf("unsupported side: %s", side)
	}

	var secMarket constant.TrdSecMarket
	if market == "CN" {
		if len(code) == 6 {
			if code[0] == '6' {
				secMarket = constant.TrdSecMarket_CN_SH
			} else {
				secMarket = constant.TrdSecMarket_CN_SZ
			}
		}
	} else if market == "HK" {
		secMarket = constant.TrdSecMarket_HK
	} else if market == "US" {
		secMarket = constant.TrdSecMarket_US
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
		secMarket,
	)
	if err != nil {
		return "", fmt.Errorf("PlaceOrder failed: %w", err)
	}

	orderID := fmt.Sprintf("%d", result.OrderID)
	log.Printf("Order placed: %s %s %s %d @ %.2f (OrderID: %s)", side, market, code, quantity, price, orderID)
	return orderID, nil
}

func (c *Client) GetQuote(ctx context.Context, market, code string) (*Quote, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Futu OpenD")
	}

	marketConst, ok := marketMap[market]
	if !ok {
		return nil, fmt.Errorf("unsupported market: %s", market)
	}

	sdkQuote, err := client.GetQuote(ctx, c.sdkClient, constant.Market(marketConst), code)
	if err != nil {
		return nil, fmt.Errorf("GetQuote failed: %w", err)
	}

	changePct := 0.0
	if sdkQuote.LastClose > 0 {
		changePct = (sdkQuote.Price - sdkQuote.LastClose) / sdkQuote.LastClose * 100
	}

	return &Quote{
		Code:         code,
		Market:       market,
		Name:         sdkQuote.Name,
		Price:        sdkQuote.Price,
		Open:         sdkQuote.Open,
		High:         sdkQuote.High,
		Low:          sdkQuote.Low,
		LastClose:    sdkQuote.LastClose,
		Volume:       sdkQuote.Volume,
		Turnover:     sdkQuote.Turnover,
		TurnoverRate: sdkQuote.TurnoverRate,
		Amplitude:    sdkQuote.Amplitude,
		ChangePct:    changePct,
	}, nil
}

func (c *Client) GetOrders(ctx context.Context, market string) ([]Order, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Futu OpenD")
	}

	var result []Order

	if market == "" || market == "ALL" {
		seen := make(map[uint64]bool)
		for _, accIDs := range c.accMap {
			for _, accID := range accIDs {
				if seen[accID] {
					continue
				}
				seen[accID] = true

				orders, err := client.GetOrderList(ctx, c.sdkClient, accID)
				if err != nil {
					log.Printf("Failed to get orders for accID %d: %v", accID, err)
					continue
				}

				for _, o := range orders {
					result = append(result, convertOrder(o))
				}
			}
		}
	} else {
		accIDs := c.accMap[market]
		if len(accIDs) == 0 {
			return result, nil
		}

		for _, accID := range accIDs {
			orders, err := client.GetOrderList(ctx, c.sdkClient, accID)
			if err != nil {
				log.Printf("Failed to get orders for accID %d: %v", accID, err)
				continue
			}

			for _, o := range orders {
				order := convertOrder(o)
				if order.Market == market {
					result = append(result, order)
				}
			}
		}
	}

	return result, nil
}

func convertOrder(o client.Order) Order {
	side := "BUY"
	if o.TrdSide == 2 {
		side = "SELL"
	}

	orderType := "NORMAL"
	switch o.OrderType {
	case 1:
		orderType = "MARKET"
	case 2:
		orderType = "STOP"
	}

	status := "UNKNOWN"
	switch o.OrderStatus {
	case 1:
		status = "PENDING"
	case 2:
		status = "SUBMITTED"
	case 3:
		status = "FILLED"
	case 4:
		status = "CANCELLED"
	case 5:
		status = "FAILED"
	case 6:
		status = "EXPIRED"
	}

	market := marketFromCode(o.Code)

	return Order{
		OrderID:    o.OrderIDEx,
		Code:       o.Code,
		Name:       o.Name,
		Market:     market,
		Side:       side,
		OrderType:  orderType,
		Status:     status,
		Price:      o.Price,
		Qty:        o.Qty,
		FillQty:    o.FillQty,
		FillPrice:  o.FillAvgPrice,
		CreateTime: o.CreateTime,
		UpdateTime: o.UpdateTime,
		Remark:     o.LastErrMsg,
	}
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
	Code         string  `json:"code"`
	Market       string  `json:"market"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Open         float64 `json:"open"`
	High         float64 `json:"high"`
	Low          float64 `json:"low"`
	LastClose    float64 `json:"last_close"`
	Volume       int64   `json:"volume"`
	Turnover     float64 `json:"turnover"`
	TurnoverRate float64 `json:"turnover_rate"`
	Amplitude    float64 `json:"amplitude"`
	ChangePct    float64 `json:"change_pct"`
}

type Order struct {
	OrderID      string  `json:"order_id"`
	Code         string  `json:"code"`
	Name         string  `json:"name"`
	Market       string  `json:"market"`
	Side         string  `json:"side"`
	OrderType    string  `json:"order_type"`
	Status       string  `json:"status"`
	Price        float64 `json:"price"`
	Qty          float64 `json:"qty"`
	FillQty      float64 `json:"fill_qty"`
	FillPrice    float64 `json:"fill_price"`
	CreateTime   string  `json:"create_time"`
	UpdateTime   string  `json:"update_time"`
	Remark       string  `json:"remark"`
}

type OrderResult struct {
	OrderID   string
	OrderIDEx string
}
