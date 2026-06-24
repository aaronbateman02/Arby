# PolyBot — Architecture Context & Decision Log

> **This file must be read at the start of every session and updated after every meaningful change.**
> It is the internal source of truth for all architectural decisions, open questions, and constraints.

---

## Session Log

| Date | Change |
|---|---|
| 2026-05-14 | Initial architecture scoping session. README and CONTEXT bootstrapped. |
| 2026-05-14 | Capital model and bankroll constraints decided. |
| 2026-05-14 | Per-strategy circuit breaker model decided. |
| 2026-05-14 | Partial-fill execution policy decided. |
| 2026-05-14 | Order sizing model decided. |
| 2026-05-14 | Slippage / entry gate model decided. |
| 2026-05-14 | Matching Engine plugin model and approval policy decided. |
| 2026-05-14 | Matching Engine semantic model decided (embeddings + LLM). |
| 2026-05-14 | V1 strategy decided: cross-market arbitrage. Build sequence confirmed. |
| 2026-05-14 | Paper mode as a first-class platform-wide principle decided. |
| 2026-05-15 | Selling Service early-exit policy decided. |
| 2026-05-15 | Alerting model decided: in-app only for v1. |
| 2026-05-15 | Infrastructure topology decided. |
| 2026-05-15 | EC2 specs, orchestration model, and domain confirmed. |
| 2026-05-15 | Recovery and restart model decided. UI dashboard requirements decided. Foundational Q&A complete. |
| 2026-05-15 | Phase 1 build: migrations, docker-compose, scripts, service skeletons, nginx, UI scaffold |
| 2026-05-15 | Phase 2 build: Market Ingestion Service fully implemented (Kalshi WS+REST, Polymarket WS+REST, RSA-PSS auth, NATS, Redis, PostgreSQL) |
| 2026-05-15 | Phase 2 build: Matching Engine host.py + sports-binary sub-matcher fully implemented (pgvector embeddings, GPT-4o structured output) |
| 2026-05-15 | Phase 2 build: Opportunity Service fully implemented (ROI gate, capital reservation req/reply, TTL expiry, tick-based monitoring) |
| 2026-05-15 | Phase 2 build: Risk Service fully implemented (capital reservation, circuit breakers, operator HTTP API) |
| 2026-05-15 | Phase 2 build: Execution Service fully implemented (Polymarket-first leg ordering, Kalshi RSA-PSS order placement, paper mode simulation, mTLS proxy) |
| 2026-05-15 | Phase 2 build: Selling Service fully implemented (early-exit ROI monitoring, market resolution handling, paper mode logging) |
| 2026-05-15 | Recovery and restart model decided. |
| 2026-05-15 | UI dashboard requirements decided. Foundational Q&A complete. |
| 2026-05-15 | Phase 1 build: migrations/001_initial_schema.sql, migrations/002_seed_data.sql |
| 2026-05-15 | Phase 1 build: docker-compose.yml (US control plane), docker-compose.proxy.yml (non-US) |
| 2026-05-15 | Phase 1 build: .gitignore, scripts/inject-secrets.sh, scripts/deploy.sh |
| 2026-05-15 | Phase 1 build: all service skeletons created (ingestion, opportunity, execution, selling, risk, reporting, proxy, matching, arb-cross-market strategy) |
| 2026-05-15 | Phase 1 build: nginx/nginx.conf, services/ui scaffold (React + Vite + TypeScript) |
| 2026-05-19 | All services deployed and running on EC2 (t3.large). Matching pipeline live. |
| 2026-05-19 | match_candidates table bulk-populated with 1,177 pairs (cosine similarity ≥ 0.80, pair_enabled categories). |
| 2026-05-19 | Gemini model updated from deprecated gemini-2.0-flash to gemini-2.5-flash in all 3 sub-matchers. |
| 2026-05-19 | Matching pipeline confirmed working: first review cycle produced 157 APPROVED match pairs. |
| 2026-05-19 | Kalshi market URL format fixed in UI (MatchReview.tsx): /markets/{series}/{event} pattern. |
| 2026-05-19 | LLM pipeline: gemini-2.5-flash primary, gpt-4o-mini fallback (RateLimitError) + second opinion (confidence < 0.90). |
| 2026-05-19 | Embed worker (BAAI/bge-large-en-v1.5) incrementally growing Kalshi embeddings (~19,889/183,900 so far). |


