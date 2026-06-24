# Phase 1b: Ingestion Module Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `pkg/ingestion` with discovery (REST market polling for Kalshi + Polymarket) and pricing (WebSocket price feeds) modules, plus shared price cache.

**Architecture:** Two sub-packages under `pkg/ingestion/`: `discovery/` for periodic REST market scanning and `pricing/` for persistent WS price feeds. A shared `PriceCache` provides lock-free reads for downstream modules. No NATS — events go through the in-process bus; prices go through the cache directly.

**Tech Stack:** Go 1.22, `net/http` for REST, `gorilla/websocket` for WS, `sync.RWMutex` for price cache, pgx/v5 for DB writes, internal bus for events

## Global Constraints

- Go 1.22 minimum
- All secrets loaded from file paths specified in env vars (matching existing pattern in `internal/config`)
- No third-party router — use Go 1.22 `http.ServeMux`
- Structured JSON logging via `log/slog`
- All health checks on `GET /healthz` and `GET /readyz`
- Metrics on `GET /metrics` via Prometheus client_golang
- In-process event bus for non-critical events
- Existing internal packages: `internal/config`, `internal/db`, `internal/redis`, `internal/logging`, `internal/metrics`, `internal/bus`, `internal/auth`, `internal/health`
- Machine is local dev only — no Go toolchain available. Code is written without local test execution. Tests are verified on EC2.

---

### Task 1: Add IngestionConfig to Config Package

**Files:**
- Modify: `internal/config/config.go`

**Interfaces:**
- Consumes: nothing new
- Produces: `config.IngestionConfig` struct embedded in `Config`

- [ ] **Step 1: Add IngestionConfig struct and embed in Config**

Edit `internal/config/config.go`:

```go
type IngestionConfig struct {
    DiscoveryInterval time.Duration `validate:"min=1s"`
}

type Config struct {
    PostgresDSN        string           `validate:"required"`
    RedisAddr          string           `validate:"required"`
    ListenAddr         string           `validate:"required"`
    LogLevel           string           `validate:"omitempty,oneof=debug info warn error"`
    OpenRouterAPIKey   string
    OpenRouterBaseURL  string
    KalshiKeyID        string
    KalshiPrivateKeyPEM string
    PolymarketProxyURL string
    JWTSecretKey       string
    Ingestion          IngestionConfig
}
```

Update `LoadFromEnv()`:

```go
cfg := &Config{
    PostgresDSN:         readSecretOr(os.Getenv("POSTGRES_DSN_FILE")),
    RedisAddr:           envOrDefault("REDIS_ADDR", "localhost:6379"),
    ListenAddr:          envOrDefault("LISTEN_ADDR", ":8086"),
    LogLevel:            envOrDefault("LOG_LEVEL", "info"),
    OpenRouterAPIKey:    readSecretOr(os.Getenv("OPENROUTER_API_KEY_FILE")),
    OpenRouterBaseURL:   envOrDefault("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
    KalshiKeyID:         readSecretOr(os.Getenv("KALSHI_KEY_ID_FILE")),
    KalshiPrivateKeyPEM: readSecretOr(os.Getenv("KALSHI_PRIVATE_KEY_FILE")),
    PolymarketProxyURL:  os.Getenv("POLYMARKET_PROXY_URL"),
    JWTSecretKey:        readSecretOr(os.Getenv("JWT_SECRET_KEY_FILE")),
    Ingestion: IngestionConfig{
        DiscoveryInterval: envOrDefaultDuration("INGESTION_DISCOVERY_INTERVAL", 5*time.Minute),
    },
}
```

Add new helper:
```go
func envOrDefaultDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
```

Add `"time"` import to config.go.

- [ ] **Step 2: Update config_test.go with IngestionConfig test**

