# Matching Engine Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build the complete Matching Engine pipeline — DB schema, embed API endpoints, Python embed worker, pgvector candidate discovery, OpenRouter LLM review, engine orchestrator, and Match Review UI page.

**Architecture:** Go monolith exposes HTTP APIs for embed worker. pgvector handles similarity search. OpenRouter provides structured-output LLM review. Next.js UI shows matches for human approval.

**Tech Stack:** Go 1.22, pgvector/pg16, BAAI/bge-large-en-v1.5, OpenRouter API, Next.js 14 Tailwind TypeScript

## Global Constraints

- Go 1.22 minimum, `http.ServeMux` for routing
- PostgreSQL 16 with pgvector extension (`vector` extension + 1024-dim vectors)
- All secrets from file paths in env vars
- JSON slog logging, Prometheus metrics, in-process event bus
- All Go code verified on EC2 via SSH (no local Go toolchain)
- Use `slog` directly (no custom logging wrappers)

---

### Task 1: DB Store Layer

**Files:**
- Create: `pkg/matching/store.go`
- Create: `pkg/matching/store_test.go`

**Interfaces:**
- Produces: `NewStore(pg *db.DB) *Store` with methods:
  - `CreateTables(ctx) error` — CREATE TABLE IF NOT EXISTS for markets/match_candidates/match_pairs
  - `UpsertMarket(ctx, Market) error`
  - `GetUnembeddedMarkets(ctx, limit) ([]Market, error)`
  - `UpsertEmbedding(ctx, id string, vector []float64) error`
  - `InsertCandidate(ctx, Candidate) error`
  - `GetPendingCandidates(ctx, limit) ([]Candidate, error)`
  - `InsertMatchPair(ctx, MatchPair) error`
  - `GetMatchPairs(ctx, status) ([]MatchPair, error)`
  - `UpdateMatchPairStatus(ctx, id, status) error`
  - `GetCandidateByID(ctx, id) (*Candidate, error)`
  - `GetMarketByID(ctx, id) (*Market, error)`

- [ ] **Step 1: Write store.go with structs, CreateTables, and all CRUD methods**
- [ ] **Step 2: Write store_test.go with joinFloats unit test**
- [ ] **Step 3: Verify** — `go vet ./pkg/matching/` and `go build ./...` on EC2
- [ ] **Step 4: Commit**

### Task 2: Embed API Endpoints + Wire Into Main

**Files:**
- Create: `pkg/matching/handler.go`
- Modify: `cmd/polybot/main.go`

**Interfaces:**
- Consumes: Store methods `GetUnembeddedMarkets`, `UpsertEmbedding`
- Produces: HTTP handlers on `/api/v1/markets/unembedded` and `/api/v1/markets/embeddings`

- [ ] **Step 1: Write handler.go** — NewHandler, GetUnembedded, PostEmbeddings, WireRoutes
- [ ] **Step 2: Update main.go** — Import matching, create tables, wire routes
- [ ] **Step 3: Verify** — `go vet ./... && go build ./...` on EC2
- [ ] **Step 4: Commit**

### Task 3: Python Embed Worker

**Files:**
- Create: `tools/embed_worker.py`

The worker polls GET /api/v1/markets/unembedded, embeds with bge-large-en-v1.5, POSTs back. No DB credentials.

- [ ] **Step 1: Write embed_worker.py** — arg parse, load model, fetch/embed/post loop
- [ ] **Step 2: Test on EC2** — `python3 tools/embed_worker.py --api-url https://arby.nostrabotus.com --batch-size 1 --fetch-limit 2`
- [ ] **Step 3: Commit**

### Task 4: Candidate Discovery (pgvector ANN)

**Files:**
- Create: `pkg/matching/candidate.go`
- Modify: `cmd/polybot/main.go` (wire discoverer loop)

- [ ] **Step 1: Write candidate.go** — CandidateDiscoverer with RunOnce that queries pgvector ANN for cross-venue pairs, inserts into match_candidates
- [ ] **Step 2: Wire into main.go** — Start CandidateDiscoverer in a goroutine on a 5-minute ticker
- [ ] **Step 3: Verify build** — `go vet ./... && go build ./...`
- [ ] **Step 4: Commit**

### Task 5: LLM Reviewer (OpenRouter Batch Review)

**Files:**
- Create: `pkg/matching/reviewer.go`
- Modify: `cmd/polybot/main.go`

Uses OpenRouter chat completions with JSON schema structured output. Sends batches of up to 40 candidates. Type-specific prompts from settings.

- [ ] **Step 1: Write reviewer.go** — NewReviewer, RunOnce: fetch PENDING candidates, batch them, call OpenRouter, parse structured output, store in match_pairs
- [ ] **Step 2: Wire into main.go** — Start reviewer in goroutine on 5-minute ticker
- [ ] **Step 3: Verify build**
- [ ] **Step 4: Commit**

### Task 6: Match Review UI Page

**Files:**
- Modify: `services/ui/src/app/review/page.tsx`
- Create: `services/ui/src/app/api/matching/pairs/route.ts`
- Create: `services/ui/src/app/api/matching/pairs/[id]/route.ts`

Shows match_pairs with PENDING_APPROVAL status. Columns: market titles, venue, category, confidence, relationship, reasoning, actions (Approve/Reject).

- [ ] **Step 1: Create UI API routes** — GET /api/matching/pairs, POST /api/matching/pairs/:id/approve, POST /api/matching/pairs/:id/reject
- [ ] **Step 2: Update review page** — Full table with search, filter by category/relationship, approve/reject buttons
- [ ] **Step 3: Build & deploy** — `next build`, commit, push, deploy on EC2
- [ ] **Step 4: Verify at** `https://arby.nostrabotus.com/review`

### Task 7 (optional): Wire Discovery Scanner Into Markets Table

The discovery scanner currently publishes events on the bus but doesn't persist markets to the DB. Add a subscriber that calls UpsertMarket for each discovered market — otherwise the markets table stays empty.

**Files:**
- Modify: `cmd/polybot/main.go`

- [ ] **Step 1: Subscribe to discovery events** — After discScanner.Run, subscribe to bus and UpsertMarket for each discovery event
- [ ] **Step 2: Verify** — Check markets table has data after a scan cycle
- [ ] **Step 3: Commit**