| 2026-05-21 | Fixed: batch result count mismatch — `raise ValueError` replaced with `log.warning` in 3 matchers; partial results now used, missing pairs retry next cycle. (commit fbe6e6c) |
| 2026-05-21 | Fixed: `entertainment` and `economics` categories had no matcher entry — now mapped to `world_events_review_batch`. (commit fbe6e6c) |
| 2026-05-21 | Fixed: index-based batch result matching — `MatchReview.pair_index` added to schema; results matched by index instead of array position, preventing cross-pair contamination when Gemini skips entries. (commit 6c0b4a6) |
| 2026-05-21 | Fixed: sports SYSTEM_PROMPT INVERSE rule — "eliminated in Final" vs "wins tournament" now explicitly UNRELATED/is_same_event=false even at the Final stage. (commit 180d361) |
| 2026-05-21 | Cleaned: 6 REJECTED match_pairs and 8 false-positive llm_validations deleted; affected markets will be re-evaluated by the fixed pipeline. |
| 2026-05-22 | Feat: Store raw LLM response text in llm_validations.raw_response_text; add model_evaluations table for cross-model replay; evaluate_model.py script to test any model against stored prompts and compare agreement rate. (commit ec90795) |
| 2026-05-22 | Fix: pre_filters.py — removed 14-day resolution date window for non-sports categories; was blocking 94.7% of candidates (34k/36k pairs). Sports keeps 14-day window. First run queued 33,465 pairs for LLM review. (commit 33e4371) |
| 2026-05-21 | Feat: LLM Rejected review tab added to MatchReview UI — browse all 22k+ AI-rejected candidates with search/category filter and ⚡ Re-evaluate button. New API: GET /api/matching/llm-rejections. (commit f537c6b) |
| 2026-06-24 | Phase 1a: Arby monolith skeleton — Go module, config, DB, Redis, logging, metrics, event bus, auth, health, main entrypoint, Docker, tests. (commit be585e2) |
| 2026-06-24 | Phase 1b: Ingestion module — config, price cache, Kalshi/Polymarket discovery REST clients, discovery scanner with event bus, Kalshi/Polymarket WS pricing clients, pricing manager with cache fan-out. (commit eb19dbd) |
---

## Decided — Platform Objective

**Decision:** The platform has no single global objective function. Objective functions are **strategy-local**.

- Each sub-strategy independently defines what it optimizes for.
- Examples: a Bitcoin-outcome speculative strategy may target high expected return at high risk; a cross-market arbitrage strategy may target guaranteed or near-guaranteed low-risk returns.
- The platform's job is to: faithfully execute each strategy's stated intent, enforce risk guardrails, report outcomes per-strategy.
- There is no global profit/IRR mandate at the platform layer.

**Implication:** The Reporting Service must support per-strategy performance attribution. The Risk & Guardrails Service must support per-strategy limit profiles.

---

## Decided — Service Topology

Eight services (see README table). Risk & Guardrails is a first-class service, not bolted on.

Key boundary decisions:
- **Market Ingestion** is separate from **Matching Engine**. Ingestion normalizes raw data; Matching does semantic reasoning.
- **Strategy Service** consumes canonical mappings from Matching Engine and live prices from Ingestion. It does not do its own semantic matching.
- **Opportunity Service** is a real-time gatekeeper — it enforces the actual entry conditions rather than having Strategy do so. Strategy declares intent; Opportunity enforces timing.
- **Execution Service** has two geographically separated agents aligned to venue access constraints.

---

## Decided — Canonical Outcome Graph

Every matched pair is represented as a graph:
- **Nodes:** Event, Market, Outcome, Leg
- **Edges:** `equivalent_to`, `inverse_of`, `subset_of`, `mutually_exclusive_with`
- Every edge carries: `confidence_score`, `evidence[]`, `source`, `created_at`, `last_validated_at`

This structure handles:
- Simple Yes/No mirrors between venues
- Inverted outcomes (Kalshi "Yes" = Polymarket "No")
- Multi-choice markets where one venue has a binary leg and the other has a multi-option market
- Naming drift over time (entity aliasing)

---

## Decided — Strategy Plugin Architecture

Sub-strategies are:
- Isolated (changes to one cannot affect another)
- Independently deployable (separate container/process preferred)
- Independently togglable (on/off at runtime without restart)
- Independently configurable (each has its own settings store)
- Independently risk-profiled (each gets its own limit parameters in Risk Service)

A strategy emits a **candidate opportunity** containing:
- Required legs (venue, market ID, outcome/leg, direction)
- Target entry conditions (max price per leg, total bundle cost ceiling)
- Minimum acceptable net ROI after fees/slippage
- Opportunity validity window (TTL before auto-expiry)
- Strategy ID + version

---

## Decided — Venues

| Venue | Access Constraint | Execution Agent |
|---|---|---|
| Kalshi | US-only | US-based AWS region |
| Polymarket | Non-US required | Non-US AWS node |

Cross-venue communication between agents goes through the central platform. Execution agents do not communicate directly with each other.

---

## Decided — Tech Stack (Proposed, Pending ADRs)

| Layer | Choice |
|---|---|
| Latency-critical services | Go |
| Strategy prototyping | Python (hot paths promoted to Go) |
| Event bus | NATS JetStream |
| Transactional DB | PostgreSQL |
| Hot state / snapshots | Redis |
| Analytics (v2) | ClickHouse |
| Container orchestration | AWS EKS |
| Observability | OpenTelemetry + Prometheus + Grafana |
| Secrets | Local secret files |

---

## Decided — Capital Model & Risk Limits

**Validation-first deployment:** The platform launches with minimum viable position sizes per venue to prove pipeline correctness before scaling capital.

| Parameter | Value / Policy |
|---|---|
| Initial capital per market | ~$100 |
| Position size | Absolute minimum required by each venue's API |
| Max open bundles | Unlimited — bankroll is the only hard ceiling |
| Scaling trigger | Manual decision after validation phase confirms outcomes are as expected |