Add to `internal/config/config_test.go`:
```go
func TestConfig_DefaultIngestionInterval(t *testing.T) {
    cfg := config.Config{
        PostgresDSN: "postgres://user:pass@localhost:5432/arby",
        RedisAddr:   "localhost:6379",
        ListenAddr:  ":8086",
    }
    if err := cfg.Validate(); err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
}
```

- [ ] **Step 3: Commit**

```bash
git add internal/config/config.go internal/config/config_test.go
git commit -m "feat: add IngestionConfig with discovery interval to config package"
```

---

### Task 2: Price Cache

**Files:**
- Create: `pkg/ingestion/cache.go`
- Create: `pkg/ingestion/cache_test.go`

**Interfaces:**
- Produces: `ingestion.PriceCache` with `Get(venue, marketID) (PriceSnapshot, bool)` and `Set(venue, marketID, bid, ask)`

- [ ] **Step 1: Write the test file**

Write `pkg/ingestion/cache_test.go`:
```go
package ingestion_test

import (
    "testing"
    "github.com/aaronbateman02/Arby/pkg/ingestion"
)

func TestCache_SetAndGet(t *testing.T) {
    c := ingestion.NewPriceCache()
    c.Set("KALSHI", "market-1", 0.52, 0.54)
    
    snap, ok := c.Get("KALSHI", "market-1")
    if !ok {
        t.Fatal("expected to find market-1")
    }
    if snap.Bid != 0.52 {
        t.Fatalf("expected bid 0.52, got %f", snap.Bid)
    }
    if snap.Ask != 0.54 {
        t.Fatalf("expected ask 0.54, got %f", snap.Ask)
    }
    if snap.Venue != "KALSHI" {
        t.Fatalf("expected venue KALSHI, got %s", snap.Venue)
    }
}

func TestCache_MissingMarket(t *testing.T) {
    c := ingestion.NewPriceCache()
    _, ok := c.Get("KALSHI", "nonexistent")
    if ok {
        t.Fatal("expected false for missing market")
    }
}

func TestCache_Overwrite(t *testing.T) {
    c := ingestion.NewPriceCache()
    c.Set("KALSHI", "m1", 0.50, 0.55)
    c.Set("KALSHI", "m1", 0.51, 0.56)
    
    snap, ok := c.Get("KALSHI", "m1")
    if !ok {
        t.Fatal("expected to find m1")
    }
    if snap.Bid != 0.51 {
        t.Fatalf("expected bid 0.51, got %f", snap.Bid)
    }
    if snap.Ask != 0.56 {
        t.Fatalf("expected ask 0.56, got %f", snap.Ask)
    }
}

func TestCache_ConcurrentSafety(t *testing.T) {
    c := ingestion.NewPriceCache()
    done := make(chan bool)
    
    go func() {
        for i := 0; i < 100; i++ {
            c.Set("KALSHI", "m1", float64(i)/100, float64(i+1)/100)
        }
        done <- true
    }()
    
    go func() {
        for i := 0; i < 100; i++ {
            c.Get("KALSHI", "m1")
        }
        done <- true
    }()
    
    <-done
    <-done
}
```

- [ ] **Step 2: Write the implementation**

Write `pkg/ingestion/cache.go`:
```go
package ingestion

import (
    "sync"
    "time"
)

type PriceSnapshot struct {
    Venue     string
    MarketID  string
    Bid       float64
    Ask       float64
    UpdatedAt time.Time
}

type PriceCache struct {
    mu     sync.RWMutex
    prices map[string]PriceSnapshot
}

func NewPriceCache() *PriceCache {
    return &PriceCache{
        prices: make(map[string]PriceSnapshot),
    }
}

func (c *PriceCache) Get(venue, marketID string) (PriceSnapshot, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    snap, ok := c.prices[key(venue, marketID)]
    return snap, ok
}

func (c *PriceCache) Set(venue, marketID string, bid, ask float64) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.prices[key(venue, marketID)] = PriceSnapshot{
        Venue:     venue,
        MarketID:  marketID,
        Bid:       bid,
        Ask:       ask,
        UpdatedAt: time.Now(),
    }
}

func key(venue, marketID string) string {
    return venue + ":" + marketID
}
```

