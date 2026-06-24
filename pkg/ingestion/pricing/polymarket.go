package pricing

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type PolymarketClient struct {
	baseURL    string
	conn       *websocket.Conn
	mu         sync.Mutex
	updates    chan PriceTick
	subscribed map[string]bool
	closed     bool
}

func NewPolymarketClient() *PolymarketClient {
	return &PolymarketClient{
		baseURL:    "wss://ws-subscriptions-clob.polymarket.com/ws/",
		updates:    make(chan PriceTick, 1000),
		subscribed: make(map[string]bool),
	}
}

func (c *PolymarketClient) Updates() <-chan PriceTick { return c.updates }

func (c *PolymarketClient) Subscribe(ctx context.Context, marketIDs []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		if err := c.connect(ctx); err != nil {
			return fmt.Errorf("polymarket ws connect: %w", err)
		}
	}

	subMsg := map[string]interface{}{
		"type":    "subscribe",
		"channel": "tickSize",
		"markets": marketIDs,
	}
	if err := c.conn.WriteJSON(subMsg); err != nil {
		return fmt.Errorf("polymarket subscribe: %w", err)
	}
	for _, id := range marketIDs {
		c.subscribed[id] = true
	}
	return nil
}

func (c *PolymarketClient) Unsubscribe(marketIDs []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	unsubMsg := map[string]interface{}{
		"type":    "unsubscribe",
		"channel": "tickSize",
		"markets": marketIDs,
	}
	if err := c.conn.WriteJSON(unsubMsg); err != nil {
		return fmt.Errorf("polymarket unsubscribe: %w", err)
	}
	for _, id := range marketIDs {
		delete(c.subscribed, id)
	}
	return nil
}

func (c *PolymarketClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *PolymarketClient) connect(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.baseURL, nil)
	if err != nil {
		return fmt.Errorf("polymarket ws dial: %w", err)
	}
	c.conn = conn
	go c.readLoop()
	return nil
}

type polymarketWSMessage struct {
	Type   string  `json:"type"`
	Market string  `json:"market"`
	Bid    float64 `json:"bid"`
	Ask    float64 `json:"ask"`
}

func (c *PolymarketClient) readLoop() {
	for {
		_, msgBytes, err := c.conn.ReadMessage()
		if err != nil {
			slog.Warn("polymarket ws read error", "error", err)
			return
		}

		var msg polymarketWSMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}

		c.updates <- PriceTick{
			Venue:     "POLYMARKET",
			MarketID:  msg.Market,
			Bid:       msg.Bid,
			Ask:       msg.Ask,
			Timestamp: time.Now(),
		}
	}
}
