# PolyBot — Ingestion Module Design (Phase 1b)

> **Status:** Draft  
> **Date:** 2026-06-24  
> **Phase 1b goal:** Implement `pkg/ingestion` (discovery + pricing) for the Arby monolith, supporting both Kalshi and Polymarket from day one.

---

## 1. Architecture

```
pkg/ingestion/
├── discovery/          # REST polling for market metadata
│   ├── client.go       # DiscoveryClient interface
│   ├── kalshi.go       # Kalshi REST implementation
│   ├── polymarket.go   # Polymarket REST implementation
│   └── scanner.go      # Scheduler driving discovery cycles
├── pricing/            # Real-time price feeds
│   ├── client.go       # PricingClient interface
│   ├── kalshi.go       # Kalshi WS implementation
│   ├── polymarket.go   # Polymarket WS implementation
│   └── manager.go      # Connection lifecycle + reconnection
└── cache.go            # Shared in-memory price cache (sync.RWMutex)
```

### Module Boundary Rules

- `pkg/ingestion` depends on `internal/db` (PostgreSQL), `internal/redis` (hot price state), `internal/bus` (event emission), `internal/config`
- `pkg/ingestion` does NOT depend on any other `pkg/` module
- Other modules read prices from the shared `PriceCache` (read-only reference injected at startup)

---

## 2. Discovery Sub-Package

### Interface

```go
type DiscoveryClient interface {
    FetchMarkets(ctx context.Context) ([]Market, error)
    Venue() string // "KALSHI" or "POLYMARKET"
}
```

### Scanner

- Runs on a configurable interval (default 5 minutes)
- For each registered venue client:
  1. Call `FetchMarkets()`
  2. Diff result against PostgreSQL to find new/modified markets
  3. Upsert new markets to DB
  4. Emit `MarketDiscovered` event on the bus for each new market
  5. Emit `MarketModified` event for changed markets (title, dates, outcomes)
- Configurable via `IngestionConfig` block:
  ```go
  type IngestionConfig struct {
      DiscoveryInterval time.Duration // default: 5m
  }
  ```

### Kalshi REST Client

- Auth: RSA-PSS key pair (`KALSHI_KEY_ID`, `KALSHI_PRIVATE_KEY`)
- Endpoint: `https://api.elections.kalshi.com/trade-api/v2`
- Paginates through `/markets` endpoint with cursor-based pagination
- Normalizes fields:
  - `ticker` → parsed into `series` + `event` (first `-`-delimited segments)
  - Outcomes derived from `yes_bid/ask` and `no_bid/ask` → two outcomes
  - Category mapped from `sector` field

### Polymarket REST Client

- Auth: none required for discovery
- Endpoint: `https://clob.polymarket.com`
- Queries `/markets` and `/events` endpoints
- Normalizes fields:
  - `slug` → `Ticker` (may be null; use `condition_id` as fallback)
  - Outcomes from `outcomes` array (typically "Yes"/"No" for binaries)
  - Category from `category` field

### Canonical Market Struct

```go
type Market struct {
    Venue        string    // "KALSHI" or "POLYMARKET"
    MarketID     string    // venue-native ID
    Ticker       string    // Kalshi ticker / Polymarket slug
    Title        string
    Series       string    // Kalshi series grouping
    Category     string
    Outcomes     []Outcome // [{Name, Price}]
    OpenTime     time.Time
    CloseTime    time.Time
    Extra        json.RawMessage // venue-specific metadata
}

type Outcome struct {
    Name  string
    Price float64 // current yes bid price (0-1)
}
```

---

## 3. Pricing Sub-Package

### Interface

```go
type PricingClient interface {
    Subscribe(ctx context.Context, marketIDs []string) error
    Unsubscribe(marketIDs []string) error
    Close() error
    Updates() <-chan PriceTick
}
```

### Manager

- Owns one `PricingClient` per venue
- On startup: connect WS, subscribe to all known open markets
- Reads `PriceTick` from each client's `Updates()` channel, writes to `PriceCache`
- Reconnection: exponential backoff (1s, 2s, 4s, 8s, max 30s)
- On reconnect: re-subscribe all tracked markets
- Graceful shutdown: close all WS connections, drain channels

### PriceTick

```go
type PriceTick struct {
    Venue     string
    MarketID  string
    Bid       float64
    Ask       float64
    Timestamp time.Time
}
```

