# Phase 1a: Monolith Infrastructure Skeleton — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Scaffold the Arby monolith Go project with all internal infrastructure (config, DB, Redis, logging, metrics, event bus, auth stubs) and a main entrypoint that starts up, connects dependencies, handles signals, and serves health checks.

**Branding:** "Arby" is the product name — marketed and sold as a standalone service, independent from Nostrabotus. `arby.nostrabotus.com` is the temporary v1 access URL. Custom domain to follow.

**Architecture:** Single Go binary with `internal/` packages for shared infrastructure and empty `pkg/` stubs for domain modules. In-process event bus replaces NATS. All service ports will be added in subsequent plans.

**Tech Stack:** Go 1.22, pgx/v5, rueidis (redis), slog (logging), prometheus/client_golang, go-playground/validator

**Access:** `https://arby.nostrabotus.com` (v1 — TLS via existing Let's Encrypt nginx on host)  
**Working directory:** `C:\Users\User\OneDrive\Desktop\Projects\Arby`

## Global Constraints

- Go 1.22 minimum (matching existing services)
- All secrets loaded from file paths specified in env vars (matching existing pattern)
- PostgreSQL via pgxpool, Redis via rueidis
- No third-party router — use Go 1.22 `http.ServeMux` with pattern routing
- Structured JSON logging via `log/slog`
- Config validation fail-fast at startup via `go-playground/validator`
- All health checks on `GET /healthz` (liveness) and `GET /readyz` (readiness)
- Metrics on `GET /metrics` via Prometheus client_golang
- In-process event bus uses `chan`-based pub/sub
- Graceful shutdown: SIGTERM/SIGINT → stop HTTP → flush → close connections → exit

---

### Task 1: Go Module Init + Project Directory Structure

**Files:**
- Create: `go.mod`
- Create: `cmd/polybot/main.go` (stub)
- Create: `.gitignore`
- Create: directories for `internal/`, `pkg/`, `cmd/`

- [ ] **Step 1: Initialize Go module and create directory structure**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go mod init github.com/aaronbateman02/Arby
```

- [ ] **Step 2: Create directory tree**

```bash
mkdir -p cmd/polybot
mkdir -p internal/config
mkdir -p internal/db
mkdir -p internal/redis
mkdir -p internal/logging
mkdir -p internal/metrics
mkdir -p internal/bus
mkdir -p internal/auth
mkdir -p internal/health
mkdir -p pkg/ingestion
mkdir -p pkg/matching
mkdir -p pkg/strategy
mkdir -p pkg/opportunity
mkdir -p pkg/execution
mkdir -p pkg/selling
mkdir -p pkg/risk
mkdir -p pkg/reporting
mkdir -p pkg/audit
mkdir -p pkg/improvement
```

- [ ] **Step 3: Create .gitignore**

Write to `.gitignore`:
```
# Binaries
/polybot
*.exe
*.test

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store
Thumbs.db

# Secrets
/secrets/
.env
*.pem

# Data
/data/

# Dependencies
/vendor/
```

- [ ] **Step 4: Create stub main.go**

Write to `cmd/polybot/main.go`:
```go
package main

import (
    "context"
    "log/slog"
    "os"
    "os/signal"
    "syscall"
)

func main() {
    logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
    slog.SetDefault(logger)
    slog.Info("starting arby")

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
    <-sigCh

    slog.Info("shutting down")
    cancel()
}
```

- [ ] **Step 5: Verify it compiles**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go build ./cmd/polybot
```

Expected: binary `polybot.exe` created, no errors

- [ ] **Step 6: Commit**

```bash
git add .
git commit -m "feat: scaffold Go module and project directory structure"
```

---

### Task 2: Config Package — Validated Config Loading

**Files:**
- Create: `internal/config/config.go`
- Create: `internal/config/config_test.go`

**Interfaces:**
- Consumes: nothing
- Produces: `config.Config` struct with `Validate()` method

- [ ] **Step 1: Write the failing test**

Write to `internal/config/config_test.go`:
```go
package config_test

import (
    "testing"
    "github.com/aaronbateman02/Arby/internal/config"
)

func TestConfig_ValidMinimal(t *testing.T) {
    cfg := config.Config{
        PostgresDSN: "postgres://user:pass@localhost:5432/arby",
        RedisAddr:   "localhost:6379",
        ListenAddr:  ":8086",
    }
    if err := cfg.Validate(); err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
}

func TestConfig_MissingPostgresDSN(t *testing.T) {
    cfg := config.Config{
        RedisAddr:  "localhost:6379",
        ListenAddr: ":8086",
    }
    if err := cfg.Validate(); err == nil {
        t.Fatal("expected validation error for missing PostgresDSN")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/config/
```

Expected: FAIL — package does not exist

- [ ] **Step 3: Write minimal implementation**

Add `github.com/go-playground/validator/v10`:
```bash
go get github.com/go-playground/validator/v10
```

Write to `internal/config/config.go`:
```go
package config

import (
    "fmt"
    "os"
    "strings"
    "github.com/go-playground/validator/v10"
)

type Config struct {
    PostgresDSN     string `validate:"required"`
    RedisAddr       string `validate:"required"`
    ListenAddr      string `validate:"required"`
    LogLevel        string `validate:"omitempty,oneof=debug info warn error"`
    OpenRouterAPIKey string
    OpenRouterBaseURL string
    KalshiKeyID     string
    KalshiPrivateKeyPEM string
    PolymarketProxyURL string
    JWTSecretKey    string
}

func (c *Config) Validate() error {
    v := validator.New()
    return v.Struct(c)
}

func LoadFromEnv() (*Config, error) {
    cfg := &Config{
        PostgresDSN:      readSecretOr(os.Getenv("POSTGRES_DSN_FILE")),
        RedisAddr:        envOrDefault("REDIS_ADDR", "localhost:6379"),
        ListenAddr:       envOrDefault("LISTEN_ADDR", ":8086"),
        LogLevel:         envOrDefault("LOG_LEVEL", "info"),
        OpenRouterAPIKey: readSecretOr(os.Getenv("OPENROUTER_API_KEY_FILE")),
        OpenRouterBaseURL: envOrDefault("OPENROUTER_BASE_URL", "https://openrouter.ai/api/v1"),
        KalshiKeyID:      readSecretOr(os.Getenv("KALSHI_KEY_ID_FILE")),
        KalshiPrivateKeyPEM: readSecretOr(os.Getenv("KALSHI_PRIVATE_KEY_FILE")),
        PolymarketProxyURL: os.Getenv("POLYMARKET_PROXY_URL"),
        JWTSecretKey:     readSecretOr(os.Getenv("JWT_SECRET_KEY_FILE")),
    }

    if err := cfg.Validate(); err != nil {
        return nil, fmt.Errorf("config validation failed: %w", err)
    }

    return cfg, nil
}

func readSecretOr(path string) string {
    if path == "" {
        return ""
    }
    data, err := os.ReadFile(path)
    if err != nil {
        return ""
    }
    return strings.TrimSpace(string(data))
}

func envOrDefault(key, def string) string {
    if v := os.Getenv(key); v != "" {
        return v
    }
    return def
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/config/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: add validated config package with env/file loading"
```

---

### Task 3: Database Package — pgxpool Wrapper

**Files:**
- Create: `internal/db/db.go`
- Create: `internal/db/db_test.go`

**Interfaces:**
- Consumes: `config.Config.PostgresDSN`
- Produces: `db.Pool` (wraps `*pgxpool.Pool`), `db.HealthCheck()`

- [ ] **Step 1: Write the failing test**

Write to `internal/db/db_test.go`:
```go
package db_test

import (
    "context"
    "testing"
    "github.com/aaronbateman02/Arby/internal/db"
)

func TestConnect_InvalidDSN(t *testing.T) {
    pool, err := db.Connect(context.Background(), "invalid-dsn")
    if err == nil {
        pool.Close()
        t.Fatal("expected error for invalid DSN")
    }
}

func TestHealthCheck_NotConnected(t *testing.T) {
    pool := &db.Pool{}
    err := pool.HealthCheck(context.Background())
    if err == nil {
        t.Fatal("expected error when pool is nil")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/db/
```

Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```bash
go get github.com/jackc/pgx/v5/pgxpool
```

Write to `internal/db/db.go`:
```go
package db

import (
    "context"
    "fmt"
    "time"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {
    p *pgxpool.Pool
}

func Connect(ctx context.Context, dsn string) (*Pool, error) {
    cfg, err := pgxpool.ParseConfig(dsn)
    if err != nil {
        return nil, fmt.Errorf("parse dsn: %w", err)
    }

    cfg.MaxConns = 10
    cfg.MinConns = 2
    cfg.MaxConnLifetime = 30 * time.Minute
    cfg.MaxConnIdleTime = 5 * time.Minute
    cfg.HealthCheckInterval = 30 * time.Second

    p, err := pgxpool.NewWithConfig(ctx, cfg)
    if err != nil {
        return nil, fmt.Errorf("create pool: %w", err)
    }

    if err := p.Ping(ctx); err != nil {
        p.Close()
        return nil, fmt.Errorf("ping: %w", err)
    }

    return &Pool{p: p}, nil
}

func (pool *Pool) HealthCheck(ctx context.Context) error {
    if pool == nil || pool.p == nil {
        return fmt.Errorf("pool not initialized")
    }
    return pool.p.Ping(ctx)
}

func (pool *Pool) Close() {
    if pool != nil && pool.p != nil {
        pool.p.Close()
    }
}

func (pool *Pool) P() *pgxpool.Pool {
    return pool.p
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/db/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: add database package with pgxpool wrapper"
```

---

### Task 4: Redis Package — Client Wrapper

**Files:**
- Create: `internal/redis/redis.go`
- Create: `internal/redis/redis_test.go`

**Interfaces:**
- Consumes: `config.Config.RedisAddr`
- Produces: `redis.Client` wrapper, `redis.HealthCheck()`

- [ ] **Step 1: Write the failing test**

Write to `internal/redis/redis_test.go`:
```go
package redis_test

import (
    "context"
    "testing"
    "github.com/aaronbateman02/Arby/internal/redis"
)

func TestConnect_InvalidAddr(t *testing.T) {
    client, err := redis.Connect(context.Background(), "")
    if err == nil {
        client.Close()
        t.Fatal("expected error for empty addr")
    }
}

func TestHealthCheck_NotConnected(t *testing.T) {
    client := &redis.Client{}
    err := client.HealthCheck(context.Background())
    if err == nil {
        t.Fatal("expected error when client is nil")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/redis/
```

Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```bash
go get github.com/redis/rueidis
```

Write to `internal/redis/redis.go`:
```go
package redis

import (
    "context"
    "fmt"
    "time"
    "github.com/redis/rueidis"
)

type Client struct {
    c rueidis.Client
}

func Connect(ctx context.Context, addr string) (*Client, error) {
    if addr == "" {
        return nil, fmt.Errorf("redis addr is required")
    }

    c, err := rueidis.NewClient(rueidis.ClientOption{
        Addr: addr,
    })
    if err != nil {
        return nil, fmt.Errorf("create client: %w", err)
    }

    if err := c.Do(ctx, c.B().Ping().Build()).Error(); err != nil {
        c.Close()
        return nil, fmt.Errorf("ping: %w", err)
    }

    return &Client{c: c}, nil
}

func (c *Client) HealthCheck(ctx context.Context) error {
    if c == nil || c.c == nil {
        return fmt.Errorf("redis client not initialized")
    }
    return c.Do(ctx, c.B().Ping().Build()).Error()
}

func (c *Client) Do(ctx context.Context, cmd rueidis.Completed) rueidis.RedisResult {
    return c.c.Do(ctx, cmd)
}

func (c *Client) Close() {
    if c != nil && c.c != nil {
        c.c.Close()
    }
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/redis/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: add redis package with rueidis client wrapper"
```

---

### Task 5: Logging Package — Structured Slog Setup

**Files:**
- Create: `internal/logging/logging.go`
- Create: `internal/logging/logging_test.go`

**Interfaces:**
- Consumes: `config.Config.LogLevel`
- Produces: configured `*slog.Logger`

- [ ] **Step 1: Write the failing test**

Write to `internal/logging/logging_test.go`:
```go
package logging_test

import (
    "testing"
    "github.com/aaronbateman02/Arby/internal/logging"
)

func TestNew_ValidLevel(t *testing.T) {
    logger, err := logging.New("debug")
    if err != nil {
        t.Fatalf("expected no error, got: %v", err)
    }
    if logger == nil {
        t.Fatal("expected non-nil logger")
    }
}

func TestNew_InvalidLevel(t *testing.T) {
    _, err := logging.New("invalid")
    if err == nil {
        t.Fatal("expected error for invalid log level")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/logging/
```

Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

Write to `internal/logging/logging.go`:
```go
package logging

import (
    "fmt"
    "io"
    "log/slog"
    "os"
)

func New(level string) (*slog.Logger, error) {
    var l slog.Level
    switch level {
    case "debug":
        l = slog.LevelDebug
    case "info":
        l = slog.LevelInfo
    case "warn":
        l = slog.LevelWarn
    case "error":
        l = slog.LevelError
    default:
        return nil, fmt.Errorf("invalid log level: %s", level)
    }

    return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level: l,
    })), nil
}

func NewWriter(logger *slog.Logger, level slog.Level) io.Writer {
    r, w := io.Pipe()
    go func() {
        buf := make([]byte, 4096)
        for {
            n, err := r.Read(buf)
            if err != nil {
                return
            }
            logger.LogAttrs(nil, level, string(buf[:n]))
        }
    }()
    return w
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/logging/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: add structured logging package with slog"
```

---

### Task 6: Metrics Package — Prometheus Registration

**Files:**
- Create: `internal/metrics/metrics.go`
- Create: `internal/metrics/metrics_test.go`

**Interfaces:**
- Consumes: nothing
- Produces: Prometheus metrics registry, HTTP handler for `/metrics`

- [ ] **Step 1: Write the failing test**

Write to `internal/metrics/metrics_test.go`:
```go
package metrics_test

import (
    "testing"
    "github.com/aaronbateman02/Arby/internal/metrics"
)

func TestRegisterAndGet(t *testing.T) {
    m := metrics.New()
    if m == nil {
        t.Fatal("expected non-nil metrics")
    }

    httpRequests := m.Counter("http_requests_total", "Total HTTP requests", "method", "path")
    if httpRequests == nil {
        t.Fatal("expected non-nil counter")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/metrics/
```

Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```bash
go get github.com/prometheus/client_golang/prometheus
go get github.com/prometheus/client_golang/prometheus/promhttp
```

Write to `internal/metrics/metrics.go`:
```go
package metrics

import (
    "net/http"
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
    reg *prometheus.Registry
    factory promauto.Factory
}

func New() *Metrics {
    reg := prometheus.NewRegistry()
    return &Metrics{
        reg:     reg,
        factory: promauto.With(reg),
    }
}

func (m *Metrics) Counter(name, help string, labels ...string) prometheus.Counter {
    return m.factory.NewCounter(prometheus.CounterOpts{
        Name: name,
        Help: help,
    })
}

func (m *Metrics) CounterVec(name, help string, labelNames []string) *prometheus.CounterVec {
    return m.factory.NewCounterVec(prometheus.CounterOpts{
        Name: name,
        Help: help,
    }, labelNames)
}

func (m *Metrics) Histogram(name, help string, buckets []float64) prometheus.Histogram {
    return m.factory.NewHistogram(prometheus.HistogramOpts{
        Name:    name,
        Help:    help,
        Buckets: buckets,
    })
}

func (m *Metrics) Gauge(name, help string) prometheus.Gauge {
    return m.factory.NewGauge(prometheus.GaugeOpts{
        Name: name,
        Help: help,
    })
}

func (m *Metrics) Handler() http.Handler {
    return promhttp.HandlerFor(m.reg, promhttp.HandlerOpts{})
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/metrics/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: add prometheus metrics package with registry"
```

---

### Task 7: In-Process Event Bus — Chan-Based Pub/Sub

**Files:**
- Create: `internal/bus/bus.go`
- Create: `internal/bus/bus_test.go`

**Interfaces:**
- Consumes: nothing
- Produces: `bus.Bus` — `Publish()`, `Subscribe()`, `Unsubscribe()`

- [ ] **Step 1: Write the failing test**

Write to `internal/bus/bus_test.go`:
```go
package bus_test

import (
    "testing"
    "time"
    "github.com/aaronbateman02/Arby/internal/bus"
)

func TestPublishSubscribe(t *testing.T) {
    b := bus.New(10)

    sub, err := b.Subscribe("test.event")
    if err != nil {
        t.Fatalf("subscribe failed: %v", err)
    }
    defer b.Unsubscribe("test.event", sub)

    payload := []byte("hello")
    b.Publish("test.event", payload)

    select {
    case msg := <-sub:
        if string(msg.Payload) != "hello" {
            t.Fatalf("expected 'hello', got '%s'", string(msg.Payload))
        }
        if msg.Topic != "test.event" {
            t.Fatalf("expected topic 'test.event', got '%s'", msg.Topic)
        }
    case <-time.After(time.Second):
        t.Fatal("timeout waiting for message")
    }
}

func TestUnsubscribe_StopsDelivery(t *testing.T) {
    b := bus.New(10)

    sub, _ := b.Subscribe("test.event")
    b.Unsubscribe("test.event", sub)

    b.Publish("test.event", []byte("should-not-receive"))

    select {
    case <-sub:
        t.Fatal("should not receive message after unsubscribe")
    case <-time.After(100 * time.Millisecond):
        // OK — no message received
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/bus/
```

Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

Write to `internal/bus/bus.go`:
```go
package bus

import (
    "fmt"
    "sync"
)

type Message struct {
    Topic   string
    Payload []byte
}

type Bus struct {
    mu          sync.RWMutex
    subscribers map[string]map[int]chan Message
    nextID      int
}

func New(channelSize int) *Bus {
    return &Bus{
        subscribers: make(map[string]map[int]chan Message),
    }
}

func (b *Bus) Subscribe(topic string) (chan Message, error) {
    b.mu.Lock()
    defer b.mu.Unlock()

    if b.subscribers[topic] == nil {
        b.subscribers[topic] = make(map[int]chan Message)
    }

    b.nextID++
    ch := make(chan Message, 100)
    b.subscribers[topic][b.nextID] = ch

    return ch, nil
}

func (b *Bus) Unsubscribe(topic string, ch chan Message) {
    b.mu.Lock()
    defer b.mu.Unlock()

    if subs, ok := b.subscribers[topic]; ok {
        for id, subCh := range subs {
            if subCh == ch {
                close(subCh)
                delete(subs, id)
                return
            }
        }
    }
}

func (b *Bus) Publish(topic string, payload []byte) {
    b.mu.RLock()
    subs := b.subscribers[topic]
    b.mu.RUnlock()

    if subs == nil {
        return
    }

    msg := Message{Topic: topic, Payload: payload}

    b.mu.RLock()
    for _, ch := range subs {
        select {
        case ch <- msg:
        default:
            // drop if channel full — backpressure strategy
        }
    }
    b.mu.RUnlock()
}

func (b *Bus) PublishTyped(topic string, v interface{}) error {
    data, err := json.Marshal(v)
    if err != nil {
        return fmt.Errorf("marshal event %s: %w", topic, err)
    }
    b.Publish(topic, data)
    return nil
}
```

Update imports in `bus.go` to include `encoding/json`.

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/bus/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: add in-process event bus with chan-based pub/sub"
```

---

### Task 8: Auth Package — JWT + RBAC Middleware

**Files:**
- Create: `internal/auth/auth.go`
- Create: `internal/auth/auth_test.go`

**Interfaces:**
- Consumes: `config.Config.JWTSecretKey`
- Produces: `auth.Authenticator`, `auth.Middleware()`

- [ ] **Step 1: Write the failing test**

Write to `internal/auth/auth_test.go`:
```go
package auth_test

import (
    "testing"
    "time"
    "github.com/aaronbateman02/Arby/internal/auth"
)

func TestGenerateAndValidate(t *testing.T) {
    a, err := auth.New("test-secret-key-32-bytes-long!!")
    if err != nil {
        t.Fatalf("new auth failed: %v", err)
    }

    token, err := a.GenerateToken("user-1", auth.RoleOperator, 15*time.Minute)
    if err != nil {
        t.Fatalf("generate token failed: %v", err)
    }

    claims, err := a.ValidateToken(token)
    if err != nil {
        t.Fatalf("validate token failed: %v", err)
    }

    if claims.UserID != "user-1" {
        t.Fatalf("expected user-1, got %s", claims.UserID)
    }
    if claims.Role != auth.RoleOperator {
        t.Fatalf("expected operator role, got %s", claims.Role)
    }
}

func TestInvalidToken(t *testing.T) {
    a, _ := auth.New("test-secret-key-32-bytes-long!!")
    _, err := a.ValidateToken("invalid-token")
    if err == nil {
        t.Fatal("expected error for invalid token")
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/auth/
```

Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

```bash
go get github.com/golang-jwt/jwt/v5
```

Write to `internal/auth/auth.go`:
```go
package auth

import (
    "context"
    "fmt"
    "net/http"
    "strings"
    "time"
    "github.com/golang-jwt/jwt/v5"
)

type Role string

const (
    RoleAdmin    Role = "admin"
    RoleOperator Role = "operator"
    RoleViewer   Role = "viewer"
)

type Claims struct {
    UserID string `json:"user_id"`
    Role   Role   `json:"role"`
    jwt.RegisteredClaims
}

type Authenticator struct {
    secret []byte
}

func New(secret string) (*Authenticator, error) {
    if len(secret) < 16 {
        return nil, fmt.Errorf("secret must be at least 16 characters")
    }
    return &Authenticator{secret: []byte(secret)}, nil
}

func (a *Authenticator) GenerateToken(userID string, role Role, ttl time.Duration) (string, error) {
    claims := &Claims{
        UserID: userID,
        Role:   role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString(a.secret)
}

func (a *Authenticator) ValidateToken(tokenStr string) (*Claims, error) {
    token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
        if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
        }
        return a.secret, nil
    })
    if err != nil {
        return nil, fmt.Errorf("parse token: %w", err)
    }

    claims, ok := token.Claims.(*Claims)
    if !ok || !token.Valid {
        return nil, fmt.Errorf("invalid token claims")
    }

    return claims, nil
}

type contextKey string

const claimsKey contextKey = "auth_claims"

func ContextWithClaims(ctx context.Context, claims *Claims) context.Context {
    return context.WithValue(ctx, claimsKey, claims)
}

func ClaimsFromContext(ctx context.Context) *Claims {
    claims, _ := ctx.Value(claimsKey).(*Claims)
    return claims
}

type Middleware struct {
    auth    *Authenticator
    allowed []Role
}

func NewMiddleware(auth *Authenticator, allowedRoles ...Role) *Middleware {
    return &Middleware{auth: auth, allowed: allowedRoles}
}

func (m *Middleware) Wrap(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        header := r.Header.Get("Authorization")
        if header == "" {
            http.Error(w, "missing authorization header", http.StatusUnauthorized)
            return
        }

        tokenStr := strings.TrimPrefix(header, "Bearer ")
        if tokenStr == header {
            http.Error(w, "invalid authorization format", http.StatusUnauthorized)
            return
        }

        claims, err := m.auth.ValidateToken(tokenStr)
        if err != nil {
            http.Error(w, "invalid token", http.StatusUnauthorized)
            return
        }

        if len(m.allowed) > 0 {
            allowed := false
            for _, role := range m.allowed {
                if claims.Role == role {
                    allowed = true
                    break
                }
            }
            if !allowed {
                http.Error(w, "insufficient permissions", http.StatusForbidden)
                return
            }
        }

        r = r.WithContext(ContextWithClaims(r.Context(), claims))
        next.ServeHTTP(w, r)
    })
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/auth/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: add auth package with JWT generation, validation, RBAC middleware"
```

---

### Task 9: Health Package — Health Check Endpoints

**Files:**
- Create: `internal/health/health.go`
- Create: `internal/health/health_test.go`

**Interfaces:**
- Consumes: DB and Redis health check functions
- Produces: HTTP handlers for `/healthz` and `/readyz`

- [ ] **Step 1: Write the failing test**

Write to `internal/health/health_test.go`:
```go
package health_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/aaronbateman02/Arby/internal/health"
)

