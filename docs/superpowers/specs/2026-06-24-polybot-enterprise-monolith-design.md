# PolyBot Enterprise Modular Monolith — Design

> **Status:** Draft  
> **Date:** 2026-06-24  
> **Drivers:** Latency reduction (primary), deployment simplicity, production hardening, multi-user support, recursive self-improvement via comprehensive auditing  
> **Infrastructure:** New t3.medium EC2 in isolated VPC, new domain, clean break from existing deployment

---

## 1. Architecture Decision

**Decision:** Consolidate the existing ~10 microservices into a single Go binary following the **modular monolith** pattern. The Python matching engine's LLM review pipeline is ported to Go. GPU-heavy embedding remains a standalone local Python tool.

### 1.1 What Changes

| Aspect | Before (Microservices) | After (Monolith) |
|--------|----------------------|------------------|
| Deployable units | 10+ Docker containers | 1 Go binary + PostgreSQL + Redis + nginx + UI |
| Inter-service comms | NATS JetStream (serialize/deserialize every message) | In-process channels + direct function calls |
| Language runtime | Go + Python + TypeScript | Go (binary) + Python (local embedding tool only) |
| Deployment | `docker compose --profile services up -d` (12 containers) | `docker compose up -d` (5 containers) |
| Critical path latency | ~800µs+ coordination overhead | ~150ns coordination overhead |

### 1.2 What Stays the Same

- PostgreSQL (+ pgvector) as the transactional DB and vector store
- Redis as the hot state / price cache
- Ireland proxy node for Polymarket order relay
- Embedding pipeline (`embed_worker.py` on local GPU machine)
- UI (Next.js, light theme, Kalshi/Polymarket-inspired design)
- nginx as TLS terminator and reverse proxy

---

## 2. Module Architecture

```
polybot/
├── cmd/polybot/             # Main binary entrypoint
│   └── main.go              # Config load, service wiring, signal handling
├── internal/
│   ├── bus/                 # In-process event bus (chan-based pub/sub)
│   ├── config/              # Config loading + validation (fail-fast)
│   ├── db/                  # PostgreSQL connection pool (pgx)
│   ├── redis/               # Redis client (rueidis)
│   ├── logging/             # Structured JSON logger (slog)
│   ├── metrics/             # Prometheus metrics registration
│   └── auth/                # JWT auth, RBAC middleware
├── pkg/
│   ├── ingestion/           # Market data — WS + REST for Kalshi & Polymarket
│   │   ├── discovery/       # REST market discovery (periodic)
│   │   └── pricing/         # Real-time pricing (WS streams)
│   ├── matching/            # Semantic market matching
│   │   ├── candidate/       # Cosine-similarity candidate generation (pgvector)
│   │   ├── llm/             # OpenRouter-based LLM review pipeline
│   │   │   ├── first_leg.go # Deepseek-V4-Flash batch review
│   │   │   └── second_leg.go# Gemini validation
│   │   └── matchers/        # Domain-specific sub-matchers (politics, sports, events)
│   ├── strategy/            # Strategy engine (plugin-based)
│   │   └── arb/             # Cross-market arbitrage strategy
│   ├── opportunity/         # Real-time gatekeeper — ROI re-check, capital reservation
│   ├── execution/           # Order placement
│   │   ├── kalshi/          # Kalshi REST API client (RSA-PSS auth)
│   │   └── polymarket/      # Polymarket CLOB client (via Ireland proxy)
│   ├── selling/             # Early-exit opportunity monitoring
│   ├── risk/                # Circuit breakers, kill switches, capital gating
│   ├── reporting/           # REST + WebSocket API for UI
│   │   ├── handler/         # HTTP handlers
│   │   └── ws/              # WebSocket push for real-time dashboard
│   ├── audit/               # Granular decision audit — every input, output, context
│   └── improvement/         # Self-improvement engine — analysis + proposal generation
└── vendor/                  # Go vendored dependencies
```

### 2.1 Module Dependency Rules