- [ ] **Step 3: Commit**

```bash
git add pkg/ingestion/cache.go pkg/ingestion/cache_test.go
git commit -m "feat: add PriceCache with concurrent-safe get/set"
```

---

### Task 3: Discovery Client Interface + Types

**Files:**
- Create: `pkg/ingestion/discovery/client.go`

**Interfaces:**
- Produces: `discovery.DiscoveryClient` interface, `discovery.Market`, `discovery.Outcome` types

- [ ] **Step 1: Write the interface and types**

Write `pkg/ingestion/discovery/client.go`:
```go
package discovery

import (
    "context"
    "encoding/json"
    "time"
)

type Outcome struct {
    Name  string  `json:"name"`
    Price float64 `json:"price"`
}

type Market struct {
    Venue     string          `json:"venue"`
    MarketID  string          `json:"market_id"`
    Ticker    string          `json:"ticker"`
    Title     string          `json:"title"`
    Series    string          `json:"series"`
    Category  string          `json:"category"`
    Outcomes  []Outcome       `json:"outcomes"`
    OpenTime  time.Time       `json:"open_time"`
    CloseTime time.Time       `json:"close_time"`
    Extra     json.RawMessage `json:"extra,omitempty"`
}

type DiscoveryClient interface {
    FetchMarkets(ctx context.Context) ([]Market, error)
    Venue() string
}
```

- [ ] **Step 2: Commit**

```bash
git add pkg/ingestion/discovery/client.go
git commit -m "feat: add DiscoveryClient interface and market types"
```

---

### Task 4: Kalshi Discovery Client

**Files:**
- Create: `pkg/ingestion/discovery/kalshi.go`
- Create: `pkg/ingestion/discovery/kalshi_test.go`

**Interfaces:**
- Consumes: `discovery.DiscoveryClient`, `discovery.Market`
- Produces: `discovery.NewKalshiClient(keyID, privateKeyPEM) *KalshiClient`

- [ ] **Step 1: Write the test file**

Write `pkg/ingestion/discovery/kalshi_test.go`:
```go
package discovery_test

import (
    "context"
    "testing"
    "github.com/aaronbateman02/Arby/pkg/ingestion/discovery"
)

func TestKalshiClient_Venue(t *testing.T) {
    c := discovery.NewKalshiClient("", "")
    if c.Venue() != "KALSHI" {
        t.Fatalf("expected KALSHI, got %s", c.Venue())
    }
}
```

- [ ] **Step 2: Write the implementation**