func TestHealthz_AlwaysOK(t *testing.T) {
    h := health.New(nil, nil)
    handler := h.LivenessHandler()

    req := httptest.NewRequest("GET", "/healthz", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rec.Code)
    }
}

func TestReadyz_AllHealthy(t *testing.T) {
    dbOK := func(ctx context.Context) error { return nil }
    redisOK := func(ctx context.Context) error { return nil }

    h := health.New(dbOK, redisOK)
    handler := h.ReadinessHandler()

    req := httptest.NewRequest("GET", "/readyz", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rec.Code)
    }
}

func TestReadyz_DBFailure(t *testing.T) {
    dbFail := func(ctx context.Context) error { return nil }
    redisFail := func(ctx context.Context) error { return nil }

    h := health.New(dbFail, redisFail)
    handler := h.ReadinessHandler()

    req := httptest.NewRequest("GET", "/readyz", nil)
    rec := httptest.NewRecorder()
    handler.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rec.Code)
    }
}
```

- [ ] **Step 2: Run test to verify it fails**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/health/
```

Expected: FAIL

- [ ] **Step 3: Write minimal implementation**

Write to `internal/health/health.go`:
```go
package health

import (
    "context"
    "encoding/json"
    "net/http"
)

type Check func(ctx context.Context) error

type Health struct {
    dbCheck   Check
    redisCheck Check
}

func New(dbCheck, redisCheck Check) *Health {
    return &Health{
        dbCheck:   dbCheck,
        redisCheck: redisCheck,
    }
}

func (h *Health) LivenessHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(`{"status":"alive"}`))
    }
}

func (h *Health) ReadinessHandler() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        status := http.StatusOK
        checks := map[string]string{}

        if h.dbCheck != nil {
            if err := h.dbCheck(r.Context()); err != nil {
                checks["database"] = err.Error()
                status = http.StatusServiceUnavailable
            } else {
                checks["database"] = "ok"
            }
        }

        if h.redisCheck != nil {
            if err := h.redisCheck(r.Context()); err != nil {
                checks["redis"] = err.Error()
                status = http.StatusServiceUnavailable
            } else {
                checks["redis"] = "ok"
            }
        }

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(status)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status": status,
            "checks": checks,
        })
    }
}
```