- `pkg/*` modules may depend on `internal/*` infrastructure
- `pkg/*` modules may **not** depend on other `pkg/*` modules directly — they communicate through the `internal/bus` event bus or through the opportunity → execution critical path interface
- The critical path (`pkg/opportunity` → `pkg/risk` → `pkg/execution`) uses direct Go interface calls defined in `pkg/execution` and `pkg/risk`

### 2.2 Hot Path Interface

```go
// pkg/risk/risk.go
type RiskService interface {
    ReserveCapital(ctx context.Context, bundle Bundle) (ReservationID, error)
    ReleaseCapital(ctx context.Context, reservationID ReservationID) error
    CheckCircuitBreaker(ctx context.Context, strategyID string) error
}

// pkg/execution/execution.go
type ExecutionService interface {
    PlaceOrders(ctx context.Context, bundle Bundle) (OrderResult, error)
    CancelOrders(ctx context.Context, bundle Bundle) error
}
```

These interfaces are wired at startup in `cmd/polybot/main.go`. The Opportunity module receives concrete implementations — no NATS, no serialization, no goroutine switch.

---

## 3. In-Process Event Bus

For non-critical-path events (market discovered, match approved, bundle resolved, etc.), the monolith uses an in-process pub/sub bus:

| Concern | Design |
|---------|--------|
| Pattern | `chan`-based with typed event structs |
| Delivery | Best-effort async (goroutine per subscriber + buffered channel) |
| Backpressure | Configurable channel depth; drop oldest when full (logged) |
| Crash safety | **Not used for state-critical events** — those use direct calls with DB writes |

**Event types:**

| Event | Publisher | Subscribers |
|-------|-----------|-------------|
| `MarketDiscovered` | Ingestion | Matching, Audit, Improvement, Reporting |
| `MatchApproved` | Matching | Strategy, Audit, Improvement, Reporting |
| `OpportunityFound` | Strategy | Opportunity, Audit, Improvement, Reporting |
| `BundleExecuted` | Execution | Selling, Risk, Audit, Improvement, Reporting |
| `BundleResolved` | Selling | Risk, Audit, Improvement, Reporting |
| `CircuitBreakerTripped` | Risk | Audit, Improvement, Reporting (alert UI) |
| `ConfigChanged` | Improvement (human) | All modules (reload), Audit, Reporting |

---

## 4. Latency-Critical Path Detail

```
WS price tick
    │
    ▼
pkg/ingestion/pricing goroutine
    │  updates shared in-memory price cache (sync.Map)
    │  (~50ns write)
    ▼
pkg/opportunity goroutine (per-watched-bundle)
    │  reads price from cache (~50ns read)
    │  computes bundle ROI
    │  (pure CPU, ~1-5µs)
    │
    ├── ROI < min_roi: sleep until next tick
    │
    └── ROI ≥ min_roi:
         │
         ▼
         risk.ReserveCapital(ctx, bundle)
         │  DB transaction: INSERT reservation row
         │  check circuit breaker state (in-memory, cached)
         │  (~1-5ms)
         │
         ├── error: alert, back to waiting
         │
         └── OK:
              │
              ▼
              execution.PlaceOrders(ctx, bundle)
              │  HTTP POST to Kalshi API
              │  HTTP POST to Ireland proxy → Polymarket CLOB
              │  (~10-100ms)
              │
              ├── both filled → bundle → COMPLETE
              │
              └── partial fill → strategy policy:
                   ├── UNWIND → CancelOrders + ReleaseCapital
                   └── HOLD_AND_GTC → bundle → PENDING_COMPLETION
```

### 4.1 Per-Bundle Goroutine Model

Each watched bundle gets its own goroutine managed by `pkg/opportunity`. This avoids shared-state contention — each goroutine owns exactly one bundle's lifecycle and reads prices from the shared read-mostly cache.

### 4.2 Price Cache Design

```go
// Shared price cache — updated by ingestion, read by opportunity/strategy
type PriceCache struct {
    mu    sync.RWMutex
    prices map[string]PriceSnapshot  // marketID → best bid/ask
}

// Update (ingestion goroutine): write-lock, ~500ns
// Read (opportunity goroutine): read-lock, ~100ns
```