**Implications:**
- The **Opportunity Service must implement atomic capital reservation** before firing any execution intent. It checks unallocated balance, reserves the required amount, then fires. If execution fails, reservation is released.
- The **Risk Service** enforces a single global limit in v1: total allocated capital ≤ available bankroll. Per-trade and per-event hard caps are not needed in v1 but the schema must support them for v2.
- No artificial cap on concurrent open bundles; throughput is naturally bounded by capital availability.
- The **Reporting Service** must expose real-time: total bankroll, allocated capital, unallocated capital, and per-bundle P&L.

---

## Decided — Circuit Breakers & Kill Switches

**Model:** Kill switches are **per-strategy and fully configurable**. There are no platform-wide loss-based circuit breakers. Each strategy defines its own conditions for being paused.

**Rationale:** ROI is the meaningful metric, not raw win/loss count. A speculative strategy may have many small losses offset by large wins; killing it on a bad run would be incorrect. A pure-arb strategy may have zero tolerance for any loss. Each strategy knows its own expected behavior.

| Parameter | Policy |
|---|---|
| Circuit breaker scope | Per-strategy only |
| Trigger metric | Strategy-defined (examples: ROI over rolling window, consecutive loss count, loss rate, drawdown % of allocated capital) |
| Trigger threshold | Strategy-defined (configured in each strategy's settings) |
| On trigger | Strategy is paused; alert sent; **manual re-enable required** |
| Auto-recovery | Not supported — human must review and re-enable |
| Global kill switch | Manual only (operator action via UI or control plane) |

**Implications:**
- Each strategy's config schema must include a `circuit_breaker` block with its own metric type, thresholds, and alert targets.
- The Risk Service enforces the circuit breaker rules and owns the paused/active state per strategy.
- The UI must clearly show which strategies are paused, why (which condition triggered), and provide a re-enable control.
- Alerting on circuit breaker trips is mandatory from day one.

---

## Decided — Partial-Fill Execution Policy

**Model:** Partial-fill behavior is **strategy-defined**. Two primary modes must be supported from day one:

| Mode | Behavior |
|---|---|
| `UNWIND_ON_PARTIAL` | If any leg fails to fill, immediately cancel/unwind all already-filled legs. Bundle is aborted. |
| `HOLD_AND_GTC` | Filled legs are held as open positions. A GTC order is placed (or maintained) for the unfilled leg(s). Bundle transitions to `PENDING_COMPLETION` state. Bundle is considered complete only when all legs fill, or the strategy's defined wait TTL expires. |

**On TTL expiry in HOLD_AND_GTC mode:** strategy config defines the fallback — unwind filled legs, or keep holding (becoming a standalone open position monitored by Selling Service).

**Implications:**
- The **Execution Service** must support both immediate cancel/unwind flows and GTC order placement per-leg.
- The **Opportunity Service** must track bundle state: `PENDING`, `EXECUTING`, `PENDING_COMPLETION`, `COMPLETE`, `UNWINDING`, `ABORTED`.
- The **Selling Service** must be aware of `PENDING_COMPLETION` bundles — legs held waiting for a GTC fill are not available for independent early-exit decisions.
- The **Reporting Service** must surface `PENDING_COMPLETION` bundles distinctly — they carry real capital exposure but are not yet resolved.
- Each strategy's config schema must include a `partial_fill_policy` block: mode, GTC TTL (if applicable), and TTL expiry fallback.

---

## Decided — Order Sizing

**Model:** Fixed share count per leg, defined in each strategy's configuration.

| Parameter | Policy |
|---|---|
| Sizing unit | Shares (not dollars) |
| Sizing scope | Per leg, set in strategy config |
| Sizing algorithm | Fixed — always buy exactly N shares on each leg |
| Validation phase default | Minimum share quantity accepted by each venue's API |

**Implications:**
- The strategy config schema includes a `shares_per_leg` field (or a per-leg override map for strategies where each leg has a different target size).
- The Opportunity Service computes the total bundle cost as `sum(price_per_share[leg] * shares[leg])` and validates it fits within available capital before reserving.
- Dollar exposure per bundle varies with market prices, but share count is always deterministic and auditable.
- Future sizing modes (Kelly, risk-budget) can be added as strategy config options without changing the execution pipeline.

---

## Decided — Slippage Tolerance & Entry Gate

**Model:** The Opportunity Service uses a **ROI floor** as the sole entry gate. Exact price targets are not used; only the net ROI of the full bundle after fees matters.

| Parameter | Policy |
|---|---|
| Entry condition | Net bundle ROI ≥ strategy-defined minimum |
| Slippage model | Implicit — price moves are tolerated as long as ROI floor is still met |
| Fee inclusion | Fees must be factored into ROI calculation before the gate check |
| Threshold definition | Each strategy defines its own `min_roi` in its config |

**ROI Calculation:**
```
net_roi = (total_payout - total_cost - total_fees) / total_cost
```
Where:
- `total_payout` = winning outcome payout across all legs (typically $1.00 per share per winning leg)
- `total_cost` = sum of current ask price × shares across all legs at moment of gate check
- `total_fees` = venue fees for all legs at current prices

**Implications:**
- The Opportunity Service re-evaluates ROI on every price tick for every watched bundle.
- As soon as `net_roi >= strategy.min_roi`, the execution intent fires (subject to capital availability).
- This means a strategy never executes a losing or sub-threshold bundle regardless of how prices moved.
- Fee structures for each venue must be modeled accurately in the Opportunity Service — this is a hard dependency before live trading.

---

## Decided — Matching Engine Architecture

**Model:** The Matching Engine is a **plugin host for domain-specific sub-matchers**, structurally identical to the Strategy Service plugin model.

### Sub-Matcher Plugin Model

Each sub-matcher:
- Is independently deployable and togglable (on/off without affecting others)
- Specializes in a specific domain or market structure type
- Owns its own confidence scoring logic and evidence model
- Has its own human-approval toggle
- Emits canonical match proposals into the shared Outcome Graph

**Example sub-matchers (illustrative, not final):**

| Sub-Matcher | Specialization |
|---|---|
| `sports-binary` | Binary markets: Team A wins / Team B wins, Yes/No game outcomes |
| `sports-prop` | Sports proposition markets (totals, spreads, player props) |
| `politics-binary` | Binary political outcomes: will X happen yes/no |
| `politics-multi` | Multi-leg political markets: who wins among N candidates |
| `crypto-binary` | Binary crypto price/event outcomes |
| `multi-leg-general` | Generic multi-option markets not covered by domain-specific matchers |

### Human Approval Policy

| State | Behavior |
|---|---|
| `approval_required = true` (default) | All match proposals from this sub-matcher queue for human review before becoming tradable |
| `approval_required = false` | High-confidence proposals auto-approved and immediately tradable |

- The toggle is **per sub-matcher**, not global.
- Transition path: start all matchers with `approval_required = true`. As each sub-matcher demonstrates near-100% accuracy, flip its toggle independently.
- Confidence score is always computed and stored regardless of approval mode — it is never discarded.
- A match approved by a human overrides the confidence floor for that specific mapping (human judgment is the highest confidence signal).

**Implications:**
- The UI must include a **Match Review Queue** showing pending proposals per sub-matcher, with evidence, confidence score, proposed leg mappings, and approve/reject controls.
- The UI must show per-sub-matcher accuracy metrics to inform the decision to flip the approval toggle.
- The Matching Engine and Strategy Service share the same plugin-host pattern — worth extracting a common plugin framework to avoid duplication.
- Sub-matchers are defined in ADR-003.

---

## Decided — Matching Engine Semantic Model

**Model:** Embeddings + LLM reasoning. No manual alias tables or synonym lists are maintained.

### Pipeline

```
Market metadata (title, description, outcomes)
        ↓
  Embedding model (text → vector)
        ↓
  Vector similarity search → candidate pairs (top-K)
        ↓
  LLM review of candidate pairs
    → confidence score (0.0–1.0)
    → leg mapping proposals (which outcome ↔ which outcome)
    → reasoning text (becomes the `evidence` field on each Outcome Graph edge)
        ↓
  Outcome Graph edge created with confidence + evidence
        ↓
  Routed to approval queue or auto-approved per sub-matcher toggle
```

### Key Properties

| Property | Detail |
|---|---|
| Entity aliasing | Handled implicitly by embedding similarity — no manual curation needed |
| Naming drift | Embeddings are robust to abbreviations, alternate names, rewordings |
| Evidence | LLM reasoning text is stored verbatim as the `evidence[]` field on each graph edge |
| Confidence | LLM produces a structured confidence score alongside its reasoning |
| Human review | LLM reasoning text is shown to reviewer in the UI to explain *why* a match was proposed |
| Re-evaluation | Matches can be re-run through the LLM pipeline when market metadata changes |

### Infrastructure Requirements

- **Vector store:** pgvector (on existing PostgreSQL) for v1; migrate to dedicated vector DB (Qdrant, Weaviate) if query latency becomes a bottleneck.
- **Embedding model:** OpenAI `text-embedding-3-small` or equivalent for v1; domain-specific fine-tuning possible per sub-matcher in v2.
- **LLM:** OpenAI GPT-4o or equivalent for candidate review; structured output (JSON) required for confidence score and leg mappings.
- **Matching is not latency-critical** — it runs as a background process. New markets are matched within minutes of appearing on a venue, not milliseconds.

## Decided — Infrastructure Topology

**Model:** US-primary control plane with a thin non-US execution proxy for Polymarket.

| Node | Location | Responsibilities |
|---|---|---|
| **US Control Plane** | US AWS region | All services: Ingestion, Matching, Strategy, Opportunity, Execution (Kalshi), Selling, Reporting, Risk, UI |
| **Non-US Proxy** | Non-US AWS node | Single responsibility: receive signed order instructions from US control plane, place orders on Polymarket, return fill confirmations |

### Key Properties

- The non-US proxy runs **zero business logic**. It is a dumb order relay.
- It accepts only authenticated, signed execution instructions from the US control plane.
- It returns order status and fill data back to the US Execution Service.
- All position tracking, P&L accounting, and state management happens on the US side.
- The proxy is stateless — it can be restarted at any time without data loss.
- Network path: `US Execution Service → encrypted channel → Non-US Proxy → Polymarket API`

### Infrastructure Specs

| Node | Instance | OS | Role |
|---|---|---|---|
| US Control Plane | t3.medium (2 vCPU, 4GB RAM) | Amazon Linux | All platform services |
| Non-US Proxy | t3.micro (2 vCPU, 1GB RAM) | Amazon Linux | Polymarket order relay only |

**Domain:** `polybot.nostrabotus.com`

### Container Orchestration: Docker Compose on EC2 (v1)

Since both instances are already provisioned, the v1 approach is **Docker + Docker Compose directly on the EC2 instances**. No ECS/EKS overhead.

| Concern | Approach |
|---|---|
| Service orchestration | Docker Compose on each node |
| Service restarts | Docker restart policies (`always` or `unless-stopped`) |
| Config/secrets | Local secret files + environment injection at container start |
| TLS termination | nginx reverse proxy on US node; Let's Encrypt cert for `polybot.nostrabotus.com` |
| Proxy auth | Mutual TLS or shared secret between US Execution Service and non-US proxy |
| Upgrade path | ECS EC2 launch type (same instances, adds orchestration layer) if horizontal scaling is needed |

**t3.medium capacity note:** 4GB RAM is sufficient for v1 with Go services (small footprint) + PostgreSQL + Redis + NATS. Monitor memory headroom; PostgreSQL `shared_buffers` should be tuned conservatively (~512MB) until load is understood.

**t3.micro capacity note:** 1GB RAM — proxy must be a single minimal Go binary. No database, no cache, no NATS. Stateless HTTP/WS relay only.

---

## Decided — Recovery & Restart Model

**Model:** All services resume from persisted state on restart. The Execution Service has an additional crash-recovery reconciliation routine.

### Per-Service Recovery Behaviour

| Service | On Restart |
|---|---|
| Market Ingestion | Reconnect to venue WS feeds; backfill any gap via REST; resume publishing |
| Matching Engine | Resume from last known match state in DB; re-queue any pending LLM reviews |
| Strategy Service | Reload strategy configs and state; resume scanning approved matches for opportunities |
| Opportunity Service | Reload all `PENDING` and `PENDING_COMPLETION` bundles from DB; resume price monitoring |
| Execution Service | **See reconciliation routine below** |
| Selling Service | Reload all open bundles flagged `early_exit_eligible`; resume monitoring |
| Reporting / Ledger | Stateless reader — always current from DB |
| Risk Service | Reload per-strategy states (paused/active, running metrics) from DB |
| UI | Stateless — reconnects to backend services |

### Execution Service: Crash-Recovery Reconciliation

On every startup, before accepting new execution intents:

```
1. Query DB for all bundles in state EXECUTING or PENDING_COMPLETION
2. For each bundle:
   a. Query actual fill state from each venue API (ground truth)
   b. Reconcile: mark legs as filled/unfilled based on venue response
3. For each partially-filled bundle:
   a. Attempt to complete unfilled legs (place orders, wait up to configurable TTL)
   b. If completion succeeds → bundle transitions to COMPLETE
   c. If completion fails within TTL → initiate unwind of all filled legs
   d. After unwind → bundle transitions to ABORTED; alert emitted
4. Only after reconciliation is complete: accept new execution intents
```

**Key properties:**
- Reconciliation is **synchronous and blocking** at startup — the service does not process new work until all in-flight state is resolved.
- Venue API is the **ground truth** for fill state, not internal DB. DB is updated to match venue state, not the other way around.
- All reconciliation actions are written to the audit log with `source: crash_recovery`.
- Unwind during recovery follows the same unwind path as runtime unwind (same code, same logging).
- Capital reservations are released on ABORTED bundles as part of the unwind flow.

### Persistence Requirements

- All bundle state transitions must be **written to PostgreSQL before being acted on** (write-ahead semantics).
- NATS JetStream provides at-least-once delivery; consumers must be idempotent.
- Redis hot state is **ephemeral** — services must be able to cold-start from PostgreSQL if Redis is empty.

---

## Decided — UI Dashboard Requirements

### Primary Dashboard (5 required panels)

| Priority | Panel | Key Data Points |
|---|---|---|
| 1 | **Capital Overview** | Total bankroll, allocated capital, unallocated capital, % deployed |
| 2 | **Open Bundles** | Per-bundle: status, venue/legs, entry cost, current exit value, current ROI, age, early-exit eligible flag |
| 3 | **Strategy Performance** | Per-strategy: opportunities identified, executed, win rate, total ROI, paper vs live indicator, circuit breaker status |
| 4 | **Historical P&L** | Resolved bundles: entry cost, payout/loss, ROI, resolution date; aggregate curves over time |
| 5 | **Execution Pipeline Latency** | p50/p95/p99 latency from opportunity fired → order submitted, per strategy and per venue |

### Additional Required UI Views (not main dashboard but mandatory)

| View | Purpose |
|---|---|
| Match Review Queue | Pending sub-matcher proposals; confidence score, LLM reasoning, approve/reject controls |
| Service Status Panel | Paper/live toggle state for every service unit; which sub-matchers and strategies are active/paused |
| Alert Log | Chronological in-app alerts: circuit breaker trips, reconciliation events, execution failures |
| Bundle Detail | Drill-down on any bundle: full leg list, fill prices, timeline of state transitions, audit trail |

### UI Technical Notes

- Real-time updates via WebSocket (no polling) for capital overview, open bundles, and latency panels.
- Historical P&L and strategy performance panels can be query-on-load (not real-time).
- Paper vs live status must be **visible on every panel** that shows live data — never ambiguous whether numbers are real or simulated.
- UI is served from the US control plane at `polybot.nostrabotus.com`.
- Single operator user in v1 — no multi-user auth required initially.

### Latency Consideration

- Cross-region round-trip (US → non-US proxy → Polymarket → back) adds latency to Polymarket legs vs Kalshi legs.
- For the arbitrage strategy, both legs must fire near-simultaneously. The Execution Service must account for this asymmetric latency when sequencing leg orders.
- Recommended: fire the higher-latency leg (Polymarket) first or simultaneously, not after Kalshi confirms.



**Strategy:** `arb-cross-market-v1`

**Logic:** Identify matched market pairs where the cost of buying all mutually exclusive outcome legs across venues sums to less than $1.00. Since exactly one outcome must resolve as the winner paying $1.00, a sub-$1.00 total cost guarantees profit regardless of outcome.

**Simple binary example:**
```
Kalshi:     YES on Cardinals win tonight  →  ask $0.52/share
Polymarket: YES on Blue Jays win tonight  →  ask $0.44/share
                                              ─────────────
Total cost per share set:                     $0.96
Guaranteed payout:                            $1.00
Gross profit:                                 $0.04/share (4.2% ROI)
```

**Strategy config schema (fields needed):**
- `min_roi` — minimum net ROI after fees before entry fires
- `shares_per_leg` — fixed share count per leg
- `partial_fill_policy` — UNWIND_ON_PARTIAL or HOLD_AND_GTC
- `circuit_breaker` — metric, threshold, alert target
- `approval_required` — whether to require human approval of new matches before this strategy uses them

**Critical dependency:** This strategy is fully blocked on the Matching Engine producing validated, high-confidence canonical pairs. No arbitrage opportunities can be identified without accurate leg mappings.

### Build Sequence (confirmed)

```
[1] Market Ingestion  ──►  [2] Matching Engine  ──►  [3] arb-cross-market-v1
         (unblocked)            (unblocked)               (blocked on #2)
```

Phase 2 (Strategy v1) cannot start until Phase 1 (Matching Engine) is producing reviewed, approved pairs.

---

## Decided — Paper Mode (Progressive Live Promotion)

**Principle:** Paper mode is a **first-class concern at every service boundary**. Each independently deployable unit (sub-matcher, sub-strategy, execution engine) has its own paper/live toggle. The platform is designed so each stage can be validated independently before any real capital moves.

### Per-Service Paper Mode Toggles

| Service / Unit | Toggle | Paper Behavior | Live Behavior |
|---|---|---|---|
| Matching Engine sub-matcher | `human_review_required` | All proposed matches queue for human review before becoming tradable | High-confidence matches auto-approved |
| Strategy sub-strategy | `paper_mode` | Identifies opportunities, logs what it would have emitted, does NOT send to Opportunity Service | Emits real candidate opportunities |
| Execution Service | `paper_mode` | Simulates fills at current market prices, records theoretical positions and P&L | Places real orders on venues |

### Promotion Path

```
Sub-Matcher:  [human_review=ON]  ─► matches reviewed ► accuracy proven ► [human_review=OFF]

Strategy:     [paper_mode=ON]    ─► opportunities logged ► logic validated ► [paper_mode=OFF]

Execution:    [paper_mode=ON]    ─► theoretical fills tracked ► P&L looks correct ► [paper_mode=OFF]
```

Promotion at each stage is a **manual operator decision** via the UI. No automatic promotion.

### Key Implications

- **Reporting Service must distinguish paper vs live positions at all times.** Paper P&L and live P&L are tracked, attributed, and displayed separately.
- **Paper fills are simulated** at the best available ask at the moment the execution intent fires. This gives a realistic (slightly optimistic) theoretical benchmark.
- **Execution paper mode is a global toggle** — it applies to all strategies simultaneously. This is intentional: it lets you validate the execution pipeline end-to-end before any real money flows, regardless of whether individual strategies have already been validated.
- **UI must surface paper mode status prominently** for every service unit so there is never ambiguity about whether real orders are being placed.
- The circuit breaker model applies equally to paper mode — a strategy can be paused in paper mode too, based on theoretical P&L.
- This principle extends to any future service where it makes sense (e.g., Selling Service can have a paper mode that logs proposed exits without acting on them).

---

## Decided — Selling Service: Early-Exit Policy

**Model:** Global ROI threshold governs the Selling Service, with per-bundle opt-out stamped by the strategy at opportunity creation time.

### Rules

| Rule | Detail |
|---|---|
| Default eligibility | All open bundles are eligible for early exit |
| Opt-out mechanism | Strategy sets `early_exit_eligible = false` on specific bundles at creation time |
| Exit trigger | Net ROI of closing all legs now ≥ configurable `selling_service.min_exit_roi` |
| Exit scope | All legs of the bundle must be closeable simultaneously (same all-or-nothing principle as entry) |
| Selling paper mode | Selling Service has its own `paper_mode` toggle — surfaces exit opportunities in UI without acting |
| Config location | `selling_service.min_exit_roi` is a global platform setting (not per-strategy) |

### Bundle Schema Additions

The candidate opportunity emitted by a strategy must include:
- `early_exit_eligible: bool` (default: `true`)
- Optional: `early_exit_min_roi_override` — if set, overrides the global threshold for this specific bundle

### Implications

- The Selling Service never needs to know which strategy created a bundle — it only reads `early_exit_eligible` and the global threshold.
- Strategies with long-term ROI intent (e.g., a multi-month political market play) set `early_exit_eligible = false` on their bundles.
- The **UI must show which open bundles are eligible vs ineligible**, and surface current exit ROI for all eligible bundles in real time.
- Selling Service also gets a `paper_mode` toggle consistent with the progressive live-promotion principle.
- Future enhancement: per-bundle `early_exit_min_roi_override` allows fine-grained control without changing global config.



Questions below are being answered one at a time. Answered questions move to the Decided sections above.

### Capital & Risk
- [x] Total bankroll: ~$100/market initially, minimum venue position sizes, validation-first
- [x] Max open bundles: unlimited (bankroll-gated only)
- [x] Circuit breakers: per-strategy, configurable metric + threshold, manual re-enable required
- [x] Global kill-switch: manual only via UI/control plane

### Execution
- [x] Partial-fill policy: strategy-defined; UNWIND_ON_PARTIAL or HOLD_AND_GTC with configurable TTL and expiry fallback
- [x] Order sizing: fixed share count per leg, defined in strategy config; minimum venue quantity during validation
- [x] Slippage tolerance: ROI floor (strategy-defined `min_roi`); fees included; no fixed price targets
- [ ] IOC/FOK preference per venue?

### Matching Engine
- [x] Confidence threshold: per sub-matcher plugin; approval_required toggle per sub-matcher (default on)
- [x] Human approval: toggleable per sub-matcher; off when accuracy proven near 100%
- [x] Sub-matcher plugins: domain-specific (sports-binary, politics-multi, crypto, etc.) mirroring Strategy plugin model
- [x] Versioned mapping history: confidence score always stored; human approvals recorded as highest-confidence signal
- [x] ML-assisted vs rule-based entity/team aliasing: embeddings + LLM (no manual tables); pgvector for v1
- [x] How quickly must mappings refresh: background process, new markets matched within minutes of appearing

### Opportunity & Selling
- [x] Minimum net ROI: strategy-defined `min_roi` enforced by Opportunity Service at entry
- [x] Early-exit: global `min_exit_roi` threshold; per-bundle `early_exit_eligible` flag (default true); strategy can opt-out at creation
- [x] Selling paper mode: yes, consistent with progressive live-promotion principle
- [ ] Minimum hold time before exit is considered?
- [ ] Annualized IRR vs absolute return for exit scoring?

### Operations & Infrastructure
- [x] Alerting: in-app only for v1; alert events emitted on bus so external channels (Slack, SMS, email) can be added later without service changes
- [x] Topology: US control plane (all services); non-US node is a thin Polymarket execution proxy only
- [x] AWS region: US-based primary; non-US secondary (proxy only)
- [x] ECS vs EKS: Docker Compose on existing EC2s for v1; ECS EC2 launch type as upgrade path if needed
- [x] Domain: polybot.nostrabotus.com (nginx + Let's Encrypt on US control plane)
- [x] Recovery: all services resume from DB state on restart; Execution Service runs blocking reconciliation routine before accepting new work
- [ ] Blue/green deploy requirement for execution-critical services?
- [ ] RPO/RTO targets?

### UI
- [x] Primary dashboard: capital overview, open bundles, strategy performance, historical P&L, execution latency
- [x] Required views: match review queue, service status panel, alert log, bundle detail drill-down
- [x] Real-time via WebSocket for live panels; paper/live status visible on every panel
- [x] Single operator user in v1

---

## Architecture Decision Records

All ADRs are in `docs/adr/`.

| ADR | Title | Status |
|---|---|---|
| [ADR-001](docs/adr/ADR-001-event-bus.md) | Event Bus — NATS JetStream | ✅ Accepted |
| [ADR-002](docs/adr/ADR-002-service-languages.md) | Service Language Choices | ✅ Accepted |
| [ADR-003](docs/adr/ADR-003-matching-engine.md) | Matching Engine — Sub-Matcher Plugin Model | ✅ Accepted |
| [ADR-004](docs/adr/ADR-004-execution-atomicity.md) | Execution Atomicity — Strategy-Defined Partial Fill Policy | ✅ Accepted |
| [ADR-005](docs/adr/ADR-005-strategy-isolation.md) | Strategy Isolation — Separate Containers | ✅ Accepted |
| [ADR-006](docs/adr/ADR-006-deployment-topology.md) | Deployment Topology — Docker Compose on EC2 | ✅ Accepted |
| [ADR-007](docs/adr/ADR-007-risk-guardrails.md) | Risk Guardrails — Centralized Risk Service | ✅ Accepted |

---

## Live Deployment State (as of 2026-05-19)

### Infrastructure

| Component | Detail |
|---|---|
| EC2 instance | `t3.large` (2 vCPU, 8GB RAM), Amazon Linux, SSH alias `polybot`, IP `3.16.140.129` |
| PostgreSQL | `pgvector/pgvector:pg16`, port `5434` on host, DB/user `polybot`, password in secrets |
| NATS | `nats:2.10-alpine`, running and healthy |
| Redis | `redis:7-alpine`, running and healthy |
| nginx | `nginx:1.27-alpine`, container `polybot-nginx`, serves UI from bind-mount `~/PolyBot/services/ui/dist` → `/usr/share/nginx/html` |
| Project root on EC2 | `~/PolyBot` |

### Running Containers (docker-compose.yml + docker-compose.proxy.yml)

| Container | Status | Notes |
|---|---|---|
| `polybot-postgres` | ✅ Up (healthy) | pgvector:pg16 |
| `polybot-nats` | ✅ Up (healthy) | nats:2.10-alpine |
| `polybot-redis` | ✅ Up (healthy) | redis:7-alpine |
| `polybot-nginx` | ✅ Up | nginx:1.27-alpine, serves UI static files |
| `polybot-ingestion` | ✅ Up | Kalshi + Polymarket WS/REST market data ingestion |
| `polybot-matching` | ✅ Up | Matching Engine (Python, APScheduler) |
| `polybot-reporting` | ✅ Up | Reporting/Ledger Service (Go) |
| `polybot-opportunity` | ✅ Up | Opportunity Service (Go) |
| `polybot-execution` | ✅ Up | Execution Service (Go, paper mode) |
| `polybot-proxy` | ✅ Up (healthy) | Non-US proxy for Polymarket |
| `polybot-risk` | ✅ Up | Risk & Guardrails Service (Go) |
| `polybot-selling` | ✅ Up | Selling Service (Go) |
| `polybot-strategy-arb-cross-market` | ✅ Up | arb-cross-market-v1 strategy (Go) |

**Note:** The UI is NOT a Docker Compose service. It is a statically-built React/Vite app deployed to `~/PolyBot/services/ui/dist/`. To deploy UI changes: build with `npm run build` (or `npx vite build`) in `services/ui/` on EC2, then run `docker exec polybot-nginx nginx -s reload`.

### Database State (as of 2026-05-19)

| Table | Count | Notes |
|---|---|---|
| `markets` (KALSHI, OPEN) | 183,900 | ~19,889 have embeddings (~11%) |
| `markets` (POLYMARKET, OPEN) | 2,479 | ~1,279 have embeddings (~52%) |
| `match_candidates` | 1,177 | Bulk-populated 2026-05-19; cosine similarity ≥ 0.80, pair_enabled categories |
| `match_pairs` (APPROVED) | 157 | From first successful review cycle |

### Matching Engine Pipeline

- **Schedule:** embed stage every 1 min, review stage every 5 min
- **Embed worker:** local Python script `tools/embed_worker.py` using `BAAI/bge-large-en-v1.5` (1024-dim); runs on dev machine when active
- **LLM:** `gemini-2.5-flash` (primary) via `https://generativelanguage.googleapis.com/v1beta/openai/`, `gpt-4o-mini` (fallback on RateLimitError + second opinion when confidence < 0.90)
- **Sub-matchers:** `politics.py`, `sports_binary.py`, `world_events.py` in `services/matching/matchers/`
- **Current cycle behavior:** `candidates_found=1177`, `already_validated=853`, `already_exists=157`, `pre_filter_failed=167`, `pairs_proposed=0` — all 1,177 candidates have been processed; new pairs will only appear as new markets are ingested and embedded
- **Key file:** `services/matching/host.py` — APScheduler, reads `GEMINI_API_KEY_FILE`, creates `gemini_client = AsyncOpenAI(api_key=..., base_url=GEMINI_BASE_URL)`

### UI

- **Stack:** React + TypeScript + Vite
- **Source:** `services/ui/src/`
- **Built output:** `services/ui/dist/` (served by `polybot-nginx`)
- **Match Review page:** `services/ui/src/components/MatchReview.tsx`
  - Shows ANN candidates (from `match_candidates`) and proposed pairs (from `match_pairs`)
  - Links to each venue's market page
- **Kalshi URL format (correct):** `https://kalshi.com/markets/{series}/{event}` where:
  - `series` = first `-`-delimited segment of ticker, lowercased (e.g. `KXPRESNOMR-28-DJTJR` → `kxpresnomr`)
  - `event` = all segments except the last, joined with `-`, lowercased (e.g. → `kxpresnomr-28`)
- **Polymarket URL format:** `https://polymarket.com/event/{slug}` — only works when `slug IS NOT NULL` in DB. Many politics/world_events markets have null slug; backfilling from Polymarket API is a known pending task.

### Secrets Layout (on EC2)

All secrets are in `~/PolyBot/secrets/` (not in repo). Key files:
- `gemini_api_key` — Gemini API key for matching LLM
- `openai_api_key` — OpenAI API key (gpt-4o-mini fallback)
- `kalshi_api_key_id` / `kalshi_private_key.pem` — Kalshi RSA-PSS auth
- `polymarket_private_key` — Polymarket private key
- `db_password` — PostgreSQL password

### Known Issues & Pending Work

| Issue | Status | Notes |
|---|---|---|
| Polymarket slugs null for most politics/world_events markets | Open | No link can be shown without slug; needs API backfill |
| Kalshi embed coverage low (~11%) | In progress | embed_worker.py runs incrementally; ~19,889/183,900 done |
| match_candidates not auto-refreshed by embed stage | Open | The 5-min review stage only processes existing candidates; bulk re-population needed as new markets are embedded |
| 157 APPROVED pairs in match_pairs not yet reviewed by human | Open | Match Review UI is built; operator needs to review and approve/reject pairs before live trading |