- [ ] **Step 4: Run tests to verify they pass**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./internal/health/ -v
```

Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add .
git commit -m "feat: add health check package with liveness and readiness endpoints"
```

---

### Task 10: Main Entrypoint — Wiring Everything Together

**Files:**
- Modify: `cmd/polybot/main.go`

**Interfaces:**
- Consumes: all internal packages
- Produces: running binary with HTTP server, signal handling, graceful shutdown

- [ ] **Step 1: Write the main.go**

Write to `cmd/polybot/main.go`:
```go
package main

import (
    "context"
    "errors"
    "log/slog"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/aaronbateman02/Arby/internal/auth"
    "github.com/aaronbateman02/Arby/internal/bus"
    "github.com/aaronbateman02/Arby/internal/config"
    "github.com/aaronbateman02/Arby/internal/db"
    "github.com/aaronbateman02/Arby/internal/health"
    "github.com/aaronbateman02/Arby/internal/logging"
    "github.com/aaronbateman02/Arby/internal/metrics"
    arbyredis "github.com/aaronbateman02/Arby/internal/redis"
)

func main() {
    // 1. Load config
    cfg, err := config.LoadFromEnv()
    if err != nil {
        slog.Error("config", "error", err)
        os.Exit(1)
    }

    // 2. Setup logger
    logger, err := logging.New(cfg.LogLevel)
    if err != nil {
        slog.Error("logging", "error", err)
        os.Exit(1)
    }
    slog.SetDefault(logger)
    slog.Info("starting arby", "listen", cfg.ListenAddr)

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // 3. Connect PostgreSQL
    pg, err := db.Connect(ctx, cfg.PostgresDSN)
    if err != nil {
        slog.Error("postgres", "error", err)
        os.Exit(1)
    }
    defer pg.Close()
    slog.Info("postgres connected")

    // 4. Connect Redis
    rdb, err := arbyredis.Connect(ctx, cfg.RedisAddr)
    if err != nil {
        slog.Error("redis", "error", err)
        os.Exit(1)
    }
    defer rdb.Close()
    slog.Info("redis connected")

    // 5. Init metrics
    met := metrics.New()
    slog.Info("metrics initialized")

    // 6. Init event bus
    eventBus := bus.New(1000)
    slog.Info("event bus initialized")

    // 7. Init auth
    authenticator, err := auth.New(cfg.JWTSecretKey)
    if err != nil {
        slog.Error("auth", "error", err)
        os.Exit(1)
    }
    slog.Info("auth initialized")

    // 8. Health checks
    h := health.New(
        func(ctx context.Context) error { return pg.HealthCheck(ctx) },
        func(ctx context.Context) error { return rdb.HealthCheck(ctx) },
    )

    // 9. Setup HTTP server
    mux := http.NewServeMux()
    mux.HandleFunc("GET /healthz", h.LivenessHandler())
    mux.HandleFunc("GET /readyz", h.ReadinessHandler())
    mux.Handle("GET /metrics", met.Handler())

    srv := &http.Server{
        Addr:    cfg.ListenAddr,
        Handler: mux,
    }

    // 10. Start HTTP in background
    go func() {
        slog.Info("http server listening", "addr", cfg.ListenAddr)
        if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
            slog.Error("http server", "error", err)
            os.Exit(1)
        }
    }()

    // 11. Wait for signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    sig := <-quit
    slog.Info("shutdown signal received", "signal", sig)

    // 12. Graceful shutdown
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()

    if err := srv.Shutdown(shutdownCtx); err != nil {
        slog.Error("http server shutdown", "error", err)
    }

    cancel() // cancel main context
    pg.Close()
    rdb.Close()
    slog.Info("shutdown complete")
}
```