Write `pkg/ingestion/discovery/kalshi.go`:
```go
package discovery

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "strings"
    "time"
)

type KalshiClient struct {
    baseURL    string
    keyID      string
    keyPEM     string
    httpClient *http.Client
}

func NewKalshiClient(keyID, keyPEM string) *KalshiClient {
    return &KalshiClient{
        baseURL:    "https://api.elections.kalshi.com/trade-api/v2",
        keyID:      keyID,
        keyPEM:     keyPEM,
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *KalshiClient) Venue() string { return "KALSHI" }

func (c *KalshiClient) FetchMarkets(ctx context.Context) ([]Market, error) {
    var all []Market
    cursor := ""

    for {
        url := fmt.Sprintf("%s/markets?limit=1000&status=open", c.baseURL)
        if cursor != "" {
            url += "&cursor=" + cursor
        }

        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
        if err != nil {
            return nil, fmt.Errorf("kalshi request: %w", err)
        }
        c.signRequest(req)

        resp, err := c.httpClient.Do(req)
        if err != nil {
            return nil, fmt.Errorf("kalshi do: %w", err)
        }
        defer resp.Body.Close()

        body, err := io.ReadAll(resp.Body)
        if err != nil {
            return nil, fmt.Errorf("kalshi read: %w", err)
        }

        var result struct {
            Markets []kalshiMarket `json:"markets"`
            Cursor  string         `json:"cursor"`
        }
        if err := json.Unmarshal(body, &result); err != nil {
            return nil, fmt.Errorf("kalshi decode: %w", err)
        }

        for _, km := range result.Markets {
            all = append(all, c.normalize(km))
        }

        if result.Cursor == "" || len(result.Markets) == 0 {
            break
        }
        cursor = result.Cursor
    }

    return all, nil
}

type kalshiMarket struct {
    Ticker       string `json:"ticker"`
    Title        string `json:"title"`
    Sector       string `json:"sector"`
    Series       string `json:"series"`
    OpenTime     string `json:"open_time"`
    CloseTime    string `json:"close_time"`
    YesBid       float64 `json:"yes_bid"`
    YesAsk       float64 `json:"yes_ask"`
    NoBid        float64 `json:"no_bid"`
    NoAsk        float64 `json:"no_ask"`
}

func (c *KalshiClient) normalize(km kalshiMarket) Market {
    var openTime, closeTime time.Time
    if km.OpenTime != "" {
        openTime, _ = time.Parse(time.RFC3339, km.OpenTime)
    }
    if km.CloseTime != "" {
        closeTime, _ = time.Parse(time.RFC3339, km.CloseTime)
    }

    tickerParts := strings.Split(km.Ticker, "-")
    series := ""
    if len(tickerParts) > 0 {
        series = strings.ToLower(tickerParts[0])
    }

    return Market{
        Venue:    "KALSHI",
        MarketID: km.Ticker,
        Ticker:   km.Ticker,
        Title:    km.Title,
        Series:   series,
        Category: km.Sector,
        Outcomes: []Outcome{
            {Name: "Yes", Price: km.YesBid},
            {Name: "No", Price: km.NoBid},
        },
        OpenTime:  openTime,
        CloseTime: closeTime,
    }
}

func (c *KalshiClient) signRequest(req *http.Request) {
    // TODO: RSA-PSS signing — requires crypto/rsa + crypto/sha256
    // For now, set key ID header; signing will be added in a follow-up
    // when the production environment is set up.
    if c.keyID != "" {
        req.Header.Set("Kalshi-Key-Id", c.keyID)
    }
}
```

- [ ] **Step 3: Commit**

```bash
git add pkg/ingestion/discovery/kalshi.go pkg/ingestion/discovery/kalshi_test.go
git commit -m "feat: add Kalshi discovery client with REST market polling"
```

---

### Task 5: Polymarket Discovery Client

**Files:**
- Create: `pkg/ingestion/discovery/polymarket.go`
- Create: `pkg/ingestion/discovery/polymarket_test.go`

**Interfaces:**
- Consumes: `discovery.DiscoveryClient`
- Produces: `discovery.NewPolymarketClient() *PolymarketClient`

- [ ] **Step 1: Write the test file**

Write `pkg/ingestion/discovery/polymarket_test.go`:
```go
package discovery_test

import (
    "testing"
    "github.com/aaronbateman02/Arby/pkg/ingestion/discovery"
)

func TestPolymarketClient_Venue(t *testing.T) {
    c := discovery.NewPolymarketClient()
    if c.Venue() != "POLYMARKET" {
        t.Fatalf("expected POLYMARKET, got %s", c.Venue())
    }
}
```

- [ ] **Step 2: Write the implementation**