---

## 5. Matching Engine (LLM Review Pipeline)

Port from Python to Go. Uses OpenRouter as the unified API provider.

| Step | Model | Purpose |
|------|-------|---------|
| First leg (batch) | `deepseek-v4-flash` via OpenRouter | Review N candidate pairs, produce structured JSON with confidence + mapping |
| Second leg (validation) | `gemini-2.5-flash` via OpenRouter | Validate low-confidence results (< 0.90) from first leg |

**OpenRouter client:**

```go
type OpenRouterClient struct {
    apiKey string
    baseURL string  // https://openrouter.ai/api/v1
    httpClient *http.Client
}

func (c *OpenRouterClient) ReviewBatch(ctx context.Context, candidates []Candidate, model string) ([]ReviewResult, error)
func (c *OpenRouterClient) Validate(ctx context.Context, candidate Candidate, model string) (ReviewResult, error)
```

**Scheduling:** `internal/cron` package (or a simple `time.Ticker` loop) replaces APScheduler. Configurable intervals per stage:
- Embedding availability check: every 1 min
- Candidate review: every 5 min
- Low-confidence validation: immediately after first leg results

---

## 6. Multi-User & Auth

| Feature | Detail |
|---------|--------|
| User store | `users` table (PostgreSQL), bcrypt password hash |
| Auth flow | `POST /api/auth/login` → JWT access (15min) + refresh (7d) |
| JWT signing | Ed25519 key pair, loaded from file at startup |
| RBAC | `admin` (all), `operator` (trade + review), `viewer` (read-only) |
| Middleware | `internal/auth/middleware.go` — extracts JWT, sets `ctx` with user+role |
| Audit | See Section 12 — comprehensive decision-level audit, separate from auth logging |

**Roles and permissions:**

| Action | admin | operator | viewer |
|--------|-------|----------|--------|
| View dashboard | ✅ | ✅ | ✅ |
| Approve/reject matches | ✅ | ✅ | ❌ |
| Pause/resume strategy | ✅ | ✅ | ❌ |
| Flip paper/live toggle | ✅ | ✅ | ❌ |
| Manage users | ✅ | ❌ | ❌ |
| View audit log | ✅ | ✅ | ❌ |
| Review improvement proposals | ✅ | ✅ | ❌ |
| Approve/deny improvement proposals | ✅ | ✅ | ❌ |

---

## 7. Production Hardening

### 7.1 Graceful Shutdown

```
SIGTERM/SIGINT received
    │
    ▼
1. Stop accepting new HTTP requests (shutdown http.Server)
2. Signal ingestion goroutines to close WS connections
3. Wait for in-flight execution to complete (configurable timeout, default 30s)
4. Flush pending DB writes
5. Cancel all remaining contexts
6. Close DB pool, Redis client
7. Exit (os.Exit(0))
```

### 7.2 Health Checks

| Endpoint | What it checks |
|----------|---------------|
| `GET /healthz` | Liveness — process is running (always 200) |
| `GET /readyz` | Readiness — DB ping, Redis ping, Ireland proxy ping |

Both are used by Docker HEALTHCHECK and (if behind a load balancer) by the ALB target group.

### 7.3 Metrics (Prometheus)

| Metric | Type | Labels |
|--------|------|--------|
| `polybot_http_requests_total` | Counter | `method`, `path`, `status` |
| `polybot_http_request_duration_ms` | Histogram | `method`, `path` (buckets: 1, 5, 25, 100, 500) |
| `polybot_goroutines_active` | Gauge | none |
| `polybot_db_connections_acquired` | Counter | `pool` |
| `polybot_db_connection_wait_ms` | Histogram | `pool` |
| `polybot_execution_latency_ms` | Histogram | `venue` (buckets: 10, 50, 200, 1000, 5000) |
| `polybot_opportunity_roi_rechecks` | Counter | `strategy_id` |
| `polybot_circuit_breaker_state` | Gauge | `strategy_id` (1=active, 0=paused) |

