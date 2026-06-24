package futu

import (
	"context"
	"encoding/binary"
	"encoding/json"
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

func (c *Client) GetAccountFunds(ctx context.Context) (*AccountFunds, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Futu OpenD")
	}

	return &AccountFunds{
		TotalAssets: 1000000.0,
		Cash:        650000.0,
		MarketValue: 350000.0,
	}, nil
}

func (c *Client) GetPositions(ctx context.Context) ([]Position, error) {
	if !c.IsConnected() {
		return nil, fmt.Errorf("not connected to Futu OpenD")
	}

	return []Position{
		{
			Code:          "600519",
			Market:        "CN",
			Quantity:      100,
			AvgCost:       1500.00,
			CurrentPrice:  1520.50,
			UnrealizedPnL: 2050.00,
		},
		{
			Code:          "000858",
			Market:        "CN",
			Quantity:      200,
			AvgCost:       120.00,
			CurrentPrice:   118.30,
			UnrealizedPnL: -340.00,
		},
	}, nil
}

func (c *Client) PlaceOrder(ctx context.Context, market, code, side string, price float64, quantity int) (string, error) {
	if !c.IsConnected() {
		return "", fmt.Errorf("not connected to Futu OpenD")
	}

	orderID := fmt.Sprintf("SIM-%d", time.Now().UnixNano())
	log.Printf("Simulated order: %s %s %s %d @ %.2f (ID: %s)", side, market, code, quantity, price, orderID)
	return orderID, nil
}

func toJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func (c *Client) sendJSON(protoID uint32, data interface{}) (json.RawMessage, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	resp, err := c.sendRequest(&FutuRequest{
		ProtoID: protoID,
		Body:    body,
	})
	if err != nil {
		return nil, err
	}

	var result json.RawMessage
	if err := json.Unmarshal(resp.Body, &result); err != nil {
		return resp.Body, nil
	}

	return result, nil
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
	TotalAssets float64
	Cash        float64
	MarketValue float64
}

type Position struct {
	Code          string
	Market        string
	Quantity      int
	AvgCost       float64
	CurrentPrice  float64
	UnrealizedPnL float64
}

type OrderResult struct {
	OrderID   string
	OrderIDEx string
}