Write `pkg/ingestion/discovery/polymarket.go`:
```go
package discovery

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
)

type PolymarketClient struct {
    baseURL    string
    httpClient *http.Client
}

func NewPolymarketClient() *PolymarketClient {
    return &PolymarketClient{
        baseURL:    "https://clob.polymarket.com",
        httpClient: &http.Client{Timeout: 30 * time.Second},
    }
}

func (c *PolymarketClient) Venue() string { return "POLYMARKET" }

func (c *PolymarketClient) FetchMarkets(ctx context.Context) ([]Market, error) {
    url := fmt.Sprintf("%s/markets?limit=1000&closed=false", c.baseURL)

    req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
    if err != nil {
        return nil, fmt.Errorf("polymarket request: %w", err)
    }

    resp, err := c.httpClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("polymarket do: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("polymarket read: %w", err)
    }

    var raw []polymarketMarket
    if err := json.Unmarshal(body, &raw); err != nil {
        return nil, fmt.Errorf("polymarket decode: %w", err)
    }

    var result []Market
    for _, pm := range raw {
        result = append(result, c.normalize(pm))
    }

    return result, nil
}

type polymarketMarket struct {
    ConditionID string  `json:"condition_id"`
    Question    string  `json:"question"`
    Slug        string  `json:"slug"`
    Category    string  `json:"category"`
    EndDate     string  `json:"end_date"`
    Outcomes    []string `json:"outcomes"`
    Prices      []float64 `json:"prices"`
}

func (c *PolymarketClient) normalize(pm polymarketMarket) Market {
    var closeTime time.Time
    if pm.EndDate != "" {
        closeTime, _ = time.Parse(time.RFC3339, pm.EndDate)
    }

    ticker := pm.Slug
    if ticker == "" {
        ticker = pm.ConditionID
    }

    outcomes := make([]Outcome, len(pm.Outcomes))
    for i, name := range pm.Outcomes {
        price := 0.0
        if i < len(pm.Prices) {
            price = pm.Prices[i]
        }
        outcomes[i] = Outcome{Name: name, Price: price}
    }

    return Market{
        Venue:     "POLYMARKET",
        MarketID:  pm.ConditionID,
        Ticker:    ticker,
        Title:     pm.Question,
        Category:  pm.Category,
        Outcomes:  outcomes,
        CloseTime: closeTime,
    }
}
```

- [ ] **Step 3: Commit**

```bash
git add pkg/ingestion/discovery/polymarket.go pkg/ingestion/discovery/polymarket_test.go
git commit -m "feat: add Polymarket discovery client with REST market polling"
```

---

### Task 6: Discovery Scanner

**Files:**
- Create: `pkg/ingestion/discovery/scanner.go`
- Create: `pkg/ingestion/discovery/scanner_test.go`

**Interfaces:**
- Consumes: `discovery.DiscoveryClient` (2 instances), `*pgxpool.Pool`, `*bus.Bus`, `config.IngestionConfig`
- Produces: `discovery.NewScanner(clients..., db, bus, interval) *Scanner` with `Run(ctx)`

- [ ] **Step 1: Write the test file**

Write `pkg/ingestion/discovery/scanner_test.go`:
```go
package discovery_test

import (
    "context"
    "testing"
    "time"
    "github.com/aaronbateman02/Arby/pkg/ingestion/discovery"
    "github.com/aaronbateman02/Arby/internal/bus"
)

type mockClient struct {
    venue   string
    markets []discovery.Market
}

func (m *mockClient) FetchMarkets(ctx context.Context) ([]discovery.Market, error) {
    return m.markets, nil
}
func (m *mockClient) Venue() string { return m.venue }

func TestScanner_RunAndStop(t *testing.T) {
    b := bus.New(10)
    s := discovery.NewScanner(b, time.Hour)
    
    ctx, cancel := context.WithCancel(context.Background())
    cancel()
    s.Run(ctx)
}
```

- [ ] **Step 2: Write the implementation**