- [ ] **Step 2: Build and verify**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go mod tidy
go build ./cmd/polybot
```

Expected: binary `polybot.exe` created, no errors

- [ ] **Step 3: Commit**

```bash
git add .
git commit -m "feat: wire up main entrypoint with config, DB, Redis, HTTP, signal handling"
```

---

### Task 11: Dockerfile + Docker Compose + Makefile

**Files:**
- Create: `Dockerfile`
- Create: `docker-compose.yml`
- Create: `Makefile`
- Create: `nginx/arby.conf`

- [ ] **Step 1: Create Dockerfile**

Write to `Dockerfile`:
```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /polybot ./cmd/polybot

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /polybot /polybot
EXPOSE 8086
ENTRYPOINT ["/polybot"]
```

- [ ] **Step 2: Create docker-compose.yml**

Write to `docker-compose.yml`:
```yaml
services:
  polybot:
    build: .
    container_name: arby-polybot
    restart: unless-stopped
    env_file: [".env"]
    ports: ["8087:8086"]
    depends_on:
      postgres: { condition: service_healthy }
      redis:    { condition: service_healthy }

  postgres:
    image: pgvector/pgvector:pg16
    container_name: arby-postgres
    restart: unless-stopped
    shm_size: 1g
    environment:
      POSTGRES_DB: arby
      POSTGRES_USER: arby
      POSTGRES_PASSWORD_FILE: /run/secrets/postgres_password
    volumes:
      - arby_postgres_data:/var/lib/postgresql/data
    ports: ["5435:5432"]
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U arby -d arby"]
      interval: 10s
      timeout: 5s
      retries: 5

  redis:
    image: redis:7-alpine
    container_name: arby-redis
    restart: unless-stopped
    command: redis-server --appendonly yes --appendfsync everysec --maxmemory 256mb --maxmemory-policy allkeys-lru
    volumes:
      - arby_redis_data:/data
    ports: ["6381:6379"]
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 10s
      timeout: 3s
      retries: 3