### Kalshi WS Client

- Auth: RSA-PSS signed JWT handshake
- URL: `wss://api.elections.kalshi.com/trade-api/ws`
- Subscribes to order book channels: `orderbook_{ticker}`
- Handles snapshots and deltas, produces `PriceTick` from best bid/ask

### Polymarket WS Client

- Auth: none required for public feeds
- URL: `wss://ws-subscriptions-clob.polymarket.com/ws/`
- Subscribes to `tickSize` channels
- Handles order book snapshots and deltas

---

## 4. Price Cache

```go
type PriceCache struct {
    mu      sync.RWMutex
    prices  map[string]PriceSnapshot // key: "venue:marketID"
}

type PriceSnapshot struct {
    Venue     string
    MarketID  string
    Bid       float64
    Ask       float64
    UpdatedAt time.Time
}

func (c *PriceCache) Get(venue, marketID string) (PriceSnapshot, bool)
func (c *PriceCache) Set(venue, marketID string, bid, ask float64)
```

- Read-lock for `Get` (~100ns)
- Write-lock for `Set` (~500ns)
- No expiry — consumers check staleness themselves
- Instantiated once in `main.go`, reference injected into pricing manager and shared with opportunity/strategy modules at startup

---

## 5. Events

| Event | Emitted By | Payload | Subscribers |
|-------|-----------|---------|-------------|
| `MarketDiscovered` | Discovery scanner | `{venue, marketID, ticker, title}` | Matching, Audit, Improvement, Reporting |
| `MarketModified` | Discovery scanner | `{venue, marketID, changedFields}` | Matching, Reporting |

---

## 6. Wiring in main.go

```go
// Discovery scanner
discScanner := discovery.NewScanner(
    discovery.NewKalshiClient(cfg.KalshiKeyID, cfg.KalshiPrivateKeyPEM),
    discovery.NewPolymarketClient(),
    pg.P(),           // *pgxpool.Pool for DB writes
    eventBus,
    cfg.Ingestion.DiscoveryInterval,
)
go discScanner.Run(ctx)

// Price cache
priceCache := ingestion.NewPriceCache()

// Pricing manager
pricingMgr := pricing.NewManager(
    priceCache,
    pricing.NewKalshiClient(cfg.KalshiKeyID, cfg.KalshiPrivateKeyPEM),
    pricing.NewPolymarketClient(),
)
go pricingMgr.Run(ctx)
```

---

## 7. Error Handling & Resilience

| Failure Mode | Handling |
|-------------|----------|
| REST API timeout (discovery) | Retry once after 5s; log warning, skip cycle on repeated failure |
| WS disconnect (pricing) | Exponential backoff reconnect; re-subscribe all tracked markets |
| DB write failure (discovery) | Log error, continue (markets will be retried on next cycle) |
| Invalid market data | Log malformed entry with venue+ID, skip that market, continue cycle |

---

## 8. Testing Strategy

| Layer | Approach |
|-------|----------|
| Venue clients | Unit tests with recorded/mocked HTTP and WS responses |
| Scanner | Unit test with in-memory DB mock; verify correct market diff and event emission |
| Price cache | Unit test concurrent read/write safety, correct snapshot semantics |
| Manager | Integration test with fake venue WS server; verify subscribe/reconnect/price flow |

---

## 9. Files to Create

```
pkg/ingestion/cache.go
pkg/ingestion/cache_test.go
pkg/ingestion/discovery/client.go
pkg/ingestion/discovery/kalshi.go
pkg/ingestion/discovery/kalshi_test.go
pkg/ingestion/discovery/polymarket.go
pkg/ingestion/discovery/polymarket_test.go
pkg/ingestion/discovery/scanner.go
pkg/ingestion/discovery/scanner_test.go
pkg/ingestion/pricing/client.go
pkg/ingestion/pricing/kalshi.go
pkg/ingestion/pricing/kalshi_test.go
pkg/ingestion/pricing/polymarket.go
pkg/ingestion/pricing/polymarket_test.go
pkg/ingestion/pricing/manager.go
pkg/ingestion/pricing/manager_test.go
```

Also modify:
- `internal/config/config.go` — add `IngestionConfig` block
- `cmd/polybot/main.go` — wire discovery scanner and pricing manager