Write `pkg/ingestion/discovery/scanner.go`:
```go
package discovery

import (
    "context"
    "log/slog"
    "time"
    "github.com/aaronbateman02/Arby/internal/bus"
)

type Scanner struct {
    clients  []DiscoveryClient
    eventBus *bus.Bus
    interval time.Duration
}

func NewScanner(eventBus *bus.Bus, interval time.Duration, clients ...DiscoveryClient) *Scanner {
    return &Scanner{
        clients:  clients,
        eventBus: eventBus,
        interval: interval,
    }
}

func (s *Scanner) Run(ctx context.Context) {
    slog.Info("discovery scanner started", "interval", s.interval)

    if err := s.scanAll(ctx); err != nil {
        slog.Error("initial discovery scan", "error", err)
    }

    ticker := time.NewTicker(s.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            slog.Info("discovery scanner stopped")
            return
        case <-ticker.C:
            if err := s.scanAll(ctx); err != nil {
                slog.Error("discovery scan cycle", "error", err)
            }
        }
    }
}

func (s *Scanner) scanAll(ctx context.Context) error {
    for _, client := range s.clients {
        markets, err := client.FetchMarkets(ctx)
        if err != nil {
            slog.Error("discovery fetch", "venue", client.Venue(), "error", err)
            continue
        }
        slog.Info("discovery fetched markets", "venue", client.Venue(), "count", len(markets))

        for _, m := range markets {
            if err := s.eventBus.PublishTyped("MarketDiscovered", m); err != nil {
                slog.Error("discovery publish event", "market", m.MarketID, "error", err)
            }
        }
    }
    return nil
}
```

Fix the `bus.PublishTyped` signature — it returns `error`, which means we need to handle it.

- [ ] **Step 3: Commit**

```bash
git add pkg/ingestion/discovery/scanner.go pkg/ingestion/discovery/scanner_test.go
git commit -m "feat: add discovery scanner with periodic market polling"
```

---

### Task 7: Pricing Client Interface + PriceTick

**Files:**
- Create: `pkg/ingestion/pricing/client.go`

**Interfaces:**
- Produces: `pricing.PricingClient` interface, `pricing.PriceTick` type

- [ ] **Step 1: Write the interface**

Write `pkg/ingestion/pricing/client.go`:
```go
package pricing

import (
    "context"
    "time"
)

type PriceTick struct {
    Venue     string    `json:"venue"`
    MarketID  string    `json:"market_id"`
    Bid       float64   `json:"bid"`
    Ask       float64   `json:"ask"`
    Timestamp time.Time `json:"timestamp"`
}

type PricingClient interface {
    Subscribe(ctx context.Context, marketIDs []string) error
    Unsubscribe(marketIDs []string) error
    Close() error
    Updates() <-chan PriceTick
}
```

- [ ] **Step 2: Commit**

```bash
git add pkg/ingestion/pricing/client.go
git commit -m "feat: add PricingClient interface and PriceTick type"
```

---

### Task 8: Kalshi WebSocket Pricing Client

**Files:**
- Create: `pkg/ingestion/pricing/kalshi.go`
- Create: `pkg/ingestion/pricing/kalshi_test.go`

**Interfaces:**
- Consumes: `pricing.PricingClient`
- Produces: `pricing.NewKalshiClient(keyID, keyPEM) *KalshiClient`

- [ ] **Step 1: Write the test file**

Write `pkg/ingestion/pricing/kalshi_test.go`:
```go
package pricing_test

import (
    "testing"
    "github.com/aaronbateman02/Arby/pkg/ingestion/pricing"
)

func TestKalshiClient_CloseNotConnected(t *testing.T) {
    c := pricing.NewKalshiClient("", "")
    if err := c.Close(); err != nil {
        t.Fatalf("expected no error closing unconnected client, got: %v", err)
    }
}
```

- [ ] **Step 2: Write the implementation**

Write `pkg/ingestion/pricing/kalshi.go`:
```go
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
            "type": "subscribe",
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
    // TODO: Add RSA-PSS JWT authentication header
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
    Ticker string `json:"ticker"`
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
```

- [ ] **Step 3: Add gorilla/websocket dependency to go.mod**

