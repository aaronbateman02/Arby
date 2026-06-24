# Matching Engine — Design Spec

## Overview

The Matching Engine identifies pairs of markets (one Kalshi, one Polymarket) that
resolve on the same real-world event, enabling cross-venue arbitrage. It uses
embedding similarity for candidate discovery and an LLM (via OpenRouter) for
validation.

## Architecture

```
Discovery Scanner ──► markets table (PostgreSQL + pgvector)
                            │
                     Embed Worker (llm.nostrabotus.com)
                            │
                     POST /api/v1/markets/unembedded  ◄─── GET fetched
                     GET  /api/v1/markets/embeddings  ───► POST written
                            │
                     pgvector ANN search ──► match_candidates
                            │
                     LLM review via OpenRouter ──► match_pairs
                            │
                     Event bus notification ──► downstream consumers
```

## Components

### 1. Database Schema (SQL migration)

Tables:

- **`markets`** — mirrors what the discovery scanner produces. Each row is one
  market from one venue. Embedding column uses `vector(1024)` for
  bge-large-en-v1.5.
  - Columns: `id UUID PK`, `venue VARCHAR(20)`, `venue_market_id VARCHAR(255)`,
    `title TEXT`, `description TEXT`, `category VARCHAR(100)`,
    `outcomes JSONB`, `open_time TIMESTAMPTZ`, `close_time TIMESTAMPTZ`,
    `embedding VECTOR(1024)`, `embedding_model VARCHAR(100)`,
    `embedding_updated_at TIMESTAMPTZ`, `first_seen_at TIMESTAMPTZ`,
    `last_updated_at TIMESTAMPTZ`
  - Unique constraint on `(venue, venue_market_id)`
  - IVFFlat index on `embedding` for cosine similarity

- **`match_candidates`** — pairs found by vector similarity, awaiting LLM review.
  - Columns: `id UUID PK`, `market_a_id UUID`, `market_b_id UUID`,
    `similarity DOUBLE PRECISION`, `category VARCHAR(100)`,
    `status VARCHAR(20)` (PENDING, REVIEWED, SKIPPED),
    `created_at TIMESTAMPTZ`
  - FK to markets, unique on `(market_a_id, market_b_id)`

- **`match_pairs`** — LLM-validated pairs with relationship + confidence.
  - Columns: `id UUID PK`, `candidate_id UUID`, `is_same_event BOOLEAN`,
    `relationship VARCHAR(20)` (EQUIVALENT, INVERSE, UNRELATED),
    `confidence DOUBLE PRECISION`, `reasoning TEXT`,
    `leg_a_model VARCHAR(100)`, `leg_b_model VARCHAR(100)`,
    `reviewed_at TIMESTAMPTZ`, `status VARCHAR(20)`
    (PENDING_APPROVAL, APPROVED, REJECTED)
  - FK to match_candidates

### 2. Go API Endpoints (on the polybot HTTP server)

**`GET /api/v1/markets/unembedded?limit=N`**
- Returns markets where `embedding IS NULL` in priority order (smallest
  categories first, matching the old embed worker strategy).
- Response: `{ markets: [{ id, venue, venue_market_id, title, description, category }] }`

**`POST /api/v1/markets/embeddings`**
- Accepts: `{ embeddings: [{ id, vector: [0.001, -0.02, ...] }] }`
- Upserts each embedding back to the markets table.

### 3. Embed Worker (Python script for llm.nostrabotus.com)

Thin HTTP client that loops:
1. `GET http://arby.nostrabotus.com/api/v1/markets/unembedded?limit=64`
2. If empty, sleep 30s, retry
3. Load `BAAI/bge-large-en-v1.5` (or configured model), encode texts
4. `POST http://arby.nostrabotus.com/api/v1/markets/embeddings`
5. Log count, repeat

Embedding text = `title + ". " + description[:400]` (mirroring the old approach).

No DB credentials needed on the LLM machine. No migration files needed there.
Single Python file, one pip install.

### 4. Go Matching Engine (`pkg/matching/`)

Two sub-components running on a timer:

**Candidate Discovery** (`candidate.go`)
- Runs on a configurable interval (default 5min)
- Queries markets that have embeddings and no candidate pair yet
- For each market, runs pgvector ANN search against the opposite venue:
  `SELECT id FROM markets WHERE venue = $2 AND embedding IS NOT NULL
   ORDER BY embedding <=> $1 LIMIT 20`
- Inserts pairs with similarity ≥ threshold (default 0.80) into match_candidates

**LLM Reviewer** (`reviewer.go`)
- Runs after candidate discovery (or on its own interval, default 5min)
- Fetches `PENDING` candidates, sends batches to OpenRouter
- Uses structured output JSON to get `{ is_same_event, relationship, confidence,
  reasoning, potential_ambiguity }`
- Type-specific prompt from the system's settings (fetched from the settings API
  or configured via env)
- After Leg A model review, sends confirmed matches to Leg B model for
  second-opinion validation
- Writes results to match_pairs

**Engine** (`engine.go`)
- Orchestrator: starts/stops the discovery and review loops
- Publishes events on the bus for downstream consumers when new matched pairs
  are created

### 5. UI — Match Review Page

New `/review` page:
- Table of `match_pairs` with status PENDING_APPROVAL
- Columns: Market A title, Market B title, Category, Confidence, Relationship,
  Model used, Reasoning, Actions (Approve / Reject)
- Approve/Reject via API call to the Go backend
- Filters: by category, by relationship type, by status
- Search by market title

## Data Flow

```
Discovery → markets table stored
       ↓
LLM machine polls GET /api/v1/markets/unembedded
       ↓
LLM machine embeds, POSTs back
       ↓
Candidate discovery queries pgvector, writes match_candidates
       ↓
LLM reviewer fetches PENDING candidates, calls OpenRouter
       ↓
Leg A model (batch screener): determines same event + relationship
       ↓
Leg B model (second opinion): validates confirmed matches
       ↓
Results stored in match_pairs, event emitted on bus
       ↓
UI shows PENDING_APPROVAL pairs for human review
```

## OpenRouter LLM Call

Uses the OpenRouter chat completions API with a JSON schema response format:

```
POST https://openrouter.ai/api/v1/chat/completions
Headers: Authorization: Bearer <key>, Content-Type: application/json
Body: {
  model: <legAModel or legBModel from settings>,
  messages: [
    { role: "system", content: <type-specific prompt> },
    { role: "user", content: <pair details> }
  ],
  response_format: {
    type: "json_schema",
    json_schema: {
      name: "MatchReview",
      schema: {
        type: "object",
        properties: {
          is_same_event: { type: "boolean" },
          relationship: { type: "string", enum: ["EQUIVALENT", "INVERSE", "UNRELATED"] },
          confidence: { type: "number", minimum: 0, maximum: 1 },
          reasoning: { type: "string" },
          potential_ambiguity: { type: "string" }
        },
        required: ["is_same_event", "relationship", "confidence", "reasoning"]
      }
    }
  }
}
```

The user prompt includes: title_a, resolves_a, description_a, scheduled_a,
close_a, title_b, resolves_b, description_b, scheduled_b, close_b, category,
similarity.

## Batch Review

Up to 40 candidates are sent in a single LLM call to reduce API costs
(mirroring the old PolyBot batch approach). Each candidate includes its
`pair_index` so results can be mapped back reliably.

## Configuration

Settings loaded from the settings file / API:
- Leg A model name
- Leg B model name
- Batch size
- Confidence threshold for auto-approval
- Type-specific prompts (domain-optimized system prompts per strategy type)
- Candidate discovery interval
- Similarity threshold for candidate discovery