volumes:
  arby_postgres_data:
  arby_redis_data:

secrets:
  postgres_password:
    file: ./secrets/postgres_password.txt
```

- [ ] **Step 3: Create nginx config**

Write to `nginx/arby.conf`:
```nginx
# Arby nginx config — mounted into the host's existing nginx container
# (polybot-nginx) or used standalone. Ports below are the host-mapped
# ports from docker-compose.yml (not container-internal ports).
#
# Branding note: "Arby" is the product name, independent from Nostrabotus.
# arby.nostrabotus.com is a temporary v1 access URL.

server {
    listen 80;
    server_name arby.nostrabotus.com;

    location /api/ {
        proxy_pass http://127.0.0.1:8087/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /ws/ {
        proxy_pass http://127.0.0.1:8087/;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }

    location / {
        root /usr/share/nginx/html/arby;
        try_files $uri $uri/ /index.html;
    }
}
```

- [ ] **Step 4: Create Makefile**

Write to `Makefile`:
```makefile
.PHONY: build test vet tidy docker-build docker-up docker-down

build:
	go build -o polybot ./cmd/polybot

test:
	go test ./... -v -count=1

vet:
	go vet ./...

tidy:
	go mod tidy

docker-build:
	docker compose build

docker-up:
	docker compose -p arby up -d

docker-down:
	docker compose -p arby down

docker-logs:
	docker compose -p arby logs -f
```

- [ ] **Step 5: Verify build still works**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go mod tidy
go build ./cmd/polybot
```

Expected: OK

- [ ] **Step 6: Commit**

```bash
git add .
git commit -m "feat: add Dockerfile, docker-compose, nginx config, and Makefile"
```

---

### Task 12: Integration Test — Startup Sequence

**Files:**
- Create: `cmd/polybot/main_test.go`

- [ ] **Step 1: Write integration test**

Write to `cmd/polybot/main_test.go`:
```go
package main

import (
    "context"
    "net/http"
    "testing"
    "time"

    "github.com/aaronbateman02/Arby/internal/config"
    "github.com/aaronbateman02/Arby/internal/db"
    "github.com/aaronbateman02/Arby/internal/redis"
)

func TestMainStartup(t *testing.T) {
    // Skip if no DSN provided
    dsn := getEnvOrDefault("TEST_POSTGRES_DSN", "")
    if dsn == "" {
        t.Skip("TEST_POSTGRES_DSN not set")
    }

    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    pg, err := db.Connect(ctx, dsn)
    if err != nil {
        t.Fatalf("db connect: %v", err)
    }
    defer pg.Close()

    if err := pg.HealthCheck(ctx); err != nil {
        t.Fatalf("db health: %v", err)
    }
}

func TestHealthEndpoint(t *testing.T) {
    resp, err := http.Get("http://localhost:8087/healthz")
    if err != nil {
        t.Skipf("server not running: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Fatalf("expected 200, got %d", resp.StatusCode)
    }
}

func getEnvOrDefault(key, def string) string {
    if v := getenv(key); v != "" {
        return v
    }
    return def
}
```

- [ ] **Step 2: Run unit tests**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
go test ./... -v -count=1 -short
```

Expected: All tests pass (with skips for integration tests without DSN)

- [ ] **Step 3: Commit**

```bash
git add .
git commit -m "test: add integration test stubs for startup sequence"
```

---

### Task 13: GitHub Remote Setup + Push

- [ ] **Step 1: Add remote and push**

```bash
cd C:\Users\User\OneDrive\Desktop\Projects\Arby
git remote add origin https://github.com/aaronbateman02/Arby.git
git push -u origin master
```

Expected: pushed successfully to GitHub

---

**End of Phase 1a Plan.** After this plan is complete, the Arby monolith will:
- Compile into a single binary (~20MB)
- Start up, load validated config, connect to PostgreSQL and Redis
- Serve `/healthz`, `/readyz`, `/metrics` HTTP endpoints
- Have an in-process event bus ready for service wiring
- Have auth middleware (JWT + RBAC) ready for service HTTP handlers
- Be deployable via Docker Compose with isolated ports/volumes alongside the existing Polybot stack

Subsequent plans (Phase 1b-1g) will port each existing Go service into the `pkg/` modules.