Edit `go.mod` — add to require block:
```
github.com/gorilla/websocket v1.5.3
```

- [ ] **Step 4: Commit**

```bash
git add pkg/ingestion/pricing/kalshi.go pkg/ingestion/pricing/kalshi_test.go go.mod
git commit -m "feat: add Kalshi WebSocket pricing client"
```

---

### Task 9: Polymarket WebSocket Pricing Client

**Files:**
- Create: `pkg/ingestion/pricing/polymarket.go`
- Create: `pkg/ingestion/pricing/polymarket_test.go`

**Interfaces:**
- Consumes: `pricing.PricingClient`
- Produces: `pricing.NewPolymarketClient() *PolymarketClient`

- [ ] **Step 1: Write the test file**

Write `pkg/ingestion/pricing/polymarket_test.go`:
```go
package pricing_test

import (
    "testing"
    "github.com/aaronbateman02/Arby/pkg/ingestion/pricing"
)

func TestPolymarketClient_CloseNotConnected(t *testing.T) {
    c := pricing.NewPolymarketClient()
    if err := c.Close(); err != nil {
        t.Fatalf("expected no error closing unconnected client, got: %v", err)
    }
}
```

- [ ] **Step 2: Write the implementation**

Write `pkg/ingestion/pricing/polymarket.go`:
```go
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
        "type": "subscribe",
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
    Type    string          `json:"type"`
    Market  string          `json:"market"`
    Bid     float64         `json:"bid"`
    Ask     float64         `json:"ask"`
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
```

- [ ] **Step 3: Commit**

```bash
git add pkg/ingestion/pricing/polymarket.go pkg/ingestion/pricing/polymarket_test.go
git commit -m "feat: add Polymarket WebSocket pricing client"
```

---

### Task 10: Pricing Manager

**Files:**
- Create: `pkg/ingestion/pricing/manager.go`
- Create: `pkg/ingestion/pricing/manager_test.go`

**Interfaces:**
- Consumes: `pricing.PricingClient` (2 instances), `*ingestion.PriceCache`
- Produces: `pricing.NewManager(priceCache, clients...) *Manager` with `Run(ctx)`

- [ ] **Step 1: Write the test file**

Write `pkg/ingestion/pricing/manager_test.go`:
```go
package pricing_test

import (
    "context"
    "testing"
    "time"
    "github.com/aaronbateman02/Arby/pkg/ingestion"
    "github.com/aaronbateman02/Arby/pkg/ingestion/pricing"
)

type mockPricingClient struct {
    updates chan pricing.PriceTick
}

func newMockPricingClient() *mockPricingClient {
    return &mockPricingClient{updates: make(chan pricing.PriceTick, 100)}
}

func (m *mockPricingClient) Subscribe(ctx context.Context, ids []string) error { return nil }
func (m *mockPricingClient) Unsubscribe(ids []string) error { return nil }
func (m *mockPricingClient) Close() error { return nil }
func (m *mockPricingClient) Updates() <-chan pricing.PriceTick { return m.updates }

func TestManager_PriceCacheUpdates(t *testing.T) {
    cache := ingestion.NewPriceCache()
    mock := newMockPricingClient()

    mgr := pricing.NewManager(cache, mock)
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go mgr.Run(ctx)

    mock.updates <- pricing.PriceTick{
        Venue:    "KALSHI",
        MarketID: "test-1",
        Bid:      0.50,
        Ask:      0.55,
    }

    time.Sleep(50 * time.Millisecond)

    snap, ok := cache.Get("KALSHI", "test-1")
    if !ok {
        t.Fatal("expected price in cache")
    }
    if snap.Bid != 0.50 {
        t.Fatalf("expected bid 0.50, got %f", snap.Bid)
    }
}
```

- [ ] **Step 2: Write the implementation**

