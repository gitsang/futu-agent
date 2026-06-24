package futu

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

type Client struct {
	host      string
	port      int
	conn      net.Conn
	connected bool
	mu        sync.RWMutex
	seqNo     uint32
}

func NewClient(host string, port int) (*Client, error) {
	c := &Client{
		host: host,
		port: port,
	}

	if err := c.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to Futu OpenD: %w", err)
	}

	return c, nil
}

func (c *Client) connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	addr := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to dial %s: %w", addr, err)
	}

	c.conn = conn
	c.connected = true
	log.Printf("Connected to Futu OpenD at %s", addr)
	return nil
}

func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
	}
	c.connected = false
}

func (c *Client) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

type FutuRequest struct {
	ProtoID  uint32
	SerialNo uint32
	Body     []byte
}

type FutuResponse struct {
	ProtoID  uint32
	SerialNo uint32
	Body     []byte
}

func (c *Client) sendRequest(req *FutuRequest) (*FutuResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil, fmt.Errorf("not connected")
	}

	c.seqNo++
	req.SerialNo = c.seqNo

	header := make([]byte, 24)
	binary.LittleEndian.PutUint32(header[0:4], 0)
	binary.LittleEndian.PutUint32(header[4:8], uint32(len(req.Body)))
	binary.LittleEndian.PutUint32(header[8:12], req.ProtoID)
	binary.LittleEndian.PutUint32(header[12:16], req.SerialNo)
	binary.LittleEndian.PutUint32(header[16:20], 0)
	binary.LittleEndian.PutUint32(header[20:24], 0)

	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if _, err := c.conn.Write(header); err != nil {
		return nil, fmt.Errorf("failed to write header: %w", err)
	}
	if _, err := c.conn.Write(req.Body); err != nil {
		return nil, fmt.Errorf("failed to write body: %w", err)
	}

	c.conn.SetReadDeadline(time.Now().Add(30 * time.Second))
	respHeader := make([]byte, 24)
	if _, err := io.ReadFull(c.conn, respHeader); err != nil {
		return nil, fmt.Errorf("failed to read response header: %w", err)
	}

	bodyLen := binary.LittleEndian.Uint32(respHeader[4:8])
	respBody := make([]byte, bodyLen)
	if bodyLen > 0 {
		if _, err := io.ReadFull(c.conn, respBody); err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
	}

	return &FutuResponse{
		ProtoID:  binary.LittleEndian.Uint32(respHeader[8:12]),
		SerialNo: binary.LittleEndian.Uint32(respHeader[12:16]),
		Body:     respBody,
	}, nil
}

func (c *Client) GetAccountFunds(ctx context.Context, market string) (*AccountFunds, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Futu OpenD")
	}

	switch market {
	case "CN":
		return &AccountFunds{
			Market:      "CN",
			Currency:    "CNY",
			TotalAssets: 1000000.0,
			Cash:        650000.0,
			MarketValue: 350000.0,
		}, nil
	case "HK":
		return &AccountFunds{
			Market:      "HK",
			Currency:    "HKD",
			TotalAssets: 500000.0,
			Cash:        300000.0,
			MarketValue: 200000.0,
		}, nil
	case "US":
		return &AccountFunds{
			Market:      "US",
			Currency:    "USD",
			TotalAssets: 200000.0,
			Cash:        150000.0,
			MarketValue: 50000.0,
		}, nil
	default:
		return &AccountFunds{
			Market:      "ALL",
			Currency:    "CNY",
			TotalAssets: 1700000.0,
			Cash:        1100000.0,
			MarketValue: 600000.0,
		}, nil
	}
}

func (c *Client) GetAllAccountFunds(ctx context.Context) ([]AccountFunds, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Futu OpenD")
	}

	return []AccountFunds{
		{Market: "CN", Currency: "CNY", TotalAssets: 1000000.0, Cash: 650000.0, MarketValue: 350000.0},
		{Market: "HK", Currency: "HKD", TotalAssets: 500000.0, Cash: 300000.0, MarketValue: 200000.0},
		{Market: "US", Currency: "USD", TotalAssets: 200000.0, Cash: 150000.0, MarketValue: 50000.0},
	}, nil
}

func (c *Client) GetPositions(ctx context.Context, market string) ([]Position, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Futu OpenD")
	}

	allPositions := []Position{
		{Code: "600519", Market: "CN", Name: "贵州茅台", Quantity: 100, AvgCost: 1500.00, CurrentPrice: 1520.50, UnrealizedPnL: 2050.00},
		{Code: "000858", Market: "CN", Name: "五粮液", Quantity: 200, AvgCost: 120.00, CurrentPrice: 118.30, UnrealizedPnL: -340.00},
		{Code: "00700", Market: "HK", Name: "腾讯控股", Quantity: 100, AvgCost: 350.00, CurrentPrice: 380.00, UnrealizedPnL: 3000.00},
		{Code: "9988", Market: "HK", Name: "阿里巴巴-SW", Quantity: 200, AvgCost: 85.00, CurrentPrice: 82.00, UnrealizedPnL: -600.00},
		{Code: "AAPL", Market: "US", Name: "苹果公司", Quantity: 50, AvgCost: 180.00, CurrentPrice: 190.00, UnrealizedPnL: 500.00},
		{Code: "TSLA", Market: "US", Name: "特斯拉", Quantity: 20, AvgCost: 250.00, CurrentPrice: 240.00, UnrealizedPnL: -200.00},
	}

	if market == "" || market == "ALL" {
		return allPositions, nil
	}

	var filtered []Position
	for _, pos := range allPositions {
		if pos.Market == market {
			filtered = append(filtered, pos)
		}
	}
	return filtered, nil
}

func (c *Client) PlaceOrder(ctx context.Context, market, code, side string, price float64, quantity int) (string, error) {
	if !c.IsConnected() {
		return "", fmt.Errorf("not connected to Futu OpenD")
	}

	orderID := fmt.Sprintf("SIM-%s-%d", market, time.Now().UnixNano())
	log.Printf("Simulated order: %s %s %s %d @ %.2f (ID: %s)", side, market, code, quantity, price, orderID)
	return orderID, nil
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

type OrderResult struct {
	OrderID   string
	OrderIDEx string
}
