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

type KalshiClient struct {
	baseURL    string
	keyID      string
	keyPEM     string
	conn       *websocket.Conn
	mu         sync.Mutex
	updates    chan PriceTick
	subscribed map[string]bool
	closed     bool
}

func NewKalshiClient(keyID, keyPEM string) *KalshiClient {
	return &KalshiClient{
		baseURL:    "wss://api.elections.kalshi.com/trade-api/ws",
		keyID:      keyID,
		keyPEM:     keyPEM,
		updates:    make(chan PriceTick, 1000),
		subscribed: make(map[string]bool),
	}
}

func (c *KalshiClient) Updates() <-chan PriceTick { return c.updates }

func (c *KalshiClient) Subscribe(ctx context.Context, marketIDs []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn == nil {
		if err := c.connect(ctx); err != nil {
			return fmt.Errorf("kalshi ws connect: %w", err)
		}
	}

	for _, id := range marketIDs {
		if c.subscribed[id] {
			continue
		}
		subMsg := map[string]interface{}{
			"type":    "subscribe",
			"channel": fmt.Sprintf("orderbook_%s", id),
		}
		if err := c.conn.WriteJSON(subMsg); err != nil {
			return fmt.Errorf("kalshi subscribe %s: %w", id, err)
		}
		c.subscribed[id] = true
		slog.Debug("kalshi subscribed", "market", id)
	}
	return nil
}

func (c *KalshiClient) Unsubscribe(marketIDs []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, id := range marketIDs {
		if !c.subscribed[id] {
			continue
		}
		unsubMsg := map[string]interface{}{
			"type":    "unsubscribe",
			"channel": fmt.Sprintf("orderbook_%s", id),
		}
		if err := c.conn.WriteJSON(unsubMsg); err != nil {
			return fmt.Errorf("kalshi unsubscribe %s: %w", id, err)
		}
		delete(c.subscribed, id)
	}
	return nil
}

func (c *KalshiClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *KalshiClient) connect(ctx context.Context) error {
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.baseURL, nil)
	if err != nil {
		return fmt.Errorf("kalshi ws dial: %w", err)
	}
	c.conn = conn
	go c.readLoop()
	return nil
}

type kalshiWSMessage struct {
	Type    string          `json:"type"`
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

type kalshiOrderbook struct {
	Ticker string  `json:"ticker"`
	YesAsk float64 `json:"yes_ask"`
	YesBid float64 `json:"yes_bid"`
	NoAsk  float64 `json:"no_ask"`
	NoBid  float64 `json:"no_bid"`
}

func (c *KalshiClient) readLoop() {
	for {
		_, msgBytes, err := c.conn.ReadMessage()
		if err != nil {
			slog.Warn("kalshi ws read error", "error", err)
			return
		}

		var msg kalshiWSMessage
		if err := json.Unmarshal(msgBytes, &msg); err != nil {
			continue
		}

		if msg.Type == "orderbook" {
			var ob kalshiOrderbook
			if err := json.Unmarshal(msg.Data, &ob); err != nil {
				continue
			}
			c.updates <- PriceTick{
				Venue:     "KALSHI",
				MarketID:  ob.Ticker,
				Bid:       ob.YesBid,
				Ask:       ob.YesAsk,
				Timestamp: time.Now(),
			}
		}
	}
}