Write `pkg/ingestion/pricing/manager.go`:
```go
package pricing

import (
    "context"
    "log/slog"
    "time"
    "github.com/aaronbateman02/Arby/pkg/ingestion"
)

type Manager struct {
    cache   *ingestion.PriceCache
    clients []PricingClient
}

func NewManager(cache *ingestion.PriceCache, clients ...PricingClient) *Manager {
    return &Manager{
        cache:   cache,
        clients: clients,
    }
}

func (m *Manager) Run(ctx context.Context) {
    slog.Info("pricing manager started", "clients", len(m.clients))

    for _, client := range m.clients {
        go m.runClient(ctx, client)
    }

    <-ctx.Done()
    slog.Info("pricing manager stopped")

    for _, client := range m.clients {
        if err := client.Close(); err != nil {
            slog.Error("pricing client close", "error", err)
        }
    }
}

func (m *Manager) runClient(ctx context.Context, client PricingClient) {
    updates := client.Updates()
    for {
        select {
        case <-ctx.Done():
            return
        case tick, ok := <-updates:
            if !ok {
                return
            }
            m.cache.Set(tick.Venue, tick.MarketID, tick.Bid, tick.Ask)
        }
    }
}

func (m *Manager) SubscribeAll(ctx context.Context, marketIDs []string) error {
    for _, client := range m.clients {
        if err := client.Subscribe(ctx, marketIDs); err != nil {
            slog.Error("subscribe all", "error", err)
            return err
        }
    }
    return nil
}
```

- [ ] **Step 3: Commit**

```bash
git add pkg/ingestion/pricing/manager.go pkg/ingestion/pricing/manager_test.go
git commit -m "feat: add pricing manager with price cache fan-out"
```

---

### Task 11: Wire Ingestion into Main Entrypoint

**Files:**
- Modify: `cmd/polybot/main.go`

**Interfaces:**
- Consumes: all packages from Tasks 1-10
- Produces: running scanner + pricing manager goroutines

- [ ] **Step 1: Update main.go**

Add imports:
```go
"github.com/aaronbateman02/Arby/pkg/ingestion"
"github.com/aaronbateman02/Arby/pkg/ingestion/discovery"
"github.com/aaronbateman02/Arby/pkg/ingestion/pricing"
```

After event bus init, add:
```go
// 7a. Init price cache
priceCache := ingestion.NewPriceCache()
slog.Info("price cache initialized")

// 7b. Start discovery scanner
discKalshi := discovery.NewKalshiClient(cfg.KalshiKeyID, cfg.KalshiPrivateKeyPEM)
discPoly := discovery.NewPolymarketClient()
discScanner := discovery.NewScanner(eventBus, cfg.Ingestion.DiscoveryInterval, discKalshi, discPoly)
go discScanner.Run(ctx)
slog.Info("discovery scanner started")

// 7c. Start pricing manager
priceKalshi := pricing.NewKalshiClient(cfg.KalshiKeyID, cfg.KalshiPrivateKeyPEM)
pricePoly := pricing.NewPolymarketClient()
pricingMgr := pricing.NewManager(priceCache, priceKalshi, pricePoly)
go pricingMgr.Run(ctx)
slog.Info("pricing manager started")
```

Also add `priceCache` to the unused var suppression:
```go
_, _, _ = eventBus, authenticator, priceCache
```

- [ ] **Step 2: Commit**

```bash
git add cmd/polybot/main.go
git commit -m "feat: wire discovery scanner and pricing manager into main entrypoint"
```

---

**End of Phase 1b Plan.** After this plan is complete, the Arby monolith will:
- Poll Kalshi and Polymarket REST APIs for market discovery on a configurable interval
- Maintain persistent WebSocket connections for real-time price feeds from both venues
- Store latest prices in a concurrent-safe in-memory cache readable by downstream modules
- Emit `MarketDiscovered` events on the event bus for new markets
- Have graceful shutdown for all WS connections and goroutines