### 7.4 Rate Limiting

- Token-bucket per IP for all HTTP API endpoints
- Default: 200 req/s burst, 100 req/s sustained
- Configurable via config file
- `/healthz` and `/readyz` are exempt

### 7.5 Config Validation

Single config struct validated at startup:

```go
type Config struct {
    HTTP           HTTPConfig           `validate:"required"`
    DB             DBConfig             `validate:"required"`
    Redis          RedisConfig          `validate:"required"`
    Auth           AuthConfig           `validate:"required"`
    OpenRouter     OpenRouterConfig     `validate:"required"`
    Kalshi         KalshiConfig         `validate:"required"`
    PolymarketProxy PolymarketProxyConfig `validate:"required"`
    Strategy       map[string]StrategyConfig `validate:"required,dive"`
    LogLevel       string               `validate:"oneof=debug info warn error"`
}

type DBConfig struct {
    DSN                string `validate:"required"`
    MaxConns           int    `validate:"min=1,max=100"`
    MinConns           int    `validate:"min=0"`
    MaxConnLifetime    time.Duration `validate:"min=1m"`
    HealthCheckInterval time.Duration `validate:"min=1s"`
}
```

Validation uses `go-playground/validator`. Any missing or invalid config causes the process to exit immediately with a descriptive error message — never start with partial config.

---

## 8. Deployment

### 8.1 New EC2 Instance (Monolith Host)

The monolith deploys to a **new** EC2 instance — fully isolated from the existing `polybot.nostrabotus.com` infrastructure:

| Property | Value |
|----------|-------|
| Instance type | `t3.medium` (2 vCPU, 4GB RAM) |
| Location | US AWS region (same as existing — `us-east-2`) |
| Network | **New VPC** — no network overlap with current deployment |
| Domain | **New domain** (TBD — provisioned alongside new VPC) |
| SSH alias | TBD |
| OS | Amazon Linux 2023 |
| Orchestration | Docker Compose (same as existing pattern) |

**Rationale for new instance:** Avoids port conflicts, dependency collisions, and configuration drift during migration. Enables clean side-by-side validation: the old microservice stack continues running while the monolith is tested against fresh infrastructure. Cutover is a DNS swap.

### 8.2 Docker Compose (Monolith Stack)

```yaml
services:
  polybot:
    build: .
    container_name: polybot
    restart: unless-stopped
    env_file: [".env"]
    ports: ["8086:8086"]   # HTTP API + WS
    depends_on:
      postgres: { condition: service_healthy }
      redis:    { condition: service_healthy }

  postgres:    # pgvector/pgvector:pg16
  redis:       # redis:7-alpine
  nginx:       # nginx:1.27-alpine (serves UI, proxies /api to polybot)
  ui-v2:       # Next.js
```

### 8.3 Dockerfile (Multi-stage Go build)

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -o /polybot ./cmd/polybot

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=builder /polybot /polybot
EXPOSE 8086
ENTRYPOINT ["/polybot"]
```

Binary size target: ~20-30MB (statically linked, no libc dependency).

### 8.4 Ireland Proxy

Unchanged from current deployment. The monolith routes Polymarket orders through the existing Ireland proxy via HTTP POST. Proxy URL configured via `POLYMARKET_PROXY_URL` env var in the monolith. The Ireland proxy's security group must be updated to allow ingress from the new VPC's CIDR range.

### 8.5 Local Embedding Tool

Unchanged. `embed_worker.py` runs on the user's GPU machine, connects directly to the **new** PostgreSQL instance (pgvector), and writes embeddings. The monolith's `pkg/matching/candidate` module queries already-embedded markets for cosine similarity. The user's machine must have network access to the new VPC's PostgreSQL endpoint.

### 8.6 Data Migration — Full Knowledge Preservation

All accumulated market intelligence is migrated to the new PostgreSQL instance. Nothing is rebuilt from scratch unless necessary.

1. **Schema:** Run `flyway migrate` against the new PostgreSQL instance — same migrations as existing DB
2. **All markets (~186k):** Export `markets` table (both Kalshi and Polymarket) via `pg_dump` / `pg_restore` — preserves all metadata, categories, slugs, tickers
3. **Embeddings (~21k):** Export vector columns directly — `pg_dump` handles `vector(1024)` columns natively. Avoids re-embedding 21k markets. Only outstanding markets (those not yet embedded) will be picked up by `embed_worker.py` against the new DB
4. **Match candidates (1,177):** Export `match_candidates` — preserves all cosine-similarity pairs already found
5. **Match pairs (157 APPROVED):** Export `match_pairs` — all human-approved and LLM-approved matches are carried over intact
6. **LLM validations:** Export `llm_validations` — preserves raw LLM responses, confidence scores, and evaluation history
7. **Strategy configs:** Export `strategy_configs` — all strategy parameter profiles (min_roi, shares_per_leg, partial_fill_policy, circuit_breaker settings)
8. **Risk state:** Export `risk_state` — circuit breaker states, capital allocation snapshots
9. **Categorization:** Export `market_categories` tables — preserves category mappings and any manual overrides
10. **What we DO skip:** `bundles`, `positions`, `ledger` — existing live trades stay on the old system. The monolith starts with a clean slate for trade execution. All market knowledge (what exists, what matches, what's been validated) is preserved.

---

## 9. UI Requirements

- **Stack:** Next.js (existing `services/ui-v2`)
- **Theme:** Light theme (white background, dark text, blue accent — inspired by Kalshi)
- **Design influence:** Kalshi (clean card-based layouts, clear typography, straightforward navigation) and Polymarket (event-card grid, probability display)
- **Real-time:** WebSocket connection to monolith's `GET /api/ws` for live dashboard updates
- **Auth:** Login page → JWT stored in `httpOnly` cookie → auto-refresh

**Required views (unchanged from CONTEXT.md):**

| View | Priority | Notes |
|------|----------|-------|
| Capital Overview | P1 | Total bankroll, allocated, unallocated, % deployed |
| Open Bundles | P1 | Per-bundle: status, venue, entry cost, current ROI, early-exit flag |
| Strategy Performance | P1 | Per-strategy: opportunities, executed, win rate, ROI, circuit breaker status |
| Historical P&L | P1 | Resolved bundles over time |
| Execution Latency | P1 | p50/p95/p99 from opportunity → order submitted |
| Match Review Queue | P1 | Approve/reject candidate pairs with LLM reasoning |
| Service Status | P1 | Paper/live toggles, active/paused states |
| Alert Log | P1 | In-app alerts (circuit breaker trips, execution failures) |
| Bundle Detail | P1 | Drill-down: legs, fill prices, state transitions, audit trail |
| Login | P1 | JWT auth, role-based view control |
| Improvement Proposals | P1 | Browse pending/approved/denied proposals; review evidence; approve/deny with comment |
| Decision Audit Log | P1 | Chronological view of all system decisions with full input/context drill-down |
| Outcome Correlation | P2 | Visual charts linking config changes → outcome improvements over time |

---

## 10. Migration Path

### Phase 1: Monolith Skeleton + Go Service Consolidation
1. Create the monolith project structure (`cmd/polybot`, `internal/*`, `pkg/*` stubs)
2. Port each Go service one at a time, preserving behavior
3. Wire up the in-process event bus
4. Deploy to the new EC2 instance in its isolated VPC
5. Validate ingesting real market data without affecting the live system

### Phase 2: Matching Engine Port
1. Port the LLM review pipeline from Python to Go (`pkg/matching/llm`)
2. Implement OpenRouter client (Deepseek-V4-Flash + Gemini)
3. Port sub-matchers (politics, sports_binary, world_events)
4. Port pre_filters and candidate generation
5. Dual-run: compare Go LLM output vs Python LLM output for same candidate pairs

### Phase 3: Auth + Audit + Production Hardening
1. Multi-user auth (JWT, RBAC)
2. Comprehensive audit data model (see Section 12)
3. Graceful shutdown, health checks, rate limiting
4. Prometheus metrics
5. Config validation

### Phase 4: Improvement Engine
1. Implement `pkg/improvement` — analysis engine that reads audit data
2. Build proposal workflow (generate → store → notify → human review → apply)
3. Wire improvement proposals into UI for review

### Phase 5: UI Refresh
1. Light theme redesign
2. Login flow
3. Improvement proposal review view
4. Auth-aware role-based views

### Phase 6: Data Migration + Cutover
1. Export/import schema and market data from old PostgreSQL to new
2. Run `embed_worker.py` against new DB for fresh embeddings
3. Deploy monolith with paper mode on new EC2
4. Validate paper-mode output against live system behavior
5. Flip execution to live on new EC2
6. Update DNS to point to new domain
7. Old infrastructure remains available for rollback

---

## 11. Key Risks & Mitigations

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| Matching LLM port misses prompt nuances | Medium | Dual-run Phase 2 — compare Go LLM output vs Python LLM output side-by-side before switching |
| Monolith OOM under load | Low | Go memory model is predictable; set `GOMEMLIMIT`; profile before prod |
| Single point of failure | Medium | Monolith restarts in <1s (binary); DB holds all state; Redis is ephemeral cache |
| Porting bugs in execution pipeline | High | Paper mode runs alongside existing microservices — compare order decisions before live cutover |
| UI redesign scope creep | Medium | Ship auth-hardened monolith first, UI refresh last |
| Improvement engine proposes bad changes | Medium | Human-in-the-loop approval required; changes are proposed, never auto-applied |
| Over-audit impacts DB performance | Low | Audit writes are async (buffered channel); retention policy auto-archives old records |

---

## 12. Recursive Self-Improvement

### 12.1 What It Means

"Recursive self-improvement" means the system continuously analyzes its own decisions and outcomes, identifies patterns, and proposes improvements. Each improvement makes the system better at identifying the *next* improvement. The recursion is:

```
Cycle N: make decisions → observe outcomes → analyze → propose improvements
         ↓
Cycle N+1: improved system makes better decisions → better outcomes → better analysis → better proposals
         ↓
Cycle N+2: compounding effect
```

The system does **not** apply changes autonomously. All improvement proposals require human review and approval. The human remains in control; the system provides the analysis.

### 12.2 Comprehensive Audit Data Model

Every decision point in the system is captured with full context. This is the fuel for the improvement engine.

**Schema (new `audit_decisions` table):**

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Primary key |
| `decision_type` | TEXT | One of: `match_approval`, `opportunity_fire`, `order_placement`, `order_fill`, `early_exit`, `circuit_breaker_trip`, `capital_reservation` |
| `decision_id` | UUID | Unique ID for this specific decision (links related audit rows) |
| `strategy_id` | TEXT | Which strategy was involved (if applicable) |
| `bundle_id` | UUID | Which bundle (if applicable) |
| `user_id` | UUID | Who or what made the decision (operator UUID for manual, `system:matching`, `system:opportunity`, etc.) |
| `inputs` | JSONB | All inputs to the decision at decision time |
| `outputs` | JSONB | The decision output (what happened) |
| `context` | JSONB | Ambient state at decision time (see below) |
| `outcome_id` | UUID | FK to `audit_outcomes` — populated when outcome is known |
| `created_at` | TIMESTAMPTZ | When the decision was made |

**`inputs` captures (per decision_type):**

| decision_type | inputs includes |
|---------------|----------------|
| `match_approval` | LLM raw response, confidence score, candidate pair IDs, market metadata snapshots, matcher name |
| `opportunity_fire` | Bundle ROI calculation, each leg's price, fee estimate, min_roi threshold, capital available, circuit breaker states |
| `order_placement` | Venue, market ID, side, quantity, limit price, order type, fill policy |
| `order_fill` | Venue order ID, fill price, fill quantity, latency from fire to fill |
| `early_exit` | Current exit ROI, min_exit_roi threshold, bundle age, each leg's current price |
| `circuit_breaker_trip` | Metric name, current value, threshold, rolling window size |
| `capital_reservation` | Requested amount, unallocated balance, strategy limit, result (reserved/rejected) |

**`context` captures (ambient state at decision time):**

| Context field | Description |
|---------------|-------------|
| `config_snapshot` | Active strategy configs, risk limits, min_roi values at this moment |
| `price_snapshots` | Current bid/ask for all relevant markets |
| `open_bundles_count` | How many bundles were open at this moment |
| `capital_state` | Total bankroll, allocated, unallocated |
| `circuit_breaker_states` | Current state of all circuit breakers |
| `active_proposals` | Any pending improvement proposals |
| `venue_status` | Any known venue API issues or latency anomalies |

**`audit_outcomes` table** (populated asynchronously when outcomes materialize):

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Primary key |
| `decision_id` | UUID | Links back to the original decision |
| `outcome_type` | TEXT | `bundle_resolved`, `order_settled`, `match_validated` |
| `result` | TEXT | `profit`, `loss`, `breakeven`, `cancelled`, `valid`, `invalid` |
| `roi` | NUMERIC | Realized ROI (nullable) |
| `pnl` | NUMERIC | Realized P&L in USD (nullable) |
| `details` | JSONB | Additional outcome context |
| `resolved_at` | TIMESTAMPTZ | When the outcome was known |

**Retention policy:** Raw audit data kept for 90 days. Aggregated/analyzed results retained indefinitely. Automatic archival of raw data older than 90 days to compressed storage.

### 12.3 Improvement Engine Architecture

```
audit_decisions ──► audit_outcomes
       │                    │
       ▼                    ▼
┌──────────────────────────────────┐
│      pkg/improvement/engine      │
│                                  │
│  1. Outcome Analyzer              │
│     - Links decisions → outcomes  │
│     - Computes ROI attribution    │
│     - Stratifies by strategy/     │
│       match type/venue/etc.       │
│                                  │
│  2. Pattern Detector              │
│     - Statistical significance    │
│     - Correlation analysis        │
│     - Anomaly detection           │
│                                  │
│  3. Proposal Generator            │
│     - Translates patterns into    │
│       concrete config changes     │
│     - Estimates impact            │
│     - Generates human-readable    │
│       rationale                   │
│                                  │
│  4. Proposal Store                │
│     - Writes to                   │
│       improvement_proposals table │
│     - Triggers UI notification    │
└──────────────┬───────────────────┘
               │
               ▼
┌──────────────────────────────────┐
│   improvement_proposals table     │
│  (status: pending/approved/denied)│
└──────────────┬───────────────────┘
               │
               ▼  (human reviews in UI)
┌──────────────────────────────────┐
│      On human approval:           │
│  - Update strategy_configs        │
│  - Update risk_limits             │
│  - Update matching thresholds     │
│  - Emit ConfigChanged event       │
│  - Log to audit_decisions         │
│    (decision_type: config_change) │
└──────────────────────────────────┘
```

### 12.4 What the Improvement Engine Analyzes

| Dimension | Questions Analyzed | Potential Proposals |
|-----------|-------------------|-------------------|
| **Strategy ROI thresholds** | Are we leaving money on the table? Are we firing too early/late? | Increase/decrease `min_roi` per strategy |
| **Match quality** | Which matchers produce the most profitable pairs? Which categories have high false-positive rates? | Tune confidence thresholds per matcher, adjust LLM prompts |
| **Execution timing** | Does latency correlate with fill quality? Should we fire Polymarket or Kalshi first? | Adjust leg ordering, adjust slippage model |
| **Circuit breaker tuning** | Are circuit breakers triggering too early (leaving profit) or too late (taking losses)? | Adjust threshold values, rolling window sizes |
| **Category focus** | Which market categories consistently produce the best risk-adjusted returns? | Shift attention to higher-performing categories |
| **Fee model accuracy** | Are fee estimates matching actual fees? | Adjust fee model parameters per venue |
| **LLM prompt effectiveness** | Which prompt variants produce the most accurate match assessments? | Propose prompt updates (requires human review of diffs) |

### 12.5 Proposal Workflow

```
1. Engine runs (configurable schedule: daily, weekly)
   │
   ▼
2. Analyzes audit_decisions + audit_outcomes
   │  aggregated by strategy_id, matcher_id, category, etc.
   ▼
3. Detects statistically significant patterns
   │  (e.g., "strategy X min_roi=0.03 passed on 40% of profitable opportunities")
   ▼
4. Generates proposal(s)
   │  Each proposal contains:
   │   - Title: "Increase arb-politics min_roi from 0.03 to 0.035"
   │   - Current value, proposed value, delta
   │   - Evidence: "Over 30 days, 40% of sub-0.035 opportunities were profitable"
   │   - Estimated impact: "+$X expected gain, Y% false-positive reduction"
   │   - Supporting data: links to audit_decisions used in analysis
   │   - Risk level: low/medium/high
   ▼
5. Proposal stored in improvement_proposals table
   │  status: pending
   ▼
6. UI shows pending proposals in "Improvement Proposals" panel
   │  Operator reviews evidence, approves or denies
   ▼
7. If approved:
   - Config is updated in DB
   - ConfigChanged event emitted
   - Affected modules re-read their config
   - Audit trail: decision_type=config_change recorded
   -
   If denied:
   - Proposal marked denied with operator comment
   - Engine learns: similar proposals deprioritized or adjusted
```

### 12.6 Improvement Proposal Schema (`improvement_proposals` table)

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Primary key |
| `title` | TEXT | Human-readable summary |
| `target_type` | TEXT | `strategy_config`, `risk_limit`, `match_threshold`, `llm_prompt`, `execution_param` |
| `target_id` | TEXT | Which strategy/config row/sub-matcher this applies to |
| `field` | TEXT | The specific config field to change (e.g., `min_roi`) |
| `current_value` | JSONB | Current value of the field |
| `proposed_value` | JSONB | Proposed new value |
| `evidence` | JSONB | Supporting data: audit query results, statistical significance, sample sizes |
| `estimated_impact` | JSONB | Expected effect on ROI/false-positives/latency |
| `risk_level` | TEXT | `low`, `medium`, `high` |
| `status` | TEXT | `pending`, `approved`, `denied`, `superseded` |
| `reviewed_by` | UUID | Operator who reviewed |
| `reviewed_at` | TIMESTAMPTZ | When reviewed |
| `review_comment` | TEXT | Optional operator feedback |
| `engine_version` | TEXT | Version of the improvement engine that generated this |
| `created_at` | TIMESTAMPTZ | When proposed |

### 12.7 Recursion in Practice

The recursion manifests as compounding analysis quality:

- **Cycle 1:** Engine analyzes simple ROI correlations → proposes threshold changes → human approves → system makes better trades
- **Cycle 2:** With better data from improved trades, engine can detect finer patterns → proposes match quality improvements → human approves → matches are more accurate
- **Cycle 3:** More accurate matches → more profitable trades → richer audit data → engine can now analyze interaction effects (e.g., "match quality × execution timing × fee model")
- **Cycle N+1:** Each cycle builds on data generated by the improved system from the previous cycle

The engine **does not** modify its own analysis code — that remains a human task. It proposes parameter changes within the existing system's configurable surface.

### 12.8 Safety Mechanisms

| Mechanism | Description |
|-----------|-------------|
| Human approval gate | No change is ever auto-applied |
| Risk level classification | High-risk proposals require extra scrutiny |
| Proposal superseding | If a proposal is pending and a better one is generated, the old one is marked `superseded` |
| Rollback support | All config changes are logged; reverting is a one-click operation |
| Sample size floor | Engine requires minimum N observations before proposing a change (configurable, default 50) |
| Statistical significance | Proposals are only generated when p ≤ 0.05 (configurable) |
| Cooldown period | After any config change, a minimum cool-down period (configurable, default 7 days) before the same field can be proposed again |
